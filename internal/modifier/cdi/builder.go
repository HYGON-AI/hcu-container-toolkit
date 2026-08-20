/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package cdi

import (
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/oci"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"
)

type builder struct {
	logger   logger.Interface
	specDirs []string
	devices  []string
	cdiSpec  *specs.Spec
}

// Option represents a functional option for creating a CDI mofifier.
type Option func(*builder)

// New creates a new CDI modifier.
func New(opts ...Option) (oci.SpecModifier, error) {
	b := &builder{}
	for _, opt := range opts {
		opt(b)
	}
	if b.logger == nil {
		b.logger = logger.New()
	}
	return b.build()
}

// build uses the applied options and constructs a CDI modifier using the builder.
func (m builder) build() (oci.SpecModifier, error) {
	if len(m.devices) == 0 && m.cdiSpec == nil {
		return nil, nil
	}

	if m.cdiSpec != nil {
		modifier := fromCDISpec{
			cdiSpec: &cdi.Spec{Spec: m.cdiSpec},
		}
		return modifier, nil
	}

	registry, err := cdi.NewCache(
		cdi.WithAutoRefresh(false),
		cdi.WithSpecDirs(m.specDirs...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CDI registry: %v", err)
	}

	modifier := fromRegistry{
		logger:   m.logger,
		registry: registry,
		devices:  m.devices,
	}

	return modifier, nil
}

// WithLogger sets the logger for the CDI modifier builder.
func WithLogger(logger logger.Interface) Option {
	return func(b *builder) {
		b.logger = logger
	}
}

// WithSpecDirs sets the spec directories for the CDI modifier builder.
func WithSpecDirs(specDirs ...string) Option {
	return func(b *builder) {
		b.specDirs = specDirs
	}
}

// WithDevices sets the devices for the CDI modifier builder.
func WithDevices(devices ...string) Option {
	return func(b *builder) {
		b.devices = devices
	}
}

// WithSpec sets the spec for the CDI modifier builder.
func WithSpec(spec *specs.Spec) Option {
	return func(b *builder) {
		b.cdiSpec = spec
	}
}
