/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"hcu-container-toolkit/internal/logger"
)

// fakeVdev is a vdev instance recorded by the fake backend.
type fakeVdev struct {
	dev, cu, mem int
}

// fakeBackend simulates the hy-smi virtual control plane. New vdev indices
// are assigned the smallest free global index, like the real driver.
type fakeBackend struct {
	confDir             string
	listErr             error
	failListAfterCreate bool
	createErr           error
	destroyNoop         bool
	devices             []PhysicalDevice
	vdevs               map[int]fakeVdev

	// createHook, when set, replaces the default create behavior (used to
	// inject ambiguity / misindexing scenarios).
	createHook func(b *fakeBackend, dev, cu, mem int)

	// destroyFails makes DestroyVDevice fail this many times before
	// succeeding (simulates "vdev in use").
	destroyFails int
	destroyCalls int
	stopCalls    int
	createCalls  []string
}

func newFakeBackend(devices ...PhysicalDevice) *fakeBackend {
	return &fakeBackend{
		devices: devices,
		vdevs:   make(map[int]fakeVdev),
	}
}

func (b *fakeBackend) PhysicalDevices() ([]PhysicalDevice, error) {
	return b.devices, nil
}

func (b *fakeBackend) ListVDevices(dev int) ([]VDeviceInfo, error) {
	if b.listErr != nil {
		return nil, b.listErr
	}
	if b.failListAfterCreate && len(b.createCalls) > 0 {
		return nil, fmt.Errorf("list failed after create")
	}
	var out []VDeviceInfo
	for idx, v := range b.vdevs {
		if v.dev == dev {
			out = append(out, VDeviceInfo{Index: idx, DeviceIndex: dev, ComputeUnits: v.cu, MemMiB: v.mem, ResourcesKnown: true})
		}
	}
	return out, nil
}

func (b *fakeBackend) CreateVDevice(dev, cu, memMiB int) error {
	b.createCalls = append(b.createCalls, fmt.Sprintf("%d:%d:%d", dev, cu, memMiB))
	if b.createHook != nil {
		b.createHook(b, dev, cu, memMiB)
	} else {
		idx := b.minFreeIndex()
		b.vdevs[idx] = fakeVdev{dev: dev, cu: cu, mem: memMiB}
	}
	if b.confDir != "" {
		if err := os.MkdirAll(b.confDir, 0700); err != nil {
			return err
		}
		for idx, v := range b.vdevs {
			path := filepath.Join(b.confDir, fmt.Sprintf("vdev%d.conf", idx))
			if _, err := os.Stat(path); os.IsNotExist(err) {
				data := fmt.Sprintf("Physical Device: %d\nCU: %d\nMem: %d\n", v.dev, v.cu, v.mem)
				if err := os.WriteFile(path, []byte(data), 0600); err != nil {
					return err
				}
			}
		}
	}
	return b.createErr
}

func (b *fakeBackend) minFreeIndex() int {
	for i := 0; ; i++ {
		if _, ok := b.vdevs[i]; !ok {
			return i
		}
	}
}

func (b *fakeBackend) StopVDevice(vdev int) error {
	b.stopCalls++
	return nil
}

func (b *fakeBackend) DestroyVDevice(vdev int) error {
	b.destroyCalls++
	if b.destroyFails > 0 {
		b.destroyFails--
		return fmt.Errorf("vdev in use")
	}
	if !b.destroyNoop {
		delete(b.vdevs, vdev)
	}
	return nil
}

func (b *fakeBackend) Remaining(dev int) (Remaining, error) {
	rem := Remaining{DeviceIndex: dev}
	for _, d := range b.devices {
		if d.Index == dev {
			rem.ComputeUnits = d.ComputeUnits
			rem.MemoryMiB = d.MemoryMiB
		}
	}
	for _, v := range b.vdevs {
		if v.dev == dev {
			rem.ComputeUnits -= v.cu
			rem.MemoryMiB -= v.mem
		}
	}
	return rem, nil
}

func testCard() PhysicalDevice {
	return PhysicalDevice{Index: 0, BusID: "0000:81:00.0", ComputeUnits: 120, MemoryMiB: 65536}
}

func testCardN(idx int) PhysicalDevice {
	return PhysicalDevice{Index: idx, BusID: fmt.Sprintf("0000:81:%02x.0", idx), ComputeUnits: 120, MemoryMiB: 65536}
}

