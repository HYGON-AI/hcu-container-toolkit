/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package root

import (
	"hcu-container-toolkit/pkg/c3000cdi/transform"
	"path/filepath"
	"strings"
)

// transformer transforms roots of paths.
type transformer struct {
	root       string
	targetRoot string
}

// New creates a root transformer using the specified options.
func New(opts ...Option) transform.Transformer {
	b := &builder{}
	for _, opt := range opts {
		opt(b)
	}
	return b.build()
}

func (t transformer) transformPath(path string) string {
	if !strings.HasPrefix(path, t.root) {
		return path
	}

	return filepath.Join(t.targetRoot, strings.TrimPrefix(path, t.root))
}
