/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package modifier

import (
	"hcu-container-toolkit/internal/oci"

	"github.com/opencontainers/runtime-spec/specs-go"
)

type List []oci.SpecModifier

// Merge merges a set of OCI specification modifiers as a list.
// This can be used to compose modifiers.
func Merge(modifiers ...oci.SpecModifier) oci.SpecModifier {
	var filteredModifiers List
	for _, m := range modifiers {
		if m == nil {
			continue
		}
		filteredModifiers = append(filteredModifiers, m)
	}

	return filteredModifiers
}

// Modify applies a list of modifiers in sequence and returns on any errors encountered.
func (m List) Modify(spec *specs.Spec) error {
	for _, mm := range m {
		if mm == nil {
			continue
		}
		err := mm.Modify(spec)
		if err != nil {
			return err
		}
	}
	return nil
}
