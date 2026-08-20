/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

import "hcu-container-toolkit/internal/logger"

// NewMOFEDDiscoverer creates a discoverer for MOFED devices.
func NewMOFEDDiscoverer(logger logger.Interface, devRoot string) (Discover, error) {
	devices := NewCharDeviceDiscoverer(
		logger,
		devRoot,
		[]string{
			"/dev/infiniband/uverbs*",
			"/dev/infiniband/rdma_cm",
		},
	)

	return devices, nil
}
