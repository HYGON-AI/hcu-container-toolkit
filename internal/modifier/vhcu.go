/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package modifier

import (
	"fmt"
	"strconv"
	"strings"

	"hcu-container-toolkit/internal/config/image"
	"hcu-container-toolkit/internal/discover"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/oci"
	"hcu-container-toolkit/internal/vhcu"

	"github.com/opencontainers/runtime-spec/specs-go"
)

// vhcuModifier owns the OCI spec edits for a dynamic vHCU allocation:
//
//  1. the env write-back that hands the new vdev index to the existing
//     VHCU_VISIBLE_DEVICES mount path (conf + DRM + PCI all follow for free),
//     or, on a full-card short circuit, HCU_VISIBLE_DEVICES for a plain
//     physical mount;
//  2. the poststop release hook for actually-created vdevs.
//
// It MUST run before the graphics modifier in the chain. It deliberately does
// not mount /etc/vdev/vdev<N>.conf itself: addVHCU already does that, and a
// second mount with the same container path would make runc fail or silently
// override.
type vhcuModifier struct {
	logger logger.Interface
	alloc  *vhcu.Allocation
	hook   discover.Hook
}

// NewVHCUModifier constructs the modifier for the given allocation.
// hcuCTKPath is the toolkit CLI path used in the poststop hook (same source
// as the hcu-tracker track hook).
func NewVHCUModifier(logger logger.Interface, alloc *vhcu.Allocation, hcuCTKPath string) (oci.SpecModifier, error) {
	if alloc == nil {
		return nil, nil
	}

	m := &vhcuModifier{
		logger: logger,
		alloc:  alloc,
	}
	if !alloc.FullCard {
		m.hook = discover.CreateVHCUReleaseHook(hcuCTKPath, alloc.ContainerID, alloc.AllocationID)
	}
	return m, nil
}

// Modify applies the env write-back and injects the poststop release hook.
func (m *vhcuModifier) Modify(spec *specs.Spec) error {
	if err := ApplyVHCUEnv(spec, m.alloc); err != nil {
		return err
	}

	if m.hook.Path == "" {
		return nil
	}
	if spec.Hooks == nil {
		spec.Hooks = &specs.Hooks{}
	}
	// Idempotent: docker start replays the create path against a spec that
	// still contains the previous edit in some setups; never add the hook
	// twice.
	for _, existing := range spec.Hooks.Poststop {
		if existing.Path == m.hook.Path && sameArgs(existing.Args, m.hook.Args) {
			m.logger.Debugf("vHCU release hook already present; skipping injection")
			return nil
		}
	}
	spec.Hooks.Poststop = append(spec.Hooks.Poststop, specs.Hook{
		Path: m.hook.Path,
		Args: m.hook.Args,
	})
	return nil
}

func sameArgs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ApplyVHCUEnv writes the allocation outcome into the spec's process env:
//   - vdev case: VHCU_VISIBLE_DEVICES is set to exactly one vdev index;
//   - full-card case: HCU_VISIBLE_DEVICES is set to the physical device index.
//
// It is exported because the runtime factory must apply it to the raw spec
// BEFORE snapshotting image.DTK (the image reads env once at construction).
func ApplyVHCUEnv(spec *specs.Spec, alloc *vhcu.Allocation) error {
	if spec == nil || spec.Process == nil {
		return fmt.Errorf("cannot apply vHCU env: OCI spec has no process")
	}
	if alloc == nil {
		return nil
	}

	if alloc.FullCard {
		index := strconv.Itoa(alloc.DeviceIndex)
		if existing := getenv(spec, image.EnvVarHCUVisibleDevices); existing != "" && existing != index {
			// ParseRequest rejected this combination up front; reaching here
			// means an internal ordering bug.
			return fmt.Errorf("internal error: full-card short circuit with %s already set to %q", image.EnvVarHCUVisibleDevices, existing)
		}
		setenv(spec, image.EnvVarHCUVisibleDevices, index)
		return nil
	}

	index := strconv.Itoa(alloc.VdevIndex)
	if alloc.VdevIndex < 0 {
		return fmt.Errorf("invalid dynamic vHCU index %d", alloc.VdevIndex)
	}
	if existing := getenv(spec, image.EnvVarVHCUVisibleDevices); existing != "" && existing != index {
		return fmt.Errorf("dynamic vHCU allocation cannot be combined with existing VHCU_VISIBLE_DEVICES=%s", existing)
	}
	setenv(spec, image.EnvVarVHCUVisibleDevices, index)
	return nil
}

func getenv(spec *specs.Spec, key string) string {
	for _, kv := range spec.Process.Env {
		if strings.HasPrefix(kv, key+"=") {
			return strings.TrimPrefix(kv, key+"=")
		}
	}
	return ""
}

func setenv(spec *specs.Spec, key, value string) {
	prefix := key + "="
	for i, kv := range spec.Process.Env {
		if strings.HasPrefix(kv, prefix) {
			spec.Process.Env[i] = prefix + value
			return
		}
	}
	spec.Process.Env = append(spec.Process.Env, prefix+value)
}
