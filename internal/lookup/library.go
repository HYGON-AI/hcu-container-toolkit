/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package lookup

import (
	"fmt"
	"hcu-container-toolkit/internal/ldcache"
	"hcu-container-toolkit/internal/logger"
)

type ldcacheLocator struct {
	logger logger.Interface
	cache  ldcache.LDCache
}

var _ Locator = (*ldcacheLocator)(nil)

// NewLibraryLocator creates a library locator using the specified options.
func NewLibraryLocator(opts ...Option) Locator {
	b := newBuilder(opts...)

	// If search paths are already specified, we return a locator for the specified search paths.
	if len(b.searchPaths) > 0 {
		return NewSymlinkLocator(
			WithLogger(b.logger),
			WithSearchPaths(b.searchPaths...),
			WithRoot("/"),
		)
	}

	opts = append(opts,
		WithSearchPaths([]string{
			"/",
			"/opt/hyhal/lib",
			"/usr/lib64",
			"/usr/lib/x86_64-linux-gnu",
			"/usr/lib/aarch64-linux-gnu",
			"/lib64",
			"/lib/x86_64-linux-gnu",
			"/lib/aarch64-linux-gnu",
		}...),
	)
	// We construct a symlink locator for expected library locations.
	symlinkLocator := NewSymlinkLocator(opts...)

	l := First(
		symlinkLocator,
		newLdcacheLocator(opts...),
	)
	return l
}

func newLdcacheLocator(opts ...Option) Locator {
	b := newBuilder(opts...)

	cache, err := ldcache.New(b.logger, b.root)
	if err != nil {
		// If we failed to open the LDCache, we default to a symlink locator.
		b.logger.Warningf("Failed to load ldcache: %v", err)
		return nil
	}

	return &ldcacheLocator{
		logger: b.logger,
		cache:  cache,
	}
}

// Locate finds the specified libraryname.
// If the input is a library name, the ldcache is searched otherwise the
// provided path is resolved as a symlink.
func (l ldcacheLocator) Locate(libname string) ([]string, error) {
	paths32, paths64 := l.cache.Lookup(libname)
	if len(paths32) > 0 {
		l.logger.Warningf("Ignoring 32-bit libraries for %v: %v", libname, paths32)
	}

	if len(paths64) == 0 {
		return nil, fmt.Errorf("64-bit library %v: %w", libname, ErrNotFound)
	}

	return paths64, nil
}
