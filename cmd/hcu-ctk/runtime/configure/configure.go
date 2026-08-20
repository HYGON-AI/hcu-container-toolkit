/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package configure

import (
	"encoding/json"
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/pkg/config/engine"
	"hcu-container-toolkit/pkg/config/engine/containerd"
	"hcu-container-toolkit/pkg/config/engine/crio"
	"hcu-container-toolkit/pkg/config/engine/docker"
	"hcu-container-toolkit/pkg/config/engine/podman"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

const (
	defaultRuntime = "docker"

	// defaultHCURuntimeName is the default name to use in configs for the HCU Container Runtime
	defaultHCURuntimeName = "hcu"

	// defaultHCURuntimeExecutable is the default HCU Container Runtime executable file name
	defaultHCURuntimeExecutable      = "hcu-container-runtime"
	defaultHCURuntimeExpecutablePath = "/usr/bin/hcu-container-runtime"

	defaultPodmanConfigFilePath     = "/etc/containers/containers.conf"
	defaultContainerdConfigFilePath = "/etc/containerd/config.toml"
	defaultCrioConfigFilePath       = "/etc/crio/crio.conf"
	defaultDockerConfigFilePath     = "/etc/docker/daemon.json"
)

type command struct {
	logger logger.Interface
}

// NewCommand constructs a configure command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

// config defines the options that can be set for the CLI through config files,
// environment variables, or command line config
type config struct {
	dryRun         bool
	runtime        string
	configFilePath string
	mode           string

	runtimeConfigOverrideJSON string

	hcuRuntime struct {
		name         string
		path         string
		setAsDefault bool
		remove       bool
		rollback     bool
	}

	// cdi-specific options
	cdi struct {
		enabled bool
	}
}

func (m command) build() *cli.Command {
	// Create a config struct to hold the parsed environment variables or command line flags
	config := config{}

	// Create the 'configure' command
	configure := cli.Command{
		Name:  "configure",
		Usage: "Add a runtime to the specified container engine",
		Before: func(c *cli.Context) error {
			return m.validateFlags(c, &config)
		},
		Action: func(c *cli.Context) error {
			return m.configureWrapper(c, &config)
		},
	}

	configure.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "dry-run",
			Usage:       "update the runtime configuration as required but don't write changes to disk",
			Destination: &config.dryRun,
		},
		&cli.StringFlag{
			Name:        "runtime",
			Usage:       "the target runtime engine; one of [containerd, podman, docker]",
			Value:       defaultRuntime,
			Destination: &config.runtime,
		},
		&cli.StringFlag{
			Name:        "config",
			Usage:       "path to the config file for the target runtime",
			Destination: &config.configFilePath,
		},
		&cli.StringFlag{
			Name:        "config-mode",
			Usage:       "the config mode for runtimes that support multiple configuration mechanisms",
			Destination: &config.mode,
		},
		&cli.StringFlag{
			Name:        "hcu-runtime-name",
			Usage:       "specify the name of the HCU runtime that will be added",
			Value:       defaultHCURuntimeName,
			Destination: &config.hcuRuntime.name,
		},
		&cli.StringFlag{
			Name:        "hcu-runtime-path",
			Aliases:     []string{"runtime-path"},
			Usage:       "specify the path to the HCU runtime executable",
			Value:       defaultHCURuntimeExecutable,
			Destination: &config.hcuRuntime.path,
		},
		&cli.BoolFlag{
			Name:        "hcu-set-as-default",
			Aliases:     []string{"set-as-default"},
			Usage:       "set the HCU runtime as the default runtime",
			Destination: &config.hcuRuntime.setAsDefault,
		},
		&cli.BoolFlag{
			Name:        "cdi.enabled",
			Aliases:     []string{"cdi.enable"},
			Usage:       "Enable CDI in the configured runtime",
			Destination: &config.cdi.enabled,
		},
		&cli.StringFlag{
			Name:        "runtime-config-override",
			Destination: &config.runtimeConfigOverrideJSON,
			Usage:       "specify additional runtime options as a JSON string. The paths are relative to the runtime config.",
			Value:       "{}",
			EnvVars:     []string{"RUNTIME_CONFIG_OVERRIDE"},
		},
		&cli.BoolFlag{
			Name:        "remove",
			Usage:       "remove the runtime configuration",
			Destination: &config.hcuRuntime.remove,
		},
		&cli.BoolFlag{
			Name:        "rollback",
			Usage:       "rollback the runtime configuration",
			Destination: &config.hcuRuntime.rollback,
		},
	}
	return &configure
}

func (m command) validateFlags(c *cli.Context, config *config) error {
	if config.mode != "" && config.mode != "config-file" {
		m.logger.Warningf("Ignoring unsupported config mode for %v: %q", config.runtime, config.mode)
	}
	config.mode = "config-file"

	switch config.runtime {
	case "containerd", "crio", "docker", "podman":
		break
	default:
		return fmt.Errorf("unrecognized runtime '%v'", config.runtime)
	}

	switch config.runtime {
	case "containerd", "crio", "podman":
		if config.hcuRuntime.path == defaultHCURuntimeExecutable {
			config.hcuRuntime.path = defaultHCURuntimeExpecutablePath
		}
		if !filepath.IsAbs(config.hcuRuntime.path) {
			return fmt.Errorf("the HCU runtime path %q is not an absolute path", config.hcuRuntime.path)
		}
	}

	if config.runtime != "containerd" && config.runtime != "docker" {
		if config.cdi.enabled {
			m.logger.Warningf("Ignoring cdi.enabled flag for %v", config.runtime)
		}
		config.cdi.enabled = false
	}

	if config.runtimeConfigOverrideJSON != "" && config.runtime != "containerd" {
		m.logger.Warningf("Ignoring runtime-config-override flag for %v", config.runtime)
		config.runtimeConfigOverrideJSON = ""
	}

	return nil
}

