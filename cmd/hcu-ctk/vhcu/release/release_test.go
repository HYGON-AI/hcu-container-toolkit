/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package release

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogRelease(t *testing.T) {
	dir := t.TempDir()
	old := releaseLogPath
	releaseLogPath = filepath.Join(dir, "release.log")
	t.Cleanup(func() { releaseLogPath = old })

	logRelease("ctr-ok", "alloc-ok", nil)
	logRelease("ctr-fail", "alloc-fail", errors.New("device busy"))

	data, err := os.ReadFile(releaseLogPath)
	if err != nil {
		t.Fatalf("log file not written: %v", err)
	}
	text := string(data)
	for _, want := range []string{
		"container=ctr-ok allocation=alloc-ok ok",
		"container=ctr-fail allocation=alloc-fail failed: device busy",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("log missing %q, got:\n%s", want, text)
		}
	}
}

func TestLogReleaseNeverPanicsOnUnwritablePath(t *testing.T) {
	old := releaseLogPath
	releaseLogPath = "/nonexistent-directory/release.log"
	t.Cleanup(func() { releaseLogPath = old })

	// A logging failure must not affect the release result.
	logRelease("ctr", "alloc", nil)
}
