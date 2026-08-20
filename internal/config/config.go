/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package config

import (
	"bufio"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/lookup"
	"os"
	"path/filepath"
	"strings"

	"tags.cncf.io/container-device-interface/pkg/cdi"
)

const (
	configOverride = "XDG_CONFIG_HOME"
	configFilePath = "hcu-container-runtime/config.toml"

	hcuCTKExecutable          = "hcu-ctk"
	hcuCTKDefaultFilePath     = "/usr/bin/hcu-ctk"
	hcuCDIHookDefaultFilePath = "/usr/bin/hcu-cdi-hook"

	// hcuContainerRuntimeHookExecutable  = "hcu-container-runtime-hook"
	// hcuContainerRuntimeHookDefaultPath = "/usr/bin/hcu-container-runtime-hook"
)

var (
	// DefaultExecutableDir specifies the default path to use for executables if they cannot be located in the path.
	DefaultExecutableDir = "/usr/bin"

	// HCUContainerRuntimeHookExecutable is the executable name for the HCU Container Runtime Hook
	HCUContainerRuntimeHookExecutable = "hcu-container-runtime-hook"
	// HCUContainerToolkitExecutable is the executable name for the HCU Container Toolkit (an alias for the HCU Container Runtime Hook)
	HCUContainerToolkitExecutable = "hcu-container-toolkit"
)

// Config represents the contents of the config.toml file for the HCU Container Toolkit
type Config struct {
	DisableRequire                 bool   `toml:"disable-require"`
	SwarmResource                  string `toml:"swarm-resource"`
	AcceptEnvvarUnprivileged       bool   `toml:"accept-hcu-visible-devices-envvar-when-unprivileged"`
	AcceptDeviceListAsVolumeMounts bool   `toml:"accept-hcu-visible-devices-as-volume-mounts"`
	// SupportedDriverCapabilities    string `toml:"supported-driver-capabilities"`

	HCUCTKConfig              CTKConfig     `toml:"hcu-ctk"`
	HCUContainerRuntimeConfig RuntimeConfig `toml:"hcu-container-runtime"`
}

// GetConfigFilePath returns the path to the config file for the configured system
func GetConfigFilePath() string {
	if XDGConfigDir := os.Getenv(configOverride); len(XDGConfigDir) != 0 {
		return filepath.Join(XDGConfigDir, configFilePath)
	}

	return filepath.Join("/etc", configFilePath)
}

// GetConfig sets up the config struct. Values are read from a toml file
// or set via the environment.
func GetConfig() (*Config, error) {
	cfg, err := New(
		WithConfigFile(GetConfigFilePath()),
	)
	if err != nil {
		return nil, err
	}

	return cfg.Config()
}

// GetDefault defines the default values for the config
func GetDefault() (*Config, error) {
	d := Config{
		AcceptEnvvarUnprivileged: true,
		HCUCTKConfig: CTKConfig{
			Path: hcuCTKExecutable,
		},
		HCUContainerRuntimeConfig: RuntimeConfig{
			DebugFilePath: "/dev/null",
			LogLevel:      "info",
			Runtimes:      []string{"docker-runc", "runc", "crun"},
			Mode:          "auto",
			Modes: modesConfig{
				CSV: csvModeConfig{
					MountSpecPath: "/etc/hcu-container-runtime/host-files-for-container.d",
				},
				CDI: cdiModeConfig{
					DefaultKind:        "hygon.com/hcu",
					AnnotationPrefixes: []string{cdi.AnnotationPrefix},
					SpecDirs:           cdi.DefaultSpecDirs,
				},
			},
		},
	}
	return &d, nil
}

func GetUserGroup() string {
	if isSuse() {
		return "root:video"
	}
	return ""
}

// isSuse returns whether a SUSE-based distribution was detected.
func isSuse() bool {
	suseDists := map[string]bool{
		"suse":     true,
		"opensuse": true,
	}

	idsLike := getDistIDLike()
	for _, id := range idsLike {
		if suseDists[id] {
			return true
		}
	}
	return false
}

// getDistIDLike returns the ID_LIKE field from /etc/os-release.
// We can override this for testing.
var getDistIDLike = func() []string {
	releaseFile, err := os.Open("/etc/os-release")
	if err != nil {
		return nil
	}
	defer releaseFile.Close()

	scanner := bufio.NewScanner(releaseFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID_LIKE=") {
			value := strings.Trim(strings.TrimPrefix(line, "ID_LIKE="), "\"")
			return strings.Split(value, " ")
		}
	}
	return nil
}

// ResolveHCUCDIHookPath resolves the path to the hcu-cdi-hook binary.
// This executable is used in hooks and needs to be an absolute path.
// If the path is specified as an absolute path, it is used directly
// without checking for existence of an executable at that path.
func ResolveHCUCDIHookPath(logger logger.Interface, hcuCDIHookPath string) string {
	if filepath.Base(hcuCDIHookPath) == "hcu-ctk" {
		return resolveWithDefault(
			logger,
			"HCU Container Toolkit CLI",
			hcuCDIHookPath,
			hcuCTKDefaultFilePath,
		)
	}
	return resolveWithDefault(
		logger,
		"HCU CDI Hook CLI",
		hcuCDIHookPath,
		hcuCDIHookDefaultFilePath,
	)
}

// resolveWithDefault resolves the path to the specified binary.
// If an absolute path is specified, it is used directly without searching for the binary.
// If the binary cannot be found in the path, the specified default is used instead.
func resolveWithDefault(logger logger.Interface, label string, path string, defaultPath string) string {
	if filepath.IsAbs(path) {
		logger.Debugf("Using specified %v path %v", label, path)
		return path
	}

	if path == "" {
		path = filepath.Base(defaultPath)
	}
	logger.Debugf("Locating %v as %v", label, path)
	lookup := lookup.NewExecutableLocator(logger, "")

	resolvedPath := defaultPath
	targets, err := lookup.Locate(path)
	if err != nil {
		logger.Warningf("Failed to locate %v: %v", path, err)
	} else {
		logger.Debugf("Found %v candidates: %v", path, targets)
		resolvedPath = targets[0]
	}
	logger.Debugf("Using %v path %v", label, resolvedPath)

	return resolvedPath
}
