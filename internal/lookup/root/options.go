/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package root

import "hcu-container-toolkit/internal/logger"

type Option func(*Driver)

func WithLogger(logger logger.Interface) Option {
	return func(d *Driver) {
		d.logger = logger
	}
}

func WithDriverRoot(root string) Option {
	return func(d *Driver) {
		d.Root = root
	}
}
