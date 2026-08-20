/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package c3000caps

// MigMinor represents the minor number of a MIG device
type MigMinor int

// MigCap represents the path to a MIG cap file
type MigCap string

// MigCaps stores a map of MIG cap file paths to MIG minors
type MigCaps map[MigCap]MigMinor
