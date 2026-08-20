/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package defaultsubcommand

import (
	"fmt"
	"hcu-container-toolkit/cmd/hcu-ctk/config/flags"
	"hcu-container-toolkit/internal/config"
	"hcu-container-toolkit/internal/logger"

	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

// NewCommand constructs a default command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

// build creates the CLI command
func (m command) build() *cli.Command {
	opts := flags.Options{}

	// Create the 'default' command
	c := cli.Command{
		Name:    "default",
		Aliases: []string{"create-default", "generate-default"},
		Usage:   "Generate the default HCU Container Toolkit configuration file",
		Before: func(c *cli.Context) error {
			return m.validateFlags(c, &opts)
		},
		Action: func(c *cli.Context) error {
			return m.run(c, &opts)
		},
	}

	c.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "output",
			Aliases:     []string{"o"},
			Usage:       "Specify the output file to write to; If not specified, the output is written to stdout",
			Destination: &opts.Output,
		},
	}

	return &c
}

func (m command) validateFlags(c *cli.Context, opts *flags.Options) error {
	return nil
}

func (m command) run(c *cli.Context, opts *flags.Options) error {
	cfgToml, err := config.New()
	if err != nil {
		return fmt.Errorf("unable to load or create config: %v", err)
	}

	if err := opts.EnsureOutputFolder(); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}
	output, err := opts.CreateOutput()
	if err != nil {
		return fmt.Errorf("failed to open output file: %v", err)
	}
	defer output.Close()

	if _, err = cfgToml.Save(output); err != nil {
		return fmt.Errorf("failed to write output: %v", err)
	}

	return nil
}
