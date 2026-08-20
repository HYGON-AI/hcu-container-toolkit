/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package oci

//go:generate moq -stub -out runtime_mock.go . Runtime

// Runtime is an interface for a runtime shim. The Exec method accepts a list
// of command line arguments, and returns an error / nil.
type Runtime interface {
	Exec([]string) error
}
