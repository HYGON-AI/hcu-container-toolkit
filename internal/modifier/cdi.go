/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package modifier

import (
	"fmt"
	"hcu-container-toolkit/internal/config"
	"hcu-container-toolkit/internal/config/image"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/modifier/cdi"
	"hcu-container-toolkit/internal/oci"
	"hcu-container-toolkit/pkg/c3000cdi"
	"hcu-container-toolkit/pkg/c3000cdi/spec"
	"strings"

	"tags.cncf.io/container-device-interface/pkg/parser"
)

// NewCDIModifier creates an OCI spec modifier that determines the modifications to make based on the
// CDI specifications available on the system. The HCU_VISIBLE_DEVICES environment variable is
// used to select the devices to include.
func NewCDIModifier(logger logger.Interface, cfg *config.Config, ociSpec oci.Spec) (oci.SpecModifier, error) {
	devices, err := getDevicesFromSpec(logger, ociSpec, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get required devices from OCI specification: %v", err)
	}
	if len(devices) == 0 {
		logger.Debugf("No devices requested; no modification required.")
		return nil, nil
	}
	logger.Debugf("Creating CDI modifier for devices: %v", devices)
	automaticDevices := filterAutomaticDevices(devices)
	if len(automaticDevices) != len(devices) && len(automaticDevices) > 0 {
		return nil, fmt.Errorf("requesting a CDI device with vendor 'runtime.hygon.cn' is not supported when requesting other CDI devices")
	}
	if len(automaticDevices) > 0 {
		automaticModifier, err := newAutomaticCDISpecModifier(logger, cfg, automaticDevices)
		if err == nil {
			return automaticModifier, nil
		}
		logger.Warningf("Failed to create the automatic CDI modifier: %w", err)
		logger.Debugf("Falling back to the standard CDI modifier")
	}
	return cdi.New(
		cdi.WithLogger(logger),
		cdi.WithDevices(devices...),
		cdi.WithSpecDirs(cfg.HCUContainerRuntimeConfig.Modes.CDI.SpecDirs...),
	)
}

func getDevicesFromSpec(logger logger.Interface, ociSpec oci.Spec, cfg *config.Config) ([]string, error) {
	rawSpec, err := ociSpec.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load OCI spec: %v", err)
	}

	annotationDevices, err := getAnnotationDevices(cfg.HCUContainerRuntimeConfig.Modes.CDI.AnnotationPrefixes, rawSpec.Annotations)
	if err != nil {
		return nil, fmt.Errorf("failed to parse container annotations: %v", err)
	}
	if len(annotationDevices) > 0 {
		return annotationDevices, nil
	}

	container, err := image.NewDTKImageFromSpec(rawSpec)
	if err != nil {
		return nil, err
	}
	if cfg.AcceptDeviceListAsVolumeMounts {
		mountDevices := container.CDIDevicesFromMounts()
		if len(mountDevices) > 0 {
			return mountDevices, nil
		}
	}

	var devices []string
	seen := make(map[string]bool)
	for _, name := range container.VisibleDevicesFromEnvVar() {
		if !parser.IsQualifiedName(name) {
			name = fmt.Sprintf("%s=%s", cfg.HCUContainerRuntimeConfig.Modes.CDI.DefaultKind, name)
		}
		if seen[name] {
			logger.Debugf("Ignoring duplicate device %q", name)
			continue
		}
		devices = append(devices, name)
	}

	if len(devices) == 0 {
		return nil, nil
	}

	if cfg.AcceptEnvvarUnprivileged || image.IsPrivileged(rawSpec) {
		return devices, nil
	}

	logger.Warningf("Ignoring devices specified in HCU_VISIBLE_DEVICES: %v", devices)

	return nil, nil
}

// getAnnotationDevices returns a list of devices specified in the annotations.
// Keys starting with the specified prefixes are considered and expected to contain a comma-separated list of
// fully-qualified CDI devices names. If any device name is not fully-quality an error is returned.
// The list of returned devices is deduplicated.
func getAnnotationDevices(prefixes []string, annotations map[string]string) ([]string, error) {
	devicesByKey := make(map[string][]string)
	for key, value := range annotations {
		for _, prefix := range prefixes {
			if strings.HasPrefix(key, prefix) {
				devicesByKey[key] = strings.Split(value, ",")
			}
		}
	}

	seen := make(map[string]bool)
	var annotationDevices []string
	for key, devices := range devicesByKey {
		for _, device := range devices {
			if !parser.IsQualifiedName(device) {
				return nil, fmt.Errorf("invalid device name %q in annotation %q", device, key)
			}
			if seen[device] {
				continue
			}
			annotationDevices = append(annotationDevices, device)
			seen[device] = true
		}
	}

	return annotationDevices, nil
}

// filterAutomaticDevices searches for "automatic" device names in the input slice.
// "Automatic" devices are a well-defined list of CDI device names which, when requested,
// trigger the generation of a CDI spec at runtime. This removes the need to generate a
// CDI spec on the system a-priori as well as keep it up-to-date.
func filterAutomaticDevices(devices []string) []string {
	var automatic []string
	for _, device := range devices {
		vendor, class, _ := parser.ParseDevice(device)
		if vendor == "runtime.hygon.cn" && class == "hcu" {
			automatic = append(automatic, device)
		}
	}
	return automatic
}

func newAutomaticCDISpecModifier(logger logger.Interface, cfg *config.Config, devices []string) (oci.SpecModifier, error) {
	logger.Debugf("Generating in-memory CDI specs for devices %v", devices)
	spec, err := generateAutomaticCDISpec(logger, cfg, devices)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CDI spec: %w", err)
	}
	cdiModifier, err := cdi.New(
		cdi.WithLogger(logger),
		cdi.WithSpec(spec.Raw()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to construct CDI modifier: %w", err)
	}

	return cdiModifier, nil
}

func generateAutomaticCDISpec(logger logger.Interface, cfg *config.Config, devices []string) (spec.Interface, error) {
	cdilib, err := c3000cdi.New(
		c3000cdi.WithLogger(logger),
		c3000cdi.WithHCUCDIHookPath(cfg.HCUCTKConfig.Path),
		c3000cdi.WithVendor("runtime.hygon.cn"),
		c3000cdi.WithClass("hcu"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to construct CDI library: %w", err)
	}

	identifiers := []string{}
	for _, device := range devices {
		_, _, id := parser.ParseDevice(device)
		identifiers = append(identifiers, id)
	}

	deviceSpecs, err := cdilib.GetDeviceSpecsByID(identifiers...)
	if err != nil {
		return nil, fmt.Errorf("failed to get CDI device specs: %w", err)
	}

	commonEdits, err := cdilib.GetCommonEdits()
	if err != nil {
		return nil, fmt.Errorf("failed to get common CDI spec edits: %w", err)
	}

	return spec.New(
		spec.WithDeviceSpecs(deviceSpecs),
		spec.WithEdits(*commonEdits.ContainerEdits),
		spec.WithVendor("runtime.hygon.cn"),
		spec.WithClass("hcu"),
	)
}
