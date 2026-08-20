/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package device

import "hcu-container-toolkit/pkg/go-c3000smi/pkg/c3000smi"

// MigDevice defines the set of extended functions associated with a MIG device.
type MigDevice interface {
	c3000smi.Device
	GetProfile() (MigProfile, error)
}
