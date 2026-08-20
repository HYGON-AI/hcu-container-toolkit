/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package modifier

import (
	"hcu-container-toolkit/internal/config"
	"hcu-container-toolkit/internal/config/image"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/oci"
)

// NewStableModifier creates the modifiers for general features.
// These includes:
//
// removes inserted nvidia-container-runtime hooks for --gpus
func NewStableModifier(logger logger.Interface, cfg *config.Config, image image.DTK) (oci.SpecModifier, error) {
	var modifiers List

	modifiers = append(modifiers, NewNvidiaContainerRuntimeHookRemover(logger))
	modifiers = append(modifiers, NewSeccompRemover(logger))

	var addCaps []string

	// For xprof
	addCaps = append(addCaps, "CAP_SYS_RAWIO")
	if image.Getenv("HCU_MOFED") == "enabled" {
		addCaps = append(addCaps, "CAP_IPC_LOCK")
	}
	modifiers = append(modifiers, NewCapModifier(logger, addCaps, []string{}))

	return modifiers, nil
}
