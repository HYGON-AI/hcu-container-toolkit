/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package c3000smi

import "fmt"

// Interface represents the interface for the library type.
type Interface interface {
	IsValid() bool
	DeviceGetCount() int
	DeviceGetIndexs() []string
	DeviceGetHandleByIndex(string) (Device, error)
}

// Device represents the interface for the c3000Device type.
type Device interface {
	GetUUID() string
	GetName() string
	GetPCIBusID() string
}

type DeviceInfo struct {
	uniqueId     string
	pciBusId     string
	serialNumber string
	cardSeries   string
	cardVendor   string
}

type Devices map[string]*DeviceInfo

var _ Device = (*DeviceInfo)(nil)

func (d *DeviceInfo) GetName() string {
	return fmt.Sprintf("%s %s", d.cardSeries, d.cardVendor)
}

func (d *DeviceInfo) GetUUID() string {
	return d.serialNumber
}

func (d *DeviceInfo) GetPCIBusID() string {
	return d.pciBusId
}
