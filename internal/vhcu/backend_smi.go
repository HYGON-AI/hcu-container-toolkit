/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// smiCommandTimeout bounds each hy-smi invocation on the container create
// path so a wedged CLI cannot stall container startup indefinitely.
const smiCommandTimeout = 15 * time.Second

// smiBackend implements Backend on top of the `hy-smi virtual` CLI.
//
// The CLI does not return the index of a newly created vdev, so callers use
// ListVDevices diffs to locate it (see manager.go). Output parsing is
// fail-closed: an unparseable field is an error, never a silent default.
//
// Observed `hy-smi virtual --show-device-info` output (DTK 6.3.x):
//
//	Device 0:
//	Physical Device: 0
//	Compute units: 120
//	Global memory: 68702699520 bytes
//	Usage memory: 12582912 bytes (0.02%)
//
// Memory is reported in bytes; that command does not report the PCI BDF, so
// BusID is filled from sysfs (see fillBusIDs).
type smiBackend struct {
	path string
	// run executes hy-smi with the given args and returns combined output.
	// Overridden in tests.
	run func(args ...string) (string, error)
}

// NewSmiBackend creates a Backend bound to the hy-smi binary at path.
func NewSmiBackend(path string) Backend {
	b := &smiBackend{path: path}
	b.run = b.exec
	return b
}

func (b *smiBackend) exec(args ...string) (string, error) {
	return runHySmi(b.path, args...)
}

// runHySmi executes hy-smi with a bounded timeout and returns combined output.
func runHySmi(path string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), smiCommandTimeout)
	defer cancel()

	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		// The binary path is part of the message on purpose: several DTK
		// trees may coexist on one host, and resolving a different (older)
		// binary is the most common cause of an "unknown argument" failure.
		return out.String(), fmt.Errorf("%s %s failed: %v: %s", path, strings.Join(args, " "), err, truncateForLog(out.String()))
	}
	if smiFailureRe.MatchString(out.String()) {
		return out.String(), fmt.Errorf("%s %s reported failure: %s", path, strings.Join(args, " "), truncateForLog(out.String()))
	}
	return out.String(), nil
}

func truncateForLog(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 256 {
		s = s[:256] + "..."
	}
	return s
}

// PhysicalDevices runs `hy-smi virtual --show-device-info`.
func (b *smiBackend) PhysicalDevices() ([]PhysicalDevice, error) {
	out, err := b.run("virtual", "--show-device-info")
	if err != nil {
		return nil, err
	}
	devices, err := parsePhysicalDevices(out)
	if err != nil {
		return nil, fmt.Errorf("failed to parse `hy-smi virtual --show-device-info` output: %v", err)
	}
	fillBusIDs(devices)
	return devices, nil
}

// ListVDevices runs `hy-smi virtual --show-vdevice-info -d <dev>`.
func (b *smiBackend) ListVDevices(dev int) ([]VDeviceInfo, error) {
	out, err := b.run("virtual", "--show-vdevice-info", "-d", strconv.Itoa(dev))
	if err != nil {
		return nil, err
	}
	return parseVDevices(out, dev)
}

// CreateVDevice runs `hy-smi virtual --create-vdevices 1 ...`.
func (b *smiBackend) CreateVDevice(dev, cu, memMiB int) error {
	_, err := b.run("virtual", "--create-vdevices", "1",
		"--vdevice-compute-units", strconv.Itoa(cu),
		"--vdevice-memory-size", strconv.Itoa(memMiB),
		"-d", strconv.Itoa(dev))
	return err
}

// StopVDevice runs `hy-smi virtual --stop-vdevices -v <idx>`.
func (b *smiBackend) StopVDevice(vdev int) error {
	_, err := b.run("virtual", "--stop-vdevices", "-v", strconv.Itoa(vdev))
	return err
}

// DestroyVDevice runs `hy-smi virtual --destroy-vdevices -v <idx>`.
//
// The -v (single vdev) form is mandatory here: the -d form would destroy
// every vdev on the card, including administratively pre-split instances and
// vdevs owned by other containers. This invariant is asserted in
// backend_smi_test.go.
func (b *smiBackend) DestroyVDevice(vdev int) error {
	_, err := b.run("virtual", "--destroy-vdevices", "-v", strconv.Itoa(vdev))
	return err
}

// Remaining runs `hy-smi virtual --show-remaining-info -d <dev>`.
//
// If that subcommand is explicitly unsupported, the
// remaining amounts are derived from `--show-device-info` (total minus the
// driver-reported usage memory) and the per-vdev resources of existing
// instances. Derived values are marked as such so callers can surface the
// degraded capacity view instead of silently trusting it.
func (b *smiBackend) Remaining(dev int) (Remaining, error) {
	if rem, err := b.driverRemaining(dev); err == nil {
		return rem, nil
	} else if !unsupportedRemaining(err) {
		return Remaining{}, err
	}
	devices, err := b.PhysicalDevices()
	if err != nil {
		return Remaining{}, fmt.Errorf("failed to derive remaining capacity: %v", err)
	}
	for _, d := range devices {
		if d.Index == dev {
			return b.deriveRemaining(d)
		}
	}
	return Remaining{}, fmt.Errorf("device %d not found while deriving remaining capacity", dev)
}

