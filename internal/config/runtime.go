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
	// VHCU configures the dynamic vHCU split-and-mount capability.
	// It is a cross-mode runtime capability and therefore sits directly
	// under [hcu-container-runtime] rather than under [hcu-container-runtime.modes].
	VHCU vhcuModeConfig `toml:"vhcu"`
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

// vhcuModeConfig stores the config options for the dynamic vHCU split
// (create-and-mount) capability.
type vhcuModeConfig struct {
	// Enabled toggles the dynamic vHCU capability. When false, HCU_VHCU_*
	// requests fail with an explicit error instead of being silently ignored.
	Enabled bool `toml:"enabled"`
	// MaxVHCUPerCard is the per-card vdev slot limit enforced by the driver
	// (a single physical device supports at most 4 vdevs by default).
	MaxVHCUPerCard int `toml:"max-vhcu-per-card"`
	// StateDir stores per-container ownership records for dynamic vHCU.
	StateDir string `toml:"state-dir"`
	// LockFile is the global flock file serializing dynamic splits.
	LockFile string `toml:"lock-file"`
	// LockTimeout bounds how long we wait on the global lock.
	LockTimeout string `toml:"lock-timeout"`
	// ConfDir must be /etc/vdev, matching the driver and mount path.
	ConfDir string `toml:"conf-dir"`
	// HysmiPath overrides the hy-smi binary path; empty means auto-detect.
	HysmiPath string `toml:"hysmi-path"`
	// DestroyRetries is the retry count for vdev destroy operations.
	DestroyRetries int `toml:"destroy-retries"`
}

// GetDefaultRuntimeConfig defines the default values for the config
func GetDefaultRuntimeConfig() (*RuntimeConfig, error) {
	cfg, err := GetDefault()
	if err != nil {
		return nil, err
	}

	return &cfg.HCUContainerRuntimeConfig, nil
}
