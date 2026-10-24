// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestSessionListCommandAndAlias(t *testing.T) {
	t.Setenv("LOOM_SOCKET_DIR", t.TempDir())
	for _, args := range [][]string{{"list"}, {"ls"}} {
		var output bytes.Buffer
		command := sessionCommand()
		command.SetOut(&output)
		command.SetArgs(args)
		if err := command.Execute(); err != nil {
			t.Fatalf("session %s: %v", strings.Join(args, " "), err)
		}
		if output.String() != "No active sessions.\n" {
			t.Errorf("session %s output = %q", strings.Join(args, " "), output.String())
		}
	}
}

func TestSessionAttachUsesPersistentSessionFlagBeforeOrAfterSubcommand(t *testing.T) {
	t.Setenv("LOOM_SOCKET_DIR", t.TempDir())
	for _, args := range [][]string{{"-s", "missing", "attach"}, {"attach", "--session", "missing"}, {"a", "-s", "missing"}} {
		command := sessionCommand()
		command.SetArgs(args)
		err := command.Execute()
		if err == nil || !strings.Contains(err.Error(), "missing") {
			t.Errorf("session %s error = %v, want missing session connection error", strings.Join(args, " "), err)
		}
	}
}
