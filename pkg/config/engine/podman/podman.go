/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package podman

import (
	"fmt"
	"hcu-container-toolkit/pkg/config/engine"

	"github.com/pelletier/go-toml"
)

const (
	defaultPodmanRuntime = "runc"
)

type Config struct {
	*toml.Tree
}

var _ engine.Interface = (*Config)(nil)

func New(opts ...Option) (engine.Interface, error) {
	b := &builder{}
	for _, opt := range opts {
		opt(b)
	}
	return b.build()
}

func (c *Config) AddRuntime(name string, path string, setAsDefault bool, _ ...map[string]interface{}) error {
	if c == nil || c.Tree == nil {
		return fmt.Errorf("config is nil")
	}

	runtime, _ := c.Tree.Get("engine.runtime").(string)
	if runtime == "" {
		c.Tree.Set("containers.default_capabilities", []string{
			"NET_RAW", "CHOWN", "DAC_OVERRIDE", "FOWNER", "FSETID",
			"KILL", "NET_BIND_SERVICE", "SETFCAP", "SETGID", "SETPCAP",
			"SETUID", "SYS_CHROOT",
		})
		c.Tree.Set("containers.default_sysctls", []string{
			"net.ipv4.ping_group_range=0 0",
		})
		c.Tree.Set("containers.log_driver", "k8s-file")
		c.Tree.Set("engine.events_logger", "file")
		c.Tree.Set("engine.runtime", defaultPodmanRuntime)
		c.Tree.Set("network.network_backend", "cni")
	}

	_ = c.Tree.Delete(fmt.Sprintf("engine.runtimes.%s", name))
	c.Tree.Set(fmt.Sprintf("engine.runtimes.%s", name), []string{path})

	if setAsDefault {
		c.Tree.Set("engine.runtime", name)
	} else {
		if v, ok := c.Tree.Get("engine.runtime").(string); ok && v == name {
			c.Tree.Set("engine.runtime", defaultPodmanRuntime)
		}
	}

	return nil
}

func (c Config) DefaultRuntime() string {
	if c.Tree == nil {
		return ""
	}
	if v, ok := c.Tree.Get("engine.runtime").(string); ok {
		return v
	}
	return ""
}

func (c *Config) RemoveRuntime(name string) error {
	if c == nil || c.Tree == nil {
		return nil
	}

	_ = c.Tree.Delete(fmt.Sprintf("engine.runtimes.%s", name))

	if v, ok := c.Tree.Get("engine.runtime").(string); ok && v == name {
		c.Tree.Set("engine.runtime", defaultPodmanRuntime)
	}
	return nil
}

func (c *Config) Set(key string, value interface{}) {
	c.Tree.Set("engine."+key, value)
}

func (c *Config) Save(path string) (int64, error) {
	config := c.Tree
	output, err := config.Marshal()
	if err != nil {
		return 0, fmt.Errorf("unable to convert to TOML: %v", err)
	}

	n, err := engine.Config(path).Write(output)
	return int64(n), err
}
