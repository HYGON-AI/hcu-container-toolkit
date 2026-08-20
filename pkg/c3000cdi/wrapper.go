/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package c3000cdi

import (
	"hcu-container-toolkit/internal/config/image"
	"hcu-container-toolkit/pkg/c3000cdi/spec"
	"hcu-container-toolkit/pkg/c3000cdi/transform"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"
)

type wrapper struct {
	Interface

	vendor string
	class  string

	mergedDeviceOptions []transform.MergedDeviceOption
}

// GetSpec combines the device specs and common edits from the wrapped Interface to a single spec.Interface.
func (l *wrapper) GetSpec() (spec.Interface, error) {
	deviceSpecs, err := l.GetAllDeviceSpecs()
	if err != nil {
		return nil, err
	}

	edits, err := l.GetCommonEdits()
	if err != nil {
		return nil, err
	}

	return spec.New(
		spec.WithDeviceSpecs(deviceSpecs),
		spec.WithEdits(*edits.ContainerEdits),
		spec.WithVendor(l.vendor),
		spec.WithClass(l.class),
		spec.WithMergedDeviceOptions(l.mergedDeviceOptions...),
	)
}

// GetAllDeviceSpecs returns the device specs for all available devices.
func (l *wrapper) GetAllDeviceSpecs() ([]specs.Device, error) {
	return l.Interface.GetAllDeviceSpecs()
}

// GetCommonEdits returns the wrapped edits and adds additional edits on top.
func (m *wrapper) GetCommonEdits() (*cdi.ContainerEdits, error) {
	edits, err := m.Interface.GetCommonEdits()
	if err != nil {
		return nil, err
	}
	edits.Env = append(edits.Env, image.EnvVarHCUVisibleDevices+"=void")

	return edits, nil
}
