/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package query

import (
	"bytes"
	"fmt"
	"hcu-container-toolkit/internal/hyhcu"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type HcuInfo struct {
	HcuId         int
	Pid           []string
	ContainerName []string
	Uuid          string
}

type HCUProcess struct {
	Pid   string
	Index string
}

func parseHCUsList(hcus string) ([]int, []string, []string, error) {
	// isHexString checks if a string contains only hexadecimal characters
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

	validHCUs := []int{}
	invalidHCUs := []string{}
	invalidHCUsRange := []string{}

	hcusInfo, err := hyhcu.GetHYHCUs()
	if err != nil {
		return []int{}, []string{}, []string{}, fmt.Errorf("Failed to get HCU info, Error: %v", err)
	}

	if hcus == "all" || hcus == "All" || hcus == "ALL" {
		for i := 0; i < len(hcusInfo); i++ {
			validHCUs = append(validHCUs, i)
		}
		return validHCUs, []string{}, []string{}, nil
	}

	uuidToHCUIdMap, err := hyhcu.GetUniqueIdToDeviceIndexMap()

	if err != nil {
		fmt.Printf("Failed to get UUID to HCU Id mappings: %v", err)
		uuidToHCUIdMap = make(map[string][]int)
	}

	for _, c := range strings.Split(hcus, ",") {
		if strings.HasPrefix(c, "0x") || strings.HasPrefix(c, "0X") ||
			(len(c) > 8 && isHexString(c)) {
			uuid := strings.ToLower(c)
			if !strings.HasPrefix(uuid, "0x") {
				uuid = "0x" + uuid
			}
			if hcuIds, exists := uuidToHCUIdMap[uuid]; exists {
				validHCUs = append(validHCUs, hcuIds...)
			} else {
				uuid = strings.TrimPrefix(uuid, "0x")
				if hcuIds, exists := uuidToHCUIdMap[uuid]; exists {
					validHCUs = append(validHCUs, hcuIds...)
				} else {
					invalidHCUs = append(invalidHCUs, c)
				}
			}
		} else if strings.Contains(c, "-") {
			devsRange := strings.SplitN(c, "-", 2)
			start, err0 := strconv.Atoi(devsRange[0])
			end, err1 := strconv.Atoi(devsRange[1])
			if err0 != nil || err1 != nil ||
				start < 0 || end < 0 || start > end {
				invalidHCUsRange = append(invalidHCUsRange, c)
			} else {
				for i := start; i <= end; i++ {
					if i < len(hcusInfo) {
						validHCUs = append(validHCUs, i)
					} else {
						invalidHCUs = append(invalidHCUs, strconv.Itoa(i))
					}
				}
			}
		} else {
			i, err := strconv.Atoi(c)
			if err == nil {
				if i >= 0 && i < len(hcusInfo) {
					validHCUs = append(validHCUs, i)
				} else {
					invalidHCUs = append(invalidHCUs, c)
				}
			} else {
				invalidHCUs = append(invalidHCUs, c)
			}
		}
	}

	sort.Ints(validHCUs)

	return validHCUs, invalidHCUs, invalidHCUsRange, nil
}

func CheckHySmi() (string, error) {
	cmdWhich := exec.Command("which", "hy-smi")
	var outWhich bytes.Buffer
	cmdWhich.Stdout = &outWhich

	if err := cmdWhich.Run(); err == nil {
		path := strings.TrimSpace(outWhich.String())
		if path != "" {
			return path, nil
		}
	}

	cmdWhereis := exec.Command("whereis", "hy-smi")
	var outWhereis bytes.Buffer
	cmdWhereis.Stdout = &outWhereis

	if err := cmdWhereis.Run(); err != nil {
		return "", err
	}

	output := outWhereis.String()
	fields := strings.Fields(output)
	if len(fields) > 1 {
		return fields[1], nil
	}

	return "", nil
}

func ExecShowPids(hysmi string) ([]HCUProcess, error) {

	cmd := exec.Command(hysmi, "--showpids")
	var out bytes.Buffer
	cmd.Stdout = &out
	var results []HCUProcess
	err := cmd.Run()
	if err != nil {
		return results, err
	}

	output := out.String()

	reBlock := regexp.MustCompile(`PID:\s*(\d+)[\s\S]*?[H|D]CU Index:\s*(.*)`)
	matches := reBlock.FindAllStringSubmatch(output, -1)

	for _, m := range matches {
		pid := m[1]
		index := strings.TrimSpace(m[2])
		index = strings.Trim(index, "[]' ")
		results = append(results, HCUProcess{Pid: pid, Index: index})
	}

	return results, nil
}

func QueryName(pid string, runtime string, namespace string) (string, error) {
	runtime = strings.ToLower(strings.TrimSpace(runtime))
	if runtime == "" {
		runtime = "docker"
	}

	cgroupPath := fmt.Sprintf("/proc/%s/cgroup", pid)
	data, err := os.ReadFile(cgroupPath)
	if err != nil {
		return "", err
	}
	output := string(data)
	var containerID string
	switch runtime {
	case "docker":
		containerID = extractID(output, `(?:docker|docker/|docker-)([0-9a-f]{64})`)
	case "podman":
		containerID = extractID(output, `(?:libpod|podman|containers?)(?:-|/)([0-9a-f]{64})`)
	case "containerd":
		patterns := []string{
			namespace + `/([0-9a-f]{64})`,
			`cri-containerd-([0-9a-f]{64})`,
			`containerd(?:-|/)([0-9a-f]{64})`,
			`containerd://([0-9a-f]{64})`,
		}
		for _, p := range patterns {
			if id := extractID(output, p); id != "" {
				containerID = id
				break
			}
		}
	}

	if containerID == "" {
		return "", nil
	}
	return queryNameByRuntime(containerID, runtime, namespace)
}

func extractID(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func queryNameByRuntime(containerID, runtime string, namespace string) (string, error) {
	switch runtime {
	case "docker", "podman":
		cmd := exec.Command(runtime, "ps", "-a", "--format", "{{.ID}} {{.Names}}")
		return parsePsOutput(cmd, containerID)
	case "containerd":
		cmd := exec.Command("nerdctl", "-n", namespace, "ps", "-a", "--format", "{{.ID}} {{.Names}}")
		name, err := parsePsOutput(cmd, containerID)
		if err != nil {
			return "", err
		}
		if name != "" {
			return name, nil
		}
		return containerID, nil
	}
	return "", nil
}

func parsePsOutput(cmd *exec.Cmd, containerID string) (string, error) {
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.HasPrefix(containerID, fields[0]) {
			return fields[1], nil
		}
	}
	return containerID, nil
}

func ShowStatus(hcus string, runtime string, namespace string) error {
	if runtime != "docker" && runtime != "containerd" && runtime != "podman" {
		return fmt.Errorf("Runtime not supported")
	}
	hySmiPath, err := CheckHySmi()
	if err != nil {
		return fmt.Errorf("Failed to check hy-smi path, Error: %v", err)
	}
	processes, err := ExecShowPids(hySmiPath)
	if err != nil {
		return fmt.Errorf("Failed to run hy-smi command, Error: %v", err)
	}
	validHCUs, _, _, err := parseHCUsList(hcus)
	if err != nil {
		return fmt.Errorf("Failed to parse HCUs list, Error: %v", err)
	}

	var hcuinfos = make(map[int]HcuInfo)
	uuidToHCUIdMap, err := hyhcu.GetUniqueIdToDeviceIndexMap()
	if err != nil {
		return fmt.Errorf("Failed to get UUID to HCU Id mappings: %v", err)
	}

	for _, hcu := range validHCUs {
		for uuid, hcuIds := range uuidToHCUIdMap {
			if strings.HasPrefix(uuid, "0x") || strings.HasPrefix(uuid, "0X") {
				uuid = uuid[2:]
			}
			uuid = "0x" + strings.ToUpper(uuid)
			if hcuIds[0] == hcu {
				hcuinfos[hcu] = HcuInfo{HcuId: hcu, Uuid: uuid}
				break
			}
		}
	}

	for _, process := range processes {
		index, err := strconv.Atoi(process.Index)
		if err != nil {
			continue
		}
		if hcu, exists := hcuinfos[index]; exists {
			hcu.Pid = append(hcu.Pid, process.Pid)
			name, err := QueryName(process.Pid, runtime, namespace)
			if err != nil {
				continue
			}
			hcu.ContainerName = append(hcu.ContainerName, name)
			hcuinfos[index] = hcu // 注意：结构体是值类型，需重新赋值
		}
	}

	fmt.Println(strings.Repeat("-", 120))
	fmt.Printf("%-40s%-50s%-20s\n", "HCU ID", "Unique ID", "Container Names")
	fmt.Println(strings.Repeat("-", 120))
	for hcuId := range hcuinfos {
		for idx, name := range hcuinfos[hcuId].ContainerName {
			if idx == 0 {
				fmt.Printf("%-40v%-50s%-20v\n", hcuId, hcuinfos[hcuId].Uuid, name)
			} else {
				fmt.Printf("%-40v%-50s%-20v\n", "", "", name)
			}
		}

	}
	fmt.Println(strings.Repeat("-", 120))
	return nil
}
