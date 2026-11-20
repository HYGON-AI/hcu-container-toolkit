/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package oci

import (
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"os"
	"path/filepath"
	"strings"
)

// pathRuntime wraps the path that a binary and defines the semantics for how to exec into it.
// This can be used to wrap an OCI-compliant low-level runtime binary, allowing it to be used through the
// Runtime internface.
type pathRuntime struct {
	logger      logger.Interface
	path        string
	execRuntime Runtime
}

// StateCommand preserves the runtime's state root, which scopes container IDs
// for Docker, Podman and containerd without guessing their daemon namespaces.
func (s pathRuntime) StateCommand(args []string) ([]string, error) {
	command := []string{s.path}
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if arg == "create" {
			break
		}
		parts := strings.SplitN(arg, "=", 2)
		if parts[0] == "--root" || parts[0] == "-root" || parts[0] == "-r" {
			var value string
			if len(parts) == 2 {
				value = parts[1]
			} else {
				i++
				if i == len(args) {
					return nil, fmt.Errorf("runtime root requires a value")
				}
				value = args[i]
			}
			if value == "" {
				return nil, fmt.Errorf("runtime root cannot be empty")
			}
			root, err := filepath.Abs(value)
			if err != nil {
				return nil, err
			}
			command = append(command, "--root", root)
		} else if len(parts) == 1 {
			switch parts[0] {
			case "--log", "--log-format", "--criu", "--rootless":
				i++
			}
		}
	}
	return append(command, "list", "--format", "json"), nil
}

var _ Runtime = (*pathRuntime)(nil)

// NewRuntimeForPath creates a Runtime for the specified logger and path
func NewRuntimeForPath(logger logger.Interface, path string) (Runtime, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path '%v': %v", path, err)
	}
	if info.IsDir() || info.Mode()&0111 == 0 {
		return nil, fmt.Errorf("specified path '%v' is not an executable file", path)
	}

	shim := pathRuntime{
		logger:      logger,
		path:        path,
		execRuntime: syscallExec{},
	}

	return &shim, nil
}

// Exec exces into the binary at the path from the pathRuntime struct, passing it the supplied arguments
// after ensuring that the first argument is the path of the target binary.
func (s pathRuntime) Exec(args []string) error {
	runtimeArgs := []string{s.path}
	if len(args) > 1 {
		runtimeArgs = append(runtimeArgs, args[1:]...)
	}

	return s.execRuntime.Exec(runtimeArgs)
}
