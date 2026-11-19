/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package oci

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestStateCommandKeepsExactRuntimeRoot(t *testing.T) {
	r := pathRuntime{path: "/usr/bin/runc"}
	for _, tc := range []struct {
		args []string
		root string
	}{
		{[]string{"runtime", "--root", "/run/docker/runtime-runc/moby", "create", "id"}, "/run/docker/runtime-runc/moby"},
		{[]string{"runtime", "--root=/run/containerd/runc/k8s.io", "create", "id"}, "/run/containerd/runc/k8s.io"},
		{[]string{"runtime", "-r", "/run/crun", "create", "id"}, "/run/crun"},
		{[]string{"runtime", "--log", "create", "--root", "/run/podman", "create", "id"}, "/run/podman"},
		{[]string{"runtime", "--root", "relative", "create", "id"}, "relative"},
		{[]string{"runtime", "create", "--bundle", "/bundle", "id"}, ""},
	} {
		command, err := r.StateCommand(tc.args)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{r.path}
		if tc.root != "" {
			abs, err := filepath.Abs(tc.root)
			if err != nil {
				t.Fatal(err)
			}
			want = append(want, "--root", abs)
		}
		want = append(want, "list", "--format", "json")
		if !reflect.DeepEqual(command, want) {
			t.Fatalf("scope changed: got %v want %v", command, want)
		}
	}
	for _, args := range [][]string{{"runtime", "--root"}, {"runtime", "--root=", "create", "id"}} {
		if _, err := r.StateCommand(args); err == nil {
			t.Fatal("invalid runtime root accepted")
		}
	}
}
