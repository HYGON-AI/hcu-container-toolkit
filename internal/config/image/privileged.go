/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package image

import "github.com/opencontainers/runtime-spec/specs-go"

const (
	capSysAdmin = "CAP_SYS_ADMIN"
)

// IsPrivileged returns true if the container is a privileged container.
func IsPrivileged(s *specs.Spec) bool {
	if s.Process.Capabilities == nil {
		return false
	}

	// We only make sure that the bounding capabibility set has
	// CAP_SYS_ADMIN. This allows us to make sure that the container was
	// actually started as '--privileged', but also allow non-root users to
	// access the privileged HCU capabilities.
	for _, c := range s.Process.Capabilities.Bounding {
		if c == capSysAdmin {
			return true
		}
	}
	return false
}
