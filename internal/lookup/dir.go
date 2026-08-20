/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package lookup

import (
	"fmt"
	"os"
)

// NewDirectoryLocator creates a Locator that can be used to find directories at the specified root.
func NewDirectoryLocator(opts ...Option) Locator {
	return NewFileLocator(
		append(
			opts,
			WithFilter(assertDirectory),
		)...,
	)
}

// assertDirectory checks wither the specified path is a directory.
func assertDirectory(filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("error getting info for %v: %v", filename, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("specified path '%v' is not a directory", filename)
	}

	return nil
}
