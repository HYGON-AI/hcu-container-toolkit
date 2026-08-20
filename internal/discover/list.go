/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

import "fmt"

// list is a discoverer that contains a list of Discoverers. The output of the
// Mounts functions is the concatenation of the output for each of the
// elements in the list.
type list struct {
	discoverers []Discover
}

var _ Discover = (*list)(nil)

// Merge creates a discoverer that is the composite of a list of discoverers.
func Merge(d ...Discover) Discover {
	l := list{
		discoverers: d,
	}

	return &l
}

// Devices returns all devices from the included discoverers
func (d list) Devices() ([]Device, error) {
	var allDevices []Device

	for i, di := range d.discoverers {
		devices, err := di.Devices()
		if err != nil {
			return nil, fmt.Errorf("error discovering devices for discoverer %v: %v", i, err)
		}
		allDevices = append(allDevices, devices...)
	}

	return allDevices, nil
}

// Mounts returns all mounts from the included discoverers
func (d list) Mounts() ([]Mount, error) {
	var allMounts []Mount

	for i, di := range d.discoverers {
		mounts, err := di.Mounts()
		if err != nil {
			return nil, fmt.Errorf("error discovering mounts for discoverer %v: %v", i, err)
		}
		allMounts = append(allMounts, mounts...)
	}

	return allMounts, nil
}

// Hooks returns all Hooks from the included discoverers
func (d list) Hooks() ([]Hook, error) {
	var allHooks []Hook

	for i, di := range d.discoverers {
		hooks, err := di.Hooks()
		if err != nil {
			return nil, fmt.Errorf("error discovering hooks for discoverer %v: %v", i, err)
		}
		allHooks = append(allHooks, hooks...)
	}

	return allHooks, nil
}

func (d list) AdditionalGIDs() ([]uint32, error) {
	var additionalGIDs []uint32

	for i, di := range d.discoverers {
		gids, err := di.AdditionalGIDs()
		if err != nil {
			return nil, fmt.Errorf("error discovering AdditionalGIDs for discoverer %v: %v", i, err)
		}
		additionalGIDs = append(additionalGIDs, gids...)
	}

	return additionalGIDs, nil
}
