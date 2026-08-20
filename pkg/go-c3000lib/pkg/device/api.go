/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package device

import "hcu-container-toolkit/pkg/go-c3000smi/pkg/c3000smi"

// Interface provides the API to the 'device' package.
type Interface interface {
	VisitDevices(func(i string, d Device) error) error
}

type devicelib struct {
	c3000smicmd    c3000smi.Interface
	skippedDevices map[string]struct{}
}

var _ Interface = &devicelib{}

// New creates a new instance of the 'device' interface.
func New(smicmd c3000smi.Interface, opts ...Option) Interface {
	d := &devicelib{
		c3000smicmd: smicmd,
	}
	for _, opt := range opts {
		opt(d)
	}
	if d.skippedDevices == nil {
		WithSkippedDevices()(d)
	}
	return d
}

// WithSkippedDevices provides an Option to set devices to be skipped by model name.
func WithSkippedDevices(names ...string) Option {
	return func(d *devicelib) {
		if d.skippedDevices == nil {
			d.skippedDevices = make(map[string]struct{})
		}
		for _, name := range names {
			d.skippedDevices[name] = struct{}{}
		}
	}
}

// Option defines a function for passing options to the New() call.
type Option func(*devicelib)
