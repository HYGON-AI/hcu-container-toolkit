/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProbeRuntime queries the exact OCI runtime/root recorded at allocation.
// Missing metadata, command failures and unknown statuses are never absence.
func ProbeRuntime(st *State) (ContainerStatus, error) {
	if len(st.RuntimeCommand) == 0 || !filepath.IsAbs(st.RuntimeCommand[0]) {
		return ContainerUnknown, fmt.Errorf("container %s has no scoped runtime probe", st.ContainerID)
	}
	ctx, cancel := context.WithTimeout(context.Background(), smiCommandTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, st.RuntimeCommand[0], st.RuntimeCommand[1:]...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	out, err := command.Output()
	if err != nil {
		return ContainerUnknown, fmt.Errorf("runtime query for %s failed: %w", st.ContainerID, err)
	}
	// runc list may skip unreadable containers, report them on stderr and
	// still exit successfully. Such an incomplete list cannot prove absence.
	if strings.TrimSpace(stderr.String()) != "" {
		return ContainerUnknown, fmt.Errorf("runtime list may be incomplete: %s", truncateForLog(stderr.String()))
	}
	return parseRuntimeStatus(out, st.ContainerID)
}

func currentCreator() (int, string, error) {
	pid := os.Getpid()
	start, err := processStartTime(pid)
	return pid, start, err
}

func processStartTime(pid int) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return "", err
	}
	return parseProcessStartTime(string(data))
}

func parseProcessStartTime(stat string) (string, error) {
	// comm may itself contain spaces and parentheses; fields after its final
	// closing parenthesis begin at field 3. Starttime is field 22.
	end := strings.LastIndex(stat, ") ")
	if end < 0 {
		return "", fmt.Errorf("invalid process stat")
	}
	fields := strings.Fields(stat[end+2:])
	if len(fields) <= 19 {
		return "", fmt.Errorf("incomplete process stat")
	}
	return fields[19], nil
}

func creatorRunning(st *State) (bool, error) {
	if st.CreatorPID == 0 {
		return false, nil
	}
	boot, err := readBootID()
	if err != nil {
		return false, err
	}
	if boot != st.BootID {
		return false, nil
	}
	start, err := processStartTime(st.CreatorPID)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return start == st.CreatorStartTime, nil
}

func parseRuntimeStatus(out []byte, id string) (ContainerStatus, error) {
	// runc encodes its nil slice as null when no containers are present.
	if strings.TrimSpace(string(out)) == "null" {
		return ContainerAbsent, nil
	}
	if !strings.HasPrefix(strings.TrimSpace(string(out)), "[") {
		return ContainerUnknown, fmt.Errorf("runtime list did not return a JSON array")
	}
	var containers []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(out, &containers); err != nil {
		return ContainerUnknown, err
	}
	status := ContainerAbsent
	seen := map[string]bool{}
	for _, container := range containers {
		if container.ID == "" || seen[container.ID] {
			return ContainerUnknown, fmt.Errorf("invalid container IDs in runtime list")
		}
		seen[container.ID] = true
		if container.ID != id {
			continue
		}
		switch container.Status {
		case "creating", "created", "running", "paused":
			status = ContainerLive
		case "stopped":
			status = ContainerStopped
		default:
			return ContainerUnknown, fmt.Errorf("unrecognized runtime status %q", container.Status)
		}
	}
	return status, nil
}
