// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"ubunatic.com/loom/internal/session"
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

func TestDaemonAttachGateFromEnv(t *testing.T) {
	for _, test := range []struct {
		name  string
		set   bool
		value string
		want  bool
	}{
		{name: "gate set", set: true, value: "1", want: true},
		{name: "other value", set: true, value: "0", want: false},
		{name: "unset", want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("LOOM_SESSION_WAIT_ATTACH", "")
			if test.set {
				t.Setenv("LOOM_SESSION_WAIT_ATTACH", test.value)
			} else {
				os.Unsetenv("LOOM_SESSION_WAIT_ATTACH")
			}
			options := daemonSessionOptions(session.Options{Debug: true})
			if options.WaitForAttach != test.want {
				t.Errorf("WaitForAttach = %v, want %v", options.WaitForAttach, test.want)
			}
			if !options.Debug {
				t.Error("daemonSessionOptions dropped existing options")
			}
			if value, ok := os.LookupEnv("LOOM_SESSION_WAIT_ATTACH"); ok {
				t.Errorf("LOOM_SESSION_WAIT_ATTACH = %q after daemon start, want it cleared for tab children", value)
			}
		})
	}
}
