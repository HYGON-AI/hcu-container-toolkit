/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package hyhcu

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type DeviceInfo struct {
	DrmDevices    []string
	PartitionType string
}

type FileSystem interface {
	Stat(name string) (os.FileInfo, error)
	Glob(pattern string) ([]string, error)
	ReadFile(name string) ([]byte, error)
	GetDeviceStat(dev string, format string) (string, error)
}

type DefaultFS struct{}

var defaultFS FileSystem = &DefaultFS{}

func (fs *DefaultFS) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }

func (fs *DefaultFS) Glob(pattern string) ([]string, error) { return filepath.Glob(pattern) }

func (fs *DefaultFS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }

func (fs *DefaultFS) GetDeviceStat(dev string, format string) (string, error) {
	out, err := exec.Command("stat", "-c", format, dev).Output()
	if err != nil {
		fmt.Printf("stat failed for %v: Error %v", dev, err)
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetHYHCUs return the list of all the HCU devices on the system.
func GetHYHCUs() ([]DeviceInfo, error) { return GetHyHCUsWithFS(defaultFS) }

func GetHyHCUsWithFS(fs FileSystem) ([]DeviceInfo, error) {

	renderDevIds := GetDevIdsFromTopology(fs)

	// Map to store devices by unique_id to maintain grouping
	uniqueIdDevices := make(map[string][]DeviceInfo)
	var uniqueIds []string // To maintain order

	pciDevs, err := fs.Glob("/sys/module/hy*cu/drivers/pci:hy*cu/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*")
	if err != nil {
		fmt.Printf("Failed to find hyhcu driver directories: %v", err)
		return nil, err
	}

	// Process platform devices for partitions
	platformDevs, _ := fs.Glob("/sys/devices/platform/hyhcu_xcp_*")

	// Combine aboth PCI and platform devices
	allDevs := append(pciDevs, platformDevs...)

	//Process all devices using the same logic
	for _, path := range allDevs {
		computePartitionFile := filepath.Join(path, "current_compute_partition")
		memoryPartitionFile := filepath.Join(path, "current_memory_partition")

		computePartitionType, memoryPartitionType, combinedPartitionType := "", "", ""

		// Read the compute partition
		if data, err := ioutil.ReadFile(computePartitionFile); err == nil {
			computePartitionType = strings.ToLower(strings.TrimSpace(string(data)))
		}
		// Read the memory partition
		if data, err := ioutil.ReadFile(memoryPartitionFile); err == nil {
			memoryPartitionType = strings.ToLower(strings.TrimSpace(string(data)))
		}

		combinedPartitionType = computePartitionType + "_" + memoryPartitionType
		if combinedPartitionType == "_" {
			combinedPartitionType = ""
		}

		drms, err := fs.Glob(path + "/drm/*")
		if err != nil {
			return nil, err
		}

		drmDevs := []string{}
		renderMinor := 0
		for _, drm := range drms {
			dev := filepath.Base(drm)
			if len(dev) >= 4 && dev[0:4] == "card" || len(dev) >= 7 && dev[0:7] == "renderD" {
				drmDevs = append(drmDevs, "/dev/dri/"+dev)
				if len(dev) >= 7 && dev[0:7] == "renderD" {
					renderMinor, _ = strconv.Atoi(dev[7:])
				}
			}
		}

		if len(drmDevs) > 0 && renderMinor > 0 {
			if devID, exists := renderDevIds[renderMinor]; exists {
				if _, exists := uniqueIdDevices[devID]; !exists {
					uniqueIds = append(uniqueIds, devID)
				}
				uniqueIdDevices[devID] = append(uniqueIdDevices[devID], DeviceInfo{DrmDevices: drmDevs, PartitionType: combinedPartitionType})
			}
		}
	}

	// Sort devices within each unique_id group by render minor number
	for _, devID := range uniqueIds {
		sort.Slice(uniqueIdDevices[devID], func(i, j int) bool {
			getRenderID := func(devInfo DeviceInfo) int {
				devs := devInfo.DrmDevices
				for _, dev := range devs {
					baseDev := filepath.Base(dev)
					if len(baseDev) >= 7 && strings.HasPrefix(baseDev, "renderD") {
						id, _ := strconv.Atoi(strings.TrimPrefix(baseDev, "renderD"))
						return id
					}
				}
				return 0
			}
			return getRenderID(uniqueIdDevices[devID][i]) < getRenderID(uniqueIdDevices[devID][j])
		})
	}

	// Combine all devices maintaining the unique_id order
	var devs []DeviceInfo
	for _, devID := range uniqueIds {
		devs = append(devs, uniqueIdDevices[devID]...)
	}

	return devs, nil
}

var topoUniqueIdRe = regexp.MustCompile(`unique_id\s(\d+)`)
var renderMinorRe = regexp.MustCompile(`drm_render_minor\s(\d+)`)

// GetDevIdsFromTopology returns a map of render minor numbers to unique_ids
func GetDevIdsFromTopology(fs FileSystem, topoRootParam ...string) map[int]string {
	topoRoot := "/sys/class/kfd/kfd"
	if len(topoRootParam) == 1 {
		topoRoot = topoRootParam[0]
	}

	renderDevIds := make(map[int]string)
	nodeFiles, err := fs.Glob(topoRoot + "/topology/nodes/*/properties")
	if err != nil {
		return renderDevIds
	}

	for _, nodeFile := range nodeFiles {
		renderMinor, err := ParseTopologyProperties(fs, nodeFile, renderMinorRe)
		if err != nil {
			continue
		}

		if renderMinor <= 0 || renderMinor > math.MaxInt32 {
			continue
		}

		devID, err := ParseTopologyPropertiesString(fs, nodeFile, topoUniqueIdRe)
		if err != nil {
			continue
		}

		renderDevIds[int(renderMinor)] = devID
	}

	return renderDevIds
}

// ParseTopologyProperties parses for a property value in kfd topology file as int64
// The format is usually one entry per line <name> <value>.
func ParseTopologyProperties(fs FileSystem, path string, re *regexp.Regexp) (int64, error) {
	content, err := fs.ReadFile(path)
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if matches != nil {
			return strconv.ParseInt(matches[1], 0, 64)
		}
	}

	return 0, fmt.Errorf("property not found in %s", path)
}

// ParseTopologyPropertiesString parses for a property value in kfd topology file as string
// The format is usually one entry per line <name> <value>.
func ParseTopologyPropertiesString(fs FileSystem, path string, re *regexp.Regexp) (string, error) {
	content, err := fs.ReadFile(path)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if matches != nil {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("property not found in %s", path)
}

// GetUniqueIdToDeviceIndexMap returns a map of unique_id (as hex string) to device indices
func GetUniqueIdToDeviceIndexMap() (map[string][]int, error) {
	return GetUniqueIdToDeviceIndexMapWithFS(defaultFS)
}

// GetUniqueIdToDeviceIndexMapWithFS creates a mapping from unique_id (hex format) to device indices
func GetUniqueIdToDeviceIndexMapWithFS(fs FileSystem) (map[string][]int, error) {
	devs, err := GetHyHCUsWithFS(fs)
	if err != nil {
		return nil, err
	}

	renderDevIds := GetDevIdsFromTopology(fs)
	uniqueIdToIndex := make(map[string][]int)

	// Process each device group and assign index
	for deviceIndex, deviceGroup := range devs {
		// Find the render minor for this device group
		for _, device := range deviceGroup.DrmDevices {
			// Extract render minor from device path like /dev/dri/renderD128
			if strings.Contains(device, "renderD") {
				renderMinorStr := strings.TrimPrefix(filepath.Base(device), "renderD")
				if renderMinor, err := strconv.Atoi(renderMinorStr); err == nil {
					if uniqueId, exists := renderDevIds[renderMinor]; exists {
						// Convert decimal unique_id to hex format (without 0x prefix)
						if uniqueIdInt, err := strconv.ParseUint(uniqueId, 10, 64); err == nil {
							hexUniqueId := fmt.Sprintf("0x%x", uniqueIdInt)
							uniqueIdToIndex[hexUniqueId] = append(uniqueIdToIndex[hexUniqueId], deviceIndex)
							// Also support without 0x prefix
							hexUniqueIdNoPrefix := fmt.Sprintf("%x", uniqueIdInt)
							uniqueIdToIndex[hexUniqueIdNoPrefix] = append(uniqueIdToIndex[hexUniqueIdNoPrefix], deviceIndex)
						}
					}
				}
				break // Only need one render device per group
			}
		}
	}

	return uniqueIdToIndex, nil
}
