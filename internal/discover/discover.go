/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

// Device represents a discovered character device.
type Device struct {
	HostPath string
	Path     string
}

// Mount represents a discovered mount.
type Mount struct {
	HostPath string
	Path     string
	Options  []string
}

// Hook represents a discovered hook.
type Hook struct {
	Lifecycle string
	Path      string
	Args      []string
}

// Discover defines an interface for discovering the devices, mounts, and hooks available on a system
//
//go:generate moq -stub -out discover_mock.go . Discover
type Discover interface {
	Devices() ([]Device, error)
	Mounts() ([]Mount, error)
	Hooks() ([]Hook, error)
	AdditionalGIDs() ([]uint32, error)
}
