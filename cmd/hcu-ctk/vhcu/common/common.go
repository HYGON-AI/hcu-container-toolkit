/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

// Package common holds helpers shared by the `hcu-ctk vhcu` subcommands
// (kept in its own package to avoid an import cycle between the command
// group and its subcommands).
package common

import (
	"fmt"
	"os"

	"hcu-container-toolkit/internal/config"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/vhcu"
)

// NewManager builds the vhcu manager from the toolkit config.
func NewManager(logger logger.Interface) (*vhcu.Manager, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}
	return vhcu.NewManagerFromConfig(logger, cfg.HCUContainerRuntimeConfig)
}

// ValidateRoot enforces root execution, mirroring hcu-tracker subcommands.
func ValidateRoot() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("Permission denied: Not running as root")
	}
	return nil
}
