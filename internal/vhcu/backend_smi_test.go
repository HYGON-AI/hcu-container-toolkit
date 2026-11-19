/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sampleDeviceInfo is captured from a real host (DTK 6.3.31):
//
//	root@host:~# hy-smi virtual --show-device-info
//	Device 0:
//	Physical Device: 0
//	Compute units: 120
//	Global memory: 68702699520 bytes
//	Usage memory: 12582912 bytes (0.02%)
const sampleDeviceInfo = `Device 0:
Physical Device: 0
Compute units: 120
Global memory: 68702699520 bytes
Usage memory: 12582912 bytes (0.02%)
`

const sampleDeviceInfoTwoCards = sampleDeviceInfo + `Device 1:
Physical Device: 1
Compute units: 120
Global memory: 68702699520 bytes
Usage memory: 0 bytes (0.00%)
`

const sampleDriverVersion = `================================= System Management Interface ==================================
================================================================================================
Driver Version: 6.3.31-V1.5.1
================================================================================================
======================================== End of SMI Log ========================================
`

// sampleVDeviceInfo / sampleRemainingInfo are captured from the same host
// (DTK 6.3.31):
//
//	root@host:~# hy-smi virtual --show-vdevice-info
//	Virtual Device 0:
//		Physical Device: 0
//		Compute units: 10
//		Global memory: 1073741824 bytes
//	root@host:~# hy-smi virtual --show-remaining-info -d 0
//	Device 0:
//		Remaining compute units: 110
//		Remaining memories: 67628957696 bytes
//
// Two details here broke the initial implementation on real hardware and are
// therefore pinned by tests: the block header is "Virtual Device N:" (not
// "vDevice Index: N"), and the memory label is the plural "Remaining
// memories". With no vdev present the CLI prints nothing at all.
const sampleVDeviceInfo = `Virtual Device 0:
	Physical Device: 0
	Compute units: 10
	Global memory: 1073741824 bytes
`

const sampleVDeviceInfoTwo = sampleVDeviceInfo + `Virtual Device 1:
	Physical Device: 0
	Compute units: 0
	Global memory: 2147483648 bytes
`

const sampleRemainingInfo = `Device 0:
	Remaining compute units: 110
	Remaining memories: 67628957696 bytes
`

func TestParsePhysicalDevicesRealFormat(t *testing.T) {
	devices, err := parsePhysicalDevices(sampleDeviceInfo)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	d := devices[0]
	if d.Index != 0 || d.ComputeUnits != 120 {
		t.Fatalf("device mismatch: %+v", d)
	}
	// 68702699520 bytes == 65520 MiB.
	if d.MemoryMiB != 65520 {
		t.Fatalf("bytes not converted to MiB: %d", d.MemoryMiB)
	}
	if d.UsageMiB != 12 {
		t.Fatalf("usage memory not converted: %d", d.UsageMiB)
	}
}

func TestParsePhysicalDevicesMultipleCards(t *testing.T) {
	devices, err := parsePhysicalDevices(sampleDeviceInfoTwoCards)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(devices) != 2 || devices[0].Index != 0 || devices[1].Index != 1 {
		t.Fatalf("unexpected devices: %+v", devices)
	}
	if devices[1].UsageMiB != 0 {
		t.Fatalf("second card usage mismatch: %+v", devices[1])
	}
}

func TestParsePhysicalDevicesLegacyFormats(t *testing.T) {
	// Older synthetic layout with explicit index and GiB units must keep
	// working (forward/backward tolerance).
	out := `Device Index: 0
  PciBusId: 0000:81:00.0
  Compute units: 120
  Global memory: 64 GiB
`
	devices, err := parsePhysicalDevices(out)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if devices[0].BusID != "0000:81:00.0" || devices[0].MemoryMiB != 65536 {
		t.Fatalf("legacy format mismatch: %+v", devices[0])
	}
}

func TestParsePhysicalDevicesFailClosed(t *testing.T) {
	if _, err := parsePhysicalDevices("hy-smi: command not found"); err == nil {
		t.Fatal("unparseable output must be an error (fail-closed)")
	}
	if _, err := parsePhysicalDevices("Device 0:\n  Physical Device: 0\n"); err == nil {
		t.Fatal("missing CU/memory fields must be an error")
	}
}

