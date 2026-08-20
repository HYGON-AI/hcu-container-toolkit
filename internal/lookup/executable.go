/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package lookup

import (
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"os"
	"strings"
)

type executable struct {
	file
}

// NewExecutableLocator creates a locator to fine executable files in the path. A logger can also be specified.
func NewExecutableLocator(logger logger.Interface, root string, paths ...string) Locator {
	paths = append(paths, GetPaths(root)...)

	return newExecutableLocator(logger, root, paths...)
}

func newExecutableLocator(logger logger.Interface, root string, paths ...string) *executable {
	f := newFileLocator(
		WithLogger(logger),
		WithRoot(root),
		WithSearchPaths(paths...),
		WithFilter(assertExecutable),
		WithCount(1),
	)

	l := executable{
		file: *f,
	}

	return &l
}

var _ Locator = (*executable)(nil)

// Locate finds executable files with the specified pattern in the path.
// If a relative or absolute path is specified, the prefix paths are not considered.
func (p executable) Locate(pattern string) ([]string, error) {
	// For absolute paths we ensure that it is executable
	if strings.Contains(pattern, "/") {
		err := assertExecutable(pattern)
		if err != nil {
			return nil, fmt.Errorf("absolute path %v is not an executable file: %v", pattern, err)
		}
		return []string{pattern}, nil
	}

	return p.file.Locate(pattern)
}

// assertExecutable checks whether the specified path is an execuable file.
func assertExecutable(filename string) error {
	err := assertFile(filename)
	if err != nil {
		return err
	}
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("specified file '%v' is not executable", filename)
	}

	return nil
}
