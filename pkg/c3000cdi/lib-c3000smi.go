/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package c3000cdi

import (
	"fmt"
	"hcu-container-toolkit/internal/config/image"
	"hcu-container-toolkit/internal/discover"
	"hcu-container-toolkit/internal/edits"
	"hcu-container-toolkit/internal/platform-support/dhcu"
	"hcu-container-toolkit/pkg/c3000cdi/spec"
	"hcu-container-toolkit/pkg/go-c3000lib/pkg/device"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"
)

type c3000smilib c3000cdilib

var _ Interface = (*c3000smilib)(nil)

// GetSpec should not be called for c3000smilib
func (l *c3000smilib) GetSpec() (spec.Interface, error) {
	return nil, fmt.Errorf("unexpected call to c3000smilib.GetSpec()")
}

// GetAllDeviceSpecs returns the device specs for all available devices.
func (l *c3000smilib) GetAllDeviceSpecs() ([]specs.Device, error) {
	var deviceSpecs []specs.Device

	hcuDeviceSpecs, err := l.getHCUDeviceSpecs()
	if err != nil {
		return nil, err
	}
	deviceSpecs = append(deviceSpecs, hcuDeviceSpecs...)

	// todo: MIG
	// migDeviceSpecs, err := l.getMigDeviceSpecs()
	// if err != nil {
	// 	return nil, err
	// }
	// deviceSpecs = append(deviceSpecs, migDeviceSpecs...)

	return deviceSpecs, nil
}

// GetCommonEdits generates a CDI specification that can be used for ANY devices
func (l *c3000smilib) GetCommonEdits() (*cdi.ContainerEdits, error) {
	common, err := l.newCommonC3000SmiDiscoverer()
	if err != nil {
		return nil, fmt.Errorf("failed to create discoverer for common entities: %v", err)
	}

	return edits.FromDiscoverer(l.logger, common)
}

// GetDeviceSpecsByID returns the CDI device specs for the HCU(s) represented by
// the provided identifiers, where an identifier is an index or UUID of a valid
// HCU device.
// Deprecated: Use GetDeviceSpecsBy instead.
func (l *c3000smilib) GetDeviceSpecsByID(ids ...string) ([]specs.Device, error) {
	var identifiers []device.Identifier
	for _, id := range ids {
		identifiers = append(identifiers, device.Identifier(id))
	}
	return l.GetDeviceSpecsBy(identifiers...)
}

// GetDeviceSpecsBy returns the device specs for devices with the specified identifiers.
func (l *c3000smilib) GetDeviceSpecsBy(identifiers ...device.Identifier) ([]specs.Device, error) {
	for _, id := range identifiers {
		if id == "all" {
			return l.GetAllDeviceSpecs()
		}
	}

	var deviceSpecs []specs.Device

	return deviceSpecs, nil
}

// GetHCUDeviceEdits returns the CDI edits for the full HCU represented by 'device'.
func (l *c3000smilib) GetHCUDeviceEdits(d device.Device) (*cdi.ContainerEdits, error) {
	device, err := l.newFullHCUDiscoverer(d)
	if err != nil {
		return nil, fmt.Errorf("failed to create device discoverer: %v", err)
	}

	editsForDevice, err := edits.FromDiscoverer(l.logger, device)
	if err != nil {
		return nil, fmt.Errorf("failed to create container edits for device: %v", err)
	}

	return editsForDevice, nil
}

// GetHCUDeviceSpecs returns the CDI device specs for the full HCU represented by 'device'.
func (l *c3000smilib) GetHCUDeviceSpecs(i string, d device.Device) ([]specs.Device, error) {
	edits, err := l.GetHCUDeviceEdits(d)
	if err != nil {
		return nil, fmt.Errorf("failed to get edits for device: %v", err)
	}

	var deviceSpecs []specs.Device
	names, err := l.deviceNamers.GetDeviceNames(i, convert{d})
	if err != nil {
		return nil, fmt.Errorf("failed to get device name: %v", err)
	}
	for _, name := range names {
		spec := specs.Device{
			Name:           name,
			ContainerEdits: *edits.ContainerEdits,
		}
		deviceSpecs = append(deviceSpecs, spec)
	}

	return deviceSpecs, nil
}

func (l *c3000smilib) getHCUDeviceSpecs() ([]specs.Device, error) {
	var deviceSpecs []specs.Device
	err := l.devicelib.VisitDevices(func(i string, d device.Device) error {
		specsForDevice, err := l.GetHCUDeviceSpecs(i, d)
		if err != nil {
			return err
		}
		deviceSpecs = append(deviceSpecs, specsForDevice...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate CDI edits for HCU devices: %v", err)
	}
	return deviceSpecs, err
}

// GetMIGDeviceEdits returns the CDI edits for the MIG device represented by 'mig' on 'parent'.
func (l *c3000smilib) GetMIGDeviceEdits(parent device.Device, mig device.MigDevice) (*cdi.ContainerEdits, error) {
	// todo
	return nil, nil
}

// GetMIGDeviceSpecs returns the CDI device specs for the full GPU represented by 'device'.
func (l *c3000smilib) GetMIGDeviceSpecs(i int, d device.Device, j int, mig device.MigDevice) ([]specs.Device, error) {
	// todo
	return nil, nil
}

// newCommonNVMLDiscoverer returns a discoverer for entities that are not associated with a specific CDI device.
// This includes driver libraries and meta devices, for example.
func (l *c3000smilib) newCommonC3000SmiDiscoverer() (discover.Discover, error) {
	d, err := discover.NewCommonHCUDiscoverer(
		l.logger,
		l.hcuCDIHookPath,
		l.driver,
		true,
		image.DTK{},
	)

	return d, err
}

// newFullHCUDiscoverer creates a discoverer for the full HCU defined by the specified device.
func (l *c3000smilib) newFullHCUDiscoverer(d device.Device) (discover.Discover, error) {
	deviceNodes, err := dhcu.NewForDevice(
		d,
		dhcu.WithDevRoot(l.devRoot),
		dhcu.WithLogger(l.logger),
		dhcu.WithHCUCDIHookPath(l.hcuCDIHookPath),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create device discoverer: %v", err)
	}

	deviceFolderPermissionHooks := newDeviceFolderPermissionHookDiscoverer(
		l.logger,
		l.devRoot,
		l.hcuCDIHookPath,
		deviceNodes,
	)

	dd := discover.Merge(
		deviceNodes,
		deviceFolderPermissionHooks,
	)

	return dd, nil
}
