// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package apitest

import (
	"os/exec"
	"strings"
	"testing"
)

func TestSpecDefaultsIndependentOfRoot(t *testing.T) {
	cmd := exec.Command("go", "list", "-deps", "ubunatic.com/loom/specdefaults")
	cmd.Dir = "../.."
	data, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("list defaults dependencies: %v\n%s", err, data)
	}
	for _, dependency := range strings.Fields(string(data)) {
		if dependency == "ubunatic.com/loom" {
			t.Fatal("specdefaults must be importable without the root package")
		}
	}
}