// RemainingForDevice implements deviceAwareRemaining so the derived fallback
// can reuse an already-fetched PhysicalDevice instead of re-running
// `--show-device-info` once per card.
func (b *smiBackend) RemainingForDevice(dev PhysicalDevice) (Remaining, error) {
	if rem, err := b.driverRemaining(dev.Index); err == nil {
		return rem, nil
	} else if !unsupportedRemaining(err) {
		return Remaining{}, err
	}
	return b.deriveRemaining(dev)
}

// driverRemaining reads the authoritative remaining amounts.
func (b *smiBackend) driverRemaining(dev int) (Remaining, error) {
	out, err := b.run("virtual", "--show-remaining-info", "-d", strconv.Itoa(dev))
	if err != nil {
		return Remaining{}, err
	}
	return parseRemaining(out, dev)
}

// deriveRemaining computes a conservative remaining view without the
// optional --show-remaining-info subcommand.
func (b *smiBackend) deriveRemaining(target PhysicalDevice) (Remaining, error) {
	if !target.UsageKnown {
		return Remaining{}, fmt.Errorf("device %d does not report usage memory; cannot safely derive capacity", target.Index)
	}
	rem := Remaining{
		DeviceIndex:  target.Index,
		ComputeUnits: target.ComputeUnits,
		MemoryMiB:    target.MemoryMiB - target.UsageMiB,
		Derived:      true,
	}

	// Subtract CU/memory consumed by existing vdev instances when the
	// per-instance figures are available. Usage memory from the driver
	// already accounts for vdev allocations, so only CU is subtracted here;
	// the memory is taken from whichever source is more conservative.
	vdevs, err := b.ListVDevices(target.Index)
	if err != nil {
		return Remaining{}, err
	}
	usedCU, usedMem := 0, 0
	for _, v := range vdevs {
		if v.DeviceIndex != target.Index {
			continue
		}
		if !v.ResourcesKnown {
			return Remaining{}, fmt.Errorf("vHCU %d lacks resource metadata; cannot safely derive capacity", v.Index)
		}
		usedCU += v.ComputeUnits
		usedMem += v.MemMiB
	}
	rem.ComputeUnits -= usedCU
	if target.MemoryMiB-usedMem < rem.MemoryMiB {
		rem.MemoryMiB = target.MemoryMiB - usedMem
	}

	if rem.ComputeUnits < 0 {
		rem.ComputeUnits = 0
	}
	if rem.MemoryMiB < 0 {
		rem.MemoryMiB = 0
	}
	return rem, nil
}

// Field patterns. Observed on DTK 6.3.31:
//
//	Device 0:                      (physical device block)
//	Virtual Device 0:              (vdev block)
//		Physical Device: 0
//		Compute units: 10
//		Global memory: 1073741824 bytes
//	Remaining compute units: 110
//	Remaining memories: 67628957696 bytes
//
// Older spellings ("Device Index:", "vDevice Index:", "MiB" units) are kept
// for forward/backward tolerance. Each parse function fails closed when a
// required field is missing.
var (
	smiFailureRe     = regexp.MustCompile(`(?im)^\s*(?:error\b|failed\b|failure\b)`)
	zeroVDevicesRe   = regexp.MustCompile(`(?i)^\s*total\s+vdevice\s+number\s*:\s*0\s*$`)
	deviceHeaderRe   = regexp.MustCompile(`(?i)^\s*device\s+([0-9]+)\s*:`)
	deviceIndexRe    = regexp.MustCompile(`(?i)^\s*device\s+index\s*:\s*([0-9]+)`)
	physicalDeviceRe = regexp.MustCompile(`(?i)^\s*physical\s+device\s*:\s*([0-9]+)`)
	vdevVirtualRe    = regexp.MustCompile(`(?i)^\s*virtual\s+device\s+([0-9]+)\s*:`)
	vdevIndexRe      = regexp.MustCompile(`(?i)^\s*vdevice\s+(?:index|id)\s*:\s*([0-9]+)`)
	vdevHeaderRe     = regexp.MustCompile(`(?i)^\s*vdevice\s+([0-9]+)\s*:`)
	pciBusIDRe       = regexp.MustCompile(`(?i)^\s*pci\s*bus\s*id\s*:\s*([0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-9a-fA-F])`)
	computeUnitsRe   = regexp.MustCompile(`(?i)^\s*(?:compute\s*units?|cu_count|computeunits?)\s*:\s*([0-9]+)`)
	memoryValueRe    = regexp.MustCompile(`(?i)^\s*(?:global\s*memory|memory\s*(?:size|total)?|memorysize)\s*:\s*([0-9]+)\s*([a-zA-Z]*)`)
	usageMemoryRe    = regexp.MustCompile(`(?i)^\s*usage\s*memory\s*:\s*([0-9]+)\s*([a-zA-Z]*)`)
	remainingCURe    = regexp.MustCompile(`(?i)^\s*(?:remaining|free|available)\s*(?:compute\s*units?|cu_count|cus?)\s*:\s*([0-9]+)`)
	remainingMemRe   = regexp.MustCompile(`(?i)^\s*(?:remaining|free|available)\s*(?:global\s*)?memor(?:y|ies)\s*(?:size)?\s*:\s*([0-9]+)\s*([a-zA-Z]*)`)
)

