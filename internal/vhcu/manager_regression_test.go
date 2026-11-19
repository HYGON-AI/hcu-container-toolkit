/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"hcu-container-toolkit/internal/config/image"

	"github.com/gofrs/flock"
)

func TestMemoryOnlyRequestUsesSharedCU(t *testing.T) {
	req, err := ParseRequest(envMap(map[string]string{image.EnvVarVHCUMem: "4096"}))
	if err != nil {
		t.Fatal(err)
	}
	m, b := testManager(t, newFakeBackend(testCard()))
	alloc, err := m.Allocate("memory", req)
	if err != nil {
		t.Fatal(err)
	}
	if alloc.ComputeUnits != 0 || b.createCalls[0] != "0:0:4096" {
		t.Fatalf("memory-only request consumed CUs: %+v %v", alloc, b.createCalls)
	}
}

func TestAllocationQueryFailurePreservesOwnership(t *testing.T) {
	m, b := testManager(t, newFakeBackend(testCard()))
	req := &Request{Cores: 30, MemMiB: 8192}
	if _, err := m.Allocate("owner", req); err != nil {
		t.Fatal(err)
	}
	b.listErr = errors.New("transient driver query failure")
	if _, err := m.Allocate("owner", req); err == nil {
		t.Fatal("query error must abort allocation")
	}
	st, err := m.States().Load("owner")
	if err != nil || st == nil || st.Phase != PhaseActive || len(b.createCalls) != 1 {
		t.Fatalf("ownership lost on query failure: %+v %v", st, err)
	}
}

