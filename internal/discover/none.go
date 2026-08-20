/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

// None is a null discoverer that returns an empty list of devices and
// mounts.
type None struct{}

var _ Discover = (*None)(nil)

// Devices returns an empty list of devices
func (e None) Devices() ([]Device, error) {
	return []Device{}, nil
}

// Mounts returns an empty list of mounts
func (e None) Mounts() ([]Mount, error) {
	return []Mount{}, nil
}

// Hooks returns an empty list of hooks
func (e None) Hooks() ([]Hook, error) {
	return []Hook{}, nil
}

// AdditionalGIDs return an empty list of additionalGIDs
func (e None) AdditionalGIDs() ([]uint32, error) {
	return []uint32{}, nil
}
