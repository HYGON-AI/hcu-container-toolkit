/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"os"
	"sort"
)

// rmFile removes a file, treating "not exist" as success.
func rmFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func sortInts(v []int) {
	sort.Ints(v)
}
