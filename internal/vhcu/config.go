/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"fmt"
	"path/filepath"
	"time"

	"hcu-container-toolkit/internal/config"
	"hcu-container-toolkit/internal/logger"
)

// OptionsFromRuntimeConfig converts the toolkit config into vhcu Options.
func OptionsFromRuntimeConfig(cfg config.RuntimeConfig) (Options, error) {
	vhcuCfg := cfg.VHCU

	timeout, err := time.ParseDuration(vhcuCfg.LockTimeout)
	if err != nil {
		return Options{}, fmt.Errorf("invalid hcu-container-runtime.vhcu.lock-timeout %q: %v", vhcuCfg.LockTimeout, err)
	}
	if timeout <= 0 || timeout > 5*time.Minute {
		return Options{}, fmt.Errorf("hcu-container-runtime.vhcu.lock-timeout must be greater than zero and at most 5m")
	}
	if vhcuCfg.MaxVHCUPerCard < 1 || vhcuCfg.MaxVHCUPerCard > 4 {
		return Options{}, fmt.Errorf("hcu-container-runtime.vhcu.max-vhcu-per-card must be between 1 and the driver limit of 4")
	}
	if filepath.Clean(vhcuCfg.ConfDir) != "/etc/vdev" {
		return Options{}, fmt.Errorf("hcu-container-runtime.vhcu.conf-dir must be /etc/vdev: driver configuration and mounts use this fixed directory")
	}
	if !filepath.IsAbs(vhcuCfg.StateDir) || !filepath.IsAbs(vhcuCfg.LockFile) {
		return Options{}, fmt.Errorf("hcu-container-runtime.vhcu.state-dir and lock-file must be absolute paths")
	}
	if vhcuCfg.DestroyRetries < 1 || vhcuCfg.DestroyRetries > 10 {
		return Options{}, fmt.Errorf("hcu-container-runtime.vhcu.destroy-retries must be between 1 and 10")
	}

	return Options{
		MaxVHCUPerCard: vhcuCfg.MaxVHCUPerCard,
		StateDir:       vhcuCfg.StateDir,
		LockFile:       vhcuCfg.LockFile,
		LockTimeout:    timeout,
		ConfDir:        vhcuCfg.ConfDir,
		DestroyRetries: vhcuCfg.DestroyRetries,
	}, nil
}

// NewManagerFromConfig builds a Manager with the hy-smi backend resolved
// from the config (explicit path, well-known DTK location, or PATH).
func NewManagerFromConfig(l logger.Interface, cfg config.RuntimeConfig) (*Manager, error) {
	opts, err := OptionsFromRuntimeConfig(cfg)
	if err != nil {
		return nil, err
	}

	hysmiPath, source := resolveHysmiPath(cfg.VHCU.HysmiPath)
	if hysmiPath == "" {
		return nil, fmt.Errorf("hy-smi not found in a well-known DTK install location or PATH; install DTK or set hcu-container-runtime.vhcu.hysmi-path")
	}
	if source == HysmiFromPATH {
		// This process may be forked by containerd/docker with a different
		// PATH than the one used by the operator; record it for diagnosis.
		l.Debugf("Resolved hy-smi through PATH: %s (consider setting hcu-container-runtime.vhcu.hysmi-path)", hysmiPath)
	}

	return NewManager(l, opts, NewSmiBackend(hysmiPath)), nil
}
