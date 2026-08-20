/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package c3000cdi

import (
	"errors"
	"fmt"
)

// UUIDer is an interface for getting UUIDs.
type UUIDer interface {
	GetUUID() string
}

// DeviceNamers represents a list of device namers
type DeviceNamers []DeviceNamer

// DeviceNamer is an interface for getting device names
type DeviceNamer interface {
	GetDeviceName(string, UUIDer) (string, error)
	GetMigDeviceName(string, UUIDer, int, UUIDer) (string, error)
}

// Supported device naming strategies
const (
	// DeviceNameStrategyIndex generates devices names such as 0 or 1:0
	DeviceNameStrategyIndex = "index"
	// DeviceNameStrategyTypeIndex generates devices names such as hcu0 or mig1:0
	DeviceNameStrategyTypeIndex = "type-index"
	// DeviceNameStrategyUUID uses the device UUID as the name
	DeviceNameStrategyUUID = "uuid"
)

type deviceNameIndex struct {
	hcuPrefix string
	migPrefix string
}
type deviceNameUUID struct{}

// NewDeviceNamer creates a Device Namer based on the supplied strategy.
// This namer can be used to construct the names for MIG and HCU devices when generating the CDI spec.
func NewDeviceNamer(strategy string) (DeviceNamer, error) {
	switch strategy {
	case DeviceNameStrategyIndex:
		return deviceNameIndex{}, nil
	case DeviceNameStrategyTypeIndex:
		return deviceNameIndex{hcuPrefix: "hcu", migPrefix: "mig"}, nil
	case DeviceNameStrategyUUID:
		return deviceNameUUID{}, nil
	}

	return nil, fmt.Errorf("invalid device name strategy: %v", strategy)
}

// GetDeviceName returns the name for the specified device based on the naming strategy
func (s deviceNameIndex) GetDeviceName(i string, _ UUIDer) (string, error) {
	return fmt.Sprintf("%s%s", s.hcuPrefix, i), nil
}

// GetMigDeviceName returns the name for the specified device based on the naming strategy
func (s deviceNameIndex) GetMigDeviceName(i string, _ UUIDer, j int, _ UUIDer) (string, error) {
	return fmt.Sprintf("%s%s:%d", s.migPrefix, i, j), nil
}

// GetDeviceName returns the name for the specified device based on the naming strategy
func (s deviceNameUUID) GetDeviceName(i string, d UUIDer) (string, error) {
	return fmt.Sprintf("hcu-%s", d.GetUUID()), nil
}

// GetMigDeviceName returns the name for the specified device based on the naming strategy
func (s deviceNameUUID) GetMigDeviceName(i string, _ UUIDer, j int, mig UUIDer) (string, error) {
	return mig.GetUUID(), nil
}

type c3000smiUUIDer interface {
	GetUUID() string
}

type convert struct {
	c3000smiUUIDer
}

func (l DeviceNamers) GetDeviceNames(i string, d UUIDer) ([]string, error) {
	var names []string
	for _, namer := range l {
		name, err := namer.GetDeviceName(i, d)
		if err != nil {
			return nil, err
		}
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil, errors.New("no names defined")
	}
	return names, nil
}
