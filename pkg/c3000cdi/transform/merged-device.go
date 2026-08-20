/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package transform

import (
	"fmt"
	"hcu-container-toolkit/internal/edits"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/pkg/parser"
	"tags.cncf.io/container-device-interface/specs-go"
)

const (
	allDeviceName = "all"
)

type mergedDevice struct {
	name         string
	skipIfExists bool
	simplifier   Transformer
}

var _ Transformer = (*mergedDevice)(nil)

// MergedDeviceOption is a function that configures a merged device
type MergedDeviceOption func(*mergedDevice)

// WithName sets the name of the merged device
func WithName(name string) MergedDeviceOption {
	return func(m *mergedDevice) {
		m.name = name
	}
}

// WithSkipIfExists sets whether to skip adding the merged device if it already exists
func WithSkipIfExists(skipIfExists bool) MergedDeviceOption {
	return func(m *mergedDevice) {
		m.skipIfExists = skipIfExists
	}
}

// NewMergedDevice creates a transformer with the specified options
func NewMergedDevice(opts ...MergedDeviceOption) (Transformer, error) {
	m := &mergedDevice{}
	for _, opt := range opts {
		opt(m)
	}
	if m.name == "" {
		m.name = allDeviceName
	}
	m.simplifier = NewSimplifier()

	if err := parser.ValidateDeviceName(m.name); err != nil {
		return nil, fmt.Errorf("invalid device name %q: %v", m.name, err)
	}

	return m, nil
}

// Transform adds a merged device to the spec
func (m mergedDevice) Transform(spec *specs.Spec) error {
	if spec == nil {
		return nil
	}

	mergedDevice, err := mergeDeviceSpecs(spec.Devices, m.name)
	if err != nil {
		return fmt.Errorf("failed to generate merged device %q: %v", m.name, err)
	}
	if mergedDevice == nil {
		if m.skipIfExists {
			return nil
		}
		return fmt.Errorf("device %q already exists", m.name)
	}

	spec.Devices = append(spec.Devices, *mergedDevice)

	if err := m.simplifier.Transform(spec); err != nil {
		return fmt.Errorf("failed to simplify spec after merging device %q: %v", m.name, err)
	}

	return nil
}

// mergeDeviceSpecs creates a device with the specified name which combines the edits from the previous devices.
// If a device of the specified name already exists, no device is created and nil is returned.
func mergeDeviceSpecs(deviceSpecs []specs.Device, mergedDeviceName string) (*specs.Device, error) {
	for _, d := range deviceSpecs {
		if d.Name == mergedDeviceName {
			return nil, nil
		}
	}

	mergedEdits := edits.NewContainerEdits()

	for _, d := range deviceSpecs {
		d := d
		edit := cdi.ContainerEdits{
			ContainerEdits: &d.ContainerEdits,
		}
		mergedEdits.Append(&edit)
	}

	merged := specs.Device{
		Name:           mergedDeviceName,
		ContainerEdits: *mergedEdits.ContainerEdits,
	}
	return &merged, nil
}
