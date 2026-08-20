/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package podman

import (
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"os"

	"github.com/pelletier/go-toml"
)

type builder struct {
	logger logger.Interface
	path   string
}

type Option func(*builder)

func WithLogger(logger logger.Interface) Option {
	return func(b *builder) {
		b.logger = logger
	}
}

func WithPath(path string) Option {
	return func(b *builder) {
		b.path = path
	}
}

func (b *builder) build() (*Config, error) {
	if b.path == "" {
		return nil, fmt.Errorf("config path is empty")
	}
	return b.loadConfig(b.path)
}

func (b *builder) loadConfig(config string) (*Config, error) {
	info, err := os.Stat(config)
	if os.IsExist(err) && info.IsDir() {
		return nil, fmt.Errorf("config file is a directory")
	}

	if os.IsNotExist(err) {
		b.logger.Infof("Config file does not exist; using empty config")
		config = "/dev/null"
	} else {
		b.logger.Infof("Loading config from %v", config)
	}

	tomlConfig, err := toml.LoadFile(config)
	if err != nil {
		return nil, err
	}

	cfg := Config{
		Tree: tomlConfig,
	}
	return &cfg, nil
}
