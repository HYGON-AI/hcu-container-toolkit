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

type hook discover.Hook

// toEdits converts a discovered hook to CDI Container Edits.
func (d hook) toEdits() *cdi.ContainerEdits {
	e := cdi.ContainerEdits{
		ContainerEdits: &specs.ContainerEdits{
			Hooks: []*specs.Hook{d.toSpec()},
		},
	}
	return &e
}

// toSpec converts a discovered Hook to a CDI Spec Hook. Note
// that missing info is filled in when edits are applied by querying the Hook node.
func (d hook) toSpec() *specs.Hook {
	s := specs.Hook{
		HookName: d.Lifecycle,
		Path:     d.Path,
		Args:     d.Args,
	}
	return &s
}
