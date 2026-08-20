/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package modifier

import (
	"fmt"
	"hcu-container-toolkit/internal/config"
	"hcu-container-toolkit/internal/config/image"
	"hcu-container-toolkit/internal/discover"
	hcuTracker "hcu-container-toolkit/internal/hcu-tracker"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/lookup"
	"hcu-container-toolkit/internal/lookup/root"
	"hcu-container-toolkit/internal/oci"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
)

// NewGraphicsModifier constructs a modifier that injects graphics-related modifications into an OCI runtime specification.
// The value of the HCU_DRIVER_CAPABILITIES environment variable is checked to determine if this modification should be made.
func NewGraphicsModifier(logger logger.Interface, cfg *config.Config, containerImage image.DTK, driver *root.Driver, isMount bool) (oci.SpecModifier, error) {
	hcuCDIHookPath := cfg.HCUCTKConfig.Path
	value := containerImage.Getenv(image.EnvVarHCUVisibleDevices)
	if len(value) == 0 {
		value = containerImage.Getenv(image.EnvVarNvidiaVisibleDevices)
	}

	if len(value) > 0 {
		hcuTracker, err := hcuTracker.New()
		if err == nil {
			_, err = hcuTracker.ReserveHCUs(value, containerImage.ContainerId)
			logger.Infof("ReserveHCUs %s", value)
			if err != nil {
				return nil, fmt.Errorf("failed to reserve HCUs: %v", err)
			}
		}
	}

	comDiscoverer, err := discover.NewCommonHCUDiscoverer(
		logger,
		hcuCDIHookPath,
		driver,
		isMount,
		containerImage,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mounts discoverer: %v", err)
	}

	visibleDevices := containerImage.DevicesFromEnvvars(image.EnvVarHCUVisibleDevices, image.EnvVarNvidiaVisibleDevices)
	vHCUDevices := containerImage.VHCUFromEnvVar(image.EnvVarVHCUVisibleDevices)
	migDevices := containerImage.MIGFromEnvVar(image.EnvVarMIGVisibleDevices)

	if len(visibleDevices.List()) == 0 && len(vHCUDevices) == 0 && len(migDevices) == 0 {
		logger.Info("No devices requested")
		return nil, nil
	}

	busIds, err := getDevicesFromDriver()
	if err != nil {
		logger.Errorf("No hcu found")
		return nil, err
	}

	err = checkRequestDevices(logger, visibleDevices, busIds)
	if err != nil {
		return nil, err
	}

	var selectedBusIds []string
	for i, busId := range busIds {
		if visibleDevices.Has(fmt.Sprintf("%d", i)) || visibleDevices.Has(busId) {
			selectedBusIds = append(selectedBusIds, busId)
		}
	}

	for _, vHCUDevice := range vHCUDevices {
		// vHCU devices are backed by physical HCUs;
		// a physical HCU must NOT be mounted together with its vHCU device.
		if slices.Contains(selectedBusIds, vHCUDevice) {
			return nil, fmt.Errorf(
				"conflict detected: physical device %s is already selected; cannot mount its vHCU device at the same time",
				vHCUDevice,
			)
		} else {
			selectedBusIds = append(selectedBusIds, vHCUDevice)
		}
	}

	for _, migDevice := range migDevices {
		if !slices.Contains(selectedBusIds, migDevice) {
			selectedBusIds = append(selectedBusIds, migDevice)
		}
	}
	// In standard usage, the devRoot is the same as the driver.Root.
	devRoot := driver.Root
	drmNodes, err := discover.NewDRMNodesDiscoverer(
		logger,
		busIds,
		selectedBusIds,
		devRoot,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to construct discoverer: %v", err)
	}

	drmByPathLinks := discover.NewCreateDRMByPathSymlinks(logger, drmNodes, devRoot, hcuCDIHookPath)

	pciMounts := discover.NewPciMounts(
		logger,
		lookup.NewDirectoryLocator(
			lookup.WithLogger(logger),
			lookup.WithCount(1),
			lookup.WithSearchPaths("/sys/bus/pci/devices"),
		),
		driver.Root,
		selectedBusIds,
	)

	d := discover.Merge(
		comDiscoverer,
		drmNodes,
		drmByPathLinks,
		pciMounts,
	)
	return NewModifierFromDiscoverer(logger, d)
}

// getDevicesFromDriver query all HCU devices bus id
func getDevicesFromDriver() ([]string, error) {
	var devices []string

	matches, err := filepath.Glob("/sys/module/hy*cu/drivers/pci:hy*cu/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*")
	if err != nil {
		return devices, fmt.Errorf("failed to find devices bus id: %v", err)
	}
	if len(matches) == 0 {
		m, err := filepath.Glob("/sys/module/hy*cu/drivers/pci:amdgpu/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*")
		if err != nil {
			return devices, fmt.Errorf("failed to find devices bus id: %v", err)
		}
		matches = append(matches, m...)
	}

	for _, path := range sort.StringSlice(matches) {
		devices = append(devices, filepath.Base(path))
	}

	return devices, nil
}

func checkRequestDevices(logger logger.Interface, devices image.VisibleDevices, busIds []string) error {
	for _, device := range devices.List() {
		if device == "all" || device == "" {
			break
		}
		deviceId, err := strconv.Atoi(device)
		if err != nil {
			found := false
			for _, busId := range busIds {
				if device == busId {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("request device %s not found", device)
			}
		} else if deviceId >= len(busIds) {
			logger.Errorf("Request device %s is invalid", device)
			return fmt.Errorf("request device %s not found", device)
		}
	}

	return nil
}
