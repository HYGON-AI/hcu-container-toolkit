/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package hcuTracker

import (
	"fmt"
	"hcu-container-toolkit/cmd/hcu-ctk/hcu-tracker/disable"
	"hcu-container-toolkit/cmd/hcu-ctk/hcu-tracker/enable"
	"hcu-container-toolkit/cmd/hcu-ctk/hcu-tracker/initialize"
	"hcu-container-toolkit/cmd/hcu-ctk/hcu-tracker/release"
	"hcu-container-toolkit/cmd/hcu-ctk/hcu-tracker/reset"
	"hcu-container-toolkit/cmd/hcu-ctk/hcu-tracker/status"
	"hcu-container-toolkit/internal/hcu-tracker"
	"hcu-container-toolkit/internal/logger"
	"os/user"

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
	hcuTrackerCmd := cli.Command{
		Name:  "hcu-tracker",
		Usage: "HCU Tracker  related commands",
		UsageText: `hcu-ctk hcu-tracker [hcu-ids] [accessibility]
    Arguments:
        hcu-ids          Comma-separated list of HCU IDs (comma separated list, range operator, all)
        accessibility    Must be either 'exclusive' or 'shared'
		
	Examples:
        hcu-ctk hcu-tracker 0,1,2 exclusive
        hcu-ctk hcu-tracker 0,1-2 shared
        hcu-ctk hcu-tracker all shared

OR

hcu-ctk hcu-tracker [command] [options]`,
		Before: func(c *cli.Context) error { return m.validateGenOptions(c) },
		Action: func(c *cli.Context) error { return m.performAction(c) },
	}

	hcuTrackerCmd.Subcommands = []*cli.Command{
		disable.NewCommand(m.logger),
		enable.NewCommand(m.logger),
		initialize.NewCommand(m.logger),
		reset.NewCommand(m.logger),
		release.NewCommand(m.logger),
		status.NewCommand(m.logger),
	}
	return &hcuTrackerCmd
}

func (m command) validateGenOptions(c *cli.Context) error {
	curUser, err := user.Current()
	if err != nil || curUser.Uid != "0" {
		return fmt.Errorf("Permission denied: Not running as root")
	}

	return nil
}

func (m command) performAction(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return cli.ShowAppHelp(c)
	}

	if c.Args().Len() > 2 {
		return cli.Exit("Error: Missing arguments. Usgae: hcu-tracker <hcu_id> <operation>", 1)
	}

	hcuIDs := c.Args().Get(0)
	operation := c.Args().Get(1)

	hcuTracker, err := hcuTracker.New()
	if err != nil {
		return fmt.Errorf("Failed to create HCU tracker, Error: %v", err)
	}

	switch operation {
	case "exclusive":
		hcuTracker.MakeHCUsExclusive(hcuIDs)
	case "shared":
		hcuTracker.MakeHCUsShared(hcuIDs)
	default:
		return cli.Exit("Error: Invalid operation. Must be 'exclusive' or 'shared", 1)
	}

	return nil
}
