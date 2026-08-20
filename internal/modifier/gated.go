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
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/lookup/root"
	"hcu-container-toolkit/internal/oci"
)

// NewFeatureGatedModifier creates the modifiers for optional features.
// These include:
//
//	HCU_MOFED=enabled
//
// If not devices are selected, no changes are made.
func NewFeatureGatedModifier(logger logger.Interface, cfg *config.Config, image image.DTK, driver *root.Driver) (oci.SpecModifier, error) {
	getters := []func() []string{
		image.VisibleDevicesFromEnvVar,
		image.VHCUVisibleDevicesFromEnvVar,
		image.MIGVisibleDevicesFromEnvVar,
	}

	var devices []string
	for _, get := range getters {
		if devices = get(); len(devices) > 0 {
			break
		}
	}

	if len(devices) == 0 {
		logger.Infof("No modification required; no devices requested")
		return nil, nil
	}

	var discoverers []discover.Discover

	if image.Getenv("HCU_MOFED") == "enabled" {
		d, err := discover.NewMOFEDDiscoverer(logger, driver.Root)
		if err != nil {
			return nil, fmt.Errorf("failed to construct discoverer for MOFED devices: %w", err)
		}
		discoverers = append(discoverers, d)
	}

	return NewModifierFromDiscoverer(logger, discover.Merge(discoverers...))
}
