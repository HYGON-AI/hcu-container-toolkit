/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"hcu-container-toolkit/cmd/hcu-ctk/vhcu/check"
	"hcu-container-toolkit/cmd/hcu-ctk/vhcu/gc"
	"hcu-container-toolkit/cmd/hcu-ctk/vhcu/list"
	"hcu-container-toolkit/cmd/hcu-ctk/vhcu/release"
	"hcu-container-toolkit/cmd/hcu-ctk/vhcu/show"
	"hcu-container-toolkit/cmd/hcu-ctk/vhcu/status"
	"hcu-container-toolkit/internal/logger"

	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

// NewCommand constructs the `hcu-ctk vhcu` command group.
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

func (m command) build() *cli.Command {
	vhcuCmd := cli.Command{
		Name:      "vhcu",
		Usage:     "Dynamic vHCU (virtual HCU) split related commands",
		UsageText: `hcu-ctk vhcu [command] [options]`,
		Action: func(c *cli.Context) error {
			return cli.ShowAppHelp(c)
		},
	}

	vhcuCmd.Subcommands = []*cli.Command{
		check.NewCommand(m.logger),
		gc.NewCommand(m.logger),
		list.NewCommand(m.logger),
		release.NewCommand(m.logger),
		show.NewCommand(m.logger),
		status.NewCommand(m.logger),
	}
	return &vhcuCmd
}
