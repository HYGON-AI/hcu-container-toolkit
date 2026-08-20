/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package lookup

import (
	"fmt"
	"hcu-container-toolkit/internal/lookup/symlinks"
	"path/filepath"
)

type symlinkChain struct {
	file
}

type symlink struct {
	file
}

// NewSymlinkChainLocator creates a locator that can be used for locating files through symlinks.
func NewSymlinkChainLocator(opts ...Option) Locator {
	f := newFileLocator(opts...)
	l := symlinkChain{
		file: *f,
	}
	return &l
}

// NewSymlinkLocator creats a locator that can be used for locating files through symlinks.
func NewSymlinkLocator(opts ...Option) Locator {
	f := newFileLocator(opts...)
	l := symlink{
		file: *f,
	}

	return &l
}

// Locate finds the specified pattern at the specified root.
// If the file is a symlink, the link is followed and all candidates to the final target are returned.
func (p symlinkChain) Locate(pattern string) ([]string, error) {
	candidates, err := p.file.Locate(pattern)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return candidates, nil
	}

	found := make(map[string]bool)
	for len(candidates) > 0 {
		candidate := candidates[0]
		candidates = candidates[:len(candidates)-1]
		if found[candidate] {
			continue
		}
		found[candidate] = true

		target, err := symlinks.Resolve(candidate)
		if err != nil {
			return nil, fmt.Errorf("error resolving symlink: %v", err)
		}

		if !filepath.IsAbs(target) {
			target, err = filepath.Abs(filepath.Join(filepath.Dir(candidate), target))
			if err != nil {
				return nil, fmt.Errorf("failed to construct absolute path: %v", err)
			}
		}

		p.logger.Debugf("Resolved link: '%v' => '%v'", candidate, target)
		if !found[target] {
			candidates = append(candidates, target)
		}
	}

	var filenames []string
	for f := range found {
		filenames = append(filenames, f)
	}
	return filenames, nil
}

// Locate finds the specified pattern at the specified root.
// If the file is a symlink, the link is resolved and the target returned.
func (p symlink) Locate(pattern string) ([]string, error) {
	candidates, err := p.file.Locate(pattern)
	if err != nil {
		return nil, err
	}

	var targets []string
	seen := make(map[string]bool)
	for _, candidate := range candidates {
		target, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve link: %w", err)
		}
		if seen[target] {
			continue
		}
		seen[target] = true
		targets = append(targets, target)
	}
	return targets, err
}
