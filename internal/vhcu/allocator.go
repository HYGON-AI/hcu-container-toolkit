/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// hcuTrackerFile mirrors internal/hcu-tracker's on-disk state. Read directly
// (fail-open) to exclude cards that the tracker has made exclusive-and-busy;
// a missing or unreadable tracker file simply means no exclusions.
const hcuTrackerFile = "/var/log/hcu-tracker.json"

// CardView is the merged capacity view of one physical card: the driver's
// remaining amounts (authoritative), the used vdev slots, and the tracker
// exclusion flag.
type CardView struct {
	Dev            PhysicalDevice
	Remaining      Remaining
	UsedSlots      int
	Excluded       bool
	ExcludedReason string
}

// Selection is the outcome of picking a card for a request.
type Selection struct {
	DeviceIndex int
	BusID       string
	// ComputeUnits is the CU count actually requested for the new vdev.
	ComputeUnits int
	// MemMiB is the memory requested for the new vdev (HCU_VHCU_MEM is
	// mandatory, so this is always the user's ask).
	MemMiB int
}

// Allocator merges the three capacity sources and implements best-fit card
// selection:
//   - driver remaining (hy-smi virtual --show-remaining-info) is the hard
//     authority; it already accounts for administratively pre-split vdevs;
//   - the toolkit state dir only records ownership (never used to compute
//     remaining amounts, so a wiped state dir degrades selection quality
//     but never causes over-allocation);
//   - vdev slot usage comes from --show-vdevice-info (includes pre-splits).
type Allocator struct {
	backend    Backend
	maxPerCard int
}

// NewAllocator creates an Allocator with the given per-card slot limit.
func NewAllocator(backend Backend, maxPerCard int) *Allocator {
	return &Allocator{backend: backend, maxPerCard: maxPerCard}
}

// Views builds the merged per-card capacity view.
func (a *Allocator) Views() ([]CardView, error) {
	devices, err := a.backend.PhysicalDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to list physical devices: %v", err)
	}

	excluded := excludedByTracker()

	views := make([]CardView, 0, len(devices))
	for _, dev := range devices {
		vdevs, err := a.backend.ListVDevices(dev.Index)
		if err != nil {
			return nil, fmt.Errorf("failed to list vdevices on device %d: %v", dev.Index, err)
		}
		usedSlots := 0
		for _, v := range vdevs {
			if v.DeviceIndex == dev.Index {
				usedSlots++
			}
		}
		rem, err := remainingFor(a.backend, dev)
		if err != nil {
			return nil, fmt.Errorf("failed to get remaining capacity of device %d: %v", dev.Index, err)
		}
		view := CardView{
			Dev:       dev,
			Remaining: rem,
			UsedSlots: usedSlots,
		}
		if reason, ok := excluded[dev.Index]; ok {
			view.Excluded = true
			view.ExcludedReason = reason
		}
		views = append(views, view)
	}
	return views, nil
}

// remainingFor queries the driver's remaining capacity, reusing the
// already-fetched PhysicalDevice when the backend supports it.
func remainingFor(b Backend, dev PhysicalDevice) (Remaining, error) {
	if d, ok := b.(deviceAwareRemaining); ok {
		return d.RemainingForDevice(dev)
	}
	return b.Remaining(dev.Index)
}