func TestParseVDevices(t *testing.T) {
	vdevs, err := parseVDevices(sampleVDeviceInfo, 0)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(vdevs) != 1 {
		t.Fatalf("expected 1 vdev, got %d: %+v", len(vdevs), vdevs)
	}
	if vdevs[0].Index != 0 || vdevs[0].DeviceIndex != 0 {
		t.Fatalf("vdev 0 mismatch: %+v", vdevs[0])
	}
	// 1073741824 bytes == 1024 MiB.
	if vdevs[0].ComputeUnits != 10 || vdevs[0].MemMiB != 1024 {
		t.Fatalf("vdev resource mismatch: %+v", vdevs[0])
	}

	// Multiple instances on one card must yield distinct indices: locating a
	// freshly created vdev relies on the before/after index diff.
	vdevs, err = parseVDevices(sampleVDeviceInfoTwo, 0)
	if err != nil || len(vdevs) != 2 {
		t.Fatalf("two-vdev output not parsed: %v %v", vdevs, err)
	}
	if vdevs[0].Index != 0 || vdevs[1].Index != 1 {
		t.Fatalf("vdev indices mixed up: %+v", vdevs)
	}
	if vdevs[1].ComputeUnits != 0 || vdevs[1].MemMiB != 2048 {
		t.Fatalf("second vdev mismatch: %+v", vdevs[1])
	}

	// A card without vdevs prints nothing at all; that is a valid empty list
	// (and the common case), not a parse failure.
	if vdevs, err = parseVDevices("", 0); err != nil || len(vdevs) != 0 {
		t.Fatalf("empty output must succeed: %v %v", vdevs, err)
	}
	if vdevs, err = parseVDevices("Total vDevice number: 0\n", 0); err != nil || len(vdevs) != 0 {
		t.Fatalf("zero-count output must succeed: %v %v", vdevs, err)
	}

	// The bare header form "vDevice 3:" must also be recognized: locating the
	// new vdev index depends on it.
	vdevs, err = parseVDevices("vDevice 3:\n  Physical Device: 0\n", 0)
	if err != nil || len(vdevs) != 1 || vdevs[0].Index != 3 {
		t.Fatalf("bare vDevice header not parsed: %v %v", vdevs, err)
	}
}

func TestParseRemaining(t *testing.T) {
	rem, err := parseRemaining(sampleRemainingInfo, 0)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	// 67628957696 bytes == 64496 MiB (65520 total minus the 1024 MiB vdev).
	if rem.ComputeUnits != 110 || rem.MemoryMiB != 64496 {
		t.Fatalf("remaining mismatch: %+v", rem)
	}

	if _, err := parseRemaining("Device 0:\n", 0); err == nil {
		t.Fatal("missing remaining fields must be an error (fail-closed)")
	}
}

// recordingBackend captures the CLI arguments passed to hy-smi and can
// script per-command outputs.
type recordingBackend struct {
	*smiBackend
	args      [][]string
	responses map[string]string
}

func newRecordingBackend() *recordingBackend {
	rb := &recordingBackend{
		smiBackend: &smiBackend{path: "/usr/bin/hy-smi"},
		responses:  make(map[string]string),
	}
	rb.run = func(args ...string) (string, error) {
		rb.args = append(rb.args, args)
		key := strings.Join(args, " ")
		if out, ok := rb.responses[key]; ok {
			return out, nil
		}
		return "", fmt.Errorf("unscripted command: %s", key)
	}
	return rb
}

func TestSmiBackendDestroyUsesSingleVdevFlag(t *testing.T) {
	// The destroy call MUST address a single vdev via -v; the -d form would
	// destroy every vdev on the card including pre-split ones.
	rb := newRecordingBackend()
	rb.responses["virtual --destroy-vdevices -v 3"] = ""
	if err := rb.DestroyVDevice(3); err != nil {
		t.Fatalf("destroy failed: %v", err)
	}
	joined := strings.Join(rb.args[0], " ")
	if !strings.Contains(joined, "-v 3") {
		t.Fatalf("destroy must use -v <idx>, got: %s", joined)
	}
	// Word-wise check: no standalone -d flag (a naive substring check would
	// false-positive on "-destroy-vdevices").
	for _, arg := range rb.args[0] {
		if arg == "-d" {
			t.Fatalf("destroy must never use the -d flag, got: %s", joined)
		}
	}
}

