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
	t.Setenv("VTE_VERSION", "8401")
	want := TerminalProfile{Term: "xterm-256color", TermProgram: "WezTerm", VTEVersion: "8401"}
	called := false
	cmd := newCommand(func() error { return nil }, func(profile TerminalProfile, _ MeasureOptions) error {
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

func TestDebugMeasureReviewFlags(t *testing.T) {
	var got MeasureOptions
	cmd := newCommand(func() error { return nil }, func(_ TerminalProfile, options MeasureOptions) error {
		got = options
		return nil
	})
	cmd.SetArgs([]string{"debug", "--measure", "--review", "--width", "2", "--differs-from-loom", "--has-comment"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("debug review flags Execute() error = %v", err)
	}
	want := MeasureOptions{Review: true, Filter: MeasureFilterWidth2, DiffersFromLoom: true, HasComment: true}
	if got != want {
		t.Errorf("review options = %#v, want %#v", got, want)
	}
}

func TestDebugMeasureRejectsInvalidReviewWidth(t *testing.T) {
	cmd := newCommand(func() error { return nil }, func(TerminalProfile, MeasureOptions) error { return nil })
	cmd.SetArgs([]string{"debug", "--measure", "--review", "--width", "5"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("debug --measure --review --width 5 succeeded, want error")
	}
}

func TestDebugGridCommand(t *testing.T) {
	cmd := newCommand(func() error { return nil }, func(TerminalProfile, MeasureOptions) error { return nil })
	cmd.SetArgs([]string{"debug", "--grid", "-W", "2"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("debug --grid -W 2 Execute() error = %v", err)
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
			cmd := newCommand(func() error { return nil }, func(TerminalProfile, MeasureOptions) error {
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

func TestConfigureMeasurePaneUsesFullTerminalWidth(t *testing.T) {
	pane := &loom.Pane{MaxCols: loom.DefaultMaxCols}
	configureMeasurePane(pane)
	if pane.MaxCols != 0 {
		t.Fatalf("measure pane MaxCols = %d, want 0 (terminal width)", pane.MaxCols)
	}
	if !pane.Resizeable {
		t.Fatal("measure pane Resizeable = false, want true")
	}
	if !pane.DisableDefaultQuit {
		t.Fatal("measure pane DisableDefaultQuit = false, want true")
	}
}

func TestMeasureSessionLoadsProfileAndShowsAllGlyphsByDefault(t *testing.T) {
	dir := t.TempDir()
	profile := TerminalProfile{Term: "screen-256color", TermProgram: "tmux"}
	jsonPath, textPath := profile.Paths(dir)
	store := NewMeasurementStore(profile)
	first := uniqueGlyphs()[0]
	store.Set(Measurement{Glyph: first, ComputedWidth: 2, MeasuredWidth: 2, Answered: true})
	if err := store.Save(jsonPath); err != nil {
		t.Fatalf("Save() existing profile error = %v", err)
	}

	widget, err := newMeasureSession(profile, dir, MeasureOptions{})
	if err != nil {
		t.Fatalf("newMeasureSession() error = %v", err)
	}
	if widget.store.Profile != profile {
		t.Errorf("loaded profile = %#v, want %#v", widget.store.Profile, profile)
	}
	if widget.CurrentGlyph() != first {
		t.Fatalf("session did not show first glyph %q in default all mode", first)
	}

	unassessedWidget, err := newMeasureSession(profile, dir, MeasureOptions{Filter: MeasureFilterUnassessed})
	if err != nil {
		t.Fatalf("newMeasureSession(unassessed) error = %v", err)
	}
	if unassessedWidget.CurrentGlyph() == first {
		t.Fatalf("unassessed session queued already answered glyph %q", first)
	}

	widget.ConsumeKey(keyText("0"))
	widget.ConsumeKey(loom.KeyEvent{Key: "enter"})
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