func testManager(t *testing.T, backend Backend) (*Manager, *fakeBackend) {
	t.Helper()
	dir := t.TempDir()
	fake, ok := backend.(*fakeBackend)
	if !ok {
		fake = nil
	}
	m := NewManager(&logger.NullLogger{}, Options{
		MaxVHCUPerCard: 4,
		StateDir:       filepath.Join(dir, "vhcu"),
		LockFile:       filepath.Join(dir, "vhcu.lock"),
		LockTimeout:    5 * time.Second,
		ConfDir:        filepath.Join(dir, "vdev"),
		DestroyRetries: 3,
	}, backend)
	m.retryDelay = time.Millisecond
	m.gcGrace = 0
	m.bootID = func() (string, error) { return "test-boot", nil }
	m.creator = func() (int, string, error) { return 0, "", nil }
	if fake != nil {
		fake.confDir = m.confDir
	}
	return m, fake
}

func TestAllocateCreatesVdev(t *testing.T) {
	m, fake := testManager(t, newFakeBackend(testCard()))

	alloc, err := m.Allocate("ctr1", &Request{Cores: 30, MemMiB: 8192})
	if err != nil {
		t.Fatalf("allocate failed: %v", err)
	}
	if alloc.VdevIndex != 0 || alloc.DeviceIndex != 0 || alloc.FullCard {
		t.Fatalf("unexpected allocation: %+v", alloc)
	}
	if alloc.ComputeUnits != 36 {
		t.Fatalf("expected 36 CU, got %d", alloc.ComputeUnits)
	}
	if len(fake.createCalls) != 1 || fake.createCalls[0] != "0:36:8192" {
		t.Fatalf("unexpected create call: %v", fake.createCalls)
	}

	st, err := m.States().Load("ctr1")
	if err != nil || st == nil {
		t.Fatalf("state must be recorded: %v %v", st, err)
	}
	if st.Phase != PhaseActive || st.VdevIndex != 0 {
		t.Fatalf("state must be active with vdev 0: %+v", st)
	}
}

func TestAllocateFullCardShortCircuit(t *testing.T) {
	m, fake := testManager(t, newFakeBackend(testCard()))

	alloc, err := m.Allocate("ctr1", &Request{Cores: 100, MemMiB: 65536})
	if err != nil {
		t.Fatalf("allocate failed: %v", err)
	}
	if !alloc.FullCard || alloc.VdevIndex != Unset {
		t.Fatalf("expected full-card short circuit, got %+v", alloc)
	}
	if len(fake.createCalls) != 0 {
		t.Fatalf("full card must not create a vdev, got %v", fake.createCalls)
	}
	if st, _ := m.States().Load("ctr1"); st != nil {
		t.Fatal("full-card allocation must not write a state file")
	}
}

func TestAllocateIdempotentReuse(t *testing.T) {
	m, fake := testManager(t, newFakeBackend(testCard()))

	if _, err := m.Allocate("ctr1", &Request{Cores: 30, MemMiB: 8192}); err != nil {
		t.Fatalf("first allocate failed: %v", err)
	}
	alloc, err := m.Allocate("ctr1", &Request{Cores: 30, MemMiB: 8192})
	if err != nil {
		t.Fatalf("second allocate failed: %v", err)
	}
	if !alloc.Reused || alloc.VdevIndex != 0 {
		t.Fatalf("expected reuse of vdev 0, got %+v", alloc)
	}
	if len(fake.createCalls) != 1 {
		t.Fatalf("reuse must not create a second vdev, got %v", fake.createCalls)
	}
}

func TestAllocateReallocatesWhenVdevVanished(t *testing.T) {
	backend := newFakeBackend(testCard())
	m, fake := testManager(t, backend)

	if _, err := m.Allocate("ctr1", &Request{Cores: 30, MemMiB: 8192}); err != nil {
		t.Fatalf("first allocate failed: %v", err)
	}

	// Simulate an external destroy (admin or GC race).
	delete(backend.vdevs, 0)

	alloc, err := m.Allocate("ctr1", &Request{Cores: 30, MemMiB: 8192})
	if err != nil {
		t.Fatalf("reallocate failed: %v", err)
	}
	if alloc.Reused {
		t.Fatal("a vanished vdev must not be reused")
	}
	if len(fake.createCalls) != 2 {
		t.Fatalf("expected a fresh create, got %v", fake.createCalls)
	}
}

