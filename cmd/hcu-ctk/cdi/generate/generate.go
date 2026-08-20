/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package generate

import (
	"fmt"
	"hcu-container-toolkit/internal/config"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/pkg/c3000cdi"
	"hcu-container-toolkit/pkg/c3000cdi/spec"
	"hcu-container-toolkit/pkg/c3000cdi/transform"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"tags.cncf.io/container-device-interface/pkg/parser"
	"tags.cncf.io/container-device-interface/specs-go"
)

const (
	allDeviceName = "all"
)

func GenerateSpec(logger logger.Interface, format, vendor, class, hookPath string, strategies []string) (*specs.Spec, error) {
	m := command{logger: logger}
	opts := options{}
	opts.mode = string(c3000cdi.ModeAuto)
	opts.format = format
	opts.vendor = vendor
	opts.class = class
	opts.hcuCDIHookPath = hookPath
	opts.deviceNameStrategies = *cli.NewStringSlice(strategies...)

	spec, err := m.generateSpec(&opts)
	if err != nil {
		return nil, err
	}
	return spec.Raw(), nil
}

type command struct {
	logger logger.Interface
}

type options struct {
	output               string
	format               string
	deviceNameStrategies cli.StringSlice
	driverRoot           string
	devRoot              string
	hcuCDIHookPath       string
	mode                 string
	vendor               string
	class                string

	csv struct {
		files          cli.StringSlice
		ignorePatterns cli.StringSlice
	}
}

// NewCommand constructs a generate-cdi command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

// build creates the CLI command
func (m command) build() *cli.Command {
	opts := options{}

	// Create the 'generate-cdi' command
	c := cli.Command{
		Name:  "generate",
		Usage: "Generate CDI specifications for use with CDI-enabled runtimes",
		Before: func(c *cli.Context) error {
			return m.validateFlags(c, &opts)
		},
		Action: func(c *cli.Context) error {
			return m.run(c, &opts)
		},
	}

	c.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "output",
			Usage:       "Specify the file to output the generated CDI specification to. If this is '' the specification is output to STDOUT",
			Destination: &opts.output,
		},
		&cli.StringFlag{
			Name:        "format",
			Usage:       "The output format for the generated spec [json | yaml]. This overrides the format defined by the output file extension (if specified).",
			Value:       spec.FormatYAML,
			Destination: &opts.format,
		},
		&cli.StringFlag{
			Name:    "mode",
			Aliases: []string{"discovery-mode"},
			Usage: "The mode to use when discovering the available entities. " +
				"One of [" + strings.Join(c3000cdi.AllModes[string](), " | ") + "]. " +
				"If mode is set to 'auto' the mode will be determined based on the system configuration.",
			Value:       string(c3000cdi.ModeAuto),
			Destination: &opts.mode,
		},
		&cli.StringSliceFlag{
			Name:        "device-name-strategy",
			Usage:       "Specify the strategy for generating device names. If this is specified multiple times, the devices will be duplicated for each strategy. One of [index | uuid | type-index]",
			Value:       cli.NewStringSlice(c3000cdi.DeviceNameStrategyIndex, c3000cdi.DeviceNameStrategyUUID),
			Destination: &opts.deviceNameStrategies,
		},
		&cli.StringFlag{
			Name:    "hcu-cdi-hook-path",
			Aliases: []string{"hcu-ctk-path"},
			Usage: "Specify the path to use for the hcu-cdi-hook in the generated CDI specification. " +
				"If not specified, the PATH will be searched for `hcu-cdi-hook`. " +
				"NOTE: That if this is specified as `hcu-ctk`, the PATH will be searched for `hcu-ctk` instead.",
			Destination: &opts.hcuCDIHookPath,
		},
		&cli.StringFlag{
			Name:        "vendor",
			Aliases:     []string{"cdi-vendor"},
			Usage:       "the vendor string to use for the generated CDI specification.",
			Value:       "hygon.cn",
			Destination: &opts.vendor,
		},
		&cli.StringFlag{
			Name:        "class",
			Aliases:     []string{"cdi-class"},
			Usage:       "the class string to use for the generated CDI specification.",
			Value:       "hcu",
			Destination: &opts.class,
		},
	}

	return &c
}

