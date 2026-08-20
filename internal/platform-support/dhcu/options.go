/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package dhcu

import (
	"hcu-container-toolkit/internal/c3000caps"
	"hcu-container-toolkit/internal/logger"
)

type options struct {
	logger         logger.Interface
	devRoot        string
	hcuCDIHookPath string

	isMigDevice bool
	// migCaps stores the MIG capabilities for the system.
	// If MIG is not available, this is nil.
	migCaps      c3000caps.MigCaps
	migCapsError error
}

type Option func(*options)

// WithDevRoot sets the root where /dev is located.
func WithDevRoot(root string) Option {
	return func(l *options) {
		l.devRoot = root
	}
}

// WithLogger sets the logger for the library
func WithLogger(logger logger.Interface) Option {
	return func(l *options) {
		l.logger = logger
	}
}

// WithHCUCDIHookPath sets the path to the HCU Container Toolkit CLI path for the library
func WithHCUCDIHookPath(path string) Option {
	return func(l *options) {
		l.hcuCDIHookPath = path
	}
}

// WithMIGCaps sets the MIG capabilities.
func WithMIGCaps(migCaps c3000caps.MigCaps) Option {
	return func(l *options) {
		l.migCaps = migCaps
	}
}
