/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package logger

// Interface defines the API for the logger package
type Interface interface {
	Debugf(string, ...interface{})
	Errorf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warning(...interface{})
	Warningf(string, ...interface{})
}
