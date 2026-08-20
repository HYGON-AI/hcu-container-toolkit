/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package main

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func isWritable(path string) bool {
	return syscall.Access(path, syscall.O_RDWR) == nil
}

func fixVolumeArg(arg string) string {
	parts := strings.Split(arg, ":")
	if len(parts) < 2 {
		return arg
	}
	hostPath := parts[0]

	writable := isWritable(hostPath)

	hasMode := len(parts) == 3

	if !writable {
		if hasMode {
			parts[len(parts)-1] = "ro"
			return strings.Join(parts, ":")
		}
		return arg + ":ro"
	}

	return arg
}
func HasRunSubcommand(args []string) bool {
	isrun := false
	for _, arg := range args {
		if arg == "run" {
			isrun = true
			return isrun
		}
	}
	return isrun

}
func main() {
	runtime := os.Args[0]
	args := os.Args[1:]
	newArgs := []string{}
	if !HasRunSubcommand(args) {
		newArgs = args

	} else {
		for i := 0; i < len(args); i++ {
			if args[i] == "-v" && i+1 < len(args) {
				fixed := fixVolumeArg(args[i+1])
				newArgs = append(newArgs, "-v", fixed)
				i++
			} else {
				newArgs = append(newArgs, args[i])
			}
		}
	}

	cmd := exec.Command("s"+runtime, newArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
