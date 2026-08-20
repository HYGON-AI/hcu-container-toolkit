/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package lookup

import "errors"

//go:generate moq -stub -out locator_mock.go . Locator

// Locator defines the interface for locating files on a system.
type Locator interface {
	Locate(string) ([]string, error)
}

// ErrNotFound indicates that a specified pattern or file could not be found.
var ErrNotFound = errors.New("not found")
