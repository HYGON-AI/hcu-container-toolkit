/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package lookup

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	envPath = "PATH"
)

var (
	defaultPath = []string{"/usr/local/sbin", "/usr/local/bin", "/usr/sbin", "/usr/bin", "/sbin", "/bin"}
)

// GetPaths returns a list of paths for a specified root. These are constructed from the
// PATH environment variable, a default path list, and the supplied root.
func GetPaths(root string) []string {
	dirs := filepath.SplitList(os.Getenv(envPath))

	inDirs := make(map[string]bool)
	for _, d := range dirs {
		inDirs[d] = true
	}

	// directories from the environment have higher precedence
	for _, d := range defaultPath {
		if inDirs[d] {
			// We don't add paths that are already included
			continue
		}
		dirs = append(dirs, d)
	}

	if root != "" && root != "/" {
		rootDirs := []string{}
		for _, dir := range dirs {
			rootDirs = append(rootDirs, path.Join(root, dir))
		}
		// directories with the root prefix have higher precedence
		dirs = append(rootDirs, dirs...)
	}

	return dirs
}

// GetPath returns a colon-separated path value that can be used to set the PATH
// environment variable
func GetPath(root string) string {
	return strings.Join(GetPaths(root), ":")
}