func TestSmiBackendCommands(t *testing.T) {
	rb := newRecordingBackend()
	rb.responses["virtual --create-vdevices 1 --vdevice-compute-units 36 --vdevice-memory-size 8192 -d 0"] = ""
	rb.responses["virtual --stop-vdevices -v 3"] = ""
	rb.responses["virtual --show-remaining-info -d 1"] = sampleRemainingInfo
	rb.responses["virtual --show-vdevice-info -d 2"] = ""

	_ = rb.CreateVDevice(0, 36, 8192)
	joined := strings.Join(rb.args[0], " ")
	for _, want := range []string{"virtual", "--create-vdevices 1", "--vdevice-compute-units 36", "--vdevice-memory-size 8192", "-d 0"} {
		if !strings.Contains(joined, want) {
			t.Errorf("create args missing %q: %s", want, joined)
		}
	}

	_ = rb.StopVDevice(3)
	if joined := strings.Join(rb.args[1], " "); !strings.Contains(joined, "-v 3") {
		t.Errorf("stop must use -v, got: %s", joined)
	}

	_, _ = rb.Remaining(1)
	if joined := strings.Join(rb.args[2], " "); !strings.Contains(joined, "-d 1") {
		t.Errorf("remaining must use -d, got: %s", joined)
	}

	_, _ = rb.ListVDevices(2)
	if joined := strings.Join(rb.args[3], " "); !strings.Contains(joined, "--show-vdevice-info") || !strings.Contains(joined, "-d 2") {
		t.Errorf("list must use --show-vdevice-info -d, got: %s", joined)
	}
}

func TestSmiBackendRemainingPrefersDriverInfo(t *testing.T) {
	rb := newRecordingBackend()
	rb.responses["virtual --show-remaining-info -d 0"] = sampleRemainingInfo

	rem, err := rb.Remaining(0)
	if err != nil {
		t.Fatalf("remaining failed: %v", err)
	}
	if rem.Derived {
		t.Fatal("driver-reported remaining must not be marked derived")
	}
	if rem.ComputeUnits != 110 || rem.MemoryMiB != 64496 {
		t.Fatalf("unexpected remaining: %+v", rem)
	}
}

func TestSmiBackendRemainingFallsBackToUsage(t *testing.T) {
	rb := newRecordingBackend()
	rb.responses["virtual --show-remaining-info -d 0"] = "ERROR: unrecognized arguments: --show-remaining-info"
	rb.responses["virtual --show-device-info"] = sampleDeviceInfo
	// --show-remaining-info is not scripted -> the call fails -> fallback.
	rb.responses["virtual --show-vdevice-info -d 0"] = ""

	rem, err := rb.Remaining(0)
	if err != nil {
		t.Fatalf("derived remaining failed: %v", err)
	}
	if !rem.Derived {
		t.Fatal("fallback remaining must be marked derived")
	}
	if rem.ComputeUnits != 120 {
		t.Fatalf("expected full CU on an empty card, got %d", rem.ComputeUnits)
	}
	// 65520 MiB total - 12 MiB usage.
	if rem.MemoryMiB != 65508 {
		t.Fatalf("expected 65508 MiB remaining, got %d", rem.MemoryMiB)
	}
}

func TestSmiBackendRemainingFallbackSubtractsVDevices(t *testing.T) {
	rb := newRecordingBackend()
	rb.responses["virtual --show-remaining-info -d 0"] = "ERROR: unrecognized arguments: --show-remaining-info"
	rb.responses["virtual --show-device-info"] = sampleDeviceInfo
	rb.responses["virtual --show-vdevice-info -d 0"] = sampleVDeviceInfoTwo

	rem, err := rb.Remaining(0)
	if err != nil {
		t.Fatalf("derived remaining failed: %v", err)
	}
	if rem.ComputeUnits != 110 { // 120 - 10 - 0
		t.Fatalf("expected 110 CU remaining, got %d", rem.ComputeUnits)
	}
	// min(total-usage=65508, total-vdev=65520-1024-2048=62448) -> 62448.
	if rem.MemoryMiB != 62448 {
		t.Fatalf("expected 62448 MiB remaining, got %d", rem.MemoryMiB)
	}
}

