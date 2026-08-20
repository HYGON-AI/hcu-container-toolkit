/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package drm

import (
	"fmt"
	"path/filepath"
)

// GetDeviceNodesByBusID returns the DRM devices associated with the specified PCI bus ID
func GetDeviceNodesByBusID(busID string) ([]string, error) {
	drmRoot := filepath.Join("/sys/bus/pci/devices", busID, "drm")
	matches, err := filepath.Glob(fmt.Sprintf("%s/*", drmRoot))
	if err != nil {
		return nil, err
	}

	var drmDeviceNodes []string
	for _, m := range matches {
		drmDeviceNode := filepath.Join("/dev/dri", filepath.Base(m))
		drmDeviceNodes = append(drmDeviceNodes, drmDeviceNode)
	}

	return drmDeviceNodes, nil
}
