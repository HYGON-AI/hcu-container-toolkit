/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package validate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"hcu-container-toolkit/cmd/hcu-ctk/cdi/generate"
	"hcu-container-toolkit/internal/logger"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
	"tags.cncf.io/container-device-interface/specs-go"
)

const (
	defaultOutputPath = "/etc/cdi"
	defaultOutputFile = "hcu.yaml"
)

type config struct {
	cdiSpecPath string
}
type command struct {
	logger logger.Interface
}

func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

func (m command) build() *cli.Command {
	cfg := config{}
	c := cli.Command{
		Name:  "validate",
		Usage: "Validate the CDI spec for HCUs",
		Before: func(c *cli.Context) error {
			return m.validateFlags(c, &cfg)
		},
		Action: func(c *cli.Context) error {
			return m.run(c, &cfg)
		},
	}

	c.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "path",
			Usage:       "full path of CDI spec file",
			Value:       defaultOutputPath + "/" + defaultOutputFile,
			Destination: &cfg.cdiSpecPath,
		},
	}
	return &c
}

func (m command) validateFlags(c *cli.Context, cfg *config) error {
	if cfg.cdiSpecPath == "" {
		return errors.New("path is required")
	}

	out, err := filepath.Abs(cfg.cdiSpecPath)
	if err != nil {
		return fmt.Errorf("Incorrect CDI spec file, Error: %v", err)
	}

	fi, err := os.Stat(out)
	if err != nil {
		return fmt.Errorf("Cannot stat CDI spec file %q: %w", out, err)
	}
	if fi.IsDir() {
		return fmt.Errorf("CDI spec path must be a file, got directory: %q", out)
	}

	ext := strings.ToLower(filepath.Ext(out))
	switch ext {
	case ".json", ".yaml", ".yml":
	default:
		return fmt.Errorf("Unsupported CDI spec file extension %q for %q (want .json/.yaml/.yml)", ext, out)
	}

	cfg.cdiSpecPath = out
	return nil
}

func extractParams(spec *specs.Spec) (vendor, class, hookPath string, strategies []string) {
	parts := strings.SplitN(spec.Kind, "/", 2)
	if len(parts) == 2 {
		vendor = parts[0]
		class = parts[1]
	}

	for _, d := range spec.Devices {
		if len(d.ContainerEdits.Hooks) > 0 {
			hookPath = d.ContainerEdits.Hooks[0].Path
			break
		}
	}

	strategyMap := make(map[string]bool)
	reTypeIndex := regexp.MustCompile(`^hcu\d+$`)

	for _, d := range spec.Devices {
		name := d.Name
		if name == "all" {
			continue
		}

		if _, err := strconv.Atoi(name); err == nil {
			strategyMap["index"] = true
		} else if reTypeIndex.MatchString(name) {
			strategyMap["type-index"] = true
		} else if strings.HasPrefix(name, "hcu-") {
			strategyMap["uuid"] = true
		}
	}

	for s := range strategyMap {
		strategies = append(strategies, s)
	}
	return
}

func detectFormatFromContent(f string) (string, error) {
	file, err := os.Open(f)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	trimmed := bytes.TrimLeft(buf[:n], " \t\r\n")
	if len(trimmed) == 0 {
		return "", fmt.Errorf("file is empty")
	}

	if trimmed[0] == '{' || trimmed[0] == '[' {
		return "json", nil
	}

	return "yaml", nil
}

func readSpecFromFile(f string) (*specs.Spec, string, error) {
	format, err := detectFormatFromContent(f)
	if err != nil {
		return nil, "", fmt.Errorf("failed to detect format: %v", err)
	}

	data, err := os.ReadFile(f)
	if err != nil {
		return nil, "", err
	}

	var spec specs.Spec
	if format == "json" {
		err = json.Unmarshal(data, &spec)
	} else {
		err = yaml.UnmarshalStrict(data, &spec)
	}

	if err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal %s content: %v", format, err)
	}

	return &spec, format, nil
}

func (m command) run(c *cli.Context, cfg *config) error {
	path := cfg.cdiSpecPath
	m.logger.Infof("Validating CDI spec at %v", path)

	lspec, format, err := readSpecFromFile(path)
	if err != nil {
		return fmt.Errorf("Failed to read CDI spec file: %v", err)
	}
	m.logger.Debugf("Detected file format: %s", format)

	vendor, class, hookPath, strategies := extractParams(lspec)
	m.logger.Debugf("Detected parameters - Vendor: %s, Class: %s, Hook: %s, Strategies: %v",
		vendor, class, hookPath, strategies)

	spec, err := generate.GenerateSpec(m.logger, format, vendor, class, hookPath, strategies)
	if err != nil {
		return fmt.Errorf("Failed to generate CDI spec: %v", err)
	}
	spec.Version = lspec.Version
	spec.ContainerEdits.AdditionalGIDs = lspec.ContainerEdits.AdditionalGIDs

	if !reflect.DeepEqual(spec, lspec) {
		fmt.Printf("CDI spec is invalid\n Please regenerate CDI spec\n")
	} else {
		fmt.Printf("CDI spec is valid\n")
	}

	return nil
}
