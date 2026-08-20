/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package root

// Option defines a functional option for configuring a transormer.
type Option func(*builder)

// WithRoot sets the (from) root for the root transformer.
func WithRoot(root string) Option {
	return func(b *builder) {
		b.root = root
	}
}

// WithTargetRoot sets the (to) target root for the root transformer.
func WithTargetRoot(root string) Option {
	return func(b *builder) {
		b.targetRoot = root
	}
}

// WithRelativeTo sets whether the specified root is relative to the host or container.
func WithRelativeTo(relativeTo string) Option {
	return func(b *builder) {
		b.relativeTo = relativeTo
	}
}
