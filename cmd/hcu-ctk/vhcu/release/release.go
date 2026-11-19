/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package release

import (
	"fmt"
	"os"
	"time"

	"hcu-container-toolkit/cmd/hcu-ctk/vhcu/common"
	"hcu-container-toolkit/internal/logger"

	"github.com/urfave/cli/v2"
)

// releaseLogPath records every poststop release attempt. The hook runs inside
// the OCI runtime's delete path, where its output only reaches the container
// runtime's log; without a dedicated log it is impossible to tell "the hook
// never ran" from "the hook ran and failed" when a vdev is left behind.
var releaseLogPath = "/var/log/hcu-container-toolkit-vhcu.log"

type command struct {
	logger logger.Interface
}

// NewCommand constructs the `hcu-ctk vhcu release` command. It is hidden:
// it is invoked by the poststop hook injected at container create time.
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

func (m command) build() *cli.Command {
	vhcuReleaseCmd := cli.Command{
		Name:   "release",
		Hidden: true,
		Usage:  "Release the dynamic vHCU allocated to a container",
		UsageText: `hcu-ctk vhcu release <container_id> <allocation_id>

	Arguments:
		container_id   container ID of the container
        allocation_id  ownership token recorded by the toolkit

Examples:
	hcu-ctk vhcu release <container_id> <allocation_id>`,
		Before: func(c *cli.Context) error {
			return validateGenOptions(c)
		},
		Action: func(c *cli.Context) error {
			return performAction(c, m.logger)
		},
	}

	return &vhcuReleaseCmd
}

func validateGenOptions(c *cli.Context) error {
	return common.ValidateRoot()
}

func performAction(c *cli.Context, logger logger.Interface) error {
	if c.Args().Len() != 2 || c.Args().Get(1) == "" {
		return cli.Exit("Usage: hcu-ctk vhcu release <container_id> <allocation_id>", 1)
	}

	manager, err := common.NewManager(logger)
	if err != nil {
		logRelease(c.Args().Get(0), c.Args().Get(1), err)
		return fmt.Errorf("Failed to create vhcu manager, Error: %v", err)
	}

	releaseErr := manager.ReleaseAllocation(c.Args().Get(0), c.Args().Get(1))
	logRelease(c.Args().Get(0), c.Args().Get(1), releaseErr)
	if releaseErr != nil {
		// The poststop hook must not alter the container exit code; the
		// error is reported and the GC sweep retries later.
		return fmt.Errorf("Failed to release vHCU, Error: %v", releaseErr)
	}

	return nil
}

// logRelease appends one line per attempt. Best-effort: a logging failure must
// never affect the release result.
func logRelease(containerID, allocationID string, releaseErr error) {
	f, err := os.OpenFile(releaseLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()

	status := "ok"
	if releaseErr != nil {
		status = "failed: " + releaseErr.Error()
	}
	fmt.Fprintf(f, "%s container=%s allocation=%s %s\n",
		time.Now().Format(time.RFC3339), containerID, allocationID, status)
}
