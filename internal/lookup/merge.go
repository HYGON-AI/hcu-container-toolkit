/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package lookup

import (
	"errors"
)

type first []Locator

// First returns a locator that returns the first non-empty match
func First(locators ...Locator) Locator {
	var f first
	for _, l := range locators {
		if l == nil {
			continue
		}
		f = append(f, l)
	}
	return f
}

// Locate returns the results for the first locator that returns a non-empty non-error result.
func (f first) Locate(pattern string) ([]string, error) {
	var allErrors []error
	for _, l := range f {
		if l == nil {
			continue
		}
		candidates, err := l.Locate(pattern)
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}
		if len(candidates) > 0 {
			return candidates, nil
		}
	}
	return nil, errors.Join(allErrors...)
}
