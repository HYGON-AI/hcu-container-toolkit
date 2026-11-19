/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"strings"
	"testing"
)

func testViews() []CardView {
	return []CardView{
		{
			Dev:       PhysicalDevice{Index: 0, BusID: "0000:81:00.0", ComputeUnits: 120, MemoryMiB: 65536},
			Remaining: Remaining{DeviceIndex: 0, ComputeUnits: 90, MemoryMiB: 49152},
			UsedSlots: 1,
		},
		{
			Dev:       PhysicalDevice{Index: 1, BusID: "0000:c1:00.0", ComputeUnits: 120, MemoryMiB: 65536},
			Remaining: Remaining{DeviceIndex: 1, ComputeUnits: 120, MemoryMiB: 65536},
			UsedSlots: 0,
		},
	}
}

func TestSelectBestFit(t *testing.T) {
	a := NewAllocator(nil, 4)
	req := &Request{Cores: 30, MemMiB: 8192}

	// Device 0 leftover: 49152-8192=40960; device 1 leftover: 65536-8192=57344.
	// Device 0 has the smaller leftover -> best fit.
	sel, err := a.Select(req, testViews())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.DeviceIndex != 0 {
		t.Fatalf("expected device 0 (best fit), got %d", sel.DeviceIndex)
	}
	if sel.ComputeUnits != 36 {
		t.Fatalf("expected 36 CU (30%% of 120), got %d", sel.ComputeUnits)
	}
}

func TestSelectTieBreaksToLowestIndex(t *testing.T) {
	a := NewAllocator(nil, 4)
	views := []CardView{
		{Dev: PhysicalDevice{Index: 2, BusID: "b", ComputeUnits: 120, MemoryMiB: 65536},
			Remaining: Remaining{DeviceIndex: 2, ComputeUnits: 120, MemoryMiB: 16384}, UsedSlots: 0},
		{Dev: PhysicalDevice{Index: 0, BusID: "a", ComputeUnits: 120, MemoryMiB: 65536},
			Remaining: Remaining{DeviceIndex: 0, ComputeUnits: 120, MemoryMiB: 16384}, UsedSlots: 0},
	}
	sel, err := a.Select(&Request{Cores: 30, MemMiB: 8192}, views)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.DeviceIndex != 0 {
		t.Fatalf("tie must break to lowest index, got %d", sel.DeviceIndex)
	}
}

func TestSelectDevicePin(t *testing.T) {
	a := NewAllocator(nil, 4)
	views := testViews()

	sel, err := a.Select(&Request{Cores: 30, MemMiB: 8192, Device: "1"}, views)
	if err != nil || sel.DeviceIndex != 1 {
		t.Fatalf("pin by index failed: %v %+v", err, sel)
	}

	if _, err := a.Select(&Request{Cores: 30, MemMiB: 8192, Device: "0000:c1:00.0"}, views); err == nil {
		t.Fatal("BDF must be rejected")
	}
	if _, err := a.Select(&Request{Cores: 30, MemMiB: 8192, Device: "0,1"}, views); err == nil {
		t.Fatal("device list must be rejected")
	}
	if _, err := a.Select(&Request{Cores: 30, MemMiB: 8192, Device: "99"}, views); err == nil {
		t.Fatal("unknown index must fail")
	}

}

func TestSelectSlotLimit(t *testing.T) {
	a := NewAllocator(nil, 1) // one slot per card for this test
	views := testViews()      // device 0 already has 1 used slot

	sel, err := a.Select(&Request{Cores: 30, MemMiB: 8192}, views)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.DeviceIndex != 1 {
		t.Fatalf("device 0 is full (slot limit 1); expected device 1, got %d", sel.DeviceIndex)
	}
}

func TestSelectInsufficientCapacityReportsDetails(t *testing.T) {
	a := NewAllocator(nil, 4)
	req := &Request{Cores: 100, MemMiB: 60000}
	views := []CardView{
		// Card 0: enough memory, not enough CU.
		{Dev: PhysicalDevice{Index: 0, BusID: "a", ComputeUnits: 120, MemoryMiB: 65536},
			Remaining: Remaining{DeviceIndex: 0, ComputeUnits: 90, MemoryMiB: 65536}, UsedSlots: 0},
		// Card 1: enough CU, not enough memory.
		{Dev: PhysicalDevice{Index: 1, BusID: "b", ComputeUnits: 120, MemoryMiB: 65536},
			Remaining: Remaining{DeviceIndex: 1, ComputeUnits: 120, MemoryMiB: 49152}, UsedSlots: 0},
	}
	_, err := a.Select(req, views)
	if err == nil {
		t.Fatal("expected capacity error")
	}
	// The error must carry per-card details, not a bare "no space".
	for _, want := range []string{"device 0", "remaining CU 90", "device 1", "remaining memory 49152"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should contain %q: %v", want, err)
		}
	}
}

func TestSelectRejectsMissingMemory(t *testing.T) {
	// HCU_VHCU_MEM is mandatory: cores alone (MemMiB unset) must not reach
	// card selection, let alone be silently filled in.
	a := NewAllocator(nil, 4)
	views := []CardView{
		{Dev: PhysicalDevice{Index: 0, BusID: "a", ComputeUnits: 120, MemoryMiB: 65536},
			Remaining: Remaining{DeviceIndex: 0, ComputeUnits: 120, MemoryMiB: 30000}, UsedSlots: 1},
	}
	if _, err := a.Select(&Request{Cores: 30}, views); err == nil {
		t.Fatal("request without memory must be rejected")
	}
}

func TestSelectExcludedCard(t *testing.T) {
	a := NewAllocator(nil, 4)
	views := testViews()
	views[0].Excluded = true
	views[0].ExcludedReason = "exclusive device in use by the HCU tracker"

	sel, err := a.Select(&Request{Cores: 30, MemMiB: 8192}, views)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.DeviceIndex != 1 {
		t.Fatalf("excluded card must be skipped, got device %d", sel.DeviceIndex)
	}
}

func TestSelectCoresZeroNeedsNoCU(t *testing.T) {
	a := NewAllocator(nil, 4)
	views := []CardView{
		{Dev: PhysicalDevice{Index: 0, BusID: "a", ComputeUnits: 120, MemoryMiB: 65536},
			Remaining: Remaining{DeviceIndex: 0, ComputeUnits: 0, MemoryMiB: 49152}, UsedSlots: 0},
	}
	sel, err := a.Select(&Request{Cores: 0, MemMiB: 8192}, views)
	if err != nil {
		t.Fatalf("memory-only split needs no CU: %v", err)
	}
	if sel.ComputeUnits != 0 {
		t.Fatalf("expected 0 CU, got %d", sel.ComputeUnits)
	}
}
