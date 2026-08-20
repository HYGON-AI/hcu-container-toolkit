/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package containerd

import (
	"fmt"
	"hcu-container-toolkit/pkg/config/engine"

	"github.com/pelletier/go-toml"
)

// ConfigV3 represents a version 3 containerd config.
type ConfigV3 Config

var _ engine.Interface = (*ConfigV3)(nil)

var configV3RuntimePath = []string{
	"plugins",
	"io.containerd.cri.v1.runtime",
	"containerd",
}

// AddRuntime adds a runtime to the containerd config.
func (c *ConfigV3) AddRuntime(name string, path string, setAsDefault bool, configOverrides ...map[string]interface{}) error {
	if c == nil || c.Tree == nil {
		return fmt.Errorf("config is nil")
	}

	config := *c.Tree
	config.Set("version", int64(3))

	runtimePath := append(append([]string{}, configV3RuntimePath...), "runtimes", name)
	runcPath := append(append([]string{}, configV3RuntimePath...), "runtimes", "runc")
	if runc, ok := config.GetPath(runcPath).(*toml.Tree); ok {
		runc, _ = toml.Load(runc.String())
		config.SetPath(runtimePath, runc)
	}

	if config.GetPath(runtimePath) == nil {
		config.SetPath(append(runtimePath, "runtime_type"), c.RuntimeType)
		config.SetPath(append(runtimePath, "runtime_root"), "")
		config.SetPath(append(runtimePath, "runtime_engine"), "")
		config.SetPath(append(runtimePath, "privileged_without_host_devices"), false)
	}

	if len(c.ContainerAnnotations) > 0 {
		annotationsPath := append(runtimePath, "container_annotations")
		annotations, err := (*Config)(c).getRuntimeAnnotations(annotationsPath)
		if err != nil {
			return err
		}
		annotations = append(c.ContainerAnnotations, annotations...)
		config.SetPath(annotationsPath, annotations)
	}

	config.SetPath(append(runtimePath, "options", "BinaryName"), path)

	defaultRuntimeNamePath := append(append([]string{}, configV3RuntimePath...), "default_runtime_name")
	if setAsDefault {
		config.SetPath(defaultRuntimeNamePath, name)
	} else if defaultRuntime, ok := config.GetPath(defaultRuntimeNamePath).(string); ok && defaultRuntime == name {
		config.DeletePath(defaultRuntimeNamePath)
	}

	runtimeSubtree := subtreeAtPath(config, runtimePath...)
	if err := runtimeSubtree.applyOverrides(configOverrides...); err != nil {
		return fmt.Errorf("failed to apply config overrides: %w", err)
	}

	*c.Tree = config
	return nil
}

// Set sets the specified containerd option.
func (c *ConfigV3) Set(key string, value interface{}) {
	config := *c.Tree
	path := append(append([]string{}, configV3RuntimePath[:2]...), key)
	config.SetPath(path, value)
	*c.Tree = config
}

// DefaultRuntime returns the default runtime for the containerd config.
func (c ConfigV3) DefaultRuntime() string {
	path := append(append([]string{}, configV3RuntimePath...), "default_runtime_name")
	if runtime, ok := c.GetPath(path).(string); ok {
		return runtime
	}
	return ""
}

// RemoveRuntime removes a runtime from the containerd config.
func (c *ConfigV3) RemoveRuntime(name string) error {
	if c == nil || c.Tree == nil {
		return nil
	}

	config := *c.Tree
	runtimePath := append(append([]string{}, configV3RuntimePath...), "runtimes", name)
	config.DeletePath(runtimePath)

	defaultRuntimeNamePath := append(append([]string{}, configV3RuntimePath...), "default_runtime_name")
	if runtime, ok := config.GetPath(defaultRuntimeNamePath).(string); ok && runtime == name {
		config.DeletePath(defaultRuntimeNamePath)
	}

	for i := 0; i < len(runtimePath); i++ {
		if runtimes, ok := config.GetPath(runtimePath[:len(runtimePath)-i]).(*toml.Tree); ok {
			if len(runtimes.Keys()) == 0 {
				config.DeletePath(runtimePath[:len(runtimePath)-i])
			}
		}
	}

	if len(config.Keys()) == 1 && config.Keys()[0] == "version" {
		config.Delete("version")
	}

	*c.Tree = config
	return nil
}

// Save writes the config to the specified path.
func (c ConfigV3) Save(path string) (int64, error) {
	return (Config)(c).Save(path)
}
