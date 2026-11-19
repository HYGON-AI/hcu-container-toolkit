/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package status

import (
	"fmt"
	"strings"

	"hcu-container-toolkit/cmd/hcu-ctk/vhcu/common"
	"hcu-container-toolkit/internal/logger"
	internalvhcu "hcu-container-toolkit/internal/vhcu"

	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

// NewCommand constructs the `hcu-ctk vhcu status` command.
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

func (m command) build() *cli.Command {
	vhcuStatusCmd := cli.Command{
		Name:  "status",
		Usage: "Summarize per-card capacity, remaining resources, and managed vHCU count",
		UsageText: `hcu-ctk vhcu status

Examples:
	hcu-ctk vhcu status`,
		Before: func(c *cli.Context) error {
			return validateGenOptions(c)
		},
		Action: func(c *cli.Context) error {
			return performAction(c, m.logger)
		},
	}

	return &vhcuStatusCmd
}

func validateGenOptions(c *cli.Context) error {
	return common.ValidateRoot()
}

func performAction(c *cli.Context, logger logger.Interface) error {
	manager, err := common.NewManager(logger)
	if err != nil {
		return fmt.Errorf("Failed to create vhcu manager, Error: %v", err)
	}

	views, err := manager.Views()
	if err != nil {
		return fmt.Errorf("Failed to gather card capacity, Error: %v", err)
	}

	states, err := manager.States().List()
	if err != nil {
		return fmt.Errorf("Failed to list vhcu states, Error: %v", err)
	}

	// Toolkit-managed vdev count per card.
	owned := make(map[int]int)
	for _, st := range states {
		if st.VdevIndex >= 0 {
			owned[st.DeviceIndex]++
		}
	}

	internalvhcu.SortViews(views)

	fmt.Fprintln(c.App.Writer, strings.Repeat("-", 140))
	fmt.Fprintf(c.App.Writer, "%-8s%-20s%-12s%-14s%-14s%-14s%-12s%-10s\n",
		"DEVICE", "PCI BUS ID", "TOTAL CU", "TOTAL (MiB)", "REMAIN CU", "REMAIN (MiB)", "SLOTS", "SOURCE")
	fmt.Fprintln(c.App.Writer, strings.Repeat("-", 140))
	for _, v := range views {
		slots := fmt.Sprintf("%d/%d", v.UsedSlots, manager.MaxVHCUPerCard())
		if v.Excluded {
			slots += " (excluded)"
		}
		source := "driver"
		if v.Remaining.Derived {
			// `hy-smi virtual --show-remaining-info` was unavailable; the
			// numbers come from device/usage info and are conservative.
			source = "derived"
		}
		fmt.Fprintf(c.App.Writer, "%-8d%-20s%-12d%-14d%-14d%-14d%-12s%-10s\n",
			v.Dev.Index, v.Dev.BusID, v.Dev.ComputeUnits, v.Dev.MemoryMiB,
			v.Remaining.ComputeUnits, v.Remaining.MemoryMiB, slots, source)
		if v.Excluded {
			fmt.Fprintf(c.App.Writer, "         reason: %s\n", v.ExcludedReason)
		}
	}
	fmt.Fprintln(c.App.Writer, strings.Repeat("-", 140))
	fmt.Fprintf(c.App.Writer, "Toolkit-managed dynamic vHCUs: %d\n", len(states))
	for dev, count := range owned {
		fmt.Fprintf(c.App.Writer, "  device %d: %d\n", dev, count)
	}

	return nil
}
