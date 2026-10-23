// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package clitest

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := RunTests(m)
	if loomBinary.dir != "" {
		if _, err := os.Stat(loomBinary.dir); !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "shared binary directory was not removed: %v\n", err)
			code = 1
		}
	}
	os.Exit(code)
}

func TestBuildLoomBinaryShared(t *testing.T) {
	if testing.Short() {
		t.Skip("subprocess integration test")
	}
	var first string
	t.Run("first caller", func(t *testing.T) {
		first = BuildLoomBinary(t)
	})
	t.Run("after first caller cleanup", func(t *testing.T) {
		if _, err := os.Stat(first); err != nil {
			t.Fatalf("shared binary must outlive its first caller: %v", err)
		}
		if got := BuildLoomBinary(t); got != first {
			t.Fatalf("binary path changed: got %q, want %q", got, first)
		}
	})
	for range 4 {
		t.Run("parallel caller", func(t *testing.T) {
			t.Parallel()
			if got := BuildLoomBinary(t); got != first {
				t.Errorf("binary path changed: got %q, want %q", got, first)
			}
		})
	}
}