func TestSmiBackendExecErrorIncludesOutput(t *testing.T) {
	b := &smiBackend{path: "/nonexistent-hy-smi"}
	b.run = b.exec
	_, err := b.run("virtual", "--show-device-info")
	if err == nil {
		t.Fatal("missing binary must error")
	}
	if !strings.Contains(err.Error(), "hy-smi") {
		t.Errorf("error should name the command: %v", err)
	}
}

func TestDriverVersionParsing(t *testing.T) {
	m := driverVersionRe.FindStringSubmatch(sampleDriverVersion)
	if m == nil || m[1] != "6.3.31-V1.5.1" {
		t.Fatalf("driver version not parsed: %v", m)
	}
	if !versionAtLeast("6.3.31-V1.5.1", minDriverVersion) {
		t.Fatal("6.3.31 must satisfy the 6.2.26 minimum")
	}
	if versionAtLeast("6.2.25", minDriverVersion) {
		t.Fatal("6.2.25 must not satisfy the minimum")
	}
	if !versionAtLeast("6.2.26", minDriverVersion) {
		t.Fatal("the minimum version itself must satisfy the check")
	}
}

func TestToMiB(t *testing.T) {
	cases := []struct {
		value int
		unit  string
		want  int
	}{
		{68702699520, "bytes", 65520},
		{12582912, "bytes", 12},
		{1048576, "B", 1},
		{2048, "KiB", 2},
		{65536, "MiB", 65536},
		{65536, "MB", 65536},
		{65536, "", 65536},
		{64, "GiB", 65536},
		{64, "GB", 64000},
	}
	for _, c := range cases {
		if got := toMiB(fmt.Sprintf("%d", c.value), c.unit); got != c.want {
			t.Errorf("toMiB(%d, %s) = %d, want %d", c.value, c.unit, got, c.want)
		}
	}
}

func TestTruncateForLog(t *testing.T) {
	long := strings.Repeat("x", 300)
	if got := truncateForLog(long); len(got) != 256+3 {
		t.Fatalf("expected truncation, got %d", len(got))
	}
	if got := truncateForLog(fmt.Sprintf("  %s  ", "short")); got != "short" {
		t.Fatalf("expected trimming, got %q", got)
	}
}

func TestVDeviceQueryRejectsUnknownAndErrorOutput(t *testing.T) {
	for _, out := range []string{"ERROR: query failed", "new unsupported output", "Total vDevice number: 1", "Virtual Device 0:\nVirtual Device 0:\n", "Virtual Device 0:\nERROR: partial query failed"} {
		if _, err := parseVDevices(out, 0); err == nil {
			t.Errorf("unsafe query accepted: %q", out)
		}
	}
}

func TestRemainingDoesNotHideQueryFailures(t *testing.T) {
	for _, out := range []string{"ERROR: driver query failed", "unknown output"} {
		rb := newRecordingBackend()
		rb.responses["virtual --show-remaining-info -d 0"] = out
		rb.responses["virtual --show-device-info"] = sampleDeviceInfo
		rb.responses["virtual --show-vdevice-info -d 0"] = ""
		if _, err := rb.Remaining(0); err == nil {
			t.Fatalf("query failure incorrectly fell back to full capacity: %q", out)
		}
		if len(rb.args) != 1 {
			t.Fatal("failed query should not run fallback commands")
		}
	}
}

func TestRemainingFallbackRequiresResourceMetadata(t *testing.T) {
	rb := newRecordingBackend()
	rb.responses["virtual --show-remaining-info -d 0"] = "ERROR: unrecognized arguments: --show-remaining-info"
	rb.responses["virtual --show-device-info"] = sampleDeviceInfo
	rb.responses["virtual --show-vdevice-info -d 0"] = "Virtual Device 0:\nPhysical Device: 0\n"
	if _, err := rb.Remaining(0); err == nil {
		t.Fatal("missing per-instance resources must not be treated as zero")
	}
}

func TestHySmiRejectsFailureWithSuccessfulExit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hy-smi")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nprintf 'ERROR: destroy failed\\n'\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := runHySmi(path, "virtual", "--destroy-vdevices", "-v", "0"); err == nil {
		t.Fatal("driver error text was ignored")
	}
}
