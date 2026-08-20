/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package device

const (
	// AttributeMediaExtensions holds the string representation for the media extension MIG profile attribute.
	AttributeMediaExtensions = "me"
)

// MigProfile represents a specific MIG profile.
// Examples include "1g.5gb", "2g.10gb", "1c.2g.10gb", or "1c.1g.5gb+me", etc.
type MigProfile interface {
	String() string
	GetInfo() MigProfileInfo
	Equals(other MigProfile) bool
	Matches(profile string) bool
}

// MigProfileInfo holds all info associated with a specific MIG profile.
type MigProfileInfo struct {
	C              int
	G              int
	GB             int
	Attributes     []string
	GIProfileID    int
	CIProfileID    int
	CIEngProfileID int
}
