/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package release

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

	hcuTrackerReleaseCmd := cli.Command{
		Name:   "release",
		Hidden: true,
		Usage:  "Release HCUs used by a container",
		UsageText: `hcu-ctk hcu-tracker release [container_id]

	Arguments:
		container_id  container ID of the container

	Examples:
		hcu-ctk hcu-tracker release a4e19862b4e2a1b04a1f793f346d0411f4a0a3857578c526a25ac6c858168fd8`,
		Before: func(c *cli.Context) error {
			return validateGenOptions(c)
		},
		Action: func(c *cli.Context) error {
			return performAction(c)
		},
	}

	return &hcuTrackerReleaseCmd
}

func validateGenOptions(c *cli.Context) error {
	curUser, err := user.Current()
	if err != nil || curUser.Uid != "0" {
		return fmt.Errorf("Permission denied: Not running as root")
	}

	return nil
}

func performAction(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return cli.ShowAppHelp(c)
	}

	hcuTracker, err := hcuTracker.New()
	if err != nil {
		return fmt.Errorf("Failed to create HCU tracker, Error: %v", err)
	}

	err = hcuTracker.ReleaseHCUs(c.Args().Get(0))
	if err != nil {
		return fmt.Errorf("Failed to release HCUs, Error: %v", err)
	}

	return nil
}
