/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package status

import (
	"fmt"
	hcuTracker "hcu-container-toolkit/internal/hcu-tracker"
	"os/user"

	"hcu-container-toolkit/internal/logger"

	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}
func (c *command) build() *cli.Command {
	hcuTrackerStatusCmd := cli.Command{
		Name:      "status",
		Usage:     "Show Status of HCUs",
		UsageText: "hcu-ctk hcu-tracker status [options]",
		Before: func(c *cli.Context) error {
			return validateGenOptions(c)
		},
		Action: func(c *cli.Context) error {
			return performAction(c)
		},
	}

	return &hcuTrackerStatusCmd
}

func validateGenOptions(c *cli.Context) error {
	curUser, err := user.Current()
	if err != nil || curUser.Uid != "0" {
		return fmt.Errorf("Permission denied: Not running as root")
	}

	return nil
}

func performAction(c *cli.Context) error {
	hcuTracker, err := hcuTracker.New()
	if err != nil {
		return fmt.Errorf("Failed to create HCU tracker, Error: %v", err)
	}

	err = hcuTracker.ShowStatus()
	if err != nil {
		return fmt.Errorf("Failed to show HCUs status, Error: %v", err)
	}

	return nil
}
