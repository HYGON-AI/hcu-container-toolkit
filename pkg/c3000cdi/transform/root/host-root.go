/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package root

import (
	"fmt"
	"hcu-container-toolkit/pkg/c3000cdi/transform"
	"strings"

	"tags.cncf.io/container-device-interface/specs-go"
)

// hostRootTransformer transforms the roots of host paths in a CDI spec.
type hostRootTransformer transformer

var _ transform.Transformer = (*hostRootTransformer)(nil)

// Transform replaces the root in a spec with a new root.
// It walks the spec and replaces all host paths that start with root with the target root.
func (t hostRootTransformer) Transform(spec *specs.Spec) error {
	if spec == nil {
		return nil
	}

	for _, d := range spec.Devices {
		d := d
		if err := t.applyToEdits(&d.ContainerEdits); err != nil {
			return fmt.Errorf("failed to apply root transform to device %s: %w", d.Name, err)
		}
	}

	if err := t.applyToEdits(&spec.ContainerEdits); err != nil {
		return fmt.Errorf("failed to apply root transform to spec: %w", err)
	}
	return nil
}

func (t hostRootTransformer) applyToEdits(edits *specs.ContainerEdits) error {
	for i, dn := range edits.DeviceNodes {
		edits.DeviceNodes[i] = t.transformDeviceNode(dn)
	}

	for i, hook := range edits.Hooks {
		edits.Hooks[i] = t.transformHook(hook)
	}

	for i, mount := range edits.Mounts {
		edits.Mounts[i] = t.transformMount(mount)
	}

	return nil
}

func (t hostRootTransformer) transformDeviceNode(dn *specs.DeviceNode) *specs.DeviceNode {
	if dn.HostPath == "" {
		dn.HostPath = dn.Path
	}
	dn.HostPath = t.transformPath(dn.HostPath)

	return dn
}

func (t hostRootTransformer) transformHook(hook *specs.Hook) *specs.Hook {
	// The Path in the startContainer hook MUST resolve in the container namespace.
	if hook.HookName != "startContainer" {
		hook.Path = t.transformPath(hook.Path)
	}

	// The createContainer and startContainer hooks MUST execute in the container namespace.
	if hook.HookName == "createContainer" || hook.HookName == "startContainer" {
		return hook
	}

	var args []string
	for _, arg := range hook.Args {
		if !strings.Contains(arg, "::") {
			args = append(args, t.transformPath(arg))
			continue
		}

		// For the 'create-symlinks' hook, special care is taken for the
		// '--link' flag argument which takes the form <target>::<link>.
		// Both paths, the target and link paths, are transformed.
		split := strings.SplitN(arg, "::", 2)
		split[0] = t.transformPath(split[0])
		split[1] = t.transformPath(split[1])
		args = append(args, strings.Join(split, "::"))
	}
	hook.Args = args

	return hook
}

func (t hostRootTransformer) transformMount(mount *specs.Mount) *specs.Mount {
	mount.HostPath = t.transformPath(mount.HostPath)
	return mount
}

func (t hostRootTransformer) transformPath(path string) string {
	return (transformer)(t).transformPath(path)
}
