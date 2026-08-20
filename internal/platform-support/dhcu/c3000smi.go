/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package dhcu

import (
	"fmt"
	"hcu-container-toolkit/internal/discover"
	"hcu-container-toolkit/internal/info/drm"
	"hcu-container-toolkit/internal/lookup"
	"hcu-container-toolkit/pkg/go-c3000lib/pkg/device"
)

func (o *options) newC3000SimDHCUDiscoverer(d device.Device) (discover.Discover, error) {
	pciBusID := d.GetPCIBusID()
	drmDeviceNodes, err := drm.GetDeviceNodesByBusID(pciBusID)
	if err != nil {
		return nil, fmt.Errorf("failed to determine DRM devices for %v: %v", pciBusID, err)
	}

	deviceNodes := discover.NewCharDeviceDiscoverer(
		o.logger,
		o.devRoot,
		drmDeviceNodes,
	)

	byPathHooks := discover.NewCreateDRMByPathSymlinks(o.logger, deviceNodes, o.devRoot, o.hcuCDIHookPath)

	pciMounts := discover.NewPciMounts(
		o.logger,
		lookup.NewDirectoryLocator(
			lookup.WithLogger(o.logger),
			lookup.WithCount(1),
			lookup.WithSearchPaths("/sys/bus/pci/devices"),
		),
		o.devRoot,
		[]string{pciBusID},
	)

	dd := discover.Merge(
		deviceNodes,
		byPathHooks,
		pciMounts,
	)
	return dd, nil
}
