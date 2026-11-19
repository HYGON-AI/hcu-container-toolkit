/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package check

import (
	"fmt"

	"hcu-container-toolkit/internal/config"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/vhcu"

	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

// NewCommand constructs the `hcu-ctk vhcu check` command.
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

func (m command) build() *cli.Command {
	vhcuCheckCmd := cli.Command{
		Name:  "check",
		Usage: "Run an environment self-check for dynamic vHCU split support",
		UsageText: `hcu-ctk vhcu check

Examples:
	hcu-ctk vhcu check`,
		Action: func(c *cli.Context) error {
			return performAction(c, m.logger)
		},
	}

	return &vhcuCheckCmd
}

func performAction(c *cli.Context, logger logger.Interface) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("Failed to load config, Error: %v", err)
	}

	vhcuCfg := cfg.HCUContainerRuntimeConfig.VHCU
	caps := vhcu.Detect(vhcuCfg.HysmiPath)

	fmt.Fprintf(c.App.Writer, "Dynamic vHCU split enabled in config : %v\n", vhcuCfg.Enabled)
	fmt.Fprintf(c.App.Writer, "hy-smi path                         : %s\n", orDash(caps.HysmiPath))
	fmt.Fprintf(c.App.Writer, "hy-smi resolved from                : %s\n", caps.HysmiSource)
	fmt.Fprintf(c.App.Writer, "Running as root                     : %v\n", caps.Root)
	fmt.Fprintf(c.App.Writer, "hy-smi virtual usable               : %v\n", caps.VirtualSupported)
	fmt.Fprintf(c.App.Writer, "Driver version                      : %s (minimum %s)\n", orDash(caps.DriverVersion), "6.2.26")
	fmt.Fprintf(c.App.Writer, "Per-card vdev limit (config)        : %d\n", vhcuCfg.MaxVHCUPerCard)
	fmt.Fprintf(c.App.Writer, "State dir                           : %s\n", vhcuCfg.StateDir)

	for _, warning := range caps.Warnings {
		fmt.Fprintf(c.App.Writer, "WARNING: %s\n", warning)
	}

	if !vhcuCfg.Enabled {
		fmt.Fprintln(c.App.Writer, "Result: dynamic vHCU split is DISABLED in config")
		return nil
	}
	if !caps.Root {
		fmt.Fprintln(c.App.Writer, "Result: FAILED (rootless mode is unsupported for dynamic vHCU split)")
		return cli.Exit("", 1)
	}
	if !caps.Usable() {
		fmt.Fprintln(c.App.Writer, "Result: FAILED")
		return cli.Exit("", 1)
	}

	// End-to-end capacity probe: exercises --show-device-info,
	// --show-vdevice-info and the remaining-amount path for every card, so a
	// broken parser is caught here instead of at the first `docker run`.
	manager, err := vhcu.NewManagerFromConfig(logger, cfg.HCUContainerRuntimeConfig)
	if err != nil {
		fmt.Fprintf(c.App.Writer, "Capacity probe                     : FAILED (%v)\n", err)
		fmt.Fprintln(c.App.Writer, "Result: FAILED")
		return cli.Exit("", 1)
	}
	views, err := manager.Views()
	if err != nil {
		fmt.Fprintf(c.App.Writer, "Capacity probe                     : FAILED (%v)\n", err)
		fmt.Fprintln(c.App.Writer, "Result: FAILED")
		return cli.Exit("", 1)
	}
	derived := 0
	for _, v := range views {
		if v.Remaining.Derived {
			derived++
		}
	}
	probe := fmt.Sprintf("%d card(s), driver-reported remaining", len(views))
	if derived > 0 {
		probe = fmt.Sprintf("%d card(s), %d with derived remaining (--show-remaining-info unavailable)", len(views), derived)
	}
	fmt.Fprintf(c.App.Writer, "Capacity probe                     : OK (%s)\n", probe)

	fmt.Fprintln(c.App.Writer, "Result: OK")
	return nil
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
