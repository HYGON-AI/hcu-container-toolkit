/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package config

// RuntimeConfig stores the config options for the HCU Container Runtime
type RuntimeConfig struct {
	DebugFilePath string `toml:"debug"`
	// LogLevel defines the logging level for the application
	LogLevel string `toml:"log-level"`
	// Runtimes defines the candidates for the low-level runtime
	Runtimes []string    `toml:"runtimes"`
	Mode     string      `toml:"mode"`
	Modes    modesConfig `toml:"modes"`
}

// modesConfig defines (optional) per-mode configs
type modesConfig struct {
	CSV csvModeConfig `toml:"csv"`
	CDI cdiModeConfig `toml:"cdi"`
}

type cdiModeConfig struct {
	// SpecDirs allows for the default spec dirs for CDI to be overridden
	SpecDirs []string `toml:"spec-dirs"`
	// DefaultKind sets the default kind to be used when constructing fully-qualified CDI device names
	DefaultKind string `toml:"default-kind"`
	// AnnotationPrefixes sets the allowed prefixes for CDI annotation-based device injection
	AnnotationPrefixes []string `toml:"annotation-prefixes"`
}

type csvModeConfig struct {
	MountSpecPath string `toml:"mount-spec-path"`
}

// GetDefaultRuntimeConfig defines the default values for the config
func GetDefaultRuntimeConfig() (*RuntimeConfig, error) {
	cfg, err := GetDefault()
	if err != nil {
		return nil, err
	}

	return &cfg.HCUContainerRuntimeConfig, nil
}
