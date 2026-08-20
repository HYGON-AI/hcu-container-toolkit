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

// AddRuntime adds a runtime to the containerd config
func (c *Config) AddRuntime(name string, path string, setAsDefault bool, configOverrides ...map[string]interface{}) error {
	if c == nil || c.Tree == nil {
		return fmt.Errorf("config is nil")
	}
	config := *c.Tree

	config.Set("version", int64(2))

	if runc, ok := config.GetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", "runc"}).(*toml.Tree); ok {
		runc, _ = toml.Load(runc.String())
		config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", name}, runc)
	}

	if config.GetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", name}) == nil {
		config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", name, "runtime_type"}, c.RuntimeType)
		config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", name, "runtime_root"}, "")
		config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", name, "runtime_engine"}, "")
		config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", name, "privileged_without_host_devices"}, false)
	}

	if len(c.ContainerAnnotations) > 0 {
		annotations, err := c.getRuntimeAnnotations([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", name, "container_annotations"})
		if err != nil {
			return err
		}
		annotations = append(c.ContainerAnnotations, annotations...)
		config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", name, "container_annotations"}, annotations)
	}

	config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", name, "options", "BinaryName"}, path)

	if setAsDefault {
		config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "default_runtime_name"}, name)
	} else {
		defaultRuntime, ok := config.GetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "default_runtime_name"}).(string)
		if ok && defaultRuntime == name {
			config.DeletePath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "default_runtime_name"})
		}
	}

	runtimeSubtree := subtreeAtPath(config, "plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", name)
	if err := runtimeSubtree.applyOverrides(configOverrides...); err != nil {
		return fmt.Errorf("failed to apply config overrides: %w", err)
	}

	*c.Tree = config
	return nil
}

func (c *Config) getRuntimeAnnotations(path []string) ([]string, error) {
	if c == nil || c.Tree == nil {
		return nil, nil
	}

	config := *c.Tree
	if !config.HasPath(path) {
		return nil, nil
	}
	annotationsI, ok := config.GetPath(path).([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid annotations: %v", annotationsI)
	}

	var annotations []string
	for _, annotation := range annotationsI {
		a, ok := annotation.(string)
		if !ok {
			return nil, fmt.Errorf("invalid annotation: %v", annotation)
		}
		annotations = append(annotations, a)
	}

	return annotations, nil
}

// Set sets the specified containerd option.
func (c *Config) Set(key string, value interface{}) {
	config := *c.Tree
	config.SetPath([]string{"plugins", "io.containerd.grpc.v1.cri", key}, value)
	*c.Tree = config
}

// DefaultRuntime returns the default runtime for the cri-o config
func (c Config) DefaultRuntime() string {
	if runtime, ok := c.GetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "default_runtime_name"}).(string); ok {
		return runtime
	}
	return ""
}

// RemoveRuntime removes a runtime from the docker config
func (c *Config) RemoveRuntime(name string) error {
	if c == nil || c.Tree == nil {
		return nil
	}

	config := *c.Tree

	config.DeletePath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", name})
	if runtime, ok := config.GetPath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "default_runtime_name"}).(string); ok {
		if runtime == name {
			config.DeletePath([]string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "default_runtime_name"})
		}
	}

	runtimePath := []string{"plugins", "io.containerd.grpc.v1.cri", "containerd", "runtimes", name}
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

// Save writes the config to the specified path
func (c Config) Save(path string) (int64, error) {
	config := c.Tree
	output, err := config.Marshal()
	if err != nil {
		return 0, fmt.Errorf("unable to convert to TOML: %v", err)
	}

	n, err := engine.Config(path).Write(output)
	return int64(n), err
}
