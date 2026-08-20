/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

import (
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"os"
)

type directoryMountDiscoverer struct {
	logger        logger.Interface
	hostPath      string
	containerPath string
	readOnly      bool
	mustExist     bool
}

var _ Discover = (*directoryMountDiscoverer)(nil)

func NewDirectoryMountDiscoverer(
	logger logger.Interface,
	hostPath, containerPath string,
	readOnly bool,
	mustExist bool,
) Discover {
	return &directoryMountDiscoverer{
		logger:        logger,
		hostPath:      hostPath,
		containerPath: containerPath,
		readOnly:      readOnly,
		mustExist:     mustExist,
	}
}

func (d *directoryMountDiscoverer) Devices() ([]Device, error)        { return nil, nil }
func (d *directoryMountDiscoverer) Hooks() ([]Hook, error)            { return nil, nil }
func (d *directoryMountDiscoverer) AdditionalGIDs() ([]uint32, error) { return nil, nil }

func (d *directoryMountDiscoverer) Mounts() ([]Mount, error) {
	info, err := os.Stat(d.hostPath)
	if err != nil {
		if os.IsNotExist(err) {
			if d.mustExist {
				return nil, fmt.Errorf("required directory %s does not exist", d.hostPath)
			}
			d.logger.Infof("Creating missing directory %s", d.hostPath)
			if mkErr := os.MkdirAll(d.hostPath, 0755); mkErr != nil {
				return nil, fmt.Errorf("failed to create %s: %v", d.hostPath, mkErr)
			}
		} else {
			return nil, fmt.Errorf("failed to stat %s: %v", d.hostPath, err)
		}
	} else if !info.IsDir() {
		return nil, fmt.Errorf("%s exists but is not a directory", d.hostPath)
	}

	opts := []string{"rbind", "nosuid", "nodev"}
	if d.readOnly {
		opts = append(opts, "ro")
	} else {
		opts = append(opts, "rw")
	}

	d.logger.Infof("Mounting %s at %s (%v)", d.hostPath, d.containerPath, opts)

	return []Mount{
		{
			HostPath: d.hostPath,
			Path:     d.containerPath,
			Options:  opts,
		},
	}, nil
}