func TestAllocationCorruptStateIsNotOverwritten(t *testing.T) {
	m, b := testManager(t, newFakeBackend(testCard()))
	req := &Request{Cores: 30, MemMiB: 8192}
	if _, err := m.Allocate("owner", req); err != nil {
		t.Fatal(err)
	}
	path := m.States().Path("owner")
	if err := os.WriteFile(path, []byte("{corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Allocate("owner", req); err == nil {
		t.Fatal("invalid ownership must abort allocation")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "{corrupt" || len(b.createCalls) != 1 {
		t.Fatal("corrupt record was replaced")
	}
}

func TestCleanupFailureRetainedAcrossReconcileAndGC(t *testing.T) {
	m, b := testManager(t, newFakeBackend(testCard()))
	req := &Request{Cores: 30, MemMiB: 8192}
	alloc, err := m.Allocate("owner", req)
	if err != nil {
		t.Fatal(err)
	}
	b.destroyFails = 99
	if err := m.ReleaseAllocation("owner", alloc.AllocationID); err == nil {
		t.Fatal("destroy failure was ignored")
	}
	if _, err := m.Allocate("owner", req); err == nil {
		t.Fatal("failed reconciliation must prevent reallocation")
	}
	st, err := m.States().Load("owner")
	if err != nil || st == nil || st.Phase != PhaseReleasing || !fileExists(m.confPath(alloc.VdevIndex)) || len(b.createCalls) != 1 {
		t.Fatalf("failed cleanup lost ownership or conf: %+v %v", st, err)
	}
	b.destroyFails = 0
	result, err := m.GC(nil)
	if err != nil || len(result.Released) != 1 {
		t.Fatalf("GC could not retry releasing state: %+v %v", result, err)
	}
	if st, err := m.States().Load("owner"); err != nil || st != nil {
		t.Fatalf("successful cleanup kept state: %+v %v", st, err)
	}
}

func TestRecycledInstancesAreNeverDestroyed(t *testing.T) {
	for _, change := range []string{"configuration", "boot", "resources", "parent", "legacy"} {
		t.Run(change, func(t *testing.T) {
			m, b := testManager(t, newFakeBackend(testCardN(0), testCardN(1)))
			alloc, err := m.Allocate("owner", &Request{Cores: 30, MemMiB: 8192, Device: "0"})
			if err != nil {
				t.Fatal(err)
			}
			switch change {
			case "configuration":
				if err := os.WriteFile(m.confPath(alloc.VdevIndex), []byte("replacement instance"), 0600); err != nil {
					t.Fatal(err)
				}
			case "boot":
				m.bootID = func() (string, error) { return "next-boot", nil }
			case "resources":
				b.vdevs[alloc.VdevIndex] = fakeVdev{dev: 0, cu: 12, mem: 1024}
			case "parent":
				b.vdevs[alloc.VdevIndex] = fakeVdev{dev: 1, cu: 36, mem: 8192}
			case "legacy":
				st, _ := m.States().Load("owner")
				st.Identity = ""
				if err := m.States().Save(st); err != nil {
					t.Fatal(err)
				}
			}
			if err := m.ReleaseAllocation("owner", alloc.AllocationID); err == nil {
				t.Fatal("changed identity was accepted")
			}
			st, err := m.States().Load("owner")
			if err != nil || st == nil || st.Phase != PhasePending || b.destroyCalls != 0 || b.stopCalls != 0 || !fileExists(m.confPath(alloc.VdevIndex)) {
				t.Fatalf("foreign instance modified: %+v %v", st, err)
			}
		})
	}
}

func TestUncertainCreatePreservesEvidence(t *testing.T) {
	for _, failure := range []string{"query", "create"} {
		t.Run(failure, func(t *testing.T) {
			m, b := testManager(t, newFakeBackend(testCard()))
			if failure == "query" {
				b.failListAfterCreate = true
			} else {
				b.createErr = errors.New("create timed out after driver side effect")
			}
			req := &Request{Cores: 30, MemMiB: 8192}
			if _, err := m.Allocate("owner", req); err == nil {
				t.Fatal("uncertain create must fail")
			}
			st, err := m.States().Load("owner")
			if err != nil || st == nil || st.Phase != PhasePending || st.LastError == "" || len(b.vdevs) != 1 {
				t.Fatalf("create evidence lost: %+v %v", st, err)
			}
			b.failListAfterCreate = false
			if _, err := m.Allocate("owner", req); err == nil {
				t.Fatal("pending record must block new create")
			}
			if _, err := m.GC(func(*State) (ContainerStatus, error) { return ContainerAbsent, nil }); err != nil {
				t.Fatal(err)
			}
			if b.destroyCalls != 0 || len(b.createCalls) != 1 {
				t.Fatal("uncertain ownership was destroyed or overwritten")
			}
		})
	}
}

func TestSilentDestroyFailureRetainsRecord(t *testing.T) {
	m, b := testManager(t, newFakeBackend(testCard()))
	if _, err := m.Allocate("owner", &Request{Cores: 30, MemMiB: 8192}); err != nil {
		t.Fatal(err)
	}
	b.destroyNoop = true
	if err := m.Release("owner"); err == nil {
		t.Fatal("successful CLI exit is not confirmed destruction")
	}
	st, err := m.States().Load("owner")
	if err != nil || st == nil || st.Phase != PhaseReleasing || len(b.vdevs) != 1 || !fileExists(m.confPath(0)) {
		t.Fatalf("silent destroy lost state: %+v %v", st, err)
	}
}

func TestOldHookCannotReleaseNewAllocation(t *testing.T) {
	m, b := testManager(t, newFakeBackend(testCard()))
	req := &Request{Cores: 30, MemMiB: 8192}
	first, err := m.Allocate("owner", req)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.ReleaseAllocation("owner", first.AllocationID); err != nil {
		t.Fatal(err)
	}
	second, err := m.Allocate("owner", req)
	if err != nil {
		t.Fatal(err)
	}
	if first.AllocationID == second.AllocationID {
		t.Fatal("allocation token reused")
	}
	if err := m.ReleaseAllocation("owner", first.AllocationID); err == nil {
		t.Fatal("old hook token was accepted")
	}
	if len(b.vdevs) != 1 || b.destroyCalls != 1 {
		t.Fatal("old hook destroyed the replacement")
	}
}

func TestGCUsesConservativeLiveStatus(t *testing.T) {
	for _, scenario := range []string{"unknown", "failure", "live", "stopped", "absent", "became-live"} {
		t.Run(scenario, func(t *testing.T) {
			m, b := testManager(t, newFakeBackend(testCard()))
			if _, err := m.Allocate("owner", &Request{Cores: 30, MemMiB: 8192}); err != nil {
				t.Fatal(err)
			}
			calls := 0
			probe := func(*State) (ContainerStatus, error) {
				calls++
				lock := flock.New(m.states.lockFile)
				locked, err := lock.TryLock()
				if locked {
					_ = lock.Unlock()
					t.Error("GC probed outside allocation lock")
				}
				if err != nil {
					t.Fatal(err)
				}
				switch scenario {
				case "unknown":
					return ContainerUnknown, nil
				case "failure":
					return ContainerUnknown, errors.New("owning runtime unavailable")
				case "live":
					return ContainerLive, nil
				case "stopped":
					return ContainerStopped, nil
				case "became-live":
					if calls > 1 {
						return ContainerLive, nil
					}
				}
				return ContainerAbsent, nil
			}
			result, err := m.GC(probe)
			if err != nil {
				t.Fatal(err)
			}
			wantReleased := scenario == "stopped" || scenario == "absent"
			if (len(result.Released) == 1) != wantReleased || (b.destroyCalls == 1) != wantReleased {
				t.Fatalf("unsafe GC decision for %s: %+v destroy=%d", scenario, result, b.destroyCalls)
			}
		})
	}
}

func TestGCRespectsCreateGraceAndStoresRuntimeScope(t *testing.T) {
	m, b := testManager(t, newFakeBackend(testCard()))
	m.gcGrace = time.Minute
	command := []string{"/usr/bin/runc", "--root", "/run/containerd/runc/k8s.io", "list", "--format", "json"}
	if _, err := m.Allocate("owner", &Request{Cores: 30, MemMiB: 8192, RuntimeCommand: command}); err != nil {
		t.Fatal(err)
	}
	st, err := m.States().Load("owner")
	if err != nil || !reflect.DeepEqual(st.RuntimeCommand, command) {
		t.Fatalf("runtime scope lost: %+v %v", st, err)
	}
	if _, err := m.GC(func(*State) (ContainerStatus, error) {
		t.Error("fresh create should not be probed")
		return ContainerAbsent, nil
	}); err != nil {
		t.Fatal(err)
	}
	if b.destroyCalls != 0 {
		t.Fatal("fresh allocation reclaimed")
	}
}

func TestAllocationRejectsDifferentRuntimeScopeBeforeReconcile(t *testing.T) {
	for _, phase := range []string{PhaseActive, PhaseReleasing, PhaseCreating, PhasePending} {
		for _, scope := range []string{"root", "runtime", "missing"} {
			t.Run(phase+"/"+scope, func(t *testing.T) {
				m, b := testManager(t, newFakeBackend(testCard()))
				command := []string{"/usr/bin/runc", "--root", "/run/owner", "list", "--format", "json"}
				req := &Request{Cores: 30, MemMiB: 8192, RuntimeCommand: command}
				if _, err := m.Allocate("owner", req); err != nil {
					t.Fatal(err)
				}
				st, err := m.States().Load("owner")
				if err != nil {
					t.Fatal(err)
				}
				st.Phase = phase
				if err := m.States().Save(st); err != nil {
					t.Fatal(err)
				}
				req.RuntimeCommand = append([]string(nil), command...)
				switch scope {
				case "root":
					req.RuntimeCommand[2] = "/run/other"
				case "runtime":
					req.RuntimeCommand[0] = "/usr/bin/crun"
				case "missing":
					req.RuntimeCommand = nil
				}
				if alloc, err := m.Allocate("owner", req); err == nil || alloc != nil {
					t.Fatal("different runtime scope reused ownership")
				}
				after, err := m.States().Load("owner")
				if err != nil || !reflect.DeepEqual(st, after) || b.destroyCalls != 0 || b.stopCalls != 0 || len(b.createCalls) != 1 {
					t.Fatalf("scope conflict modified existing owner: %+v %v", after, err)
				}
			})
		}
	}
}

func TestReuseRefreshesCreateProtection(t *testing.T) {
	m, b := testManager(t, newFakeBackend(testCard()))
	req := &Request{Cores: 30, MemMiB: 8192}
	first, err := m.Allocate("owner", req)
	if err != nil {
		t.Fatal(err)
	}
	st, err := m.States().Load("owner")
	if err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Hour).UTC()
	st.CreatedAt, st.LeaseAt = old, old
	if err := m.States().Save(st); err != nil {
		t.Fatal(err)
	}
	m.creator = func() (int, string, error) { return 123, "new-process-start", nil }
	reused, err := m.Allocate("owner", req)
	if err != nil || !reused.Reused || reused.AllocationID != first.AllocationID {
		t.Fatalf("failed to reuse original allocation: %+v %v", reused, err)
	}
	st, err = m.States().Load("owner")
	if err != nil || !st.CreatedAt.Equal(old) || !st.LeaseAt.After(old) || st.CreatorPID != 123 || st.CreatorStartTime != "new-process-start" {
		t.Fatalf("reuse did not persist new create protection: %+v %v", st, err)
	}
	m.gcGrace = time.Minute
	if _, err := m.GC(func(*State) (ContainerStatus, error) {
		t.Error("reused allocation should remain in create grace")
		return ContainerAbsent, nil
	}); err != nil {
		t.Fatal(err)
	}
	if b.destroyCalls != 0 || len(b.createCalls) != 1 {
		t.Fatal("reuse created or destroyed hardware")
	}
}

func TestReuseProtectionSaveFailurePreservesAllocation(t *testing.T) {
	m, b := testManager(t, newFakeBackend(testCard()))
	req := &Request{Cores: 30, MemMiB: 8192}
	if _, err := m.Allocate("owner", req); err != nil {
		t.Fatal(err)
	}
	before, err := m.States().Load("owner")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(m.States().Path("owner")+".tmp", 0700); err != nil {
		t.Fatal(err)
	}
	if alloc, err := m.Allocate("owner", req); err == nil || alloc != nil {
		t.Fatal("reuse without persistent create protection succeeded")
	}
	after, err := m.States().Load("owner")
	if err != nil || !reflect.DeepEqual(before, after) || b.destroyCalls != 0 || len(b.vdevs) != 1 {
		t.Fatalf("failed reuse lost existing allocation: %+v %v", after, err)
	}
}

// DTK 6.3.31 changes shared-CU reporting from 0 to the physical CU count
// after a container enables the vdev; the original configuration stays intact.
func TestSharedCUReportingAfterActivation(t *testing.T) {
	for _, scenario := range []string{"reuse", "release", "invalid-cu", "changed-memory", "create-active"} {
		t.Run(scenario, func(t *testing.T) {
			m, b := testManager(t, newFakeBackend(testCard()))
			req := &Request{Cores: 0, MemMiB: 2048}
			if scenario == "create-active" {
				b.createHook = func(b *fakeBackend, dev, cu, mem int) {
					b.vdevs[0] = fakeVdev{dev: dev, cu: 120, mem: mem}
				}
			}
			alloc, err := m.Allocate("shared", req)
			if err != nil {
				t.Fatal(err)
			}
			b.vdevs[alloc.VdevIndex] = fakeVdev{dev: 0, cu: 120, mem: 2048}
			if scenario == "invalid-cu" {
				b.vdevs[alloc.VdevIndex] = fakeVdev{dev: 0, cu: 36, mem: 2048}
			}
			if scenario == "changed-memory" {
				b.vdevs[alloc.VdevIndex] = fakeVdev{dev: 0, cu: 120, mem: 1024}
			}
			if scenario == "reuse" {
				reused, err := m.Allocate("shared", req)
				if err != nil || !reused.Reused || reused.ComputeUnits != 0 {
					t.Fatalf("shared-CU activation prevented reuse: %+v %v", reused, err)
				}
			}
			err = m.ReleaseAllocation("shared", alloc.AllocationID)
			if scenario == "invalid-cu" || scenario == "changed-memory" {
				if err == nil || b.destroyCalls != 0 {
					t.Fatalf("unexpected resources were destroyed: %v", err)
				}
				return
			}
			if err != nil || b.destroyCalls != 1 || len(b.vdevs) != 0 {
				t.Fatalf("shared-CU activation prevented cleanup: %v", err)
			}
		})
	}
}
