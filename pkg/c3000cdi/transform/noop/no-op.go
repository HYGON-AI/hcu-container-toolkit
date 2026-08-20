/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package noop

import (
	"hcu-container-toolkit/pkg/c3000cdi/transform"

	"tags.cncf.io/container-device-interface/specs-go"
)

type noop struct{}

var _ transform.Transformer = (*noop)(nil)

// New returns a no-op transformer.
func New() transform.Transformer {
	return noop{}
}

// Transform is a no-op for a noop transformer.
func (n noop) Transform(spec *specs.Spec) error {
	return nil
}