func TestAllocateAmbiguousDiffPreservesUnconfirmedDevices(t *testing.T) {
	backend := newFakeBackend(testCard())
	// Create TWO vdevs for one request: the diff is ambiguous.
	backend.createHook = func(b *fakeBackend, dev, cu, mem int) {
		for i := 0; i < 2; i++ {
			idx := b.minFreeIndex()
			b.vdevs[idx] = fakeVdev{dev: dev, cu: cu, mem: mem}
		}
	}
	m, _ := testManager(t, backend)

	_, err := m.Allocate("ctr1", &Request{Cores: 30, MemMiB: 8192})
	if err == nil {
		t.Fatal("ambiguous diff must fail")
	}
	if len(backend.vdevs) != 2 || backend.destroyCalls != 0 {
		t.Fatalf("ambiguous diff must preserve all candidates: %v", backend.vdevs)
	}
	st, err := m.States().Load("ctr1")
	if err != nil || st == nil || st.Phase != PhasePending || len(st.Candidates) != 2 {
		t.Fatalf("pending evidence missing: %+v %v", st, err)
	}

}

func TestAllocateEmptyDiffFails(t *testing.T) {
	backend := newFakeBackend(testCard())
	// Create nothing (driver silently no-ops).
	backend.createHook = func(b *fakeBackend, dev, cu, mem int) {}
	m, _ := testManager(t, backend)

	_, err := m.Allocate("ctr1", &Request{Cores: 30, MemMiB: 8192})
	if err == nil {
		t.Fatal("empty diff must fail")
	}
}

func TestAllocateAcceptsUniqueNonMinimalIndex(t *testing.T) {
	backend := newFakeBackend(testCard())
	// Pre-seed a foreign vdev so min free index is 1, then let the "driver"
	// hand out index 5 instead.
	backend.vdevs[0] = fakeVdev{dev: 0, cu: 0, mem: 1024}
	backend.createHook = func(b *fakeBackend, dev, cu, mem int) {
		b.vdevs[5] = fakeVdev{dev: dev, cu: cu, mem: mem}
	}
	m, _ := testManager(t, backend)

	alloc, err := m.Allocate("ctr1", &Request{Cores: 30, MemMiB: 8192})
	if err != nil || alloc.VdevIndex != 5 {
		t.Fatalf("unique index need not be minimal: %+v %v", alloc, err)
	}
	if backend.destroyCalls != 0 {
		t.Fatal("valid nonminimal instance must not be destroyed")
	}

}

func TestAllocateSlotLimit(t *testing.T) {
	backend := newFakeBackend(testCard())
	// Pre-split by the admin: 4 vdevs already present.
	for i := 0; i < 4; i++ {
		backend.vdevs[i] = fakeVdev{dev: 0, cu: 10, mem: 1024}
	}
	m, _ := testManager(t, backend)

	_, err := m.Allocate("ctr1", &Request{Cores: 30, MemMiB: 8192})
	if err == nil {
		t.Fatal("slot exhaustion must fail with details")
	}
}

func TestReleaseIdempotent(t *testing.T) {
	m, _ := testManager(t, newFakeBackend(testCard()))
	if err := m.Release("unknown"); err != nil {
		t.Fatalf("release of an unknown container must be a no-op: %v", err)
	}
}

func TestReleaseDestroysAndCleans(t *testing.T) {
	backend := newFakeBackend(testCard())
	m, _ := testManager(t, backend)

	if _, err := m.Allocate("ctr1", &Request{Cores: 30, MemMiB: 8192}); err != nil {
		t.Fatalf("allocate failed: %v", err)
	}

	// Simulate the driver-created conf files.
	confHost := m.confPath(0)
	if _, err := os.Stat(confHost); err != nil {
		t.Fatal(err)
	}

	if err := m.Release("ctr1"); err != nil {
		t.Fatalf("release failed: %v", err)
	}
	if _, exists := backend.vdevs[0]; exists {
		t.Fatal("vdev must be destroyed")
	}
	if st, _ := m.States().Load("ctr1"); st != nil {
		t.Fatal("state must be deleted after release")
	}
	if fileExists(confHost) {
		t.Fatal("conf file must be removed after release")
	}
}

