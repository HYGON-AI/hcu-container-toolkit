/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package runtime

import (
	"fmt"
	"hcu-container-toolkit/internal/config"
	"hcu-container-toolkit/internal/config/image"
	"hcu-container-toolkit/internal/info"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/lookup/root"
	"hcu-container-toolkit/internal/modifier"
	"hcu-container-toolkit/internal/oci"
)

const (
	loadHyhalMethodEnvvar = "HCU_LOAD_HYHAL_METHOD"
)

// newHCUContainerRuntime is a factory method that constructs a runtime based on the selected configuration and specified logger
func newHCUContainerRuntime(logger logger.Interface, cfg *config.Config, argv []string, driver *root.Driver) (oci.Runtime, error) {
	lowLevelRuntime, err := oci.NewLowLevelRuntime(logger, cfg.HCUContainerRuntimeConfig.Runtimes)
	if err != nil {
		return nil, fmt.Errorf("error constructing low-level runtime: %v", err)
	}

	if !oci.HasCreateSubcommand(argv) {
		logger.Debugf("Skipping modifier for non-create subcommand")
		return lowLevelRuntime, nil
	}

	ociSpec, err := oci.NewSpec(logger, argv)
	if err != nil {
		return nil, fmt.Errorf("error constructing OCI specification: %v", err)
	}

	specModifier, err := newSpecModifier(logger, cfg, ociSpec, driver, argv[len(argv)-1])
	if err != nil {
		return nil, fmt.Errorf("failed to construct OCI spec modifier: %v", err)
	}

	// Create the wrapping runtime with the specified modifier
	r := oci.NewModifyingRuntimeWrapper(
		logger,
		lowLevelRuntime,
		ociSpec,
		specModifier,
	)

	return r, nil
}

// newSpecModifier is a factory method that creates constructs an OCI spec modifer based on the provided config.
func newSpecModifier(logger logger.Interface, cfg *config.Config, ociSpec oci.Spec, driver *root.Driver, containerId string) (oci.SpecModifier, error) {
	rawSpec, err := ociSpec.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load OCI spec: %v", err)
	}

	image, err := image.NewDTKImageFromSpec(rawSpec)
	if err != nil {
		return nil, err
	}
	image.ContainerId = containerId
	mode := info.ResolveAutoMode(logger, cfg.HCUContainerRuntimeConfig.Mode, image)
	// We update the mode here so that we can continue passing just the config to other functions.
	cfg.HCUContainerRuntimeConfig.Mode = mode
	modeModifier, err := newModeModifier(logger, mode, cfg, ociSpec, image)
	if err != nil {
		return nil, err
	}

	isMount := image.LoadHyhalMethod(loadHyhalMethodEnvvar)

	var modifiers modifier.List
	for _, modifierType := range supportedModifierTypes(mode) {
		switch modifierType {
		case "mode":
			modifiers = append(modifiers, modeModifier)
		case "graphics":
			graphicsModifier, err := modifier.NewGraphicsModifier(logger, cfg, image, driver, isMount)
			if err != nil {
				return nil, err
			}
			modifiers = append(modifiers, graphicsModifier)
		case "feature-gated":
			featureGatedModifier, err := modifier.NewFeatureGatedModifier(logger, cfg, image, driver)
			if err != nil {
				return nil, err
			}
			modifiers = append(modifiers, featureGatedModifier)
		}
	}

	var copyModifier oci.SpecModifier
	if !isMount {
		copyModifier, err = modifier.NewCopyModifier(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create copy modifier: %v", err)
		}
		modifiers = append(modifiers, copyModifier)
	}
	return modifiers, nil
}

func newModeModifier(logger logger.Interface, mode string, cfg *config.Config, ociSpec oci.Spec, image image.DTK) (oci.SpecModifier, error) {
	switch mode {
	case "legacy":
		return modifier.NewStableModifier(logger, cfg, image)
	case "cdi":
		return modifier.NewCDIModifier(logger, cfg, ociSpec)
	}
	return nil, fmt.Errorf("invalid runtime mode: %v", cfg.HCUContainerRuntimeConfig.Mode)
}

// supportedModifierTypes returns the modifiers supported for a specific runtime mode.
func supportedModifierTypes(mode string) []string {
	switch mode {
	case "cdi":
		// For CDI mode we make no additional modifications.
		return []string{"mode"}
	default:
		return []string{"feature-gated", "graphics", "mode"}
	}
}

// newStable
