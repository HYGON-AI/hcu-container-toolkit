/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"hcu-container-toolkit/internal/config"
	"testing"
)

func TestRuntimeConfigValidation(t *testing.T) {
	defaults, err := config.GetDefaultRuntimeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := OptionsFromRuntimeConfig(*defaults); err != nil {
		t.Fatalf("defaults invalid: %v", err)
	}
	for _, tc := range []struct {
		name   string
		change func(*config.RuntimeConfig)
	}{
		{"slots-zero", func(c *config.RuntimeConfig) { c.VHCU.MaxVHCUPerCard = 0 }},
		{"slots-over-driver", func(c *config.RuntimeConfig) { c.VHCU.MaxVHCUPerCard = 5 }},
		{"negative-timeout", func(c *config.RuntimeConfig) { c.VHCU.LockTimeout = "-1s" }},
		{"large-timeout", func(c *config.RuntimeConfig) { c.VHCU.LockTimeout = "6m" }},
		{"invalid-timeout", func(c *config.RuntimeConfig) { c.VHCU.LockTimeout = "invalid" }},
		{"other-conf-dir", func(c *config.RuntimeConfig) { c.VHCU.ConfDir = "/custom/vdev" }},
		{"relative-state", func(c *config.RuntimeConfig) { c.VHCU.StateDir = "state" }},
		{"relative-lock", func(c *config.RuntimeConfig) { c.VHCU.LockFile = "lock" }},
		{"zero-retries", func(c *config.RuntimeConfig) { c.VHCU.DestroyRetries = 0 }},
		{"large-retries", func(c *config.RuntimeConfig) { c.VHCU.DestroyRetries = 11 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := *defaults
			tc.change(&cfg)
			if _, err := OptionsFromRuntimeConfig(cfg); err == nil {
				t.Fatal("invalid config accepted")
			}
		})
	}
}
