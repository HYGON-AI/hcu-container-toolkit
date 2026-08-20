/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package cdi

import (
	"errors"
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/oci"

	"github.com/opencontainers/runtime-spec/specs-go"
	"tags.cncf.io/container-device-interface/pkg/cdi"
)

// fromRegistry represents the modifications performed using a CDI registry.
type fromRegistry struct {
	logger   logger.Interface
	registry *cdi.Cache
	devices  []string
}

var _ oci.SpecModifier = (*fromRegistry)(nil)

// Modify applies the modifications defined by the CDI registry to the incoming OCI spec.
func (m fromRegistry) Modify(spec *specs.Spec) error {
	if err := m.registry.Refresh(); err != nil {
		m.logger.Debugf("The following error was triggered when refreshing the CDI registry: %v", err)
	}

	m.logger.Debugf("Injecting devices using CDI: %v", m.devices)
	unresolvedDevices, err := m.registry.InjectDevices(spec, m.devices...)
	if unresolvedDevices != nil {
		m.logger.Warningf("could not resolve CDI devices: %v", unresolvedDevices)
	}
	if err != nil {
		var refreshErrors []error
		for _, rerrs := range m.registry.GetErrors() {
			refreshErrors = append(refreshErrors, rerrs...)
		}
		if rerr := errors.Join(refreshErrors...); rerr != nil {
			// We log the errors that may have been generated while refreshing the CDI registry.
			// These may be due to malformed specifications or device name conflicts that could be
			// the cause of an injection failure.
			m.logger.Warningf("Refreshing the CDI registry generated errors: %v", rerr)
		}

		return fmt.Errorf("failed to inject CDI devices: %v", err)
	}

	return nil
}
