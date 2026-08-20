/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package root

import (
	"hcu-container-toolkit/pkg/c3000cdi/transform"
	"hcu-container-toolkit/pkg/c3000cdi/transform/noop"
)

type builder struct {
	transformer
	relativeTo string
}

func (b *builder) build() transform.Transformer {
	if b.root == b.targetRoot {
		return noop.New()
	}

	if b.relativeTo == "container" {
		return containerRootTransformer(b.transformer)
	}
	return hostRootTransformer(b.transformer)
}
