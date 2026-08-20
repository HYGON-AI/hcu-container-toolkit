/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

import (
	"bufio"
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/lookup"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// mounts is a generic discoverer for Mounts. It is customized by specifying the
// required entities as a list and a Locator that is used to find the target mounts
// based on the entry in the list.
type mounts struct {
	None
	logger   logger.Interface
	lookup   lookup.Locator
	root     string
	required []string
	sync.Mutex
	cache []Mount
}

var _ Discover = (*mounts)(nil)

// NewMounts creates a discoverer for the required mounts using the specified locator.
func NewMounts(logger logger.Interface, lookup lookup.Locator, root string, required []string) Discover {
	return newMounts(logger, lookup, root, required)
}

// newMounts creates a discoverer for the required mounts using the specified locator.
func newMounts(logger logger.Interface, lookup lookup.Locator, root string, required []string) *mounts {
	return &mounts{
		logger:   logger,
		lookup:   lookup,
		root:     filepath.Join("/", root),
		required: required,
	}
}

func (d *mounts) Mounts() ([]Mount, error) {
	if d.lookup == nil {
		return nil, fmt.Errorf("no lookup defined")
	}

	var temps []Mount
	if d.cache != nil {
		temps = d.cache
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
			if _, ok := uniqueMounts[p]; ok {
				d.logger.Debugf("Skipping duplicate mount %v", p)
				continue
			}

			r := d.relativeTo(p)
			if r == "" {
				r = p
			}

			d.logger.Infof("Selecting %v as %v", p, r)
			uniqueMounts[p] = Mount{
				HostPath: p,
				Path:     r,
				Options: []string{
					"ro",
					"nosuid",
					"nodev",
					"bind",
				},
			}
		}
	}

	var mounts []Mount
	for _, m := range uniqueMounts {
		mounts = append(mounts, m)
	}

	for _, item := range temps {
		mounts = append(mounts, item)
	}
	d.cache = mounts

	return d.cache, nil
}

// relativeTo returns the path relative to the root for the file locator
func (d *mounts) relativeTo(path string) string {
	if d.root == "/" {
		return path
	}

	return strings.TrimPrefix(path, d.root)
}

func (d *mounts) addVHCU(indexs string) {
	var mounts []Mount
	for _, index := range strings.Split(indexs, ",") {
		var mount = Mount{
			HostPath: fmt.Sprintf("/etc/vdev/vdev%s.conf", index),
			Path:     fmt.Sprintf("/etc/vdev/docker/vdev%s.conf", index),
			Options: []string{
				"rbind",
				"ro",
				"rprivate",
			},
		}
		mounts = append(mounts, mount)
	}
	d.cache = mounts
}

func (d *mounts) addMIG(migUUIDs string) {
	const migConfigDir = "/etc/dmi_mig_config/ci"

	want := map[string]struct{}{}
	for _, raw := range strings.Split(migUUIDs, ",") {
		t := strings.TrimSpace(raw)
		if t == "" {
			continue
		}
		if strings.HasPrefix(strings.ToUpper(t), "MIG-") {
			u := strings.ToLower(strings.TrimPrefix(t, "MIG-"))
			if len(u) == 36 {
				want[u] = struct{}{}
			}
		}
	}
	if len(want) == 0 {
		d.cache = nil
		return
	}

	entries, err := os.ReadDir(migConfigDir)
	if err != nil {
		d.cache = nil
		return
	}

	uuidRe := regexp.MustCompile(`^\s*mig_uuid:\s*([0-9a-fA-F-]{36})\s*$`)

	seenHostPath := map[string]struct{}{}

	var mounts []Mount

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".conf") {
			continue
		}

		migConfig := filepath.Join(migConfigDir, e.Name())

		f, err := os.Open(migConfig)
		if err != nil {
			continue
		}

		var fileUUID string
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := sc.Text()
			if m := uuidRe.FindStringSubmatch(line); len(m) == 2 {
				fileUUID = strings.ToLower(m[1])
				break
			}
		}
		_ = f.Close()

		if fileUUID == "" {
			continue
		}
		if _, ok := want[fileUUID]; !ok {
			continue
		}
		if _, dup := seenHostPath[migConfig]; dup {
			continue
		}
		seenHostPath[migConfig] = struct{}{}

		mounts = append(mounts, Mount{
			HostPath: migConfig,
			Path:     migConfig,
			Options: []string{
				"rbind",
				"ro",
				"rprivate",
			},
		})
	}

	d.cache = mounts
}
