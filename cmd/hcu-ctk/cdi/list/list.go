/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package list

import (
	"errors"
	"fmt"
	"hcu-container-toolkit/internal/logger"

	"github.com/urfave/cli/v2"
	"tags.cncf.io/container-device-interface/pkg/cdi"
)

type command struct {
	logger logger.Interface
}

type config struct {
	cdiSpecDirs cli.StringSlice
}

// NewCommand constructs a cdi list command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

// build creates the CLI command
func (m command) build() *cli.Command {
	cfg := config{}

	// Create the command
	c := cli.Command{
		Name:  "list",
		Usage: "List the available CDI devices",
		Before: func(c *cli.Context) error {
			return m.validateFlags(c, &cfg)
		},
		Action: func(c *cli.Context) error {
			return m.run(c, &cfg)
		},
	}

	c.Flags = []cli.Flag{
		&cli.StringSliceFlag{
			Name:        "spec-dir",
			Usage:       "specify the directories to scan for CDI specifications",
			Value:       cli.NewStringSlice(cdi.DefaultSpecDirs...),
			Destination: &cfg.cdiSpecDirs,
		},
	}

	return &c
}

func (m command) validateFlags(c *cli.Context, cfg *config) error {
	if len(cfg.cdiSpecDirs.Value()) == 0 {
		return errors.New("at least one CDI specification directory must be specified")
	}
	return nil
}

func (m command) run(c *cli.Context, cfg *config) error {
	registry, err := cdi.NewCache(
		cdi.WithAutoRefresh(false),
		cdi.WithSpecDirs(cfg.cdiSpecDirs.Value()...),
	)
	if err != nil {
		return fmt.Errorf("failed to create CDI cache: %v", err)
	}

	_ = registry.Refresh()
	if errors := registry.GetErrors(); len(errors) > 0 {
		m.logger.Warningf("The following registry errors were reported:")
		for k, err := range errors {
			m.logger.Warningf("%v: %v", k, err)
		}
	}

	devices := registry.ListDevices()
	m.logger.Infof("Found %d CDI devices", len(devices)/2)
	for _, device := range devices {
		fmt.Printf("%s\n", device)
	}

	return nil
}
