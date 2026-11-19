/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package list

import (
	"fmt"
	"strings"

	"hcu-container-toolkit/cmd/hcu-ctk/vhcu/common"
	"hcu-container-toolkit/internal/logger"

	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

// NewCommand constructs the `hcu-ctk vhcu list` command.
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

func (m command) build() *cli.Command {
	vhcuListCmd := cli.Command{
		Name:  "list",
		Usage: "List all dynamically managed vHCUs with their owning containers",
		UsageText: `hcu-ctk vhcu list

Examples:
	hcu-ctk vhcu list`,
		Before: func(c *cli.Context) error {
			return validateGenOptions(c)
		},
		Action: func(c *cli.Context) error {
			return performAction(c, m.logger)
		},
	}

	return &vhcuListCmd
}

func validateGenOptions(c *cli.Context) error {
	return common.ValidateRoot()
}

func performAction(c *cli.Context, logger logger.Interface) error {
	manager, err := common.NewManager(logger)
	if err != nil {
		return fmt.Errorf("Failed to create vhcu manager, Error: %v", err)
	}

	states, err := manager.States().List()
	if err != nil {
		return fmt.Errorf("Failed to list vhcu states, Error: %v", err)
	}

	fmt.Fprintln(c.App.Writer, strings.Repeat("-", 130))
	fmt.Fprintf(c.App.Writer, "%-20s%-12s%-10s%-10s%-20s%-10s%-12s\n", "CONTAINER", "PHASE", "DEVICE", "VDEV", "PCI BUS ID", "CORES", "MEM (MiB)")
	fmt.Fprintln(c.App.Writer, strings.Repeat("-", 130))
	if len(states) == 0 {
		fmt.Fprintf(c.App.Writer, "%-20s\n", "None")
	}
	for _, st := range states {
		vdev := fmt.Sprintf("%d", st.VdevIndex)
		if st.VdevIndex < 0 {
			vdev = "-"
		}
		fmt.Fprintf(c.App.Writer, "%-20s%-12s%-10d%-10s%-20s%-10d%-12d\n",
			shortID(st.ContainerID), st.Phase, st.DeviceIndex, vdev, st.PciBusID, st.Cores, st.MemMiB)
	}
	fmt.Fprintln(c.App.Writer, strings.Repeat("-", 130))

	return nil
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
