/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"hcu-container-toolkit/internal/logger"
)

// Options carries the resolved runtime options for the vhcu Manager.
type Options struct {
	MaxVHCUPerCard int
	StateDir       string
	LockFile       string
	LockTimeout    time.Duration
	ConfDir        string
	DestroyRetries int
}

// defaultRetryDelay is the base backoff between vdev destroy retries
// (100ms, 200ms, 400ms, ...), matching the reference implementation.
const defaultRetryDelay = 100 * time.Millisecond

// Allocation is the result of a dynamic vHCU allocation.
type Allocation struct {
	ContainerID  string
	AllocationID string
	DeviceIndex  int
	VdevIndex    int // -1 when the request short-circuited to a full-card mount
	PciBusID     string
	Cores        int
	MemMiB       int
	ComputeUnits int
	FullCard     bool
	Reused       bool
}

// Manager orchestrates dynamic vHCU create / release / GC. All mutating
// operations run under the global state-store lock.
type Manager struct {
	logger       logger.Interface
	backend      Backend
	allocator    *Allocator
	states       *StateStore
	confDir      string
	destroyRetry int
	retryDelay   time.Duration
	gcGrace      time.Duration
	bootID       func() (string, error)
	creator      func() (int, string, error)
}

// NewManager creates a Manager with the given backend.
func NewManager(l logger.Interface, opts Options, backend Backend) *Manager {
	if opts.DestroyRetries <= 0 {
		opts.DestroyRetries = 5
	}
	return &Manager{
		logger:       l,
		backend:      backend,
		allocator:    NewAllocator(backend, opts.MaxVHCUPerCard),
		states:       NewStateStore(opts.StateDir, opts.LockFile, opts.LockTimeout),
		confDir:      opts.ConfDir,
		destroyRetry: opts.DestroyRetries,
		retryDelay:   defaultRetryDelay,
		gcGrace:      2 * time.Minute,
		bootID:       readBootID,
		creator:      currentCreator,
	}
}

// States exposes the state store (read paths for the CLI).
func (m *Manager) States() *StateStore {
	return m.states
}

// Views returns the merged per-card capacity view (read path for the CLI
// status command).
func (m *Manager) Views() ([]CardView, error) {
	return m.allocator.Views()
}

// MaxVHCUPerCard returns the configured per-card vdev slot limit.
func (m *Manager) MaxVHCUPerCard() int {
	return m.allocator.maxPerCard
}

// Backend exposes the backend (read paths for the CLI status command).
func (m *Manager) Backend() Backend {
	return m.backend
}

// Allocate creates (or reuses) a dynamic vHCU for containerID and returns the
// allocation result. It is idempotent: a container whose recorded vdev is
// still alive reuses it without creating a second one.
func (m *Manager) Allocate(containerID string, req *Request) (*Allocation, error) {
	var alloc *Allocation
	err := m.states.WithLock(func() error {
		var err error
		alloc, err = m.allocateLocked(containerID, req)
		return err
	})
	if err != nil {
		return nil, err
	}
	return alloc, nil
}

