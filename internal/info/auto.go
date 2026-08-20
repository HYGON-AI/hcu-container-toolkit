/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package info

import (
	"hcu-container-toolkit/internal/config/image"
	"hcu-container-toolkit/internal/logger"
)

// ResolveAutoMode determines the correct mode for the platform if set to "auto"
func ResolveAutoMode(logger logger.Interface, mode string, image image.DTK) (rmode string) {
	return resolveMode(logger, mode, image)
}

func resolveMode(logger logger.Interface, mode string, image image.DTK) (rmode string) {
	if mode != "auto" {
		logger.Infof("Using requested mode '%s'", mode)
		return mode
	}
	defer func() {
		logger.Infof("Auto-detected mode as '%v'", rmode)
	}()

	if image.OnlyFullyQualifiedCDIDevices() {
		return "cdi"
	}

	return "legacy"
}
