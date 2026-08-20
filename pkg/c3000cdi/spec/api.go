/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package spec

import (
	"io"

	"tags.cncf.io/container-device-interface/specs-go"
)

const (
	// FormatJSON indicates a JSON output format
	FormatJSON = "json"
	// FormatYAML indicates a YAML output format
	FormatYAML = "yaml"
)

// Interface is the interface for the spec API
type Interface interface {
	io.WriterTo
	Save(string) error
	Raw() *specs.Spec
}