func (m *Manager) allocateLocked(containerID string, req *Request) (*Allocation, error) {
	if err := validateContainerID(containerID); err != nil {
		return nil, err
	}
	if req == nil || req.Cores < 0 || req.Cores > 100 || req.MemMiB <= 0 || (req.Device != "" && !isDeviceIndex(req.Device)) {
		return nil, fmt.Errorf("invalid dynamic vHCU request")
	}
	st, err := m.reconcileExistingState(containerID, req)
	if err != nil {
		return nil, err
	}
	if st != nil {
		if st.Cores != req.Cores || st.MemMiB != req.MemMiB || (req.Device != "" && req.Device != strconv.Itoa(st.DeviceIndex)) {
			return nil, fmt.Errorf("container %s already owns a different vHCU allocation", containerID)
		}
		pid, start, err := m.creator()
		if err != nil {
			return nil, err
		}
		st.LeaseAt = time.Now().UTC()
		st.CreatorPID, st.CreatorStartTime = pid, start
		if err := m.states.Save(st); err != nil {
			return nil, fmt.Errorf("failed to protect reused vHCU during create: %w", err)
		}
		return allocationFromState(st, true), nil
	}
	views, err := m.allocator.Views()
	if err != nil {
		return nil, err
	}
	sel, err := m.allocator.Select(req, views)
	if err != nil {
		return nil, err
	}
	if IsFullCardRequest(req.Cores, sel.MemMiB, totalMemoryOf(views, sel.DeviceIndex)) {
		return &Allocation{ContainerID: containerID, DeviceIndex: sel.DeviceIndex, VdevIndex: Unset, PciBusID: sel.BusID, Cores: req.Cores, MemMiB: sel.MemMiB, ComputeUnits: sel.ComputeUnits, FullCard: true}, nil
	}
	before, err := m.globalVdevIndices()
	if err != nil {
		return nil, err
	}
	bootID, err := m.bootID()
	if err != nil {
		return nil, err
	}
	pid, start, err := m.creator()
	if err != nil {
		return nil, err
	}
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		return nil, err
	}
	st = &State{ContainerID: containerID, AllocationID: hex.EncodeToString(token), Phase: PhaseCreating, DeviceIndex: sel.DeviceIndex, VdevIndex: Unset, PciBusID: sel.BusID, Cores: req.Cores, ComputeUnits: sel.ComputeUnits, MemMiB: sel.MemMiB, CreatedAt: time.Now().UTC(), BootID: bootID, RuntimeCommand: append([]string(nil), req.RuntimeCommand...)}
	st.LeaseAt = st.CreatedAt
	st.CreatorPID, st.CreatorStartTime = pid, start
	for idx := range before {
		st.Before = append(st.Before, idx)
	}
	sortInts(st.Before)
	if err := m.states.Save(st); err != nil {
		return nil, err
	}
	if err := m.backend.CreateVDevice(sel.DeviceIndex, sel.ComputeUnits, sel.MemMiB); err != nil {
		return nil, m.keepPending(st, fmt.Errorf("create may have changed the driver; operator review required: %w", err))
	}
	after, err := m.globalVdevIndices()
	if err != nil {
		return nil, m.keepPending(st, fmt.Errorf("cannot locate created vHCU: %w", err))
	}
	for idx := range after {
		if !before[idx] {
			st.Candidates = append(st.Candidates, idx)
		}
	}
	sortInts(st.Candidates)
	removed := false
	for idx := range before {
		if !after[idx] {
			removed = true
		}
	}
	if len(st.Candidates) != 1 || removed {
		return nil, m.keepPending(st, fmt.Errorf("ambiguous vHCU creation (before=%v candidates=%v removed=%v); no unconfirmed devices were destroyed", st.Before, st.Candidates, removed))
	}
	idx := st.Candidates[0]
	vdevs, err := m.backend.ListVDevices(sel.DeviceIndex)
	if err != nil {
		return nil, m.keepPending(st, err)
	}
	found := false
	for _, v := range vdevs {
		if v.Index == idx && v.DeviceIndex == sel.DeviceIndex {
			match, err := m.resourcesMatch(v, sel.ComputeUnits, sel.MemMiB)
			if err != nil {
				return nil, m.keepPending(st, err)
			}
			if !match {
				return nil, m.keepPending(st, fmt.Errorf("new vHCU %d resources do not match request", idx))
			}
			found = true
		}
	}
	if !found {
		return nil, m.keepPending(st, fmt.Errorf("new vHCU %d is not on requested device %d", idx, sel.DeviceIndex))
	}
	identity, err := m.confIdentity(idx)
	if err != nil {
		return nil, m.keepPending(st, fmt.Errorf("cannot identify vHCU configuration: %w", err))
	}
	st.VdevIndex = idx
	st.Identity = identity
	st.Phase = PhaseActive
	st.Candidates = nil
	if err := m.states.Save(st); err != nil {
		return nil, errors.Join(fmt.Errorf("failed to persist active vHCU: %w", err), m.releaseState(st))
	}
	m.logger.Infof("Created dynamic vHCU %d on device %d (%d CU, %d MiB) for container %s", idx, sel.DeviceIndex, sel.ComputeUnits, sel.MemMiB, containerID)
	return allocationFromState(st, false), nil
}

