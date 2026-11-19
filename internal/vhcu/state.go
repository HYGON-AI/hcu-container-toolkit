/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"
)

// Pending records require operator review: an index alone cannot establish
// ownership after an ambiguous create, external replacement or host reboot.
const (
	PhaseCreating  = "creating"
	PhaseActive    = "active"
	PhaseReleasing = "releasing"
	PhasePending   = "pending"
)

// State is the on-disk ownership record for one dynamically created vHCU.
// The file name is <containerID>.json under the state dir; it is the single
// source of truth for "which container owns this vdev".
type State struct {
	ContainerID      string    `json:"containerID"`
	Phase            string    `json:"phase"`
	DeviceIndex      int       `json:"deviceIndex"`
	VdevIndex        int       `json:"vdevIndex"`
	PciBusID         string    `json:"pciBusID"`
	Cores            int       `json:"cores"`
	MemMiB           int       `json:"memMiB"`
	CreatedAt        time.Time `json:"createdAt"`
	LeaseAt          time.Time `json:"leaseAt,omitempty"`
	AllocationID     string    `json:"allocationID,omitempty"`
	ComputeUnits     int       `json:"computeUnits"`
	Identity         string    `json:"identity,omitempty"`
	BootID           string    `json:"bootID,omitempty"`
	RuntimeCommand   []string  `json:"runtimeCommand,omitempty"`
	CreatorPID       int       `json:"creatorPID,omitempty"`
	CreatorStartTime string    `json:"creatorStartTime,omitempty"`
	Before           []int     `json:"before,omitempty"`
	Candidates       []int     `json:"candidates,omitempty"`
	LastError        string    `json:"lastError,omitempty"`
}

func validateContainerID(id string) error {
	if id == "" || id == "." || id == ".." || strings.ContainsAny(id, "/\\") {
		return fmt.Errorf("invalid container ID %q", id)
	}
	return nil
}

// StateStore persists State records and provides the global flock that
// serializes all dynamic splits and releases on a single host.
type StateStore struct {
	dir         string
	lockFile    string
	lockTimeout time.Duration
}

// NewStateStore creates a StateStore. dir is created on first use.
func NewStateStore(dir, lockFile string, lockTimeout time.Duration) *StateStore {
	return &StateStore{
		dir:         dir,
		lockFile:    lockFile,
		lockTimeout: lockTimeout,
	}
}

// Path returns the state file path for a container.
func (s *StateStore) Path(containerID string) string {
	return filepath.Join(s.dir, containerID+".json")
}

// ensureDir creates the state dir; it is NOT acceptable to continue without
// it (an unrecorded vdev leaks), so failures here must abort allocation.
func (s *StateStore) ensureDir() error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return fmt.Errorf("failed to create vhcu state dir %s: %v", s.dir, err)
	}
	return nil
}

// Load reads the state for a container. A missing file returns (nil, nil).
func (s *StateStore) Load(containerID string) (*State, error) {
	if err := validateContainerID(containerID); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(s.Path(containerID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read vhcu state for container %s: %v", containerID, err)
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("failed to parse vhcu state for container %s: %v", containerID, err)
	}
	if st.ContainerID != containerID || st.DeviceIndex < 0 || st.VdevIndex < Unset {
		return nil, fmt.Errorf("invalid vhcu ownership record for container %s", containerID)
	}
	switch st.Phase {
	case PhaseCreating, PhaseActive, PhaseReleasing, PhasePending:
	default:
		return nil, fmt.Errorf("invalid vhcu state phase %q for container %s", st.Phase, containerID)
	}
	return &st, nil
}

// Save atomically writes the state for a container.
func (s *StateStore) Save(st *State) error {
	if err := validateContainerID(st.ContainerID); err != nil {
		return err
	}
	if err := s.ensureDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal vhcu state: %v", err)
	}
	path := s.Path(st.ContainerID)
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(data)
	if writeErr == nil {
		writeErr = f.Sync()
	}
	closeErr := f.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to write vhcu state tmp file: %v / %v", writeErr, closeErr)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to move vhcu state into place: %v", err)
	}
	return nil
}

// Delete removes the state for a container (missing file is not an error).
func (s *StateStore) Delete(containerID string) error {
	if err := validateContainerID(containerID); err != nil {
		return err
	}
	if err := os.Remove(s.Path(containerID)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete vhcu state for container %s: %v", containerID, err)
	}
	return nil
}

// List returns all states in the state dir.
func (s *StateStore) List() ([]*State, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list vhcu state dir %s: %v", s.dir, err)
	}
	var states []*State
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		st, err := s.Load(trimExt(entry.Name()))
		if err != nil {
			return nil, err
		}
		states = append(states, st)
	}
	return states, nil
}

func trimExt(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}

// WithLock runs fn while holding the global vhcu flock. The global (not
// per-card) lock is deliberate: concurrent splits would corrupt the
// before/after diff used to locate new vdev indices, and single-host split
// frequency is low enough that lock contention is irrelevant.
func (s *StateStore) WithLock(fn func() error) error {
	if err := s.ensureDir(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.lockFile), 0o700); err != nil {
		return err
	}
	lock := flock.New(s.lockFile)
	deadline := time.Now().Add(s.lockTimeout)
	for {
		locked, err := lock.TryLock()
		if err == nil && locked {
			defer func() { _ = lock.Unlock() }()
			return fn()
		}
		if err != nil {
			return fmt.Errorf("failed to acquire vhcu lock: %v", err)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out (%s) waiting for vhcu lock %s; concurrent split in progress, please retry", s.lockTimeout, s.lockFile)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
