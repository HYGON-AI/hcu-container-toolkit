/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"os"

	"hcu-container-toolkit/internal/query"
)

// HysmiSource identifies how the hy-smi binary was located.
//
// The distinction matters because the CLI and the OCI runtime are started with
// different environments: `hcu-ctk` runs in an interactive shell, while the
// runtime is forked by containerd/docker with the daemon's PATH. A PATH lookup
// can therefore resolve to different binaries in each process, which shows up
// as a "hy-smi virtual ... unknown" failure at container create time.
type HysmiSource string

const (
	// HysmiFromConfig means hcu-container-runtime.vhcu.hysmi-path was set.
	HysmiFromConfig HysmiSource = "config"
	// HysmiFromWellKnown means a known DTK install location matched.
	HysmiFromWellKnown HysmiSource = "well-known"
	// HysmiFromPATH means the binary was found through PATH.
	HysmiFromPATH HysmiSource = "PATH"
	// HysmiNotFound means no usable candidate was found.
	HysmiNotFound HysmiSource = "not found"
)

// hysmiCandidates lists the DTK install locations probed before PATH, in
// priority order. Several DTK trees can coexist on one host (/opt/hyhal,
// /usr/local/hyhal, /opt/rocm); pinning a well-known location keeps the CLI and
// the container runtime on the same binary regardless of the environment each
// process inherits.
var hysmiCandidates = []string{
	"/opt/hyhal/bin/hy-smi",
	"/usr/local/hyhal/bin/hy-smi",
	"/opt/rocm/bin/hy-smi",
}

// pathLookup resolves hy-smi through the process environment. It is a variable
// so tests can inject a fake lookup.
var pathLookup = query.CheckHySmi

// resolveHysmiPath resolves the hy-smi binary: explicit configuration wins,
// then a well-known DTK install location, then PATH.
func resolveHysmiPath(override string) (string, HysmiSource) {
	if override != "" {
		return override, HysmiFromConfig
	}
	for _, candidate := range hysmiCandidates {
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() || info.Mode()&0111 == 0 {
			continue
		}
		return candidate, HysmiFromWellKnown
	}
	if path, err := pathLookup(); err == nil && path != "" {
		return path, HysmiFromPATH
	}
	return "", HysmiNotFound
}