func allocationFromState(st *State, reused bool) *Allocation {
	return &Allocation{ContainerID: st.ContainerID, AllocationID: st.AllocationID, DeviceIndex: st.DeviceIndex, VdevIndex: st.VdevIndex, PciBusID: st.PciBusID, Cores: st.Cores, MemMiB: st.MemMiB, ComputeUnits: st.ComputeUnits, Reused: reused}
}

// Read/query failures never imply that the recorded instance disappeared.
func (m *Manager) reconcileExistingState(containerID string, req *Request) (*State, error) {
	st, err := m.states.Load(containerID)
	if err != nil || st == nil {
		return nil, err
	}
	// Scope must match before reconciliation can mutate or release a record.
	if !slices.Equal(st.RuntimeCommand, req.RuntimeCommand) {
		return nil, fmt.Errorf("container %s already owns a vHCU in a different OCI runtime scope", containerID)
	}
	if st.Phase == PhasePending || st.Phase == PhaseCreating {
		return nil, m.keepPending(st, fmt.Errorf("unfinished or ambiguous allocation for %s requires operator review", containerID))
	}
	if st.Phase == PhaseReleasing {
		if err := m.releaseState(st); err != nil {
			return nil, err
		}
		return nil, nil
	}
	exists, err := m.ownedVdevExists(st)
	if err != nil {
		return nil, err
	}
	if exists {
		return st, nil
	}
	// An absent old instance may already have had its index reused elsewhere;
	// remove its record, never an unverified configuration file.
	if err := m.states.Delete(containerID); err != nil {
		return nil, err
	}
	return nil, nil
}

func (m *Manager) keepPending(st *State, cause error) error {
	st.Phase = PhasePending
	st.LastError = cause.Error()
	return errors.Join(cause, m.states.Save(st))
}

// ReleaseAllocation only rolls back the allocation created by this attempt.
// Reused instances and subsequent allocations with the same container ID are
// unaffected by a failed create or a delayed poststop hook.
func (m *Manager) ReleaseAllocation(containerID, allocationID string) error {
	return m.states.WithLock(func() error {
		st, err := m.states.Load(containerID)
		if err != nil || st == nil {
			return err
		}
		if allocationID == "" || st.AllocationID != allocationID {
			return fmt.Errorf("allocation token mismatch for container %s; preserving vHCU", containerID)
		}
		return m.releaseState(st)
	})
}

func (m *Manager) Release(containerID string) error {
	return m.states.WithLock(func() error { return m.releaseLocked(containerID) })
}

func (m *Manager) releaseLocked(containerID string) error {
	st, err := m.states.Load(containerID)
	if err != nil || st == nil {
		return err
	}
	return m.releaseState(st)
}

// All cleanup paths retain ownership and configuration until destruction is
// confirmed. Unresolved creates are deliberately never guessed at.
func (m *Manager) releaseState(st *State) error {
	if st.Phase == PhasePending || st.Phase == PhaseCreating || st.VdevIndex == Unset {
		return m.keepPending(st, fmt.Errorf("allocation for %s is unresolved; operator review required (candidates=%v)", st.ContainerID, st.Candidates))
	}
	exists, err := m.ownedVdevExists(st)
	if err != nil {
		return err
	}
	st.Phase = PhaseReleasing
	if err := m.states.Save(st); err != nil {
		return err
	}
	if exists {
		if err := m.destroyWithRetries(st); err != nil {
			return err
		}
	}
	if err := m.removeOwnedConfFiles(st); err != nil {
		return err
	}
	return m.states.Delete(st.ContainerID)
}

func (m *Manager) destroyWithRetries(st *State) error {
	var lastErr error
	for attempt := 0; attempt < m.destroyRetry; attempt++ {
		if attempt > 0 {
			time.Sleep(m.retryDelay * (1 << (attempt - 1)))
		}
		exists, err := m.ownedVdevExists(st)
		if err != nil {
			return err
		}
		if !exists {
			return nil
		}
		if err := m.backend.StopVDevice(st.VdevIndex); err != nil {
			m.logger.Debugf("Stop vHCU %d: %v", st.VdevIndex, err)
		}
		if err := m.backend.DestroyVDevice(st.VdevIndex); err != nil {
			lastErr = err
			continue
		}
		exists, err = m.ownedVdevExists(st)
		if err != nil {
			return err
		}
		if !exists {
			return nil
		}
		lastErr = fmt.Errorf("driver still reports vHCU %d after destroy", st.VdevIndex)
	}
	return fmt.Errorf("destroy vHCU %d failed after %d attempts (state kept): %w", st.VdevIndex, m.destroyRetry, lastErr)
}

