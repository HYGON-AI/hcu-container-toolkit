/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package modifier

import (
	"bufio"
	"errors"
	"fmt"
	"hcu-container-toolkit/internal/config/image"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/oci"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
)

var reDrmRenderMinor = regexp.MustCompile(`drm_render_minor\s(\d+)`)

// sysfsMountModifier is a spec modifier that handle subdirectory mount in /sys
type sysfsMountModifier struct {
	logger         logger.Interface
	devices        image.VisibleDevices
	busIds         []string
	hcuCDIHookPath string
}

var _ oci.SpecModifier = (*sysfsMountModifier)(nil)

func NewSysfsMountModifier(logger logger.Interface, devices image.VisibleDevices, busIds []string, hcuCDIHookPath string) oci.SpecModifier {
	m := sysfsMountModifier{
		logger:         logger,
		devices:        devices,
		busIds:         busIds,
		hcuCDIHookPath: hcuCDIHookPath,
	}

	return &m
}

func (m sysfsMountModifier) Modify(spec *specs.Spec) error {
	if spec == nil {
		return nil
	}

	var selectedBusIds []string
	isAll := true
	for i, busId := range m.busIds {
		if m.devices.Has(fmt.Sprintf("%d", i)) || m.devices.Has(busId) {
			selectedBusIds = append(selectedBusIds, busId)
		} else {
			isAll = false
		}
	}

	if isAll {
		m.logger.Debugf("All devices requested, no need to handle /sys mount")
		return nil
	}

	var mounts []specs.Mount

	mounted := make(map[string]bool)
	for _, mount := range spec.Mounts {
		mount := mount
		if mount.Destination == "/sys" {
			continue
		}
		mounts = append(mounts, mount)

		if strings.HasPrefix(mount.Source, "/sys") {
			mounted[mount.Source] = true
		}
	}

	selectRender := make(map[string]bool)

	for _, busId := range selectedBusIds {
		drmRoot := filepath.Join("/sys/bus/pci/devices", busId, "drm")
		renderNodes, err := filepath.Glob(fmt.Sprintf("%s/renderD*", drmRoot))
		if err != nil {
			return fmt.Errorf("failed to determine DRM render devices for %v: %v", busId, err)
		}

		for _, renderNode := range renderNodes {
			selectRender[filepath.Base(renderNode)] = true
		}
	}

	nodeRoot := "/sys/devices/virtual/kfd/kfd/topology/nodes"

	matches, err := filepath.Glob(fmt.Sprintf("%s/*", nodeRoot))
	if err != nil {
		m.logger.Warningf("Failed to found topology nodes")
		return err
	}

	for _, path := range matches {
		render_minor, err := ParseTopologyProperties(filepath.Join(path, "properties"), reDrmRenderMinor)
		if err != nil {
			return err
		}

		if int(render_minor) == 0 || selectRender[fmt.Sprintf("renderD%d", int(render_minor))] {
			mounts = append(mounts, specs.Mount{
				Destination: path,
				Type:        "bind",
				Source:      path,
				Options:     []string{"rbind", "rprivate"},
			})
		}
	}

	var links []string

	curPath := filepath.Dir(nodeRoot)
	base := filepath.Base(nodeRoot)
	for {
		matches, err := filepath.Glob(fmt.Sprintf("%s/*", curPath))
		if err != nil {
			m.logger.Warningf("failed to find subdirecties for %s: %v", curPath, err)
			return nil
		}

		for _, path := range matches {
			if filepath.Base(path) == base || mounted[path] {
				continue
			}

			lpath, err := os.Readlink(path)
			if err != nil {
				mounts = append(mounts, specs.Mount{
					Destination: path,
					Type:        "bind",
					Source:      path,
					Options:     []string{"rbind", "rprivate"},
				})
			} else {
				m.logger.Debugf("adding symlink %v -> %v", path, lpath)
				links = append(links, fmt.Sprintf("%v::%v", lpath, path))
			}
		}

		base = filepath.Base(curPath)
		curPath = filepath.Dir(curPath)

		if curPath == "/" {
			break
		}
	}

	spec.Mounts = mounts

	if len(links) != 0 {
		var args []string
		args = append(args, "hcu-ctk", "hook", "create-symlinks")
		for _, l := range links {
			args = append(args, "--link", l)
		}

		var hooks []specs.Hook

		for _, hook := range spec.Hooks.CreateContainer {
			hook := hook
			hooks = append(hooks, hook)
		}

		hooks = append(hooks, specs.Hook{
			Path: m.hcuCDIHookPath,
			Args: args,
		})

		spec.Hooks.CreateContainer = hooks
	}

	return nil
}

// ParseTopologyProperties parse for a property value in kfd topology file
// The format is usually one entry per line <name> <value>.
func ParseTopologyProperties(path string, re *regexp.Regexp) (int64, error) {
	f, e := os.Open(path)
	if e != nil {
		return 0, e
	}

	e = errors.New("Topology property not found.  Regex: " + re.String())
	v := int64(0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		m := re.FindStringSubmatch(scanner.Text())
		if m == nil {
			continue
		}

		v, e = strconv.ParseInt(m[1], 0, 64)
		break
	}
	f.Close()

	return v, e
}
