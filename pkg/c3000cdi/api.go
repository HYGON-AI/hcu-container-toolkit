/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package c3000cdi

import (
	"hcu-container-toolkit/pkg/c3000cdi/spec"
	"hcu-container-toolkit/pkg/go-c3000lib/pkg/device"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"
)

// Interface defines the API for the c3000cdi package
type Interface interface {
	GetSpec() (spec.Interface, error)
	GetCommonEdits() (*cdi.ContainerEdits, error)
	GetAllDeviceSpecs() ([]specs.Device, error)
	GetHCUDeviceEdits(device.Device) (*cdi.ContainerEdits, error)
	GetHCUDeviceSpecs(string, device.Device) ([]specs.Device, error)
	GetMIGDeviceEdits(device.Device, device.MigDevice) (*cdi.ContainerEdits, error)
	GetMIGDeviceSpecs(int, device.Device, int, device.MigDevice) ([]specs.Device, error)
	GetDeviceSpecsByID(...string) ([]specs.Device, error)
}
