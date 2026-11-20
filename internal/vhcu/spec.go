/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

// Package vhcu implements dynamic vHCU split-and-mount: selecting a physical
// HCU, creating a vdev on it via `hy-smi virtual`, handing the vdev index back
// to the existing VHCU_VISIBLE_DEVICES mount path, and releasing the vdev when
// the owning container goes away.
//
// The semantics mirror k8s-hcu-device-plugin's HAMi mode (AllocateHAMiHCUs):
//   - cores is a 0-100 percentage; 0 means a memory-only split sharing all CUs;
//   - mem is mandatory (MiB): a request carrying cores without mem is rejected
//     at parse time rather than silently filled in;
//   - percentage is converted to CU count with ceil(pct * CU / 100), minimum 1
//     when pct > 0;
//   - a request of 100% cores plus full memory short-circuits to mounting the
//     physical card without creating a vdev;
//   - at most one dynamic vHCU per container.
package vhcu

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"hcu-container-toolkit/internal/config/image"
)

// Unset marks a request field as not provided by the user.
const Unset = -1

// Request is a parsed dynamic vHCU allocation request.
type Request struct {
	// Cores is the requested compute percentage (0-100), defaulting to 0
	// for a memory-only split.
	Cores int
	// MemMiB is the requested memory in MiB. It is mandatory: a parsed
	// request (non-nil) always carries a positive value.
	MemMiB int
	// Device pins a single physical card by its non-negative integer index.
	// Empty means auto-select.
	Device string
	// RuntimeCommand lists containers in the owning OCI runtime's root.
	RuntimeCommand []string
}

// ParseRequest extracts a dynamic vHCU request from the container environment.
// getenv is expected to be image.DTK.Getenv or an equivalent lookup on the
// OCI spec's Process.Env.
//
// It returns (nil, nil) when no dynamic split was requested (neither
// HCU_VHCU_CORES nor HCU_VHCU_MEM set) so callers can keep their legacy paths
// untouched. HCU_VHCU_MEM is mandatory: cores alone cannot express a split.
// Validation errors are returned eagerly so that allocation never happens for
// a request that will be rejected later in the modifier chain.
func ParseRequest(getenv func(string) string) (*Request, error) {
	coresRaw := getenv(image.EnvVarVHCUCores)
	memRaw := getenv(image.EnvVarVHCUMem)
	coresSet := coresRaw != ""
	memSet := memRaw != ""

	if !coresSet && !memSet {
		if getenv(image.EnvVarVHCUDevice) != "" {
			return nil, fmt.Errorf("%s requires %s", image.EnvVarVHCUDevice, image.EnvVarVHCUMem)
		}
		return nil, nil
	}
	if !memSet {
		return nil, fmt.Errorf("%s is required for a dynamic vHCU split: set it to the requested memory in MiB (cores alone cannot express a split)", image.EnvVarVHCUMem)
	}

	req := &Request{
		Cores:  0,
		MemMiB: Unset,
	}

	if coresSet {
		cores, err := strconv.Atoi(strings.TrimSpace(coresRaw))
		if err != nil {
			return nil, fmt.Errorf("invalid %s value %q: must be an integer between 0 and 100", image.EnvVarVHCUCores, coresRaw)
		}
		if cores < 0 || cores > 100 {
			return nil, fmt.Errorf("invalid %s value %d: must be between 0 and 100", image.EnvVarVHCUCores, cores)
		}
		req.Cores = cores
	}

	if memSet {
		mem, err := strconv.Atoi(strings.TrimSpace(memRaw))
		if err != nil || mem <= 0 {
			return nil, fmt.Errorf("invalid %s value %q: must be a positive integer (MiB)", image.EnvVarVHCUMem, memRaw)
		}
		req.MemMiB = mem
	}

	if device := getenv(image.EnvVarVHCUDevice); device != "" {
		device = strings.TrimSpace(device)
		if !isDeviceIndex(device) {
			return nil, fmt.Errorf("invalid %s value %q: must be a single non-negative integer device index (for example 0); lists and PCI BDFs are unsupported", image.EnvVarVHCUDevice, device)
		}
		index, _ := strconv.Atoi(device)
		req.Device = strconv.Itoa(index)
	}

	// The dynamic split drives the full card selection itself; requesting
	// specific physical devices at the same time is ambiguous. Validate this
	// before any vdev is created so we never "create first, fail later".
	for _, envVar := range []string{image.EnvVarHCUVisibleDevices, image.EnvVarNvidiaVisibleDevices, image.EnvVarVHCUVisibleDevices, image.EnvVarMIGVisibleDevices} {
		if value := getenv(envVar); value != "" {
			return nil, fmt.Errorf("%s and %s cannot be set at the same time", envVar, image.EnvVarVHCUCores+"/"+image.EnvVarVHCUMem)
		}
	}

	// When the HAMi device plugin is responsible for splitting, this toolkit
	// must stay a pure mount role; a second allocation here would double-book
	// the card.
	if strategy := getenv(image.EnvVarDevicePluginMode); strings.EqualFold(strings.TrimSpace(strategy), "hami") {
		return nil, fmt.Errorf("%s=hami and %s cannot be set at the same time: the device plugin owns vHCU creation in HAMi mode",
			image.EnvVarDevicePluginMode, image.EnvVarVHCUCores+"/"+image.EnvVarVHCUMem)
	}

	return req, nil
}

func isDeviceIndex(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	i, err := strconv.Atoi(s)
	return err == nil && i >= 0
}

// CoresToComputeUnits converts a core percentage into a CU count using the
// same rule as k8s-hcu-device-plugin: ceil(pct * totalCU / 100), minimum 1
// when pct > 0. A pct of 0 stays 0 (memory-only split sharing all CUs).
func CoresToComputeUnits(pct, totalCU int) int {
	if pct == 0 {
		return 0
	}
	if pct > 100 {
		pct = 100
	}
	units := int(math.Ceil(float64(pct) * float64(totalCU) / 100.0))
	if units < 1 {
		units = 1
	}
	return units
}

// IsFullCardRequest reports whether the request amounts to the whole card and
// should short-circuit to a plain physical device mount (no vdev created),
// matching plugin.go's Usedcores == 100 && Usedmem == total branch. memMiB is
// the mandatory HCU_VHCU_MEM request and is always positive.
func IsFullCardRequest(cores, memMiB, totalMemMiB int) bool {
	if cores != 100 {
		return false
	}
	return memMiB >= totalMemMiB
}
