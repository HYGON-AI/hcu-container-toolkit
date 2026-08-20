/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package oci

import (
	"fmt"
	"hcu-container-toolkit/internal/logger"

	"github.com/opencontainers/runtime-spec/specs-go"
)

// SpecModifier defines an interface for modifying a (raw) OCI spec
type SpecModifier interface {
	// Modify is a method that accepts a pointer to an OCI Srec and returns an
	// error. The intention is that the function would modify the spec in-place.
	Modify(*specs.Spec) error
}

// Spec defines the operations to be performed on an OCI specification
//
//go:generate moq -stub -out spec_mock.go . Spec
type Spec interface {
	Load() (*specs.Spec, error)
	Flush() error
	Modify(SpecModifier) error
	LookupEnv(string) (string, bool)
}

// NewSpec creates fileSpec based on the command line arguments passed to the
// application using the specified logger.
func NewSpec(logger logger.Interface, args []string) (Spec, error) {
	bundleDir, err := GetBundleDir(args)
	if err != nil {
		return nil, fmt.Errorf("error getting bundle directory: %v", err)
	}
	logger.Debugf("Using bundle directory: %v", bundleDir)

	ociSpecPath := GetSpecFilePath(bundleDir)
	logger.Infof("Using OCI specification file path: %v", ociSpecPath)
	ociSpec := NewFileSpec(ociSpecPath)
	return ociSpec, nil
}
