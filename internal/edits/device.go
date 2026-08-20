/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package edits

import (
	"hcu-container-toolkit/internal/discover"
	"os"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"
)

type device discover.Device

// toEdits converts a discovered device to CDI Container Edits.
func (d device) toEdits() (*cdi.ContainerEdits, error) {
	deviceNode, err := d.toSpec()
	if err != nil {
		return nil, err
	}

	e := cdi.ContainerEdits{
		ContainerEdits: &specs.ContainerEdits{
			DeviceNodes: []*specs.DeviceNode{deviceNode},
		},
	}
	return &e, nil
}

// toSpec converts a discovered Device to a CDI Spec Device. Note
// that missing info is filled in when edits are applied by querying the Device node.
func (d device) toSpec() (*specs.DeviceNode, error) {
	// The HostPath field was added in the v0.5.0 CDI specification.
	// The cdi package uses strict unmarshalling when loading specs from file causing failures for
	// unexpected fields.
	// Since the behaviour for HostPath == "" and HostPath == Path are equivalent, we clear HostPath
	// if it is equal to Path to ensure compatibility with the widest range of specs.
	hostPath := d.HostPath
	if hostPath == d.Path {
		hostPath = ""
	}
	s := specs.DeviceNode{
		HostPath: hostPath,
		Path:     d.Path,
	}
	if hostPath == "/dev/mem" {
		fileMode := os.FileMode(0666)
		s.FileMode = &fileMode
		s.Permissions = "rw"
	}

	return &s, nil
}