// Boot identity and the original configuration fingerprint guard against
// recycled indices. Driver resources are checked when the CLI reports them.
func (m *Manager) ownedVdevExists(st *State) (bool, error) {
	vdevs, err := m.backend.ListVDevices(st.DeviceIndex)
	if err != nil {
		return false, err
	}
	for _, v := range vdevs {
		if v.Index != st.VdevIndex || v.DeviceIndex != st.DeviceIndex {
			continue
		}
		boot, err := m.bootID()
		if err != nil {
			return false, err
		}
		identity, err := m.confIdentity(st.VdevIndex)
		if err != nil {
			return false, err
		}
		if st.BootID == "" || st.Identity == "" || st.BootID != boot || st.Identity != identity {
			return false, m.keepPending(st, fmt.Errorf("vHCU %d identity changed or is unknown; refusing reuse/destruction", st.VdevIndex))
		}
		match, err := m.resourcesMatch(v, st.ComputeUnits, st.MemMiB)
		if err != nil {
			return false, err
		}
		if !match {
			return false, m.keepPending(st, fmt.Errorf("vHCU %d resources changed; refusing reuse/destruction", st.VdevIndex))
		}
		return true, nil
	}
	global, err := m.globalVdevIndices()
	if err != nil {
		return false, err
	}
	if global[st.VdevIndex] {
		return false, m.keepPending(st, fmt.Errorf("vHCU index %d moved to another physical device", st.VdevIndex))
	}
	return false, nil
}

func (m *Manager) resourcesMatch(v VDeviceInfo, computeUnits, memMiB int) (bool, error) {
	if !v.ResourcesKnown {
		return true, nil
	}
	if v.MemMiB != memMiB {
		return false, nil
	}
	if v.ComputeUnits == computeUnits {
		return true, nil
	}
	if computeUnits != 0 {
		return false, nil
	}
	// A shared-CU request is encoded as 0 in the original configuration.
	// DTK 6.3.31 reports 0 before activation, then the parent's full CU count.
	// Only accept those two representations; ownership still requires the
	// unchanged boot ID and configuration fingerprint in ownedVdevExists.
	devices, err := m.backend.PhysicalDevices()
	if err != nil {
		return false, err
	}
	for _, dev := range devices {
		if dev.Index == v.DeviceIndex {
			return dev.ComputeUnits > 0 && v.ComputeUnits == dev.ComputeUnits, nil
		}
	}
	return false, fmt.Errorf("physical device %d not found while checking shared CU", v.DeviceIndex)
}

func readBootID() (string, error) {
	data, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return "", fmt.Errorf("cannot establish host boot identity: %w", err)
	}
	id := strings.TrimSpace(string(data))
	if id == "" {
		return "", fmt.Errorf("empty host boot identity")
	}
	return id, nil
}

func (m *Manager) confIdentity(vdev int) (string, error) {
	path := m.confPath(vdev)
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	digest := sha256.New()
	if _, err := io.Copy(digest, f); err != nil {
		return "", err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("cannot identify configuration inode")
	}
	return fmt.Sprintf("%v:%v:%d:%x", stat.Dev, stat.Ino, info.ModTime().UnixNano(), digest.Sum(nil)), nil
}

func (m *Manager) removeOwnedConfFiles(st *State) error {
	// A missing source is already clean. A replaced source may belong to an
	// external instance reusing the same global index and must be preserved.
	identity, err := m.confIdentity(st.VdevIndex)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	boot, err := m.bootID()
	if err != nil {
		return err
	}
	if st.Identity == "" || identity != st.Identity || st.BootID != boot {
		return m.keepPending(st, fmt.Errorf("configuration identity changed for vHCU %d; refusing removal", st.VdevIndex))
	}
	// /etc/vdev/docker is a container mount destination, not an owned host
	// configuration. Only remove the driver-created source file.
	return rmFile(m.confPath(st.VdevIndex))
}

