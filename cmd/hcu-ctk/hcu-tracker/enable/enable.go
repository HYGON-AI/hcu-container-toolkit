/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package enable

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

func (m command) build() *cli.Command {

	hcuTrackerEnableCmd := cli.Command{
		Name:      "enable",
		Usage:     "Enable the HCU Tracker",
		UsageText: "hcu-ctk hcu-tracker enable [options]",
		Before:    func(c *cli.Context) error { return validateGenOptions(c) },
		Action:    func(c *cli.Context) error { return performAction(c) },
	}

	return &hcuTrackerEnableCmd
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

	err = hcuTracker.Enable()
	if err != nil {
		return fmt.Errorf("Failed to Reset HCU Tracker, Error: %v", err)
	}

	return nil
}
