/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"os"
	"path/filepath"
	"testing"
)

// stubResolution replaces the candidate list and the PATH lookup for one test
// and restores them afterwards.
func stubResolution(t *testing.T, candidates []string, lookup func() (string, error)) {
	t.Helper()
	oldCandidates, oldLookup := hysmiCandidates, pathLookup
	hysmiCandidates, pathLookup = candidates, lookup
	t.Cleanup(func() {
		hysmiCandidates, pathLookup = oldCandidates, oldLookup
	})
}

func noLookup() (string, error) { return "", nil }

func TestResolveHysmiPathPrefersConfig(t *testing.T) {
	stubResolution(t, nil, func() (string, error) { return "/usr/bin/hy-smi", nil })

	path, source := resolveHysmiPath("/custom/hy-smi")
	if path != "/custom/hy-smi" || source != HysmiFromConfig {
		t.Fatalf("config override must win, got %q (%s)", path, source)
	}
}

func TestResolveHysmiPathPrefersWellKnownOverPATH(t *testing.T) {
	// Two DTK trees plus a different binary on PATH: the well-known install
	// location must win, otherwise the CLI and the containerd-spawned runtime
	// can end up on different binaries.
	dir := t.TempDir()
	candidate := filepath.Join(dir, "hy-smi")
	if err := os.WriteFile(candidate, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	stubResolution(t, []string{filepath.Join(dir, "missing"), candidate}, func() (string, error) {
		return "/usr/bin/hy-smi", nil
	})

	path, source := resolveHysmiPath("")
	if path != candidate {
		t.Fatalf("expected the well-known candidate %q, got %q", candidate, path)
	}
	if source != HysmiFromWellKnown {
		t.Fatalf("expected source %q, got %q", HysmiFromWellKnown, source)
	}
}

func TestResolveHysmiPathSkipsDirectories(t *testing.T) {
	dir := t.TempDir()
	stubResolution(t, []string{dir}, noLookup)

	if path, source := resolveHysmiPath(""); path != "" || source != HysmiNotFound {
		t.Fatalf("a directory must not be used as hy-smi, got %q (%s)", path, source)
	}
}

func TestResolveHysmiPathSkipsNonExecutable(t *testing.T) {
	dir := t.TempDir()
	plain := filepath.Join(dir, "hy-smi")
	if err := os.WriteFile(plain, []byte("not a program"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	stubResolution(t, []string{plain}, func() (string, error) { return "/usr/bin/hy-smi", nil })

	path, source := resolveHysmiPath("")
	if path != "/usr/bin/hy-smi" || source != HysmiFromPATH {
		t.Fatalf("a non-executable file must be skipped, got %q (%s)", path, source)
	}
}

func TestResolveHysmiPathFallsBackToPATH(t *testing.T) {
	stubResolution(t, []string{filepath.Join(t.TempDir(), "missing")}, func() (string, error) {
		return "/usr/bin/hy-smi", nil
	})

	path, source := resolveHysmiPath("")
	if path != "/usr/bin/hy-smi" || source != HysmiFromPATH {
		t.Fatalf("PATH fallback failed, got %q (%s)", path, source)
	}
}

func TestResolveHysmiPathNotFound(t *testing.T) {
	stubResolution(t, nil, noLookup)

	if path, source := resolveHysmiPath(""); path != "" || source != HysmiNotFound {
		t.Fatalf("expected not-found, got %q (%s)", path, source)
	}
}

func TestDetectRecordsConfiguredSource(t *testing.T) {
	// A configured path that does not exist must still be reported as coming
	// from config, so the operator can tell a typo from a lookup problem.
	caps := Detect("/nonexistent/hy-smi")
	if caps.HysmiSource != HysmiFromConfig {
		t.Fatalf("expected config source, got %q", caps.HysmiSource)
	}
	if caps.Usable() {
		t.Fatal("a missing binary must not be reported as usable")
	}
}
