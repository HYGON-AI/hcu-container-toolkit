/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package dhcu

import (
	"errors"
	"hcu-container-toolkit/internal/discover"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/pkg/go-c3000lib/pkg/device"
)

// NewForDevice creates a discoverer for the specified Device.
func NewForDevice(d device.Device, opts ...Option) (discover.Discover, error) {
	o := new(opts...)

	var discoverers []discover.Discover
	var errs error

	c3000smiDiscoverer, err := o.newC3000SimDHCUDiscoverer(d)
	if err != nil {
		errs = errors.Join(errs, err)
	} else if c3000smiDiscoverer != nil {
		discoverers = append(discoverers, c3000smiDiscoverer)
	}

	if len(discoverers) == 0 {
		return nil, errs
	}

	return discover.WithCache(
		discover.FirstValid(
			discoverers...,
		),
	), nil
}

func new(opts ...Option) *options {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	if o.logger == nil {
		o.logger = logger.New()
	}

	return o
}
