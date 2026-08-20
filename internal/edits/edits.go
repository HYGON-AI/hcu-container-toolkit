/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package edits

import (
	"fmt"
	"hcu-container-toolkit/internal/discover"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/oci"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"

	ociSpecs "github.com/opencontainers/runtime-spec/specs-go"
)

type edits struct {
	cdi.ContainerEdits
	logger logger.Interface
}

// NewSpecEdits creates a SpecModifier that defines the required OCI spec edits (as CDI ContainerEdits) from the specified
// discoverer.
func NewSpecEdits(logger logger.Interface, d discover.Discover) (oci.SpecModifier, error) {
	c, err := FromDiscoverer(logger, d)
	if err != nil {
		return nil, fmt.Errorf("error constructing container edits: %v", err)
	}
	e := edits{
		ContainerEdits: *c,
		logger:         logger,
	}
	return &e, nil
}

// FromDiscoverer creates CDI container edits for the specified discoverer.
func FromDiscoverer(logger logger.Interface, d discover.Discover) (*cdi.ContainerEdits, error) {
	devices, err := d.Devices()
	if err != nil {
		return nil, fmt.Errorf("failed to discover devices: %v", err)
	}

	mounts, err := d.Mounts()
	if err != nil {
		return nil, fmt.Errorf("failed to discover mounts: %v", err)
	}

	hooks, err := d.Hooks()
	if err != nil {
		return nil, fmt.Errorf("failed to discover hooks: %v", err)
	}

	gids, err := d.AdditionalGIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to discover additionalGids: %v", err)
	}

	c := NewContainerEdits()
	for _, d := range devices {
		edits, err := device(d).toEdits()
		if err != nil {
			return nil, fmt.Errorf("failed to created container edits for device: %v", err)
		}
		c.Append(edits)
	}

	for _, m := range mounts {
		c.Append(mount(m).toEdits())
	}

	for _, h := range hooks {
		c.Append(hook(h).toEdits())
	}

	c.Append(&cdi.ContainerEdits{
		ContainerEdits: &specs.ContainerEdits{
			AdditionalGIDs: gids,
		},
	})

	return c, nil
}

// NewContainerEdits is a utility function to create a CDI ContainerEdits struct.
func NewContainerEdits() *cdi.ContainerEdits {
	c := cdi.ContainerEdits{
		ContainerEdits: &specs.ContainerEdits{},
	}
	return &c
}

// Modify applies the defined edits to the incoming OCI spec
func (e *edits) Modify(spec *ociSpecs.Spec) error {
	if e == nil || e.ContainerEdits.ContainerEdits == nil {
		return nil
	}

	e.logger.Info("Mounts:")
	for _, mount := range e.Mounts {
		e.logger.Infof("Mounting %v at %v", mount.HostPath, mount.ContainerPath)
	}
	e.logger.Infof("Devices:")
	for _, device := range e.DeviceNodes {
		e.logger.Infof("Injecting %v", device.Path)
	}
	e.logger.Infof("Hooks:")
	for _, hook := range e.Hooks {
		e.logger.Infof("Injecting %v %v", hook.Path, hook.Args)
	}

	return e.Apply(spec)
}
