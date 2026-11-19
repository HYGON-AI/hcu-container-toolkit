/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func testStore(t *testing.T) *StateStore {
	t.Helper()
	dir := t.TempDir()
	return NewStateStore(filepath.Join(dir, "vhcu"), filepath.Join(dir, "vhcu.lock"), 2*time.Second)
}

func TestStateSaveLoadDelete(t *testing.T) {
	store := testStore(t)

	st := &State{
		ContainerID: "a4e19862",
		Phase:       PhaseCreating,
		DeviceIndex: 0,
		VdevIndex:   Unset,
		PciBusID:    "0000:81:00.0",
		Cores:       30,
		MemMiB:      8192,
	}
	if err := store.Save(st); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load("a4e19862")
	if err != nil || loaded == nil {
		t.Fatalf("load failed: %v %v", loaded, err)
	}
	if loaded.Phase != PhaseCreating || loaded.DeviceIndex != 0 || loaded.PciBusID != "0000:81:00.0" {
		t.Fatalf("roundtrip mismatch: %+v", loaded)
	}

	// Missing state is (nil, nil).
	missing, err := store.Load("nope")
	if err != nil || missing != nil {
		t.Fatalf("missing state must be nil/nil, got %v %v", missing, err)
	}

	if err := store.Delete("a4e19862"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if err := store.Delete("a4e19862"); err != nil {
		t.Fatalf("delete must be idempotent: %v", err)
	}
	if _, err := os.Stat(store.Path("a4e19862")); !os.IsNotExist(err) {
		t.Fatal("state file should be gone")
	}
}

func TestStateListReportsCorruptRecords(t *testing.T) {
	store := testStore(t)

	if err := store.Save(&State{ContainerID: "c1", Phase: PhaseActive, VdevIndex: 3}); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(store.dir, "garbage.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store.dir, "notes.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := store.List(); err == nil {
		t.Fatal("corrupt ownership must be reported, not silently skipped")
	}

}

func TestStateWithLockSerializes(t *testing.T) {
	store := testStore(t)
	var inside, maxInside atomic.Int32
	start := make(chan struct{})
	results := make(chan error, 2)
	body := func() error {
		n := inside.Add(1)
		defer inside.Add(-1)
		for previous := maxInside.Load(); n > previous; previous = maxInside.Load() {
			if maxInside.CompareAndSwap(previous, n) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		return nil
	}
	for i := 0; i < 2; i++ {
		go func() { <-start; results <- store.WithLock(body) }()
	}
	close(start)
	for i := 0; i < 2; i++ {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	if maxInside.Load() != 1 {
		t.Fatalf("observed %d concurrent lock holders", maxInside.Load())
	}
}

func TestStateRejectsMismatchedOwnerAndPathTraversal(t *testing.T) {
	store := testStore(t)
	if err := store.Save(&State{ContainerID: "owner", Phase: PhaseActive, VdevIndex: 0}); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(store.Path("owner"), store.Path("other")); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load("other"); err == nil {
		t.Fatal("mismatched ownership record accepted")
	}
	if err := store.Save(&State{ContainerID: "../escaped", Phase: PhaseActive}); err == nil {
		t.Fatal("container ID escaped state directory")
	}
}

func TestFailedStateWritePreservesPreviousRecord(t *testing.T) {
	store := testStore(t)
	st := &State{ContainerID: "owner", Phase: PhaseActive, VdevIndex: 0}
	if err := store.Save(st); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(store.Path("owner")+".tmp", 0700); err != nil {
		t.Fatal(err)
	}
	st.Phase = PhaseReleasing
	if err := store.Save(st); err == nil {
		t.Fatal("write failure was ignored")
	}
	saved, err := store.Load("owner")
	if err != nil || saved.Phase != PhaseActive {
		t.Fatalf("failed write damaged previous ownership: %+v %v", saved, err)
	}
}
