/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package device

import (
	"fmt"
	"hcu-container-toolkit/pkg/go-c3000smi/pkg/c3000smi"
)

// Device defines the set of extended functions associated with a device.Device.
type Device interface {
	c3000smi.Device
}

type device struct {
	c3000smi.Device
	lib *devicelib
}

var _ Device = &device{}

// newDevice creates a device from an c3000smi.Device
func (d *devicelib) newDevice(dev c3000smi.Device) (*device, error) {
	return &device{dev, d}, nil
}

// isSkipped checks whether the device should be skipped.
func (d *device) isSkipped() (bool, error) {
	name := d.GetName()
	if _, exists := d.lib.skippedDevices[name]; exists {
		return true, nil
	}
	return false, nil
}

// VisitDevices visits each top-level device and invokes a callback function for it.
func (d *devicelib) VisitDevices(visit func(string, Device) error) error {
	indexs := d.c3000smicmd.DeviceGetIndexs()
	for _, index := range indexs {
		device, err := d.c3000smicmd.DeviceGetHandleByIndex(index)
		if err != nil {
			return fmt.Errorf("error getting device hanlde for index '%v': %v", index, err)
		}
		dev, err := d.newDevice(device)
		if err != nil {
			return fmt.Errorf("error creating new device wrapper: %v", err)
		}

		isSkipped, err := dev.isSkipped()
		if err != nil {
			return fmt.Errorf("error checking whether device is skipped: %v", err)
		}
		if isSkipped {
			continue
		}

		err = visit(index, dev)
		if err != nil {
			return fmt.Errorf("error visiting device: %v", err)
		}
	}
	return nil
}
