/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package rootless

import (
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"os"
	"os/exec"
	"os/user"

	"github.com/urfave/cli/v2"
)

type rootlessCommand struct {
	logger logger.Interface
}
type options struct {
	runtime string
}

// NewCommand constructs a rootless command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := rootlessCommand{
		logger: logger,
	}
	return c.build()
}

// build
func (m rootlessCommand) build() *cli.Command {
	// Create the 'rootless' command
	opts := options{}

	c := cli.Command{
		Name:   "rootless",
		Usage:  "Provide tools for configuring rootless Docker container runtimes",
		Before: func(ctx *cli.Context) error { return validateGenOptions(ctx) },
		Action: func(ctx *cli.Context) error { return run(ctx, &opts) },
	}

	c.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "runtime",
			Aliases:     []string{"r"},
			Usage:       "docker",
			Value:       "docker",
			Destination: &opts.runtime,
		},
	}

	return &c
}

func validateGenOptions(c *cli.Context) error {
	curUser, err := user.Current()
	if err != nil || curUser.Uid != "0" {
		return fmt.Errorf("Permission denied: Not running as root")
	}

	return nil
}

func run(c *cli.Context, opts *options) error {
	runtime := opts.runtime
	if runtime != "docker" {
		return fmt.Errorf("invalid runtime %s", runtime)
	}

	_, err := os.Stat("/usr/bin/s" + runtime)

	if os.IsNotExist(err) {
		cmd := exec.Command("cp", "/usr/bin/"+runtime, "/usr/bin/s"+runtime)
		cmd.Run()
	}
	_, err = os.Stat("/usr/bin/" + runtime)

	if err == nil {
		os.Remove("/usr/bin/" + runtime)
	}

	cmd := exec.Command("cp", "/usr/bin/hcu-docker", "/usr/bin/"+runtime)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to copy hcu-docker to /usr/bin/%s: %w", runtime, err)
	}
	return nil
}
