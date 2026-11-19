/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"testing"

	"hcu-container-toolkit/internal/config/image"
)

func envMap(m map[string]string) func(string) string {
	return func(key string) string {
		return m[key]
	}
}

func TestParseRequestNotEnabled(t *testing.T) {
	req, err := ParseRequest(envMap(nil))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req != nil {
		t.Fatalf("expected nil request, got %+v", req)
	}
}

func TestParseRequestCoresAndMem(t *testing.T) {
	req, err := ParseRequest(envMap(map[string]string{
		image.EnvVarVHCUCores: "30",
		image.EnvVarVHCUMem:   "8192",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Cores != 30 || req.MemMiB != 8192 || req.Device != "" {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestParseRequestCoresWithoutMemRejected(t *testing.T) {
	// HCU_VHCU_MEM is mandatory: cores alone cannot express a split and must
	// fail before any vdev is created.
	_, err := ParseRequest(envMap(map[string]string{
		image.EnvVarVHCUCores: "30",
	}))
	if err == nil {
		t.Fatal("cores without mem must be rejected")
	}
}

func TestParseRequestMemOnly(t *testing.T) {
	req, err := ParseRequest(envMap(map[string]string{
		image.EnvVarVHCUMem: "4096",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Cores != 0 || req.MemMiB != 4096 {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestParseRequestInvalidValues(t *testing.T) {
	cases := []map[string]string{
		{image.EnvVarVHCUDevice: "0"},
		{image.EnvVarVHCUCores: "101"},
		{image.EnvVarVHCUCores: "-1"},
		{image.EnvVarVHCUCores: "abc"},
		{image.EnvVarVHCUCores: "30"}, // cores without mem
		{image.EnvVarVHCUMem: "0"},
		{image.EnvVarVHCUMem: "-8"},
		{image.EnvVarVHCUMem: "8gib"},
		{image.EnvVarVHCUCores: "50", image.EnvVarVHCUDevice: "not-a-bdf"},
		{image.EnvVarVHCUCores: "50", image.EnvVarVHCUDevice: "0000:81:00.0"},
		{image.EnvVarVHCUCores: "50", image.EnvVarVHCUDevice: "0,1"},
		{image.EnvVarVHCUCores: "50", image.EnvVarVHCUDevice: "-1"},
		{image.EnvVarVHCUCores: "50", image.EnvVarVHCUDevice: "+1"},
		{image.EnvVarVHCUCores: "50", image.EnvVarVHCUDevice: "1 2"},
		{image.EnvVarVHCUCores: "50", image.EnvVarVHCUDevice: "9999999999999999999999999999999"},
		{image.EnvVarVHCUCores: "50", image.EnvVarVHCUDevice: "0000:81:0.0"},
	}
	for _, env := range cases {
		if _, err := ParseRequest(envMap(env)); err == nil {
			t.Fatalf("expected error for env %v", env)
		}
	}
}

func TestParseRequestDeviceForms(t *testing.T) {
	for _, device := range []string{"0", "3", "10"} {
		req, err := ParseRequest(envMap(map[string]string{
			image.EnvVarVHCUCores:  "30",
			image.EnvVarVHCUMem:    "8192",
			image.EnvVarVHCUDevice: device,
		}))
		if err != nil {
			t.Fatalf("device %q should be accepted: %v", device, err)
		}
		if req.Device != device {
			t.Fatalf("device not preserved: %+v", req)
		}
	}
}

func TestParseRequestMutualExclusions(t *testing.T) {
	cases := []map[string]string{
		{image.EnvVarVHCUCores: "30", image.EnvVarHCUVisibleDevices: "0"},
		{image.EnvVarVHCUMem: "1024", image.EnvVarVHCUVisibleDevices: "0"},
		{image.EnvVarVHCUMem: "1024", image.EnvVarMIGVisibleDevices: "MIG-abc"},
		{image.EnvVarVHCUMem: "1024", image.EnvVarHCUVisibleDevices: "all"},
		{image.EnvVarVHCUCores: "30", image.EnvVarNvidiaVisibleDevices: "0"},
		{image.EnvVarVHCUCores: "30", image.EnvVarDevicePluginMode: "hami"},
		{image.EnvVarVHCUCores: "30", image.EnvVarDevicePluginMode: "HAMI"},
	}
	for _, env := range cases {
		if _, err := ParseRequest(envMap(env)); err == nil {
			t.Fatalf("expected mutual-exclusion error for env %v", env)
		}
	}

	// Non-hami device-plugin strategies stay compatible (toolkit is the
	// mount-only role there, which does not conflict with a local split).
	if _, err := ParseRequest(envMap(map[string]string{
		image.EnvVarVHCUCores:        "30",
		image.EnvVarVHCUMem:          "8192",
		image.EnvVarDevicePluginMode: "vhcu",
	})); err != nil {
		t.Fatalf("vhcu strategy should not conflict: %v", err)
	}
}

func TestCoresToComputeUnits(t *testing.T) {
	cases := []struct {
		pct, total, want int
	}{
		{0, 120, 0},   // memory-only split
		{30, 120, 36}, // exact
		{25, 120, 30}, // exact
		{1, 120, 2},   // ceil(1.2)
		{33, 3, 1},    // ceil(0.99) clamps to 1
		{100, 120, 120},
		{50, 7, 4}, // ceil(3.5)
	}
	for _, c := range cases {
		got := CoresToComputeUnits(c.pct, c.total)
		if got != c.want {
			t.Errorf("CoresToComputeUnits(%d, %d) = %d, want %d", c.pct, c.total, got, c.want)
		}
	}
}

func TestIsFullCardRequest(t *testing.T) {
	if !IsFullCardRequest(100, 65536, 65536) {
		t.Error("100% + full memory must be a full-card request")
	}
	if !IsFullCardRequest(100, 70000, 65536) {
		t.Error("100% + over-full memory must be a full-card request")
	}
	if IsFullCardRequest(100, 8192, 65536) {
		t.Error("100% cores with partial memory is NOT a full-card request (plugin creates a vdev)")
	}
	if IsFullCardRequest(50, 65536, 65536) {
		t.Error("partial cores is never a full-card request")
	}
}
