/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

import (
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/lookup"
	"os"
	"path/filepath"
)

type drmDevicesByPath struct {
	None
	logger         logger.Interface
	hcuCDIHookPath string
	devRoot        string
	devicesFrom    Discover
}

// NewCreateDRMByPathSymlinks creates a discoverer for a hook to create the by-path symlinks for DRM devices discovered by the specified devices discoverer
func NewCreateDRMByPathSymlinks(logger logger.Interface, devices Discover, devRoot string, hcuCDIHookPath string) Discover {
	d := drmDevicesByPath{
		logger:         logger,
		hcuCDIHookPath: hcuCDIHookPath,
		devRoot:        devRoot,
		devicesFrom:    devices,
	}

	return &d
}

// Hooks returns a hook to create the symlinks from the required CSV files
func (d drmDevicesByPath) Hooks() ([]Hook, error) {
	devices, err := d.devicesFrom.Devices()
	if err != nil {
		return nil, fmt.Errorf("failed to discover devices for by-path symlinks: %v", err)
	}
	if len(devices) == 0 {
		return nil, nil
	}
	links, err := d.getSpecificLinkArgs(devices)
	if err != nil {
		return nil, fmt.Errorf("failed to determine specific links: %v", err)
	}
	if len(links) == 0 {
		return nil, nil
	}

	var hooks []Hook

	hook := CreateSymlinkHook(d.hcuCDIHookPath, links)
	hooks = append(hooks, hook)

	return hooks, nil
}

// getSpecificLinkArgs returns the required specific links that need to be created
func (d drmDevicesByPath) getSpecificLinkArgs(devices []Device) ([]string, error) {
	selectedDevices := make(map[string]bool)
	for _, d := range devices {
		selectedDevices[filepath.Base(d.HostPath)] = true
	}

	linkLocator := lookup.NewFileLocator(
		lookup.WithLogger(d.logger),
		lookup.WithRoot(d.devRoot),
	)
	candidates, err := linkLocator.Locate("/dev/dri/by-path/pci-*-*")
	if err != nil {
		d.logger.Warningf("Failed to locate by-path links: %v; ignoring", err)
		return nil, nil
	}

	var links []string
	for _, c := range candidates {
		device, err := os.Readlink(c)
		if err != nil {
			d.logger.Warningf("Failed to evaluate symlink %v; ignoring", c)
			continue
		}

		if selectedDevices[filepath.Base(device)] {
			d.logger.Debugf("adding device symlink %v -> %v", c, device)
			links = append(links, fmt.Sprintf("%v::%v", device, c))
		}
	}

	return links, nil
}
