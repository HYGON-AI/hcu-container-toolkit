/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package spec

import (
	"encoding/json"
	"fmt"
	"hcu-container-toolkit/pkg/c3000cdi/transform"
	"io"
	"os"
	"path/filepath"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"
)

type spec struct {
	*specs.Spec
	format          string
	permissions     os.FileMode
	transformOnSave transform.Transformer
}

var _ Interface = (*spec)(nil)

// New creates a new spec with the specified options.
func New(opts ...Option) (Interface, error) {
	return newBuilder(opts...).Build()
}

// Save writes the spec to the specified path and overwrites the file if it exists.
func (s *spec) Save(path string) error {
	s.ContainerEdits.AdditionalGIDs = nil
	if s.transformOnSave != nil {
		err := s.transformOnSave.Transform(s.Raw())
		if err != nil {
			return fmt.Errorf("error applying transform: %w", err)
		}
	}
	path, err := s.normalizePath(path)
	if err != nil {
		return fmt.Errorf("failed to normalize path: %w", err)
	}

	specDir := filepath.Dir(path)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		return fmt.Errorf("failed to create spec directory %q: %w", specDir, err)
	}

	// The upstream CDI library writes JSON without indentation, which hurts readability.
	// For JSON output, write a pretty-printed file ourselves.
	if s.format == FormatJSON {
		if err := s.writePrettyJSON(path); err != nil {
			return err
		}
		if err := os.Chmod(path, s.permissions); err != nil {
			return fmt.Errorf("failed to set permissions on spec file: %w", err)
		}
		return nil
	}

	cache, _ := cdi.NewCache(
		cdi.WithAutoRefresh(false),
		cdi.WithSpecDirs(specDir),
	)
	if err := cache.WriteSpec(s.Raw(), filepath.Base(path)); err != nil {
		return fmt.Errorf("failed to write spec: %w", err)
	}

	if err := os.Chmod(path, s.permissions); err != nil {
		return fmt.Errorf("failed to set permissions on spec file: %w", err)
	}

	return nil
}

func (s *spec) writePrettyJSON(path string) error {
	pretty, err := json.MarshalIndent(s.Raw(), "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal CDI spec as JSON: %w", err)
	}
	pretty = append(pretty, '\n')

	dir := filepath.Dir(path)
	base := filepath.Base(path)

	tmp, err := os.CreateTemp(dir, "."+base+".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp spec file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(pretty); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to write temp spec file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp spec file: %w", err)
	}

	// Best-effort set perms before rename; we also chmod after.
	_ = os.Chmod(tmpName, s.permissions)

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("failed to replace spec file %q: %w", path, err)
	}
	return nil
}

// WriteTo writes the spec to the specified writer.
func (s *spec) WriteTo(w io.Writer) (int64, error) {
	name, err := cdi.GenerateNameForSpec(s.Raw())
	if err != nil {
		return 0, err
	}

	path, _ := s.normalizePath(name)
	tmpFile, err := os.CreateTemp("", "*"+filepath.Base(path))
	if err != nil {
		return 0, err
	}
	defer os.Remove(tmpFile.Name())

	if err := s.Save(tmpFile.Name()); err != nil {
		return 0, err
	}

	err = tmpFile.Close()
	if err != nil {
		return 0, fmt.Errorf("failed to close temporary file: %w", err)
	}

	r, err := os.Open(tmpFile.Name())
	if err != nil {
		return 0, fmt.Errorf("failed to open temporary file: %w", err)
	}
	defer r.Close()

	return io.Copy(w, r)
}

// Raw returns a pointer to the raw spec.
func (s *spec) Raw() *specs.Spec {
	return s.Spec
}

// normalizePath ensures that the specified path has a supported extension
func (s *spec) normalizePath(path string) (string, error) {
	if ext := filepath.Ext(path); ext != ".yaml" && ext != ".json" {
		path += s.extension()
	}

	if filepath.Clean(filepath.Dir(path)) == "." {
		pwd, err := os.Getwd()
		if err != nil {
			return path, fmt.Errorf("failed to get current working directory: %v", err)
		}
		path = filepath.Join(pwd, path)
	}

	return path, nil
}

func (s *spec) extension() string {
	switch s.format {
	case FormatJSON:
		return ".json"
	case FormatYAML:
		return ".yaml"
	}

	return ".yaml"
}
