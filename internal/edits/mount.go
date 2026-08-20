/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package edits

import (
	"hcu-container-toolkit/internal/discover"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"
)

type mount discover.Mount

// toEdits converts a discovered mount to CDI Container Edits.
func (d mount) toEdits() *cdi.ContainerEdits {
	e := cdi.ContainerEdits{
		ContainerEdits: &specs.ContainerEdits{
			Mounts: []*specs.Mount{d.toSpec()},
		},
	}
	return &e
}

// toSpec converts a discovered Mount to a CDI Spec Mount. Note
// that missing info is filled in when edits are applied by querying the Mount node.
func (d mount) toSpec() *specs.Mount {
	s := specs.Mount{
		HostPath:      d.HostPath,
		ContainerPath: d.Path,
		Options:       d.Options,
	}

	return &s
}
