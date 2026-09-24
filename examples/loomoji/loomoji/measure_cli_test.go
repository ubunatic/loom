// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomoji

import (
	"os"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestDebugMeasureCommandParsesAndUsesEnvironmentProfile(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "WezTerm")
	want := TerminalProfile{Term: "xterm-256color", TermProgram: "WezTerm"}
	called := false
	cmd := newCommand(func() error { return nil }, func(profile TerminalProfile) error {
		called = true
		if profile != want {
			t.Errorf("runMeasure profile = %#v, want %#v", profile, want)
		}
		return nil
	})
	cmd.SetArgs([]string{"debug", "--measure"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("debug --measure Execute() error = %v", err)
	}
	if !called {
		t.Fatal("debug --measure did not invoke measure workflow")
	}
}

func TestDebugCommandRequiresMeasureFlagAndNoExtraArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing flag", args: []string{"debug"}},
		{name: "extra argument", args: []string{"debug", "--measure", "extra"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			cmd := newCommand(func() error { return nil }, func(TerminalProfile) error {
				called = true
				return nil
			})
			cmd.SetArgs(tt.args)
			if err := cmd.Execute(); err == nil {
				t.Fatal("Execute() succeeded, want parsing/usage error")
			}
			if called {
				t.Fatal("invalid debug invocation ran the measure workflow")
			}
		})
	}
}

func TestMeasureSessionLoadsProfileAndSkipsAnsweredGlyphs(t *testing.T) {
	dir := t.TempDir()
	profile := TerminalProfile{Term: "screen-256color", TermProgram: "tmux"}
	jsonPath, textPath := profile.Paths(dir)
	store := NewMeasurementStore(profile)
	first := uniqueGlyphs()[0]
	store.Set(Measurement{Glyph: first, ComputedWidth: 2, MeasuredWidth: 2, Answered: true})
	if err := store.Save(jsonPath); err != nil {
		t.Fatalf("Save() existing profile error = %v", err)
	}

	widget, err := newMeasureSession(profile, dir)
	if err != nil {
		t.Fatalf("newMeasureSession() error = %v", err)
	}
	if widget.store.Profile != profile {
		t.Errorf("loaded profile = %#v, want %#v", widget.store.Profile, profile)
	}
	if widget.CurrentGlyph() == first {
		t.Fatalf("session queued already answered glyph %q", first)
	}
	widget.HandleKey(keyText("1"))
	loaded, err := LoadMeasurementStore(jsonPath)
	if err != nil {
		t.Fatalf("LoadMeasurementStore() after answer error = %v", err)
	}
	if loaded.Profile != profile {
		t.Errorf("saved profile = %#v, want %#v", loaded.Profile, profile)
	}
	report, err := os.ReadFile(textPath)
	if err != nil {
		t.Fatalf("read text report: %v", err)
	}
	if !strings.Contains(string(report), widget.glyphs[0]) {
		t.Errorf("text report omitted newly measured glyph %q", widget.glyphs[0])
	}
}

func keyText(text string) loom.KeyEvent {
	return loom.KeyEvent{Text: text}
}
