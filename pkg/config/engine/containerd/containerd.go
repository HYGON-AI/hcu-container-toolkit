/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package containerd

import (
	"github.com/pelletier/go-toml"

	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/pkg/config/engine"
)

// Config represents the containerd config
type Config struct {
	*toml.Tree
	RuntimeType           string
	UseDefaultRuntimeName bool
	ContainerAnnotations  []string
}

var _ engine.Interface = (*Config)(nil)

// New creates a containerd config with the specified options
func New(opts ...Option) (engine.Interface, error) {
	b := &builder{}
	for _, opt := range opts {
		opt(b)
	}

	if b.logger == nil {
		b.logger = logger.New()
	}

	return b.build()
}
