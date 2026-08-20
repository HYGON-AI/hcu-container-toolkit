/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package lookup

import (
	"fmt"
	"os"
)

const (
	devRoot = "/dev"
)

// NewCharDeviceLocator creates a Locator that can be used to find char devices at the specified root. A logger is
// also specified.
func NewCharDeviceLocator(opts ...Option) Locator {
	opts = append(opts,
		WithSearchPaths("", devRoot),
		WithFilter(assertCharDevice),
	)
	return NewFileLocator(
		opts...,
	)
}

// assertCharDevice checks whether the specified path is a char device and returns an error if this is not the case.
func assertCharDevice(filename string) error {
	info, err := os.Lstat(filename)
	if err != nil {
		return fmt.Errorf("error getting info: %v", err)
	}
	if info.Mode()&os.ModeCharDevice == 0 {
		return fmt.Errorf("%v is not a char device", filename)
	}
	return nil
}
