/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package show

import (
	"encoding/json"
	"fmt"

	"hcu-container-toolkit/cmd/hcu-ctk/vhcu/common"
	"hcu-container-toolkit/internal/logger"

	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

// NewCommand constructs the `hcu-ctk vhcu show` command.
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

func (m command) build() *cli.Command {
	vhcuShowCmd := cli.Command{
		Name:  "show",
		Usage: "Show the vHCU state details of a container",
		UsageText: `hcu-ctk vhcu show [container_id]

	Arguments:
		container_id  container ID of the container

Examples:
	hcu-ctk vhcu show a4e19862b4e2a1b04a1f793f346d0411f4a0a3857578c526a25ac6c858168fd8`,
		Before: func(c *cli.Context) error {
			return validateGenOptions(c)
		},
		Action: func(c *cli.Context) error {
			return performAction(c, m.logger)
		},
	}

	return &vhcuShowCmd
}

func validateGenOptions(c *cli.Context) error {
	return common.ValidateRoot()
}

func performAction(c *cli.Context, logger logger.Interface) error {
	if c.Args().Len() != 1 {
		return cli.Exit("Error: Missing arguments. Usage: hcu-ctk vhcu show <container_id>", 1)
	}

	manager, err := common.NewManager(logger)
	if err != nil {
		return fmt.Errorf("Failed to create vhcu manager, Error: %v", err)
	}

	st, err := manager.States().Load(c.Args().Get(0))
	if err != nil {
		return fmt.Errorf("Failed to load vhcu state, Error: %v", err)
	}
	if st == nil {
		return cli.Exit(fmt.Sprintf("No vHCU state found for container %s", c.Args().Get(0)), 1)
	}

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to marshal vhcu state, Error: %v", err)
	}
	fmt.Fprintln(c.App.Writer, string(data))

	return nil
}
