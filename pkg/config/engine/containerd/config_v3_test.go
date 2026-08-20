/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package containerd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml"
)

func TestConfigV3RuntimeLifecycle(t *testing.T) {
	path := writeConfigV3(t, `version = 3

[plugins."io.containerd.cri.v1.images"]
  snapshotter = "overlayfs"

[plugins."io.containerd.cri.v1.runtime".containerd]
  default_runtime_name = "runc"

[plugins."io.containerd.cri.v1.runtime".containerd.runtimes.runc]
  runtime_type = "io.containerd.runc.v2"
`)

	cfg, err := New(WithPath(path))
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.AddRuntime("hcu", "/usr/bin/hcu-container-runtime", true); err != nil {
		t.Fatal(err)
	}
	if got := cfg.DefaultRuntime(); got != "hcu" {
		t.Fatalf("default runtime = %q, want hcu", got)
	}
	if _, err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}

	configured := loadConfigV3(t, path)
	if got := configured.Get("version"); got != int64(3) {
		t.Fatalf("version = %v, want 3", got)
	}
	if got := configured.GetPath([]string{"plugins", "io.containerd.cri.v1.runtime", "containerd", "runtimes", "hcu", "options", "BinaryName"}); got != "/usr/bin/hcu-container-runtime" {
		t.Fatalf("BinaryName = %v, want /usr/bin/hcu-container-runtime", got)
	}
	if got := configured.GetPath([]string{"plugins", "io.containerd.cri.v1.images", "snapshotter"}); got != "overlayfs" {
		t.Fatalf("images snapshotter = %v, want overlayfs", got)
	}

	if err := cfg.RemoveRuntime("hcu"); err != nil {
		t.Fatal(err)
	}
	if got := cfg.DefaultRuntime(); got != "" {
		t.Fatalf("default runtime = %q, want empty", got)
	}
}

func writeConfigV3(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func loadConfigV3(t *testing.T, path string) *toml.Tree {
	t.Helper()

	config, err := toml.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return config
}
