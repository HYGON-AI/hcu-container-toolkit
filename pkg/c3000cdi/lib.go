/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package c3000cdi

import (
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/lookup/root"
	"hcu-container-toolkit/pkg/c3000cdi/transform"
	"hcu-container-toolkit/pkg/go-c3000lib/pkg/device"
	"hcu-container-toolkit/pkg/go-c3000smi/pkg/c3000smi"
)

type c3000cdilib struct {
	logger         logger.Interface
	c3000smicmd    c3000smi.Interface
	mode           Mode
	devicelib      device.Interface
	deviceNamers   DeviceNamers
	driverRoot     string
	devRoot        string
	hcuCDIHookPath string

	vendor string
	class  string

	driver *root.Driver

	mergedDeviceOptions []transform.MergedDeviceOption
}

// New creates a new c3000cdi library
func New(opts ...Option) (Interface, error) {
	l := &c3000cdilib{}
	for _, opt := range opts {
		opt(l)
	}
	if l.mode == "" {
		l.mode = ModeAuto
	}
	if l.logger == nil {
		l.logger = logger.New()
	}
	if len(l.deviceNamers) == 0 {
		indexNamer, _ := NewDeviceNamer(DeviceNameStrategyIndex)
		l.deviceNamers = []DeviceNamer{indexNamer}
	}
	if l.hcuCDIHookPath == "" {
		l.hcuCDIHookPath = "/usr/bin/hcu-cdi-hook"
	}
	if l.driverRoot == "" {
		l.driverRoot = "/"
	}
	if l.devRoot == "" {
		l.devRoot = "/"
	}

	l.driver = root.New(
		root.WithLogger(l.logger),
		root.WithDriverRoot(l.driverRoot),
	)
	if l.c3000smicmd == nil {
		smicmd, err := c3000smi.NewSmiCommand(l.logger)
		if err != nil {
			return nil, fmt.Errorf("failed to new smi commd: %v", err)
		}
		l.c3000smicmd = smicmd
	}
	if l.devicelib == nil {
		l.devicelib = device.New(l.c3000smicmd)
	}

	var lib Interface
	switch l.resolveMode() {
	case ModeSmi:
		lib = (*c3000smilib)(l)
	default:
		return nil, fmt.Errorf("unknown mode %q", l.mode)
	}

	w := wrapper{
		Interface:           lib,
		vendor:              l.vendor,
		class:               l.class,
		mergedDeviceOptions: l.mergedDeviceOptions,
	}
	return &w, nil
}
