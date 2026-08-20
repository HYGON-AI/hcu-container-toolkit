/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package engine

// Interface defines the API for a runtime config updater.
type Interface interface {
	DefaultRuntime() string
	AddRuntime(string, string, bool, ...map[string]interface{}) error
	Set(string, interface{})
	RemoveRuntime(string) error
	Save(string) (int64, error)
}
