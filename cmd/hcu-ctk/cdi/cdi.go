/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package cdi

import (
	"hcu-container-toolkit/cmd/hcu-ctk/cdi/generate"
	"hcu-container-toolkit/cmd/hcu-ctk/cdi/list"
	"hcu-container-toolkit/cmd/hcu-ctk/cdi/transform"
	"hcu-container-toolkit/cmd/hcu-ctk/cdi/validate"
	"hcu-container-toolkit/internal/logger"

	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

// NewCommand constructs an info command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

// build
func (m command) build() *cli.Command {
	// Create the 'hook' command
	hook := cli.Command{
		Name:  "cdi",
		Usage: "Provide tools for interacting with Container Device Interface specifications",
	}

	hook.Subcommands = []*cli.Command{
		generate.NewCommand(m.logger),
		transform.NewCommand(m.logger),
		list.NewCommand(m.logger),
		validate.NewCommand(m.logger),
	}

	return &hook
}
