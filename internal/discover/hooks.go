/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

import (
	"path/filepath"

	"tags.cncf.io/container-device-interface/pkg/cdi"
)

var _ Discover = (*Hook)(nil)

// Devices returns an empty list of devices for a Hook discoverer.
func (h Hook) Devices() ([]Device, error) {
	return nil, nil
}

// Mounts returns an empty list of mounts for a Hook discoverer.
func (h Hook) Mounts() ([]Mount, error) {
	return nil, nil
}

// Hooks allows the Hook type to also implement the Discoverer interface.
// It returns a single hook
func (h Hook) Hooks() ([]Hook, error) {
	return []Hook{h}, nil
}

func (h Hook) AdditionalGIDs() ([]uint32, error) {
	return []uint32{}, nil
}

// CreateSymlinkHook creates a hook which creates a symlink from link -> target.
func CreateSymlinkHook(hcuCDIHookPath string, links []string) Hook {
	if len(links) == 0 {
		return Hook{}
	}

	var args []string
	for _, link := range links {
		args = append(args, "--link", link)
	}
	return CreateHcuCDIHook(
		hcuCDIHookPath,
		"create-symlinks",
		args...,
	)
}

func CreateTrackHook(hcuCDIHookPath string, containerId string) Hook {
	return CreateHcuTrackHook(
		hcuCDIHookPath,
		"hcu-tracker",
		"release",
		containerId,
	)
}

func CreateHcuTrackHook(path string, s string, s2 string, id string) Hook {
	return Hook{
		Lifecycle: cdi.PoststopHook,
		Path:      path,
		Args:      []string{"hcu-ctk", s, s2, id},
	}
}

// CreateHcuCDIHook creates a hook which invokes the HCU Container CLI hook subcommand.
func CreateHcuCDIHook(hcuCDIHookPath string, hookName string, additionalArgs ...string) Hook {
	return cdiHook(hcuCDIHookPath).Create(hookName, additionalArgs...)
}

type cdiHook string

func (c cdiHook) Create(name string, args ...string) Hook {
	return Hook{
		Lifecycle: cdi.CreateContainerHook,
		Path:      string(c),
		Args:      append(c.requiredArgs(name), args...),
	}
}
func (c cdiHook) requiredArgs(name string) []string {
	base := filepath.Base(string(c))
	if base == "hcu-ctk" {
		return []string{base, "hook", name}
	}
	return []string{base, name}
}
