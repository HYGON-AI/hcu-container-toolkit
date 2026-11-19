/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package vhcu

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRuntimeStatusRequiresValidScopedResult(t *testing.T) {
	for _, tc := range []struct {
		out  string
		want ContainerStatus
		fail bool
	}{
		{`[]`, ContainerAbsent, false},
		{`[{"id":"other","status":"running"}]`, ContainerAbsent, false},
		{`[{"id":"owner","status":"stopped"}]`, ContainerStopped, false},
		{`[{"id":"owner","status":"running"}]`, ContainerLive, false},
		{`[{"id":"owner","status":"created"}]`, ContainerLive, false},
		{`[{"id":"owner","status":"paused"}]`, ContainerLive, false},
		{`[{"id":"owner","status":"unknown"}]`, ContainerUnknown, true},
		{`null`, ContainerAbsent, false},
		{`ERROR`, ContainerUnknown, true},
		{`{}`, ContainerUnknown, true},
		{`[{"id":"owner","status":"stopped"},{"id":"owner","status":"running"}]`, ContainerUnknown, true},
		{`[{"status":"running"}]`, ContainerUnknown, true},
	} {
		t.Run(tc.out, func(t *testing.T) {
			status, err := parseRuntimeStatus([]byte(tc.out), "owner")
			if status != tc.want || (err != nil) != tc.fail {
				t.Fatalf("status=%v err=%v", status, err)
			}
		})
	}
}

func TestProbeFailureNeverMeansAbsent(t *testing.T) {
	for _, command := range [][]string{nil, {"runc", "list"}, {"/does-not-exist/runc", "list"}} {
		status, err := ProbeRuntime(&State{ContainerID: "owner", RuntimeCommand: command})
		if status != ContainerUnknown || err == nil {
			t.Fatalf("unavailable owner was considered absent: %v %v", status, err)
		}
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "runtime")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf '[]\\n'\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	status, err := ProbeRuntime(&State{ContainerID: "owner", RuntimeCommand: []string{script, "list", "--format", "json"}})
	if status != ContainerUnknown || err == nil {
		t.Fatalf("failed runtime with partial output was trusted: %v %v", status, err)
	}
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf '[]\\n'\nprintf 'cannot load container\\n' >&2\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	status, err = ProbeRuntime(&State{ContainerID: "owner", RuntimeCommand: []string{script, "list", "--format", "json"}})
	if status != ContainerUnknown || err == nil {
		t.Fatalf("incomplete successful list was trusted: %v %v", status, err)
	}
}

func TestProcessStartTimeHandlesParenthesesAndSpaces(t *testing.T) {
	fields := []string{}
	for i := 3; i <= 22; i++ {
		fields = append(fields, strconv.Itoa(i))
	}
	start, err := parseProcessStartTime("123 (runtime ) with spaces) " + strings.Join(fields, " "))
	if err != nil || start != "22" {
		t.Fatalf("incorrect process identity: %q %v", start, err)
	}
	if _, err := parseProcessStartTime("123 (bad) S 1"); err == nil {
		t.Fatal("incomplete process identity accepted")
	}
}
