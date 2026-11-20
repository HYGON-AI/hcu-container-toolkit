/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte("fake conf\n"), 0o600)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TestCoresZeroMemoryOnlySplit covers the cu_count=0 "pure memory split"
// semantics from the reference implementation.
func TestCoresZeroMemoryOnlySplit(t *testing.T) {
	m, fake := testManager(t, newFakeBackend(testCard()))

	alloc, err := m.Allocate("ctr1", &Request{Cores: 0, MemMiB: 8192})
	if err != nil {
		t.Fatalf("memory-only split failed: %v", err)
	}
	if alloc.ComputeUnits != 0 {
		t.Fatalf("expected 0 CU, got %d", alloc.ComputeUnits)
	}
	if len(fake.createCalls) != 1 || fake.createCalls[0] != "0:0:8192" {
		t.Fatalf("expected cu=0 create call, got %v", fake.createCalls)
	}
}
