/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

import (
	"fmt"
	"hcu-container-toolkit/internal/config/image"
	"hcu-container-toolkit/internal/info/drm"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/lookup"
	"hcu-container-toolkit/internal/lookup/root"
	"strings"
)

// NewDRMNodesDiscoverer returns a discoverer for the DRM device nodes associated with the specified visible devices.
//
// TODO: The logic for creating DRM devices should be consolidated between this
// and the logic for generating CDI specs for a single device. This is only used
// when applying OCI spec modifications to an incoming spec in "legacy" mode.
func NewDRMNodesDiscoverer(logger logger.Interface, busIds []string, requestBusIds []string, devRoot string) (Discover, error) {
	drmDeviceNodes, err := newDRMDeviceDiscoverer(logger, busIds, requestBusIds, devRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to create DRM device discoverer: %v", err)
	}

	return drmDeviceNodes, nil
}

// newDRMDeviceDiscoverer creates a discoverer for the DRM devices associated with the requested devices.
func newDRMDeviceDiscoverer(logger logger.Interface, busIds []string, requestBusIds []string, devRoot string) (Discover, error) {
	allDevices := NewCharDeviceDiscoverer(
		logger,
		devRoot,
		[]string{
			"/dev/dri/card*",
			"/dev/dri/renderD*",
		},
	)

	filter := make(selectDeviceByPath)
	for _, busId := range requestBusIds {
		drmDeviceNodes, err := drm.GetDeviceNodesByBusID(busId)
		if err != nil {
			return nil, fmt.Errorf("failed to determine DRM devices for %v: %v", busId, err)
		}
		logger.Infof("selected drm nodes from bus %s: %v", busId, drmDeviceNodes)
		for _, drmDeviceNode := range drmDeviceNodes {
			filter[drmDeviceNode] = true
		}
	}

	// We return a discoverer that applies the DRM device filter created above to all discovered DRM device nodes.
	d := newFilteredDiscoverer(
		logger,
		allDevices,
		filter,
	)

	return d, nil
}

// selectDeviceByPath is a filter that allows devices to be selected by the path
type selectDeviceByPath map[string]bool

var _ Filter = (*selectDeviceByPath)(nil)

// DeviceIsSelected determines whether the device's path has been selected
func (s selectDeviceByPath) DeviceIsSelected(device Device) bool {
	return s[device.Path]
}

// MountIsSelected is always true
func (s selectDeviceByPath) MountIsSelected(Mount) bool {
	return true
}

// HookIsSelected is always true
func (s selectDeviceByPath) HookIsSelected(Hook) bool {
	return true
}

// NewCommonHCUDiscoverer creates a discoverer for the mounts required by HCU.
func NewCommonHCUDiscoverer(logger logger.Interface, hcuCDIHookPath string, driver *root.Driver, isMount bool, containerImage image.DTK) (Discover, error) {
	metaDevices := NewCharDeviceDiscoverer(
		logger,
		driver.Root,
		[]string{
			"/dev/kfd",
			"/dev/mkfd",
			"/dev/mem",
		},
	)

	var directory []string
	if isMount {
		directory = append(directory, "hyhal")
	}
	libraries := NewMounts(
		logger,
		lookup.NewDirectoryLocator(
			lookup.WithLogger(logger),
			lookup.WithCount(1),
			lookup.WithSearchPaths("/opt"),
		),
		driver.Root,
		directory,
	)

	debug := NewMounts(
		logger,
		lookup.NewDirectoryLocator(
			lookup.WithLogger(logger),
			lookup.WithCount(1),
			lookup.WithSearchPaths(""),
		),
		"/",
		[]string{"/sys/kernel/debug"},
	)

	var trackHook Hook
	hcuDevices := containerImage.Getenv(image.EnvVarHCUVisibleDevices)
	if len(hcuDevices) == 0 {
		hcuDevices = containerImage.Getenv(image.EnvVarNvidiaVisibleDevices)
	}
	vHCUDevices := containerImage.Getenv(image.EnvVarVHCUVisibleDevices)
	migDevices := containerImage.Getenv(image.EnvVarMIGVisibleDevices)
	devicePluginMode := containerImage.Getenv(image.EnvVarDevicePluginMode)
	if len(migDevices) > 0 && len(vHCUDevices) > 0 {
		return nil, fmt.Errorf(
			"%s and %s cannot be set at the same time",
			image.EnvVarVHCUVisibleDevices,
			image.EnvVarMIGVisibleDevices)
	}

	if len(hcuDevices) > 0 {
		if len(migDevices) > 0 {
			return nil, fmt.Errorf(
				"%s and %s cannot be set at the same time",
				image.EnvVarHCUVisibleDevices,
				image.EnvVarMIGVisibleDevices)
		}
		trackHook = CreateTrackHook(hcuCDIHookPath, containerImage.ContainerId)
	}

	var strategyMount Discover

	if len(devicePluginMode) > 0 {
		dpModeToLower := strings.ToLower(devicePluginMode)
		switch dpModeToLower {
		case "vhcu", "mixed":
			strategyMount = NewDirectoryMountDiscoverer(
				logger,
				"/etc/vdev",
				"/etc/vdev",
				true,
				false,
			)
		case "hami":
			strategyMount = NewDirectoryMountDiscoverer(
				logger,
				"/etc/vdev",
				"/etc/vdev",
				false,
				false,
			)
		case "mig":
			strategyMount = NewDirectoryMountDiscoverer(
				logger,
				"/etc/dmi_mig_config",
				"/etc/dmi_mig_config",
				true,
				true,
			)
		}
	}

	if len(vHCUDevices) > 0 {
		m, ok := libraries.(*mounts)
		if ok {
			m.addVHCU(vHCUDevices)
		}
	}

	if len(migDevices) > 0 {
		m, ok := libraries.(*mounts)
		if ok {
			m.addMIG(migDevices)
		}
	}

	var d Discover
	var parts []Discover

	parts = append(parts,
		metaDevices,
		libraries,
		NewUserGroupDiscover(logger),
	)

	if trackHook.Lifecycle != "" {
		parts = append(parts, trackHook)
	}

	if strategyMount != nil {
		parts = append(parts, strategyMount)
	}

	parts = append(parts, debug)

	d = Merge(parts...)

	return d, nil
}
