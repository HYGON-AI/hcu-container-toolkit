/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package image

import (
	"bufio"
	"fmt"
	"hcu-container-toolkit/internal/hyhcu"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
	"tags.cncf.io/container-device-interface/pkg/parser"
)

// DTK represents a DTK image that can be used for HCU computing. This wraps
// a map of environment variable to values that can be used to perform lookups
// such as requirements.
type DTK struct {
	env         map[string]string
	mounts      []specs.Mount
	ContainerId string
}

// NewDTKImageFromSpec creates a DTK image from the input OCI runtime spec.
// The process environment is read (if present) to construct the DTK Image.
func NewDTKImageFromSpec(spec *specs.Spec) (DTK, error) {
	var env []string
	if spec != nil && spec.Process != nil {
		env = spec.Process.Env
	}

	return New(
		WithEnv(env),
		WithMounts(spec.Mounts),
	)
}

// LoadHyhalMethod
func (i DTK) LoadHyhalMethod(envVar string) bool {
	if devs, ok := i.env[envVar]; ok {
		if devs == "copy" {
			return false
		}
	}
	return true
}

func (i DTK) IsLegacy() bool {
	rocm_version := i.env[EnvVarROCmVersion]
	return len(rocm_version) > 0
}

// Getenv returns the value of the specified environment variable.
// If the environment variable is not specified, an empty string is returned.
func (i DTK) Getenv(key string) string {
	return i.env[key]
}