func TestReleaseKeepsStateWhenDestroyFails(t *testing.T) {
	backend := newFakeBackend(testCard())
	m, _ := testManager(t, backend)

	if _, err := m.Allocate("ctr1", &Request{Cores: 30, MemMiB: 8192}); err != nil {
		t.Fatalf("allocate failed: %v", err)
	}

	backend.destroyFails = 99 // every destroy attempt fails ("vdev in use")
	if err := m.Release("ctr1"); err == nil {
		t.Fatal("release must fail when the vdev cannot be destroyed")
	}

	st, _ := m.States().Load("ctr1")
	if st == nil || st.Phase != PhaseReleasing {
		t.Fatalf("state must be kept as releasing for GC retry: %+v", st)
	}

	// The container's users are gone; destroy now succeeds.
	backend.destroyFails = 0
	if err := m.Release("ctr1"); err != nil {
		t.Fatalf("retry release failed: %v", err)
	}
	if st, _ := m.States().Load("ctr1"); st != nil {
		t.Fatal("state must be gone after successful retry")
	}
}

func TestGCReclaimsDeadContainersAndWarnsOnOrphans(t *testing.T) {
	backend := newFakeBackend(testCard())
	m, _ := testManager(t, backend)

	if _, err := m.Allocate("alive", &Request{Cores: 30, MemMiB: 8192}); err != nil {
		t.Fatalf("allocate failed: %v", err)
	}
	if _, err := m.Allocate("dead", &Request{Cores: 30, MemMiB: 8192}); err != nil {
		t.Fatalf("allocate failed: %v", err)
	}

	// Administratively pre-split vdev without a toolkit record: an orphan.
	backend.vdevs[100] = fakeVdev{dev: 0, cu: 10, mem: 1024}

	probe := func(st *State) (ContainerStatus, error) {
		if st.ContainerID == "alive" {
			return ContainerLive, nil
		}
		return ContainerAbsent, nil
	}
	result, err := m.GC(probe)
	if err != nil {
		t.Fatalf("gc failed: %v", err)
	}
	if len(result.Released) != 1 || result.Released[0] != "dead" {
		t.Fatalf("expected only 'dead' released, got %v", result.Released)
	}
	if _, exists := backend.vdevs[0]; !exists {
		t.Fatal("'alive' container's vdev must be kept")
	}
	if _, exists := backend.vdevs[100]; !exists {
		t.Fatal("orphan vdev must NOT be destroyed (fail-safe rule)")
	}
	if len(result.Orphans) != 1 || result.Orphans[0] != 100 {
		t.Fatalf("orphan must be reported, got %v", result.Orphans)
	}
}

func TestGCPreservesUnresolvedCreatingState(t *testing.T) {
	backend := newFakeBackend(testCard())
	m, _ := testManager(t, backend)
	if err := m.States().Save(&State{ContainerID: "crashed", Phase: PhaseCreating, DeviceIndex: 0, VdevIndex: Unset}); err != nil {
		t.Fatal(err)
	}
	result, err := m.GC(func(*State) (ContainerStatus, error) { return ContainerAbsent, nil })
	if err != nil {
		t.Fatal(err)
	}
	st, err := m.States().Load("crashed")
	if err != nil || st == nil || st.Phase != PhasePending || len(result.Released) != 0 || len(result.Deferred) != 1 {
		t.Fatalf("unresolved create lost: %+v %+v %v", st, result, err)
	}
}

func TestConcurrentAllocate(t *testing.T) {
	// 3 cards x 4 slots = 12 slots for 10 concurrent containers.
	backend := newFakeBackend(testCard(), testCardN(1), testCardN(2))
	m, _ := testManager(t, backend)

	const n = 10
	allocs := make([]*Allocation, n)
	errs := make([]error, n)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			allocs[i], errs[i] = m.Allocate(fmt.Sprintf("ctr%d", i), &Request{Cores: 5, MemMiB: 2048})
		}(i)
	}
	wg.Wait()

	seen := make(map[int]bool)
	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Fatalf("concurrent allocate %d failed: %v", i, errs[i])
		}
		if seen[allocs[i].VdevIndex] {
			t.Fatalf("duplicate vdev index %d across containers", allocs[i].VdevIndex)
		}
		seen[allocs[i].VdevIndex] = true
	}

	states, err := m.States().List()
	if err != nil || len(states) != n {
		t.Fatalf("expected %d state files, got %d (%v)", n, len(states), err)
	}
}
