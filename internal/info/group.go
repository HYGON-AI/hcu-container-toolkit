/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package info

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Get groups to add to container
func GetAdditionalGroups() ([]string, error) {
	groups := []string{"video"}

	subGrps, err := GetSubsystemGroups("kfd")
	if err != nil {
		return nil, err
	}

	groups = append(groups, subGrps...)

	return groups, nil
}

func GetSubsystemGroups(subsystem string) ([]string, error) {
	ruleFiles, err := filepath.Glob("/lib/udev/rules.d/*.rules")
	if err != nil {
		return nil, err
	}

	var groups []string

	for _, f := range ruleFiles {
		g, err := parseRuleFile(f, subsystem)
		if err != nil {
			return nil, err
		}
		if g != "" {
			groups = append(groups, g)
		}
	}

	return groups, nil
}

func parseRuleFile(path string, subsystem string) (string, error) {
	infoFile, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open %v: %v", path, err)
	}
	defer infoFile.Close()

	key := fmt.Sprintf(`SUBSYSTEM=="%s"`, subsystem)
	reg := regexp.MustCompile(`GROUP="(\w+)"`)

	scanner := bufio.NewScanner(infoFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, key) {
			found := reg.FindStringSubmatch(line)
			if len(found) < 2 {
				continue
			}
			return found[1], nil
		}
	}
	return "", nil
}
