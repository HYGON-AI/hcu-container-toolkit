/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package modifier

import (
	"testing"

	"hcu-container-toolkit/internal/config/image"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/vhcu"

	"github.com/opencontainers/runtime-spec/specs-go"
)

func newSpec(env []string) *specs.Spec {
	return &specs.Spec{
		Process: &specs.Process{Env: env},
	}
}

func TestApplyVHCUEnvSetsVHCUVisibleDevices(t *testing.T) {
	spec := newSpec([]string{"PATH=/usr/bin"})
	alloc := &vhcu.Allocation{VdevIndex: 3}

	if err := ApplyVHCUEnv(spec, alloc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := getenv(spec, image.EnvVarVHCUVisibleDevices); got != "3" {
		t.Fatalf("expected VHCU_VISIBLE_DEVICES=3, got %q", got)
	}
}

func TestApplyVHCUEnvRejectsAdditionalVdev(t *testing.T) {
	spec := newSpec([]string{"VHCU_VISIBLE_DEVICES=0"})
	alloc := &vhcu.Allocation{VdevIndex: 3}
	if err := ApplyVHCUEnv(spec, alloc); err == nil {
		t.Fatal("existing virtual card must be rejected")
	}
	if got := getenv(spec, image.EnvVarVHCUVisibleDevices); got != "0" {
		t.Fatal("conflicting env was mutated")
	}
	spec = newSpec(nil)
	for i := 0; i < 2; i++ {
		if err := ApplyVHCUEnv(spec, alloc); err != nil {
			t.Fatal(err)
		}
	}
	if got := getenv(spec, image.EnvVarVHCUVisibleDevices); got != "3" {
		t.Fatalf("expected single vHCU, got %q", got)
	}
}

func TestApplyVHCUEnvFullCard(t *testing.T) {
	spec := newSpec([]string{"PATH=/usr/bin"})
	alloc := &vhcu.Allocation{FullCard: true, DeviceIndex: 2}

	if err := ApplyVHCUEnv(spec, alloc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := getenv(spec, image.EnvVarHCUVisibleDevices); got != "2" {
		t.Fatalf("expected HCU_VISIBLE_DEVICES=2, got %q", got)
	}

	if err := ApplyVHCUEnv(spec, alloc); err != nil {
		t.Fatalf("repeated full-card write must be idempotent: %v", err)
	}

	// The combination was already rejected at request parsing; an
	// already-set variable here is an internal ordering bug.
	spec = newSpec([]string{"HCU_VISIBLE_DEVICES=all"})
	if err := ApplyVHCUEnv(spec, alloc); err == nil {
		t.Fatal("full-card write onto a set HCU_VISIBLE_DEVICES must fail")
	}
}

func TestVHCUModifierInjectsPoststopHook(t *testing.T) {
	spec := newSpec(nil)
	spec.Hooks = &specs.Hooks{}

	alloc := &vhcu.Allocation{ContainerID: "a4e1", AllocationID: "token", VdevIndex: 3}
	m, err := NewVHCUModifier(&logger.NullLogger{}, alloc, "/usr/bin/hcu-ctk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil {
		t.Fatal("modifier must not be nil for a vdev allocation")
	}

	if err := m.Modify(spec); err != nil {
		t.Fatalf("modify failed: %v", err)
	}

	if got := getenv(spec, image.EnvVarVHCUVisibleDevices); got != "3" {
		t.Fatalf("expected VHCU_VISIBLE_DEVICES=3, got %q", got)
	}
	if len(spec.Hooks.Poststop) != 1 {
		t.Fatalf("expected exactly one poststop hook, got %d", len(spec.Hooks.Poststop))
	}
	hook := spec.Hooks.Poststop[0]
	if hook.Path != "/usr/bin/hcu-ctk" {
		t.Fatalf("unexpected hook path: %q", hook.Path)
	}
	if len(hook.Args) != 5 || hook.Args[0] != "hcu-ctk" || hook.Args[1] != "vhcu" || hook.Args[2] != "release" || hook.Args[3] != "a4e1" || hook.Args[4] != "token" {
		t.Fatalf("unexpected hook args: %v", hook.Args)
	}

	// Re-applying must not duplicate the hook (idempotency).
	if err := m.Modify(spec); err != nil {
		t.Fatalf("second modify failed: %v", err)
	}
	if len(spec.Hooks.Poststop) != 1 {
		t.Fatalf("hook must not be injected twice, got %d", len(spec.Hooks.Poststop))
	}
}

func TestVHCUModifierNilHooks(t *testing.T) {
	spec := newSpec(nil)

	alloc := &vhcu.Allocation{ContainerID: "a4e1", AllocationID: "token", VdevIndex: 3}
	m, _ := NewVHCUModifier(&logger.NullLogger{}, alloc, "/usr/bin/hcu-ctk")

	if err := m.Modify(spec); err != nil {
		t.Fatalf("modify failed: %v", err)
	}
	if spec.Hooks == nil || len(spec.Hooks.Poststop) != 1 {
		t.Fatalf("hooks section must be created on demand: %+v", spec.Hooks)
	}
}

func TestVHCUModifierFullCardHasNoHook(t *testing.T) {
	spec := newSpec(nil)
	spec.Hooks = &specs.Hooks{}

	alloc := &vhcu.Allocation{ContainerID: "a4e1", FullCard: true, DeviceIndex: 1}
	m, err := NewVHCUModifier(&logger.NullLogger{}, alloc, "/usr/bin/hcu-ctk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.Modify(spec); err != nil {
		t.Fatalf("modify failed: %v", err)
	}
	if len(spec.Hooks.Poststop) != 0 {
		t.Fatal("full-card allocation must not inject a release hook")
	}
	if got := getenv(spec, image.EnvVarHCUVisibleDevices); got != "1" {
		t.Fatalf("expected HCU_VISIBLE_DEVICES=1, got %q", got)
	}
}

func TestVHCUModifierNilAllocation(t *testing.T) {
	m, err := NewVHCUModifier(&logger.NullLogger{}, nil, "/usr/bin/hcu-ctk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m != nil {
		t.Fatal("nil allocation must yield a nil modifier")
	}
}
