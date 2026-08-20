/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package oci

import (
	"fmt"
	"os"
	"syscall"
)

type syscallExec struct{}

var _ Runtime = (*syscallExec)(nil)

func (r syscallExec) Exec(args []string) error {
	//nolint:gosec // TODO: Can we harden this so that there is less risk of command injection
	err := syscall.Exec(args[0], args, os.Environ())
	if err != nil {
		return fmt.Errorf("cound not exec '%v': %v", args[0], err)
	}

	// syscall.Exec is not expected to return. This is an error state regardless of whether
	// err is nil or not.
	return fmt.Errorf("unexpected return from exec '%v'", args[0])
}