// Select applies best-fit: among eligible cards pick the one whose remaining
// memory after the allocation is smallest (least fragmentation), ties broken
// by lowest device index for reproducibility.
func (a *Allocator) Select(req *Request, views []CardView) (*Selection, error) {
	if req.MemMiB <= 0 {
		return nil, fmt.Errorf("dynamic vHCU request requires a positive memory value (MiB)")
	}
	deviceIndex, err := resolveDeviceFilter(req.Device, views)
	if err != nil {
		return nil, err
	}

	var best *CardView
	var bestSel Selection
	var reasons []string

	for i := range views {
		view := &views[i]

		if view.Excluded {
			reasons = append(reasons, fmt.Sprintf("device %d: excluded (%s)", view.Dev.Index, view.ExcludedReason))
			continue
		}
		if deviceIndex != Unset && view.Dev.Index != deviceIndex {
			continue
		}

		if view.UsedSlots >= a.maxPerCard {
			reasons = append(reasons, fmt.Sprintf("device %d: %d/%d vdev slots used", view.Dev.Index, view.UsedSlots, a.maxPerCard))
			continue
		}

		needCU := CoresToComputeUnits(req.Cores, view.Dev.ComputeUnits)
		if view.Remaining.ComputeUnits < needCU {
			reasons = append(reasons, fmt.Sprintf("device %d: remaining CU %d < requested %d", view.Dev.Index, view.Remaining.ComputeUnits, needCU))
			continue
		}

		needMem := req.MemMiB
		if view.Remaining.MemoryMiB < needMem {
			reasons = append(reasons, fmt.Sprintf("device %d: remaining memory %d MiB < requested %d MiB", view.Dev.Index, view.Remaining.MemoryMiB, needMem))
			continue
		}

		candidate := Selection{
			DeviceIndex:  view.Dev.Index,
			BusID:        view.Dev.BusID,
			ComputeUnits: needCU,
			MemMiB:       needMem,
		}

		// Best fit: minimize leftover memory; ties -> lowest device index
		// for reproducibility (views arrive in arbitrary order).
		if best == nil || leftover(view, &candidate) < leftover(best, &bestSel) ||
			(leftover(view, &candidate) == leftover(best, &bestSel) && view.Dev.Index < best.Dev.Index) {
			best = view
			bestSel = candidate
		}
	}

	if best == nil {
		return nil, fmt.Errorf("no physical device can satisfy the request (%s)", strings.Join(reasons, "; "))
	}
	return &bestSel, nil
}

func leftover(view *CardView, sel *Selection) int {
	return view.Remaining.MemoryMiB - sel.MemMiB
}

// resolveDeviceFilter maps the optional HCU_VHCU_DEVICE index
// to a device index; Unset means no pin.
func resolveDeviceFilter(device string, views []CardView) (int, error) {
	if device == "" {
		return Unset, nil
	}
	if isDeviceIndex(device) {
		idx, _ := strconv.Atoi(device)
		for _, v := range views {
			if v.Dev.Index == idx {
				return idx, nil
			}
		}
		return Unset, fmt.Errorf("device %s not found", device)
	}
	return Unset, fmt.Errorf("invalid physical device index %q: expected a single non-negative integer", device)
}

// trackerStatus is the subset of internal/hcu-tracker's state file needed to
// keep dynamic splits away from exclusive / whole-card-occupied devices.
type trackerStatus struct {
	Enabled    bool `json:"enabled"`
	HCUsStatus map[string]struct {
		Accessibility int      `json:"accessibility"`
		ContainerIds  []string `json:"containerIds"`
	} `json:"hcusStatus"`
}

// excludedByTracker returns device indices that must not host dynamic vHCU
// splits, with a human-readable reason. Fail-open on any read/parse error.
func excludedByTracker() map[int]string {
	data, err := os.ReadFile(hcuTrackerFile)
	if err != nil {
		return nil
	}
	var tracker trackerStatus
	if err := json.Unmarshal(data, &tracker); err != nil {
		return nil
	}
	if !tracker.Enabled {
		return nil
	}
	// EXCLUSIVE_ACCESS == 1 in internal/hcu-tracker.
	excluded := make(map[int]string)
	for id, status := range tracker.HCUsStatus {
		idx, err := strconv.Atoi(id)
		if err != nil {
			continue
		}
		if status.Accessibility == 1 && len(status.ContainerIds) > 0 {
			excluded[idx] = "exclusive device in use by the HCU tracker"
		}
	}
	return excluded
}

// SortViews orders card views by device index (used by status output).
func SortViews(views []CardView) {
	sort.Slice(views, func(i, j int) bool { return views[i].Dev.Index < views[j].Dev.Index })
}
