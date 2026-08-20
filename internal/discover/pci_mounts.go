/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

import (
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/lookup"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type pciMounts struct {
	None
	logger   logger.Interface
	lookup   lookup.Locator
	root     string
	required []string
	sync.Mutex
	cache []Mount
}

var _ Discover = (*pciMounts)(nil)

// NewPciMounts creates a discoverer for the required pci mounts using the specified locator.
func NewPciMounts(logger logger.Interface, lookup lookup.Locator, root string, required []string) Discover {
	return &pciMounts{
		logger:   logger,
		lookup:   lookup,
		root:     root,
		required: required,
	}
}

func (d *pciMounts) Mounts() ([]Mount, error) {
	if d.lookup == nil {
		return nil, fmt.Errorf("no lookup defined")
	}

	if d.cache != nil {
		d.logger.Debugf("returning cached mounts")
		return d.cache, nil
	}

	d.Lock()
	defer d.Unlock()

	d.logger.Infof("Locating %v in %v", d.required, d.root)

	uniqueMounts := make(map[string]Mount)

	for _, candidate := range d.required {
		d.logger.Infof("Locating %v", candidate)
		located, err := d.lookup.Locate(candidate)
		if err != nil {
			d.logger.Warningf("Could not locate %v: %v", candidate, err)
			continue
		}
		if len(located) == 0 {
			d.logger.Warningf("Missing %v", candidate)
			continue
		}
		d.logger.Debugf("Located %v as %v", candidate, located)
		for _, p := range located {
			node, err := os.Readlink(p)
			if err != nil {
				d.logger.Warningf("Failed to evaluate symlink %v; ignoring", p)
				continue
			}

			if !filepath.IsAbs(node) {
				node = filepath.Join(filepath.Dir(p), node)
			}

			if _, ok := uniqueMounts[node]; ok {
				d.logger.Debugf("Skipping duplicate mount %v", node)
				continue
			}

			r := d.relativeTo(node)
			if r == "" {
				r = node
			}

			d.logger.Infof("Selecting %v as %v", node, r)
			uniqueMounts[node] = Mount{
				HostPath: node,
				Path:     r,
				Options: []string{
					"rbind",
					"rprivate",
				},
			}
		}
	}

	var mounts []Mount
	for _, m := range uniqueMounts {
		mounts = append(mounts, m)
	}
	d.cache = mounts

	return d.cache, nil
}

// relativeTo returns the path relative to the root for the file locator
func (d *pciMounts) relativeTo(path string) string {
	if d.root == "/" {
		return path
	}

	return strings.TrimPrefix(path, d.root)
}
