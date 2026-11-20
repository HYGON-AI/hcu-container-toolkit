/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package gc

import (
	"fmt"

	"hcu-container-toolkit/cmd/hcu-ctk/vhcu/common"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/vhcu"

	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

// NewCommand constructs the `hcu-ctk vhcu gc` command.
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

func (m command) build() *cli.Command {
	vhcuGCCmd := cli.Command{
		Name:  "gc",
		Usage: "Reclaim dynamic vHCUs whose owning containers are stopped or absent",
		UsageText: `hcu-ctk vhcu gc

Examples:
	hcu-ctk vhcu gc`,
		Before: func(c *cli.Context) error {
			return validateGenOptions(c)
		},
		Action: func(c *cli.Context) error {
			return performAction(c, m.logger)
		},
	}

	return &vhcuGCCmd
}

func validateGenOptions(c *cli.Context) error {
	return common.ValidateRoot()
}

func performAction(c *cli.Context, logger logger.Interface) error {
	manager, err := common.NewManager(logger)
	if err != nil {
		return fmt.Errorf("Failed to create vhcu manager, Error: %v", err)
	}

	result, err := manager.GC(vhcu.ProbeRuntime)
	if err != nil {
		return fmt.Errorf("Failed to run vhcu GC, Error: %v", err)
	}

	if len(result.Released) == 0 {
		fmt.Fprintln(c.App.Writer, "No dynamic vHCU to reclaim")
	} else {
		for _, id := range result.Released {
			fmt.Fprintf(c.App.Writer, "Released vHCU of container %s\n", id)
		}
	}

	// Orphan vdevs are never destroyed automatically: they may be
	// administratively pre-split. Report them for manual handling.
	for _, vdev := range result.Orphans {
		fmt.Fprintf(c.App.Writer, "WARNING: vdev %d exists in the driver but has no toolkit ownership record (administratively pre-split or leaked); not destroyed\n", vdev)
	}

	for _, id := range result.Deferred {
		fmt.Fprintf(c.App.Writer, "WARNING: allocation for %s preserved; inspect with hcu-ctk vhcu show %s\n", id, id)
	}
	if len(result.Deferred) > 0 {
		return fmt.Errorf("%d allocations require retry or operator review", len(result.Deferred))
	}
	return nil
}
