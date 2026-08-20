/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package modifier

import (
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/oci"
	"path/filepath"

	"github.com/opencontainers/runtime-spec/specs-go"
)

func NewNvidiaContainerRuntimeHookRemover(logger logger.Interface) oci.SpecModifier {
	m := nvidiaContainerRuntimeHookRemover{
		logger: logger,
	}

	return &m
}

// nvidiaContainerRuntimeHookRemover is a spec modifier that detects and removes inserted nvidia-container-runtime hooks
type nvidiaContainerRuntimeHookRemover struct {
	logger logger.Interface
}

var _ oci.SpecModifier = (*nvidiaContainerRuntimeHookRemover)(nil)

// Modify removes any NVIDIA Container Runtime hooks from the provided spec
func (m nvidiaContainerRuntimeHookRemover) Modify(spec *specs.Spec) error {
	if spec == nil || spec.Hooks == nil {
		return nil
	}

	if len(spec.Hooks.Prestart) == 0 {
		return nil
	}

	var hooks []specs.Hook
	for _, hook := range spec.Hooks.Prestart {
		hook := hook
		if isNVIDIAContainerRuntimeHook(&hook) {
			m.logger.Debugf("Removing hook %v", hook)
			continue
		}
		hooks = append(hooks, hook)
	}

	if len(hooks) != len(spec.Hooks.Prestart) {
		m.logger.Debugf("Updating 'prestart' hooks to %v", hooks)
		spec.Hooks.Prestart = hooks
	}
	return nil
}

// isNVIDIAContainerRuntimeHook checks if the provided hook is an nvidia-container-runtime-hook
// or nvidia-container-toolkit hook. These are included, for example, by the non-experimental
// nvidia-container-runtime or docker when specifying the --gpus flag.
func isNVIDIAContainerRuntimeHook(hook *specs.Hook) bool {
	bins := map[string]struct{}{
		"nvidia-container-runtime-hook": {},
		"nvidia-container-toolkit":      {},
	}

	_, exists := bins[filepath.Base(hook.Path)]

	return exists
}

func NewSeccompRemover(logger logger.Interface) oci.SpecModifier {
	m := seccompRemover{
		logger: logger,
	}

	return &m
}

// seccompRemover is a spec modifer that disable seccomp
type seccompRemover struct {
	logger logger.Interface
}

var _ oci.SpecModifier = (*seccompRemover)(nil)

func (m seccompRemover) Modify(spec *specs.Spec) error {
	if spec == nil || spec.Linux == nil {
		return nil
	}
	spec.Linux.Seccomp = nil
	spec.Process.ApparmorProfile = ""
	m.logger.Info("Remove linux.seccomp in OCI spec and Process.ApparmorProfile")
	return nil
}
