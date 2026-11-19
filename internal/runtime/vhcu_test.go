/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package runtime

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hcu-container-toolkit/internal/config"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/oci"

	"github.com/opencontainers/runtime-spec/specs-go"
)

type failModifier struct{ err error }

func (m failModifier) Modify(*specs.Spec) error { return m.err }

func TestCreateErrorsRunScopedCleanup(t *testing.T) {
	for _, stage := range []string{"modify", "flush", "exec", "success"} {
		t.Run(stage, func(t *testing.T) {
			cause := errors.New(stage + " failed")
			cleanupErr := errors.New("cleanup deferred")
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(`{"process":{"env":[]}}`), 0600); err != nil {
				t.Fatal(err)
			}
			spec := oci.NewFileSpec(path)
			low := &oci.RuntimeMock{ExecFunc: func([]string) error {
				if stage == "exec" {
					return cause
				}
				return nil
			}}
			var modifierErr error
			if stage == "modify" {
				modifierErr = cause
			}
			modifier := failModifier{err: modifierErr}
			if stage == "flush" {
				// Inject a spec write failure after successful loading/modification.
				spec = &oci.SpecMock{
					LoadFunc:   func() (*specs.Spec, error) { return &specs.Spec{}, nil },
					ModifyFunc: func(oci.SpecModifier) error { return nil },
					FlushFunc:  func() error { return cause },
				}
			}
			calls := 0
			r := allocationRuntime{Runtime: oci.NewModifyingRuntimeWrapper(&logger.NullLogger{}, low, spec, modifier), rollback: func() error { calls++; return cleanupErr }}
			err := r.Exec([]string{"runtime", "create", "owner"})
			if stage == "success" {
				if err != nil || calls != 0 {
					t.Fatalf("successful create rolled back: %v calls=%d", err, calls)
				}
			} else {
				if err == nil || calls != 1 || !errors.Is(err, cleanupErr) {
					t.Fatalf("create failure did not retain cleanup result: %v calls=%d", err, calls)
				}
			}
		})
	}
}

func TestInvalidDynamicRequestsFailBeforeDriverAccess(t *testing.T) {
	cfg, err := config.GetDefault()
	if err != nil {
		t.Fatal(err)
	}
	for _, env := range [][]string{
		{"HCU_VHCU_MEM=4096", "HCU_VHCU_DEVICE=0,1"},
		{"HCU_VHCU_MEM=4096", "HCU_VHCU_DEVICE=0000:81:00.0"},
		{"HCU_VHCU_MEM=4096", "VHCU_VISIBLE_DEVICES=0"},
		{"HCU_VHCU_MEM=4096", "MIG_VISIBLE_DEVICES=MIG-abc"},
	} {
		spec := &specs.Spec{Process: &specs.Process{Env: env}}
		m, cleanup, err := newVHCUModifier(&logger.NullLogger{}, cfg, spec, "owner", nil)
		if err == nil || m != nil || cleanup != nil {
			t.Fatalf("invalid request was not rejected before allocation: %v", env)
		}
		if strings.Contains(err.Error(), "hy-smi") {
			t.Fatalf("invalid request reached driver: %v", err)
		}
	}
	m, cleanup, err := newVHCUModifier(&logger.NullLogger{}, cfg, &specs.Spec{Process: &specs.Process{}}, "owner", nil)
	if err != nil || m != nil || cleanup != nil {
		t.Fatal("non-dynamic request should be a no-op")
	}
}

func TestDynamicReleaseHookRequiresExecutableCTK(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hcu-ctk")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := dynamicVHCUHookPath(&logger.NullLogger{}, path); err == nil {
		t.Fatal("non-executable release hook accepted")
	}
	if err := os.Chmod(path, 0700); err != nil {
		t.Fatal(err)
	}
	resolved, err := dynamicVHCUHookPath(&logger.NullLogger{}, path)
	if err != nil || resolved != path {
		t.Fatalf("wrong release executable: %s %v", resolved, err)
	}
	t.Setenv("PATH", filepath.Dir(path))
	resolved, err = dynamicVHCUHookPath(&logger.NullLogger{}, "/usr/bin/hcu-cdi-hook")
	if err != nil || resolved != path {
		t.Fatalf("CDI executable incorrectly used for vhcu release: %s %v", resolved, err)
	}
}