func isDeviceBlockStart(line string) (int, bool) {
	if m := deviceHeaderRe.FindStringSubmatch(line); m != nil {
		idx, _ := strconv.Atoi(m[1])
		return idx, true
	}
	if m := deviceIndexRe.FindStringSubmatch(line); m != nil {
		idx, _ := strconv.Atoi(m[1])
		return idx, true
	}
	return 0, false
}

// vdevBlockStart recognizes the vdev block headers across DTK releases:
// "Virtual Device 3:" (observed on 6.3.31), "vDevice Index: 3", and the
// bare "vDevice 3:".
func vdevBlockStart(line string) (int, bool) {
	if m := vdevVirtualRe.FindStringSubmatch(line); m != nil {
		idx, _ := strconv.Atoi(m[1])
		return idx, true
	}
	if m := vdevIndexRe.FindStringSubmatch(line); m != nil {
		idx, _ := strconv.Atoi(m[1])
		return idx, true
	}
	if m := vdevHeaderRe.FindStringSubmatch(line); m != nil {
		idx, _ := strconv.Atoi(m[1])
		return idx, true
	}
	return 0, false
}

// parsePhysicalDevices parses `--show-device-info` output into device blocks.
func parsePhysicalDevices(out string) ([]PhysicalDevice, error) {
	var devices []PhysicalDevice
	var cur *PhysicalDevice

	flush := func() {
		if cur != nil {
			devices = append(devices, *cur)
			cur = nil
		}
	}

	for _, line := range strings.Split(out, "\n") {
		if idx, ok := isDeviceBlockStart(line); ok {
			flush()
			cur = &PhysicalDevice{Index: idx}
			continue
		}
		if cur == nil {
			continue
		}
		// The physical index is authoritative when both are present.
		if m := physicalDeviceRe.FindStringSubmatch(line); m != nil {
			cur.Index, _ = strconv.Atoi(m[1])
			continue
		}
		if m := pciBusIDRe.FindStringSubmatch(line); m != nil {
			cur.BusID = m[1]
			continue
		}
		if m := computeUnitsRe.FindStringSubmatch(line); m != nil {
			cur.ComputeUnits, _ = strconv.Atoi(m[1])
			continue
		}
		if m := usageMemoryRe.FindStringSubmatch(line); m != nil {
			cur.UsageMiB = toMiB(m[1], m[2])
			cur.UsageKnown = true
			continue
		}
		if m := memoryValueRe.FindStringSubmatch(line); m != nil {
			cur.MemoryMiB = toMiB(m[1], m[2])
			continue
		}
	}
	flush()

	if len(devices) == 0 {
		return nil, fmt.Errorf("no device entries found in output: %s", truncateForLog(out))
	}
	for _, d := range devices {
		if d.ComputeUnits <= 0 || d.MemoryMiB <= 0 {
			return nil, fmt.Errorf("incomplete entry for device %d (CU=%d, MiB=%d)", d.Index, d.ComputeUnits, d.MemoryMiB)
		}
	}
	return devices, nil
}

