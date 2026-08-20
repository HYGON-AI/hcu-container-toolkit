/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package modifier

import (
	"fmt"
	"hcu-container-toolkit/internal/discover"
	"hcu-container-toolkit/internal/edits"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/oci"

	"github.com/opencontainers/runtime-spec/specs-go"
)

type discoverModifier struct {
	logger     logger.Interface
	discoverer discover.Discover
}

// NewModifierFromDiscoverer creates a modifier that applies the discovered
// modifications to an OCI spec if required by the runtime wrapper.
func NewModifierFromDiscoverer(logger logger.Interface, d discover.Discover) (oci.SpecModifier, error) {
	m := discoverModifier{
		logger:     logger,
		discoverer: d,
	}
	return &m, nil
}

// Modify applies the modifications required by discoverer to the incomming OCI spec.
// These modifications are applied in-place.
func (m discoverModifier) Modify(spec *specs.Spec) error {
	specEdits, err := edits.NewSpecEdits(m.logger, m.discoverer)
	if err != nil {
		return fmt.Errorf("failed to get required container edits: %v", err)
	}

	return specEdits.Modify(spec)
}