// configureWrapper updates the specified container engine config to enable the HCU runtime
func (m command) configureWrapper(c *cli.Context, config *config) error {
	switch config.mode {
	case "config-file":
		return m.configureConfigFile(c, config)
	}
	return fmt.Errorf("unsupported config-mode: %v", config.mode)
}

// configureConfigFile updates the specified container engine config file to enable the HCU runtime.
func (m command) configureConfigFile(c *cli.Context, config *config) error {
	outputPath := config.getOuputConfigPath()

	var backupPath string
	// Backup original config only once; do not overwrite existing .hcu-ctk.bak
	backupPath = outputPath + ".hcu-ctk.bak"

	if config.hcuRuntime.rollback {
		st, err := os.Stat(backupPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("the backup file %s does not exist", backupPath)
			}
		}
		if st.IsDir() {
			return fmt.Errorf("the backup file %s is a directory", backupPath)
		}
		_ = os.Remove(outputPath)
		err = restoreBackupIfExists(backupPath, outputPath)
		if err != nil {
			return fmt.Errorf("could not restore the backup file %s: %v", backupPath, err)
		}
		_ = os.Remove(backupPath)
		m.logger.Infof("Restored the backup file %v", backupPath)
		return nil
	}

	if err := backupFileIfExists(outputPath, backupPath); err != nil {
		return fmt.Errorf("unable to backup %s to %s: %w", outputPath, backupPath, err)
	}

	configFilePath := config.resolveConfigFilePath()

	var cfg engine.Interface
	var err error
	switch config.runtime {
	case "containerd":
		cfg, err = containerd.New(
			containerd.WithLogger(m.logger),
			containerd.WithPath(configFilePath),
		)
	case "crio":
		cfg, err = crio.New(
			crio.WithLogger(m.logger),
			crio.WithPath(configFilePath),
		)
	case "docker":
		cfg, err = docker.New(
			docker.WithLogger(m.logger),
			docker.WithPath(configFilePath),
		)
	case "podman":
		cfg, err = podman.New(
			podman.WithLogger(m.logger),
			podman.WithPath(configFilePath),
		)
	default:
		err = fmt.Errorf("unrecognized runtime '%v'", config.runtime)
	}
	if err != nil || cfg == nil {
		return fmt.Errorf("unable to load config for runtime %v: %v", config.runtime, err)
	}

	if config.hcuRuntime.remove {
		_ = cfg.RemoveRuntime(config.hcuRuntime.name)
	} else {
		runtimeConfigOverride, err := config.runtimeConfigOverride()
		if err != nil {
			return fmt.Errorf("unable to parse config overrides: %w", err)
		}

		err = cfg.AddRuntime(
			config.hcuRuntime.name,
			config.hcuRuntime.path,
			config.hcuRuntime.setAsDefault,
			runtimeConfigOverride,
		)
		if err != nil {
			return fmt.Errorf("unable to update config: %v", err)
		}

		err = enableCDI(config, cfg)
		if err != nil {
			return fmt.Errorf("failed to enable CDI in %s: %w", config.runtime, err)
		}
	}

	n, err := cfg.Save(outputPath)
	if err != nil {
		if _, stErr := os.Stat(backupPath); stErr == nil {
			if rmErr := os.Remove(outputPath); rmErr != nil && !os.IsNotExist(rmErr) {
				return fmt.Errorf("save failed (%v) and unable to remove %s before restore: %w", err, outputPath, rmErr)
			}
			_ = restoreBackupIfExists(backupPath, outputPath)
		}
		return fmt.Errorf("unable to flush config: %v", err)
	}

	if outputPath != "" {
		if n == 0 {
			m.logger.Infof("Removed empty config from %v", outputPath)
		} else {
			m.logger.Infof("Wrote updated config to %v", outputPath)
		}
		if config.runtime != "podman" {
			m.logger.Infof("It is recommended that %v daemon be restarted.", config.runtime)
		}
	}

	return nil
}

// resolveConfigFilePath returns the default config file path for the configured container engine
func (c *config) resolveConfigFilePath() string {
	if c.configFilePath != "" {
		return c.configFilePath
	}
	switch c.runtime {
	case "containerd":
		return defaultContainerdConfigFilePath
	case "crio":
		return defaultCrioConfigFilePath
	case "docker":
		return defaultDockerConfigFilePath
	case "podman":
		return defaultPodmanConfigFilePath
	}
	return ""
}

// getOuputConfigPath returns the configured config path or "" if dry-run is enabled
func (c *config) getOuputConfigPath() string {
	if c.dryRun {
		return ""
	}
	return c.resolveConfigFilePath()
}

// runtimeConfigOverride converts the specified runtimeConfigOverride JSON string to a map.
func (o *config) runtimeConfigOverride() (map[string]interface{}, error) {
	if o.runtimeConfigOverrideJSON == "" {
		return nil, nil
	}

	runtimeOptions := make(map[string]interface{})
	if err := json.Unmarshal([]byte(o.runtimeConfigOverrideJSON), &runtimeOptions); err != nil {
		return nil, fmt.Errorf("failed to read %v as JSON: %w", o.runtimeConfigOverrideJSON, err)
	}

	return runtimeOptions, nil
}

// enableCDI enables the use of CDI in the corresponding container engine
func enableCDI(config *config, cfg engine.Interface) error {
	if !config.cdi.enabled {
		return nil
	}
	switch config.runtime {
	case "containerd":
		cfg.Set("enable_cdi", true)
	case "docker":
		cfg.Set("features", map[string]bool{"cdi": true})
	default:
		return fmt.Errorf("enabling CDI in %s is not supported", config.runtime)
	}
	return nil
}