// parseVDevices parses `--show-vdevice-info` output. An empty output is a
// valid "no vdevs on this card" result, not an error. CU/memory are captured
// opportunistically (used by the Remaining fallback) and may be zero when a
// DTK release does not print them.
func parseVDevices(out string, dev int) ([]VDeviceInfo, error) {
	if smiFailureRe.MatchString(out) {
		return nil, fmt.Errorf("vHCU query failed: %s", truncateForLog(out))
	}
	var vdevs []VDeviceInfo
	var cur *VDeviceInfo
	var cuKnown, memKnown bool

	flush := func() {
		if cur != nil {
			cur.ResourcesKnown = cuKnown && memKnown
			if cur.DeviceIndex == Unset {
				cur.DeviceIndex = dev
			}
			vdevs = append(vdevs, *cur)
			cur = nil
		}
	}

	for _, line := range strings.Split(out, "\n") {
		if idx, ok := vdevBlockStart(line); ok {
			flush()
			cur = &VDeviceInfo{Index: idx, DeviceIndex: Unset}
			cuKnown, memKnown = false, false
			continue
		}
		if cur == nil {
			continue
		}
		if m := physicalDeviceRe.FindStringSubmatch(line); m != nil {
			cur.DeviceIndex, _ = strconv.Atoi(m[1])
			continue
		}
		if m := deviceIndexRe.FindStringSubmatch(line); m != nil {
			cur.DeviceIndex, _ = strconv.Atoi(m[1])
			continue
		}
		if m := computeUnitsRe.FindStringSubmatch(line); m != nil {
			cur.ComputeUnits, _ = strconv.Atoi(m[1])
			cuKnown = true
			continue
		}
		if m := memoryValueRe.FindStringSubmatch(line); m != nil {
			cur.MemMiB = toMiB(m[1], m[2])
			memKnown = true
			continue
		}
	}
	flush()
	if len(vdevs) == 0 && strings.TrimSpace(out) != "" && !zeroVDevicesRe.MatchString(strings.TrimSpace(out)) {
		return nil, fmt.Errorf("unrecognized vHCU query output: %s", truncateForLog(out))
	}
	seen := make(map[int]bool)
	for _, v := range vdevs {
		if seen[v.Index] {
			return nil, fmt.Errorf("duplicate vHCU index %d in driver output", v.Index)
		}
		seen[v.Index] = true
	}

	return vdevs, nil
}

func unsupportedRemaining(err error) bool {
	message := strings.ToLower(err.Error())
	for _, text := range []string{"unrecognized arguments", "unrecognized option", "unknown option", "not supported", "not implemented", "invalid choice"} {
		if strings.Contains(message, text) {
			return true
		}
	}
	return false
}

// parseRemaining parses `--show-remaining-info` output for a single device.
func parseRemaining(out string, dev int) (Remaining, error) {
	r := Remaining{DeviceIndex: dev, ComputeUnits: Unset, MemoryMiB: Unset}

	for _, line := range strings.Split(out, "\n") {
		if m := remainingCURe.FindStringSubmatch(line); m != nil && r.ComputeUnits == Unset {
			r.ComputeUnits, _ = strconv.Atoi(m[1])
			continue
		}
		if m := remainingMemRe.FindStringSubmatch(line); m != nil && r.MemoryMiB == Unset {
			r.MemoryMiB = toMiB(m[1], m[2])
			continue
		}
	}

	if r.ComputeUnits == Unset || r.MemoryMiB == Unset {
		return r, fmt.Errorf("no remaining CU/memory entries found in output: %s", truncateForLog(out))
	}
	return r, nil
}

// toMiB converts a "value unit" pair to MiB. hy-smi reports global memory in
// bytes on DTK 6.3.x; MiB/MB and GiB/GB are also accepted.
func toMiB(value, unit string) int {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "BYTES", "B":
		return n / (1024 * 1024)
	case "KIB", "KB":
		return n / 1024
	case "GIB":
		return n * 1024
	case "GB":
		return n * 1000
	default:
		// MiB, MB, or unitless: assume MiB.
		return n
	}
}

// fillBusIDs maps physical device indices to PCI BDFs using the same sysfs
// scan and ordering as the whole-card mount path (internal/modifier). The
// `--show-device-info` output does not carry the BDF. When the sysfs list
// size disagrees with the driver-reported device count the mapping is
// ambiguous, so BusIDs are left empty rather than guessed (they are used for
// display only; mounting derives the parent
// BDF from /etc/vdev/vdev<N>.conf).
func fillBusIDs(devices []PhysicalDevice) {
	needsFill := false
	for _, d := range devices {
		if d.BusID == "" {
			needsFill = true
			break
		}
	}
	if !needsFill {
		return
	}

	busIDs := sysfsBusIDs()
	if len(busIDs) != len(devices) {
		return
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i].Index < devices[j].Index })
	for i := range devices {
		if devices[i].BusID == "" {
			devices[i].BusID = busIDs[i]
		}
	}
}

// sysfsBusIDs returns the PCI BDFs of HCU devices in sorted order.
func sysfsBusIDs() []string {
	patterns := []string{
		"/sys/module/hy*cu/drivers/pci:hy*cu/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*",
		"/sys/module/hy*cu/drivers/pci:amdgpu/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*",
	}
	var busIDs []string
	seen := make(map[string]bool)
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		sort.Strings(matches)
		for _, path := range matches {
			base := filepath.Base(path)
			if !seen[base] {
				seen[base] = true
				busIDs = append(busIDs, base)
			}
		}
		if len(busIDs) > 0 {
			break
		}
	}
	return busIDs
}
