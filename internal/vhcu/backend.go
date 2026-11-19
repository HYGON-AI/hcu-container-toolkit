/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import "fmt"

// PhysicalDevice describes a physical HCU as reported by the driver.
type PhysicalDevice struct {
	Index        int
	BusID        string
	ComputeUnits int
	MemoryMiB    int
	// UsageMiB is the driver-reported usage memory. Used only to derive a
	// remaining view when `--show-remaining-info` is unavailable.
	UsageMiB   int
	UsageKnown bool
}

// VDeviceInfo describes an existing vdev instance (toolkit-created or
// administratively pre-split). ComputeUnits/MemMiB are captured when the
// listing prints them and may be zero otherwise.
type VDeviceInfo struct {
	// Index is the global vdev index used in /etc/vdev/vdev<N>.conf.
	Index int
	// DeviceIndex is the parent physical device index.
	DeviceIndex int
	// ComputeUnits / MemMiB are the instance's resources when reported.
	ComputeUnits   int
	MemMiB         int
	ResourcesKnown bool
}

// Remaining is the driver's view of what is still allocatable on a card.
type Remaining struct {
	DeviceIndex  int
	ComputeUnits int
	MemoryMiB    int
	// Derived reports that the values were computed from device/usage info
	// (fallback) rather than read from `--show-remaining-info`.
	Derived bool
}

// Backend abstracts the vdev control plane. The only production
// implementation is the `hy-smi virtual` CLI (backend_smi.go); tests inject
// fakes. Keeping this interface is what makes Create/Release/GC testable
// without a driver. A future CGO-free hcu-dcgm implementation can slot in
// as an additional Backend without touching manager.go.
type Backend interface {
	// PhysicalDevices lists all physical HCUs with their total CU count and
	// memory. Required for percentage conversion and full-card detection.
	PhysicalDevices() ([]PhysicalDevice, error)
	// ListVDevices returns the vdev instances on the given physical device.
	ListVDevices(dev int) ([]VDeviceInfo, error)
	// CreateVDevice creates one vdev with the given CU count and memory
	// (MiB) on device dev. The CLI does not return the new vdev index;
	// callers locate it via ListVDevices before/after diffs.
	CreateVDevice(dev, cu, memMiB int) error
	// StopVDevice stops the vdev with the given global index.
	StopVDevice(vdev int) error
	// DestroyVDevice destroys the vdev with the given global index.
	// Implementations MUST target a single vdev (-v <idx>); destroying by
	// device (-d) would take out administratively pre-split vdevs and vdevs
	// owned by other containers.
	DestroyVDevice(vdev int) error
	// Remaining returns the driver-reported remaining CU / memory for dev.
	Remaining(dev int) (Remaining, error)
}

// deviceAwareRemaining is an optional Backend extension: it lets the caller
// hand over the already-fetched PhysicalDevice so the derived-remaining
// fallback does not re-query `--show-device-info` once per card.
type deviceAwareRemaining interface {
	RemainingForDevice(dev PhysicalDevice) (Remaining, error)
}

// ErrNoBackend is a sentinel for backends that cannot be constructed.
var ErrNoBackend = fmt.Errorf("no vhcu backend available")
