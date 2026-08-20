/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package commands

import (
	"hcu-container-toolkit/cmd/hcu-cdi-hook/chmod"
	symlinks "hcu-container-toolkit/cmd/hcu-cdi-hook/create-symlinks"
	"hcu-container-toolkit/internal/logger"

	"github.com/urfave/cli/v2"
)

// New creates the commands associated with supported CDI hooks.
// These are shared by the hcu-cdi-hook and hcu-ctk hook commands.
func New(logger logger.Interface) []*cli.Command {
	return []*cli.Command{
		symlinks.NewCommand(logger),
		chmod.NewCommand(logger),
	}
}
