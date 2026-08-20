/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package transform

import (
	"fmt"

	"tags.cncf.io/container-device-interface/specs-go"
)

type simplify struct{}

var _ Transformer = (*simplify)(nil)

// NewSimplifier creates a simplifier transformer.
// This transoformer ensures that entities in the spec are deduplicated and that common edits are removed from device-specific edits.
func NewSimplifier() Transformer {
	return Merge(
		dedupe{},
		simplify{},
		sorter{},
	)
}

// Transform simplifies the supplied spec.
// Edits that are present in the common edits are removed from device-specific edits.
func (s simplify) Transform(spec *specs.Spec) error {
	if spec == nil {
		return nil
	}

	dedupe := dedupe{}
	if err := dedupe.Transform(spec); err != nil {
		return err
	}

	commonEntityIDs, err := (*containerEdits)(&spec.ContainerEdits).getEntityIds()
	if err != nil {
		return err
	}

	toRemove := newRemover(commonEntityIDs...)
	var updatedDevices []specs.Device
	for _, device := range spec.Devices {
		deviceAsSpec := specs.Spec{
			ContainerEdits: device.ContainerEdits,
		}
		err := toRemove.Transform(&deviceAsSpec)
		if err != nil {
			return fmt.Errorf("failed to transform device edits: %w", err)
		}

		if !(containerEdits)(deviceAsSpec.ContainerEdits).IsEmpty() {
			// Devices with empty edits are invalid.
			// We only update the container edits for the device if this would
			// result in a valid device.
			device.ContainerEdits = deviceAsSpec.ContainerEdits
		}
		updatedDevices = append(updatedDevices, device)
	}
	spec.Devices = updatedDevices

	return nil
}
