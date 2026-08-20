/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package spec

import (
	"fmt"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"
)

type setMinimumRequiredVersion struct{}

func (d setMinimumRequiredVersion) Transform(spec *specs.Spec) error {
	minVersion, err := cdi.MinimumRequiredVersion(spec)
	if err != nil {
		return fmt.Errorf("failed to get minimum required CDI spec version: %v", err)
	}
	spec.Version = minVersion
	return nil
}
