/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package main

import (
	"hcu-container-toolkit/cmd/hcu-ctk/cdi"
	"hcu-container-toolkit/cmd/hcu-ctk/config"
	"hcu-container-toolkit/cmd/hcu-ctk/container"
	"hcu-container-toolkit/cmd/hcu-ctk/hcu-tracker"
	"hcu-container-toolkit/cmd/hcu-ctk/hook"
	"hcu-container-toolkit/cmd/hcu-ctk/rootless"
	"hcu-container-toolkit/cmd/hcu-ctk/runtime"
	"hcu-container-toolkit/internal/info"
	"os"

	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

// options defines the options that can be set for the CLI through config files,
// environment variables, or command line flags
type options struct {
	// Debug indicates whether the CLI is started in "debug" mode
	Debug bool
	// Quiet indicates whether the CLI is started in "quiet" mode
	Quiet bool
}

func main() {
	logger := logrus.New()

	// Create a options struct to hold the parsed environment variables or command line flags
	opts := options{}

	// Create the top-level CLI
	c := cli.NewApp()
	c.Name = "Hygon HCU Container Toolkit CLI"
	c.UseShortOptionHandling = true
	c.EnableBashCompletion = true
	c.Usage = "Tools to configure the Hygon HCU Container Toolkit"
	c.Version = info.GetVersionString()

	// Setup the flags for this command
	c.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "debug",
			Aliases:     []string{"d"},
			Usage:       "Enable debug-level logging",
			Destination: &opts.Debug,
			EnvVars:     []string{"HCU_CTK_DEBUG"},
		},
		&cli.BoolFlag{
			Name:        "quiet",
			Usage:       "Suppress all output except for errors; overrides --debug",
			Destination: &opts.Quiet,
			EnvVars:     []string{"HCU_CTK_QUIET"},
		},
	}

	// Set log-level for all subcommands
	c.Before = func(c *cli.Context) error {
		logLevel := logrus.InfoLevel
		if opts.Debug {
			logLevel = logrus.DebugLevel
		}
		if opts.Quiet {
			logLevel = logrus.ErrorLevel
		}
		logger.SetLevel(logLevel)
		return nil
	}

	// Define the subcommands
	c.Commands = []*cli.Command{
		rootless.NewCommand(logger),
		hcuTracker.NewCommand(logger),
		runtime.NewCommand(logger),
		config.NewCommand(logger),
		container.NewCommand(logger),
		hook.NewCommand(logger),
		cdi.NewCommand(logger),
	}

	// Run the CLI
	err := c.Run(os.Args)
	if err != nil {
		logger.Errorf("%v", err)
		os.Exit(1)
	}
}
