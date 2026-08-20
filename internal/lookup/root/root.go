/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package root

import (
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/lookup"
	"os"
	"path/filepath"
)

// Driver represents a filesystem in which a set of drivers or devices is defined.
type Driver struct {
	logger logger.Interface
	// Root represents the root from the perspective of the driver libraries and binaries.
	Root string
}

// New creates a new Driver root using the specified options.
func New(opts ...Option) *Driver {
	d := &Driver{}
	for _, opt := range opts {
		opt(d)
	}
	if d.logger == nil {
		d.logger = logger.New()
	}
	return d
}

// Files returns a Locator for arbitrary driver files.
func (r *Driver) Files(opts ...lookup.Option) lookup.Locator {
	return lookup.NewFileLocator(
		append(
			opts,
			lookup.WithLogger(r.logger),
			lookup.WithRoot(r.Root),
		)...,
	)
}

// Libraries returns a Locator for driver libraries.
func (r *Driver) Libraries() lookup.Locator {
	return lookup.NewLibraryLocator(
		lookup.WithLogger(r.logger),
		lookup.WithRoot(r.Root),
	)
}

// normalizeSearchPaths takes a list of paths and normalized these.
// Each of the elements in the list is expanded if it is a path list and the
// resultant list is returned.
// This allows, for example, for the contents of `PATH` or `LD_LIBRARY_PATH` to
// be passed as a search path directly.
func normalizeSearchPaths(paths ...string) []string {
	var normalized []string
	for _, path := range paths {
		normalized = append(normalized, filepath.SplitList(path)...)
	}
	return normalized
}

// xdgDataDirs finds the paths as specified in the environment variable XDG_DATA_DIRS.
// See https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html.
func xdgDataDirs() []string {
	if dirs, exists := os.LookupEnv("XDG_DATA_DIRS"); exists && dirs != "" {
		return normalizeSearchPaths(dirs)
	}

	return []string{"/usr/local/share", "/usr/share"}
}
