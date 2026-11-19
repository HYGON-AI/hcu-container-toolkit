/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package runtime

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"hcu-container-toolkit/internal/config"
	"hcu-container-toolkit/internal/config/image"
	"hcu-container-toolkit/internal/info"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/modifier"
	"hcu-container-toolkit/internal/oci"
	"hcu-container-toolkit/internal/vhcu"

	"github.com/opencontainers/runtime-spec/specs-go"
)

// newVHCUModifier runs the dynamic vHCU split for this container (if
// requested) and returns the modifier carrying the env write-back and the
// poststop release hook. It returns (nil, nil, nil) when no dynamic split was
// requested.
//
// Ordering constraints:
//   - the allocation runs here, at spec-modifier construction time, because
//     only the create stage may add devices/mounts to the spec;
//   - the env write-back is applied to rawSpec BEFORE image.DTK is snapshotted
//     by the caller, so the existing graphics modifier chain picks up
//     VHCU_VISIBLE_DEVICES (or HCU_VISIBLE_DEVICES on a full-card short
//     circuit) without any change;
//   - CDI mode is rejected loudly: a static, install-time-generated CDI spec
//     cannot express a vdev index that only exists at container-create time.
func newVHCUModifier(logger logger.Interface, cfg *config.Config, rawSpec *specs.Spec, containerId string, runtimeCommand []string) (oci.SpecModifier, func() error, error) {
	env := func(key string) string {
		if rawSpec == nil || rawSpec.Process == nil {
			return ""
		}
		for _, kv := range rawSpec.Process.Env {
			if len(kv) > len(key) && kv[:len(key)] == key && kv[len(key)] == '=' {
				return kv[len(key)+1:]
			}
		}
		return ""
	}

	req, err := vhcu.ParseRequest(env)
	if err != nil {
		return nil, nil, err
	}
	if req == nil {
		// No dynamic split requested: legacy paths stay untouched.
		return nil, nil, nil
	}

	if !cfg.HCUContainerRuntimeConfig.VHCU.Enabled {
		return nil, nil, fmt.Errorf("dynamic vHCU split requested via HCU_VHCU_CORES/HCU_VHCU_MEM but disabled in config (hcu-container-runtime.vhcu.enabled=false)")
	}

	// Mode check on the pre-mutation env snapshot (mode resolution does not
	// depend on the VHCU env we are about to write).
	preImage, err := image.NewDTKImageFromSpec(rawSpec)
	if err != nil {
		return nil, nil, err
	}
	mode := info.ResolveAutoMode(logger, cfg.HCUContainerRuntimeConfig.Mode, preImage)
	if mode == "cdi" {
		return nil, nil, fmt.Errorf("HCU_VHCU_CORES/HCU_VHCU_MEM require legacy runtime mode: CDI specs are generated at install time and cannot express dynamically created vHCU devices; set hcu-container-runtime.mode to \"legacy\" or drop the HCU_VHCU_* variables")
	}
	hookPath, err := dynamicVHCUHookPath(logger, cfg.HCUCTKConfig.Path)
	if err != nil {
		return nil, nil, err
	}

	caps := vhcu.Detect(cfg.HCUContainerRuntimeConfig.VHCU.HysmiPath)
	if err := caps.Error(); err != nil {
		// Fail closed: the container must not start without the requested
		// fractional card (silently falling back to a whole card would
		// defeat the resource isolation the user asked for).
		return nil, nil, err
	}
	if !caps.Root {
		return nil, nil, fmt.Errorf("dynamic vHCU split is unsupported in rootless mode: hy-smi virtual requires root")
	}

	manager, err := vhcu.NewManagerFromConfig(logger, cfg.HCUContainerRuntimeConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to construct vhcu manager: %v", err)
	}

	req.RuntimeCommand = append([]string(nil), runtimeCommand...)
	alloc, err := manager.Allocate(containerId, req)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() error {
		if alloc.Reused || alloc.FullCard {
			return nil
		}
		return manager.ReleaseAllocation(containerId, alloc.AllocationID)
	}

	// Write the outcome into the raw spec so the image snapshot taken by the
	// caller sees it (image.DTK reads env exactly once, at construction).
	if err := modifier.ApplyVHCUEnv(rawSpec, alloc); err != nil {
		return nil, nil, errors.Join(err, cleanup())
	}

	m, err := modifier.NewVHCUModifier(logger, alloc, hookPath)
	if err != nil {
		return nil, nil, errors.Join(err, cleanup())
	}
	return m, cleanup, nil
}

func dynamicVHCUHookPath(l logger.Interface, path string) (string, error) {
	// hcu-cdi-hook only provides CDI hooks; vhcu release belongs to hcu-ctk.
	if filepath.Base(path) != "hcu-ctk" {
		path = "hcu-ctk"
	}
	path = config.ResolveHCUCDIHookPath(l, path)
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("dynamic vHCU requires an executable hcu-ctk release hook at %s: %w", path, err)
	}
	if info.IsDir() || info.Mode()&0111 == 0 {
		return "", fmt.Errorf("dynamic vHCU release hook is not executable: %s", path)
	}
	return path, nil
}
