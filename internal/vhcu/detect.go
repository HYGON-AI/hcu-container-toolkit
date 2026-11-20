/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// minDriverVersion is the minimum DTK/driver version supporting
// `hy-smi virtual --create-vdevices`.
const minDriverVersion = "6.2.26"

// Capabilities is the outcome of the environment self-check.
type Capabilities struct {
	// HysmiPath is the resolved hy-smi binary path ("" when not found).
	HysmiPath string
	// HysmiSource reports how HysmiPath was resolved (config, well-known,
	// PATH, not found).
	HysmiSource HysmiSource
	// Root reports whether the current process runs as root; hy-smi
	// virtual requires root, so rootless usage is unsupported.
	Root bool
	// VirtualSupported reports whether `hy-smi virtual` is usable.
	VirtualSupported bool
	// DriverVersion is the detected driver version, if any.
	DriverVersion string
	// DriverVersionOK reports whether the driver meets minDriverVersion.
	// Unknown versions fail open (OK, with a warning) — the virtual
	// subcommand check is the hard gate.
	DriverVersionOK bool
	// DriverVersionKnown reports whether a version could be detected.
	DriverVersionKnown bool
	// Warnings carries non-fatal findings.
	Warnings []string
}

// Detect probes the environment for dynamic vHCU support.
func Detect(hysmiPathOverride string) *Capabilities {
	caps := &Capabilities{Root: os.Geteuid() == 0}

	caps.HysmiPath, caps.HysmiSource = resolveHysmiPath(hysmiPathOverride)
	if caps.HysmiPath == "" {
		caps.Warnings = append(caps.Warnings, "hy-smi not found in a well-known DTK install location or PATH; install DTK or set hcu-container-runtime.vhcu.hysmi-path")
		return caps
	}
	if _, err := os.Stat(caps.HysmiPath); err != nil {
		caps.Warnings = append(caps.Warnings, fmt.Sprintf("hy-smi not accessible at %s: %v", caps.HysmiPath, err))
		return caps
	}
	if caps.HysmiSource == HysmiFromPATH {
		// The container runtime is forked by containerd/docker and inherits
		// the daemon's PATH, not this shell's. A PATH-resolved binary is
		// therefore not guaranteed to be the one used at container create
		// time (a second/older hy-smi earlier in the daemon PATH wins).
		caps.Warnings = append(caps.Warnings, fmt.Sprintf("hy-smi was resolved through PATH (%s): the OCI runtime is started by containerd/docker with a different PATH and may use another binary; set hcu-container-runtime.vhcu.hysmi-path to pin it", caps.HysmiPath))
	}

	if !caps.Root {
		caps.Warnings = append(caps.Warnings, "not running as root: hy-smi virtual requires root; dynamic vHCU split is unsupported in rootless mode")
	}

	// Detect the driver version first so a failing virtual probe still
	// reports the real version in `hcu-ctk vhcu check`.
	caps.DriverVersion, caps.DriverVersionKnown = detectDriverVersion(caps.HysmiPath)

	// Probe the virtual subcommand with a real (cheap) query; this also
	// fail-closes on unparseable output formats.
	backend := NewSmiBackend(caps.HysmiPath)
	if _, err := backend.PhysicalDevices(); err != nil {
		caps.Warnings = append(caps.Warnings, fmt.Sprintf("hy-smi virtual is not usable: %v", err))
		return caps
	}
	caps.VirtualSupported = true

	if !caps.DriverVersionKnown {
		caps.DriverVersionOK = true
		caps.Warnings = append(caps.Warnings, "driver version could not be detected; assuming it meets "+minDriverVersion)
		return caps
	}
	caps.DriverVersionOK = versionAtLeast(caps.DriverVersion, minDriverVersion)
	if !caps.DriverVersionOK {
		caps.Warnings = append(caps.Warnings, fmt.Sprintf("driver version %s is below the minimum %s required for dynamic vHCU split", caps.DriverVersion, minDriverVersion))
	}

	return caps
}

// Usable reports whether dynamic vHCU allocation can run in this environment.
func (c *Capabilities) Usable() bool {
	return c.VirtualSupported && c.DriverVersionOK
}

// Error summarizes why the capability is unusable.
func (c *Capabilities) Error() error {
	if c.Usable() {
		return nil
	}
	return fmt.Errorf("dynamic vHCU split is unavailable: %s", strings.Join(c.Warnings, "; "))
}

// driverVersionRe matches `hy-smi --showdriverversion` output such as
// "Driver Version: 6.3.31-V1.5.1".
var driverVersionRe = regexp.MustCompile(`(?i)driver\s*version\s*:\s*([0-9][0-9A-Za-z._-]*)`)

// detectDriverVersion queries the driver version, preferring
// `hy-smi --showdriverversion` and falling back to the kernel module's
// sysfs version file.
func detectDriverVersion(hysmiPath string) (string, bool) {
	if out, err := runHySmi(hysmiPath, "--showdriverversion"); err == nil {
		if m := driverVersionRe.FindStringSubmatch(out); m != nil {
			return m[1], true
		}
	}
	return readDriverVersion()
}

// readDriverVersion reads the version of the hy*cu kernel module.
func readDriverVersion() (string, bool) {
	matches, err := filepath.Glob("/sys/module/hy*cu/version")
	if err != nil || len(matches) == 0 {
		return "", false
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		return "", false
	}
	version := strings.TrimSpace(string(data))
	if version == "" {
		return "", false
	}
	return version, true
}

// versionAtLeast compares dotted numeric versions segment by segment, so
// "6.3.31-V1.5.1" is treated as 6.3.31.
func versionAtLeast(version, minimum string) bool {
	parse := func(s string) []int {
		var out []int
		for _, part := range strings.Split(strings.TrimSpace(s), ".") {
			n, err := strconv.Atoi(strings.SplitN(part, "-", 2)[0])
			if err != nil {
				return out
			}
			out = append(out, n)
		}
		return out
	}
	v, m := parse(version), parse(minimum)
	for i := 0; i < len(v) || i < len(m); i++ {
		var vi, mi int
		if i < len(v) {
			vi = v[i]
		}
		if i < len(m) {
			mi = m[i]
		}
		if vi != mi {
			return vi > mi
		}
	}
	return true
}