func (m *Manager) confPath(vdev int) string {
	return fmt.Sprintf("%s/vdev%d.conf", m.confDir, vdev)
}

// GCResult summarizes a garbage-collection sweep.
type GCResult struct {
	// Released holds the container IDs whose vHCU was reclaimed.
	Released []string
	// Deferred requires retry or operator review; ownership was preserved.
	Deferred []string
	// Orphans are vdev indices present in the driver but not owned by any
	// toolkit state. They are only reported, never destroyed: they may be
	// administratively pre-split vHCU (fail-safe rule: never destroy what
	// we have no record for).
	Orphans []int
}

// ContainerStatus is deliberately separate from query errors: unknown owners
// are preserved, including allocations created before runtime scoping existed.
type ContainerStatus int

const (
	ContainerUnknown ContainerStatus = iota
	ContainerLive
	ContainerStopped
	ContainerAbsent
)

type ContainerProbe func(*State) (ContainerStatus, error)

// GC probes each owner while holding the allocation lock. The grace period
// covers the gap between Toolkit allocation and the OCI runtime's create.
func (m *Manager) GC(probe ContainerProbe) (*GCResult, error) {
	result := &GCResult{}
	err := m.states.WithLock(func() error {
		states, err := m.states.List()
		if err != nil {
			return err
		}
		for _, st := range states {
			if st.Phase == PhasePending {
				result.Deferred = append(result.Deferred, st.ContainerID)
				continue
			}
			leaseAt := st.LeaseAt
			if leaseAt.IsZero() {
				leaseAt = st.CreatedAt
			}
			if time.Since(leaseAt) < m.gcGrace {
				continue
			}
			if st.Phase == PhaseActive || st.Phase == PhaseCreating {
				creating, err := creatorRunning(st)
				if err != nil {
					result.Deferred = append(result.Deferred, st.ContainerID)
					m.logger.Warningf("GC: creator query failed for %s: %v", st.ContainerID, err)
					continue
				}
				if creating {
					continue
				}
				if probe == nil {
					result.Deferred = append(result.Deferred, st.ContainerID)
					continue
				}
				status, err := probe(st)
				if err != nil || status == ContainerUnknown {
					result.Deferred = append(result.Deferred, st.ContainerID)
					m.logger.Warningf("GC: cannot determine status of %s: %v", st.ContainerID, err)
					continue
				}
				if status == ContainerLive {
					continue
				}
				// A fresh final query avoids releasing an allocation whose runtime
				// changed state while an earlier status query was being processed.
				status, err = probe(st)
				if err != nil || (status != ContainerAbsent && status != ContainerStopped) {
					if err != nil || status == ContainerUnknown {
						result.Deferred = append(result.Deferred, st.ContainerID)
					}
					continue
				}
			}
			if err := m.releaseState(st); err != nil {
				m.logger.Warningf("GC: preserving vHCU of %s: %v", st.ContainerID, err)
				result.Deferred = append(result.Deferred, st.ContainerID)
				continue
			}
			result.Released = append(result.Released, st.ContainerID)
		}
		remaining, err := m.states.List()
		if err != nil {
			return err
		}
		owned := map[int]bool{}
		for _, st := range remaining {
			if st.VdevIndex != Unset {
				owned[st.VdevIndex] = true
			}
		}
		global, err := m.globalVdevIndices()
		if err != nil {
			return err
		}
		for idx := range global {
			if !owned[idx] {
				result.Orphans = append(result.Orphans, idx)
			}
		}
		sortInts(result.Orphans)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// globalVdevIndices returns the set of all vdev indices across all devices.
func (m *Manager) globalVdevIndices() (map[int]bool, error) {
	devices, err := m.backend.PhysicalDevices()
	if err != nil {
		return nil, err
	}
	indices := make(map[int]bool)
	for _, dev := range devices {
		vdevs, err := m.backend.ListVDevices(dev.Index)
		if err != nil {
			return nil, err
		}
		for _, v := range vdevs {
			indices[v.Index] = true
		}
	}
	return indices, nil
}

func totalMemoryOf(views []CardView, deviceIndex int) int {
	for _, v := range views {
		if v.Dev.Index == deviceIndex {
			return v.Dev.MemoryMiB
		}
	}
	return 0
}
