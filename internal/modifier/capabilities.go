/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package modifier

import (
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/oci"

	"github.com/opencontainers/runtime-spec/specs-go"
)

func NewCapModifier(logger logger.Interface, toAddCaps []string, toRemoveCaps []string) oci.SpecModifier {
	return &capModifier{
		logger:       logger,
		toAddCaps:    toAddCaps,
		toRemoveCaps: toRemoveCaps,
	}
}

type capModifier struct {
	logger       logger.Interface
	toAddCaps    []string
	toRemoveCaps []string
}

var _ oci.SpecModifier = (*capModifier)(nil)

func (m *capModifier) Modify(spec *specs.Spec) error {
	if spec == nil || spec.Process == nil || spec.Process.Capabilities == nil {
		return nil
	}

	bounding := spec.Process.Capabilities.Bounding
	effective := spec.Process.Capabilities.Effective
	permitted := spec.Process.Capabilities.Permitted

	if len(m.toAddCaps) != 0 {
		for _, cap := range m.toAddCaps {
			bounding = append(bounding, cap)
			effective = append(effective, cap)
			permitted = append(permitted, cap)
		}
		bounding = unqiueCaps(bounding)
		effective = unqiueCaps(effective)
		permitted = unqiueCaps(permitted)
	}

	if len(m.toRemoveCaps) != 0 {
		bounding = removeCaps(bounding, m.toRemoveCaps)
		effective = removeCaps(effective, m.toRemoveCaps)
		permitted = removeCaps(permitted, m.toRemoveCaps)
	}

	spec.Process.Capabilities.Bounding = bounding
	spec.Process.Capabilities.Effective = effective
	spec.Process.Capabilities.Permitted = permitted

	return nil
}

func unqiueCaps(caps []string) []string {
	var uCaps []string

	mCaps := make(map[string]bool)

	for _, v := range caps {
		if _, ok := mCaps[v]; !ok {
			mCaps[v] = true
			uCaps = append(uCaps, v)
		}
	}

	return uCaps
}

func removeCaps(orgCaps []string, reCaps []string) []string {
	toRemove := make(map[string]bool)
	for _, cap := range reCaps {
		toRemove[cap] = true
	}

	var caps []string
	for _, v := range orgCaps {
		if _, ok := toRemove[v]; ok {
			continue
		}
		caps = append(caps, v)
	}

	return caps
}
