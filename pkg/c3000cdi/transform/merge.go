/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package transform

import "tags.cncf.io/container-device-interface/specs-go"

type merged []Transformer

// Merge creates a merged transofrmer from the specified transformers.
func Merge(transformers ...Transformer) Transformer {
	return merged(transformers)
}

// Transform applies all the transformers in the merged set.
func (t merged) Transform(spec *specs.Spec) error {
	for _, transformer := range t {
		if err := transformer.Transform(spec); err != nil {
			return err
		}
	}
	return nil
}