func Contains[T comparable](slice []T, val T) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// DevicesFromEnvvars returns the devices requested by the image through environment variables
func (i DTK) DevicesFromEnvvars(envVars ...string) VisibleDevices {
	isHexString := func(s string) bool {
		if len(s) == 0 {
			return false
		}
		for _, c := range s {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
		return true
	}

	// We concantenate all the devices from the specified env.
	var isSet bool
	var devices []string
	requested := make(map[string]bool)
	for _, envVar := range envVars {
		if devs, ok := i.env[envVar]; ok {
			isSet = true
			for _, d := range strings.Split(devs, ",") {
				trimmed := strings.TrimSpace(d)
				if len(trimmed) == 0 || requested[trimmed] {
					continue
				}
				devices = append(devices, trimmed)
				requested[trimmed] = true
			}
		}
	}

	uuidToHCUIdMap, err := hyhcu.GetUniqueIdToDeviceIndexMap()
	if err != nil {
		uuidToHCUIdMap = make(map[string][]int)
	}
	for key, value := range i.env {
		if strings.HasPrefix(key, "DOCKER_RESOURCE_") {
			for _, c := range strings.Split(value, ",") {
				if strings.HasPrefix(c, "0x") || strings.HasPrefix(c, "0X") ||
					(len(c) > 8 && isHexString(c)) {
					uuid := strings.ToLower(c)
					if !strings.HasPrefix(uuid, "0x") {
						uuid = "0x" + uuid
					}
					if hcuId, exists := uuidToHCUIdMap[uuid]; exists {
						if !Contains(devices, strconv.Itoa(hcuId[0])) {
							devices = append(devices, strconv.Itoa(hcuId[0]))
						}
					} else {
						uuid = strings.TrimPrefix(uuid, "0x")
						if hcuId, exists := uuidToHCUIdMap[uuid]; exists {
							devices = append(devices, strconv.Itoa(hcuId[0]))
						}
					}
				}
			}
			break
		}
	}

	// Environment variable unset with legacy image: default to "all".
	if !isSet && len(devices) == 0 && i.IsLegacy() {
		return NewVisibleDevices("all")
	}

	// Environment variable unset or empty or "void": return nil
	if len(devices) == 0 || requested["void"] {
		return NewVisibleDevices("void")
	}

	return NewVisibleDevices(devices...)
}

// OnlyFullyQualifiedCDIDevices returns true if all devices requested in the image are requested as CDI devices/
func (i DTK) OnlyFullyQualifiedCDIDevices() bool {
	var hasCDIdevice bool
	for _, device := range i.DevicesFromEnvvars(EnvVarHCUVisibleDevices, EnvVarNvidiaVisibleDevices).List() {
		if !parser.IsQualifiedName(device) {
			return false
		}
		hasCDIdevice = true
	}

	for _, device := range i.DevicesFromMounts() {
		if !strings.HasPrefix(device, "cdi/") {
			return false
		}
		hasCDIdevice = true
	}
	return hasCDIdevice
}

const (
	deviceListAsVolumeMountsRoot = "/var/run/hcu-container-devices"
)

// DevicesFromMounts returns a list of device specified as mounts.
// TODO: This should be merged with getDevicesFromMounts used in the HCU Container Runtime
func (i DTK) DevicesFromMounts() []string {
	root := filepath.Clean(deviceListAsVolumeMountsRoot)
	seen := make(map[string]bool)
	var devices []string
	for _, m := range i.mounts {
		source := filepath.Clean(m.Source)
		// Only consider mounts who's host volume is /dev/null
		if source != "/dev/null" {
			continue
		}

		destination := filepath.Clean(m.Destination)
		if seen[destination] {
			continue
		}
		seen[destination] = true

		// Only consider container mount points that begin with 'root'
		if !strings.HasPrefix(destination, root) {
			continue
		}

		// Grab the full path beyond 'root' and add it to the list of devices
		device := strings.Trim(strings.TrimPrefix(destination, root), "/")
		if len(device) == 0 {
			continue
		}
		devices = append(devices, device)
	}
	return devices
}

// CDIDevicesFromMounts returns a list of CDI devices specified as mounts on the image.
func (i DTK) CDIDevicesFromMounts() []string {
	var devices []string
	for _, mountDevice := range i.DevicesFromMounts() {
		if !strings.HasPrefix(mountDevice, "cdi/") {
			continue
		}
		parts := strings.SplitN(strings.TrimPrefix(mountDevice, "cdi/"), "/", 3)
		if len(parts) != 3 {
			continue
		}
		vendor := parts[0]
		class := parts[1]
		device := parts[2]
		devices = append(devices, fmt.Sprintf("%s/%s=%s", vendor, class, device))
	}
	return devices
}

// VisibleDevicesFromEnvVar returns the set of visible devices requested through
// the HCU_VISIBLE_DEVICES environment variable.
func (i DTK) VisibleDevicesFromEnvVar() []string {
	return i.DevicesFromEnvvars(EnvVarHCUVisibleDevices, EnvVarNvidiaVisibleDevices).List()
}

func (i DTK) VHCUFromEnvVar(envVar string) []string {
	getPci := func(path string) string {
		var pciIds []string
		pciRegex := regexp.MustCompile(`PciBusId:\s*([0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-9a-fA-F])`)
		file, err := os.Open(path)
		if err != nil {
			return ""
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			matches := pciRegex.FindStringSubmatch(line)
			if len(matches) > 0 {
				pciIds = append(pciIds, matches[1])
			}
		}
		if len(pciIds) > 0 {
			return pciIds[0]
		}
		return ""
	}

	var vhcuDevices []string
	if value, ok := i.env[envVar]; ok {
		for _, index := range strings.Split(value, ",") {
			var path = fmt.Sprintf("/etc/vdev/vdev%s.conf", index)
			var pciId = getPci(path)
			if len(pciId) > 0 {
				vhcuDevices = append(vhcuDevices, pciId)
			}
		}
	}
	return vhcuDevices
}

func (i DTK) VHCUVisibleDevicesFromEnvVar() []string {
	return i.VHCUFromEnvVar(EnvVarVHCUVisibleDevices)
}

func (i DTK) MIGFromEnvVar(envVar string) []string {
	const migConfigDir = "/etc/dmi_mig_config/ci"

	uniqAppend := func(out []string, v string) []string {
		for _, x := range out {
			if x == v {
				return out
			}
		}
		return append(out, v)
	}

	value, ok := i.env[envVar]
	if !ok || strings.TrimSpace(value) == "" {
		return nil
	}

	var migUUIDs []string
	for _, raw := range strings.Split(value, ",") {
		token := strings.TrimSpace(raw)
		if token == "" {
			continue
		}
		if strings.HasPrefix(strings.ToUpper(token), "MIG-") {
			u := strings.ToLower(strings.TrimPrefix(token, "MIG-"))
			if len(u) == 36 {
				migUUIDs = uniqAppend(migUUIDs, u)
			}
		}
	}
	if len(migUUIDs) == 0 {
		return nil
	}

	entries, err := os.ReadDir(migConfigDir)
	if err != nil {
		return nil
	}

	pciRe := regexp.MustCompile(`^\s*pci:\s*([0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-9a-fA-F])\s*$`)
	uuidRe := regexp.MustCompile(`^\s*mig_uuid:\s*([0-9a-fA-F-]{36})\s*$`)

	var out []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".conf") {
			continue
		}
		path := filepath.Join(migConfigDir, e.Name())

		f, err := os.Open(path)
		if err != nil {
			continue
		}

		var pci string
		var uuid string

		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := sc.Text()

			if pci == "" {
				if m := pciRe.FindStringSubmatch(line); len(m) == 2 {
					pci = m[1]
				}
			}
			if uuid == "" {
				if m := uuidRe.FindStringSubmatch(line); len(m) == 2 {
					uuid = strings.ToLower(m[1])
				}
			}

			if pci != "" && uuid != "" {
				break
			}
		}
		_ = f.Close()

		if pci == "" || uuid == "" {
			continue
		}

		for _, w := range migUUIDs {
			if uuid == w {
				out = uniqAppend(out, pci)
				break
			}
		}
	}

	return out
}

func (i DTK) MIGVisibleDevicesFromEnvVar() []string {
	return i.MIGFromEnvVar(EnvVarMIGVisibleDevices)
}
