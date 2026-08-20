/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

import (
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/lookup"
	"path/filepath"
	"strings"
)

// mountsToContainerPath defines a Discoverer for a required set of mounts.
// When these are discovered by a locator the specified container root is used
// to construct the container path for the mount returned.
type mountsToContainerPath struct {
	None
	logger        logger.Interface
	locator       lookup.Locator
	required      []string
	containerRoot string
}

func (d *mountsToContainerPath) Mounts() ([]Mount, error) {
	seen := make(map[string]bool)
	var mounts []Mount
	for _, target := range d.required {
		if strings.Contains(target, "*") {
			// TODO: We could relax this condition.
			return nil, fmt.Errorf("wildcard patterns are not supported: %s", target)
		}
		candidates, err := d.locator.Locate(target)
		if err != nil {
			d.logger.Warningf("Could not locate %v: %v", target, err)
			continue
		}
		if len(candidates) == 0 {
			d.logger.Warningf("Missing %v", target)
			continue
		}
		hostPath := candidates[0]
		if seen[hostPath] {
			d.logger.Debugf("Skipping duplicate mount %v", hostPath)
			continue
		}
		seen[hostPath] = true
		d.logger.Debugf("Located %v as %v", target, hostPath)

		containerPath := filepath.Join(d.containerRoot, target)
		d.logger.Infof("Selecting %v as %v", hostPath, containerPath)

		mount := Mount{
			HostPath: hostPath,
			Path:     containerPath,
			Options: []string{
				"ro",
				"nosuid",
				"nodev",
				"bind",
			},
		}
		mounts = append(mounts, mount)
	}

	return mounts, nil
}
