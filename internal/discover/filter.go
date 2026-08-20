/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

import (
	"hcu-container-toolkit/internal/logger"
)

// Filter defines an interface for filtering discovered entities
type Filter interface {
	DeviceIsSelected(device Device) bool
}

// filtered represents a filtered discoverer
type filtered struct {
	Discover
	logger logger.Interface
	filter Filter
}

// newFilteredDiscoverer creates a discoverer that applies the specified filter to the returned entities of the discoverer
func newFilteredDiscoverer(logger logger.Interface, applyTo Discover, filter Filter) Discover {
	return filtered{
		Discover: applyTo,
		logger:   logger,
		filter:   filter,
	}
}

// Devices returns a filtered list of devices based on the specified filter.
func (d filtered) Devices() ([]Device, error) {
	devices, err := d.Discover.Devices()
	if err != nil {
		return nil, err
	}

	if d.filter == nil {
		return devices, nil
	}

	var selected []Device
	for _, device := range devices {
		if d.filter.DeviceIsSelected(device) {
			selected = append(selected, device)
		} else {
			d.logger.Infof("Skipping device %v", device)
		}
	}

	return selected, nil
}
