/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"hcu-container-toolkit/internal/config"
	"hcu-container-toolkit/internal/info"
	"hcu-container-toolkit/internal/lookup/root"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
)

// Run is an entry point that allows for idiomatic handling of errors
// when calling from the main function.
func (r rt) Run(argv []string) (rerr error) {
	defer func() {
		if rerr != nil {
			r.logger.Errorf("%v", rerr)
		}
	}()

	printVersion := hasVersionFlag(argv)
	if printVersion {
		fmt.Printf("%v version %v\n", "HCU Container Runtime", info.GetVersionString(fmt.Sprintf("spec: %v", specs.Version)))
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}
	r.logger.Update(
		cfg.HCUContainerRuntimeConfig.DebugFilePath,
		cfg.HCUContainerRuntimeConfig.LogLevel,
		argv,
	)
	defer func() {
		if rerr != nil {
			r.logger.Errorf("%v", rerr)
		}
		if err := r.logger.Reset(); err != nil {
			rerr = errors.Join(rerr, fmt.Errorf("failed to reset logger: %v", err))
		}
	}()

	//nolint:staticcheck  // TODO(elezar): We should switch the hcu-container-runtime from using hcu-ctk to using hcu-cdi-hook.
	cfg.HCUCTKConfig.Path = config.ResolveHCUCDIHookPath(r.logger, cfg.HCUCTKConfig.Path)

	// Print the config to the output.
	configJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err == nil {
		r.logger.Debugf("Running with config:\n%v", string(configJSON))
	} else {
		r.logger.Debugf("Running with config:\n%+v", cfg)
	}

	driver := root.New(
		root.WithLogger(r.logger),
		root.WithDriverRoot(""),
	)

	r.logger.Infof("Command line arguments: %v", argv)
	runtime, err := newHCUContainerRuntime(r.logger, cfg, argv, driver)
	if err != nil {
		return fmt.Errorf("failed to create HCU Container Runtime: %v", err)
	}

	if printVersion {
		fmt.Print("\n")
	}
	return runtime.Exec(argv)
}

func (r rt) Errorf(format string, args ...interface{}) {
	r.logger.Errorf(format, args...)
}

// TODO: This should be refactored / combined with parseArgs in logger.
func hasVersionFlag(args []string) bool {
	for i := 0; i < len(args); i++ {
		param := args[i]

		parts := strings.SplitN(param, "=", 2)
		trimmed := strings.TrimLeft(parts[0], "-")
		// If this is not a flag we continue
		if parts[0] == trimmed {
			continue
		}

		// Check the version flag
		if trimmed == "version" {
			return true
		}
	}

	return false
}
