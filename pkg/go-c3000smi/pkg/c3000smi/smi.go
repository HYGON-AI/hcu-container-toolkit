/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package c3000smi

import (
	"bytes"
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/lookup"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type runCmd func(cmd *exec.Cmd) error

const (
	DefaultHySmiCommand = "hy-smi"
)

var (
	defaultRunCmd = func(cmd *exec.Cmd) error {
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error running command: %w", err)
		}

		return nil
	}

	serialNumberRe = regexp.MustCompile(`[D|H]CU\[(\d+)\].*Serial Number:\s*(\w+)`)
	cardSeriesRe   = regexp.MustCompile(`[D|H]CU\[(\d+)\].*Card Series:\s*([\w].*)`)
	cardVendorRe   = regexp.MustCompile(`[D|H]CU\[(\d+)\].*Card Vendor:\s*([\w].*)`)
	uniqueIdRe     = regexp.MustCompile(`[D|H]CU\[(\d+)\].*Unique ID:\s*([\w]+)`)
	pciBusIdRe     = regexp.MustCompile(`[D|H]CU\[(\d+)\].*PCI Bus:\s*([\w:.]+)`)
)

type smiCommand struct {
	logger  logger.Interface
	path    string
	Command runCmd
	devices Devices
}

var _ Interface = (*smiCommand)(nil)

// NewSmiCommand creates a Command for the specified logger and path
func NewSmiCommand(logger logger.Interface) (Interface, error) {
	path, err := findSmiBinary(logger)
	if err != nil {
		return nil, fmt.Errorf("error locating binary: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path '%v': %v", path, err)
	}
	if info.IsDir() || info.Mode()&0111 == 0 {
		return nil, fmt.Errorf("specified path '%v' is not an executable file", path)
	}

	smi := smiCommand{
		logger:  logger,
		path:    path,
		Command: defaultRunCmd,
	}

	err = smi.buildDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	return &smi, nil
}

func (s *smiCommand) IsValid() bool {
	_, err := os.Stat(s.path)
	if err != nil {
		s.logger.Errorf("invalid path: '%v': %v", s.path, err)
		return false
	}
	return true
}

func (s *smiCommand) DeviceGetCount() int {
	return len(s.devices)
}

func (s *smiCommand) DeviceGetIndexs() []string {
	indexs := make([]string, 0, s.DeviceGetCount())
	for index := range s.devices {
		indexs = append(indexs, index)
	}

	return indexs
}

func (s *smiCommand) DeviceGetHandleByIndex(index string) (Device, error) {
	if s.devices[index] == nil {
		return nil, fmt.Errorf("%s device not exist", index)
	}
	return s.devices[index], nil
}

func (s *smiCommand) buildDevices() error {
	s.devices = make(Devices)

	cmdArgs := strings.Fields(s.path)
	cmdArgs = append(cmdArgs, "--showuniqueid")
	cmdArgs = append(cmdArgs, "--showbus")
	cmdArgs = append(cmdArgs, "--showproductname")
	cmdArgs = append(cmdArgs, "--showserial")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := s.Command(cmd)
	if err != nil {
		rerr := fmt.Errorf(
			"exec failed: %s | stdout: %s | stderr: %s: %w",
			strings.Join(cmdArgs, " "),
			stdout.String(),
			stderr.String(),
			err,
		)
		s.logger.Errorf("%w", rerr)
		return rerr
	}

	text := stdout.String()

	serialNumbers := parse(text, serialNumberRe)
	for index, serialNumber := range serialNumbers {
		if s.devices[index] == nil {
			s.devices[index] = &DeviceInfo{}
		}
		s.devices[index].serialNumber = serialNumber
	}

	cardSeriess := parse(text, cardSeriesRe)
	for index, cardSeries := range cardSeriess {
		if s.devices[index] == nil {
			s.devices[index] = &DeviceInfo{}
		}
		s.devices[index].cardSeries = cardSeries
	}

	cardVendors := parse(text, cardVendorRe)
	for index, cardVendor := range cardVendors {
		if s.devices[index] == nil {
			s.devices[index] = &DeviceInfo{}
		}
		s.devices[index].cardVendor = cardVendor
	}

	uniqueIds := parse(text, uniqueIdRe)
	for index, uniqueId := range uniqueIds {
		if s.devices[index] == nil {
			s.devices[index] = &DeviceInfo{}
		}
		s.devices[index].uniqueId = uniqueId
	}

	pciBusIds := parse(text, pciBusIdRe)
	for index, pciBusId := range pciBusIds {
		if s.devices[index] == nil {
			s.devices[index] = &DeviceInfo{}
		}
		s.devices[index].pciBusId = pciBusId
	}

	return nil
}

func parse(text string, re *regexp.Regexp) map[string]string {
	matches := make(map[string]string)
	for _, match := range re.FindAllStringSubmatch(text, -1) {
		matches[match[1]] = match[2]
	}
	return matches
}

func findSmiBinary(logger logger.Interface) (string, error) {
	locator := lookup.NewExecutableLocator(logger, "", "/usr/local/hyhal/bin", "/opt/hyhal/bin")

	targets, err := locator.Locate(DefaultHySmiCommand)
	if err == nil && len(targets) > 0 {
		logger.Debugf("Found binary '%v'", targets)
		return targets[0], err
	}
	return "", fmt.Errorf("no binary found from %s", DefaultHySmiCommand)
}
