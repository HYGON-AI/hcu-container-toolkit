/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package c3000cdi

import "sync"

type Mode string

const (
	// ModeAuto configures the CDI spec generator to automatically detect the system configuration
	ModeAuto = Mode("auto")
	// ModeSmi configures the CDI spec generator to use c3000 smi.
	ModeSmi = Mode("smi")
)

type modeConstraint interface {
	string | Mode
}

type modes struct {
	lookup map[Mode]bool
	all    []Mode
}

var validModes modes
var validModesOnce sync.Once

func getModes() modes {
	validModesOnce.Do(func() {
		all := []Mode{
			ModeAuto,
			ModeSmi,
		}
		lookup := make(map[Mode]bool)

		for _, m := range all {
			lookup[m] = true
		}

		validModes = modes{
			lookup: lookup,
			all:    all,
		}
	})
	return validModes
}

// AllModes returns the set of valid modes.
func AllModes[T modeConstraint]() []T {
	var output []T
	for _, m := range getModes().all {
		output = append(output, T(m))
	}
	return output
}

// IsValidMode checks whether a specified mode is valid.
func IsValidMode[T modeConstraint](mode T) bool {
	return getModes().lookup[Mode(mode)]
}

// resolveMode resolves the mode for CDI spec generation based on the current system.
func (l *c3000cdilib) resolveMode() (rmode Mode) {
	if l.mode != ModeAuto {
		return l.mode
	}
	defer func() {
		l.logger.Infof("Auto-detected mode as '%v'", rmode)
	}()
	// todo: 更合理的方式决定用什么mode
	if l.c3000smicmd.IsValid() {
		return ModeSmi
	}
	return ModeSmi
}
