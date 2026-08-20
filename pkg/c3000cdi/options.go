/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package c3000cdi

import (
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/pkg/c3000cdi/transform"
	"hcu-container-toolkit/pkg/go-c3000lib/pkg/device"
	"hcu-container-toolkit/pkg/go-c3000smi/pkg/c3000smi"
)

// Option is a function that configures the c3000cdilib
type Option func(*c3000cdilib)

// WithDeviceLib sets the device library for the library
func WithDeviceLib(devicelib device.Interface) Option {
	return func(l *c3000cdilib) {
		l.devicelib = devicelib
	}
}

// WithDeviceNamers sets the device namer for the library
func WithDeviceNamers(namers ...DeviceNamer) Option {
	return func(l *c3000cdilib) {
		l.deviceNamers = namers
	}
}

// WithDriverRoot sets the driver root for the library
func WithDriverRoot(root string) Option {
	return func(l *c3000cdilib) {
		l.driverRoot = root
	}
}

// WithDevRoot sets the root where /dev is located.
func WithDevRoot(root string) Option {
	return func(l *c3000cdilib) {
		l.devRoot = root
	}
}

// WithLogger sets the logger for the library
func WithLogger(logger logger.Interface) Option {
	return func(l *c3000cdilib) {
		l.logger = logger
	}
}

// WithHCUCDIHookPath sets the path to the HCU Container Toolkit CLI path for the library
func WithHCUCDIHookPath(path string) Option {
	return func(l *c3000cdilib) {
		l.hcuCDIHookPath = path
	}
}

// WithNvmlLib sets the nvml library for the library
func WithNvmlLib(c3000smicmd c3000smi.Interface) Option {
	return func(l *c3000cdilib) {
		l.c3000smicmd = c3000smicmd
	}
}

// WithMode sets the discovery mode for the library
func WithMode[m modeConstraint](mode m) Option {
	return func(l *c3000cdilib) {
		l.mode = Mode(mode)
	}
}

// WithVendor sets the vendor for the library
func WithVendor(vendor string) Option {
	return func(o *c3000cdilib) {
		o.vendor = vendor
	}
}

// WithClass sets the class for the library
func WithClass(class string) Option {
	return func(o *c3000cdilib) {
		o.class = class
	}
}

// WithMergedDeviceOptions sets the merged device options for the library
// If these are not set, no merged device will be generated.
func WithMergedDeviceOptions(opts ...transform.MergedDeviceOption) Option {
	return func(o *c3000cdilib) {
		o.mergedDeviceOptions = opts
	}
}