func (m command) validateFlags(c *cli.Context, opts *options) error {
	opts.format = strings.ToLower(opts.format)

	switch opts.format {
	case spec.FormatJSON:
	case spec.FormatYAML:
	default:
		return fmt.Errorf("invalid output format: %v", opts.format)
	}

	opts.mode = strings.ToLower(opts.mode)
	if !c3000cdi.IsValidMode(opts.mode) {
		return fmt.Errorf("invalid discovery mode: %v", opts.mode)
	}

	for _, strategy := range opts.deviceNameStrategies.Value() {
		_, err := c3000cdi.NewDeviceNamer(strategy)
		if err != nil {
			return err
		}
	}

	opts.hcuCDIHookPath = config.ResolveHCUCDIHookPath(m.logger, opts.hcuCDIHookPath)

	if outputFileFormat := formatFromFilename(opts.output); outputFileFormat != "" {
		m.logger.Debugf("Inferred output format as %q from output file name", outputFileFormat)
		if !c.IsSet("format") {
			opts.format = outputFileFormat
		} else if outputFileFormat != opts.format {
			m.logger.Warningf("Requested output format %q does not match format implied by output file name: %q", opts.format, outputFileFormat)
		}
	}

	if err := parser.ValidateVendorName(opts.vendor); err != nil {
		return fmt.Errorf("invalid CDI vendor name: %v", err)
	}
	if err := parser.ValidateClassName(opts.class); err != nil {
		return fmt.Errorf("invalid CDI class name: %v", err)
	}
	return nil
}

func (m command) run(c *cli.Context, opts *options) error {
	spec, err := m.generateSpec(opts)
	if err != nil {
		return fmt.Errorf("failed to generate CDI spec: %v", err)
	}
	m.logger.Infof("Generated CDI spec with version %v", spec.Raw().Version)

	if opts.output == "" {
		_, err := spec.WriteTo(os.Stdout)
		if err != nil {
			return fmt.Errorf("failed to write CDI spec to STDOUT: %v", err)
		}
		return nil
	}

	return spec.Save(opts.output)
}

func formatFromFilename(filename string) string {
	ext := filepath.Ext(filename)
	switch strings.ToLower(ext) {
	case ".json":
		return spec.FormatJSON
	case ".yaml", ".yml":
		return spec.FormatYAML
	}

	return ""
}

func (m command) generateSpec(opts *options) (spec.Interface, error) {
	var deviceNamers []c3000cdi.DeviceNamer
	for _, strategy := range opts.deviceNameStrategies.Value() {
		deviceNamer, err := c3000cdi.NewDeviceNamer(strategy)
		if err != nil {
			return nil, fmt.Errorf("failed to create device namer: %v", err)
		}
		deviceNamers = append(deviceNamers, deviceNamer)
	}

	cdilib, err := c3000cdi.New(
		c3000cdi.WithLogger(m.logger),
		c3000cdi.WithDriverRoot(opts.driverRoot),
		c3000cdi.WithDevRoot(opts.devRoot),
		c3000cdi.WithHCUCDIHookPath(opts.hcuCDIHookPath),
		c3000cdi.WithDeviceNamers(deviceNamers...),
		c3000cdi.WithMode(opts.mode),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CDI library: %v", err)
	}

	deviceSpecs, err := cdilib.GetAllDeviceSpecs()
	if err != nil {
		return nil, fmt.Errorf("failed to create device CDI specs: %v", err)
	}

	commonEdits, err := cdilib.GetCommonEdits()
	if err != nil {
		return nil, fmt.Errorf("failed to create edits common for entities: %v", err)
	}

	return spec.New(
		spec.WithVendor(opts.vendor),
		spec.WithClass(opts.class),
		spec.WithDeviceSpecs(deviceSpecs),
		spec.WithEdits(*commonEdits.ContainerEdits),
		spec.WithFormat(opts.format),
		spec.WithMergedDeviceOptions(
			transform.WithName(allDeviceName),
			transform.WithSkipIfExists(true),
		),
		spec.WithPermissions(0644),
	)
}
