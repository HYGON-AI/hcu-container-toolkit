/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package hook

import (
	"hcu-container-toolkit/cmd/hcu-cdi-hook/commands"
	"hcu-container-toolkit/internal/logger"

	"github.com/urfave/cli/v2"
)

type hookCommand struct {
	logger logger.Interface
}

// NewCommand constructs a hook command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := hookCommand{
		logger: logger,
	}
	return c.build()
}

// build
func (m hookCommand) build() *cli.Command {
	// Create the 'hook' command
	hook := cli.Command{
		Name:  "hook",
		Usage: "A collection of hooks that may be injected into an OCI spec",
	}

	hook.Subcommands = commands.New(m.logger)

	return &hook
}
