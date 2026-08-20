/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package config

// CTKConfig stores the config options for the HCU Container Toolkit CLI (hcu-ctk)
type CTKConfig struct {
	Path string `toml:"path"`
}
