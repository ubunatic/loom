// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomoji

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestMeasurementStoreRoundTripMergeAndUnmeasured(t *testing.T) {
	profile := TerminalProfile{Term: "xterm-256color", TermProgram: "WezTerm"}
	path := filepath.Join(t.TempDir(), "terminal.json")
	store := NewMeasurementStore(profile)
	store.Merge(Measurement{Glyph: "😀", Codepoints: []string{"U+1F600"}, ComputedWidth: 2, MeasuredWidth: 1, Comment: "narrow here"})
	if err := store.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := LoadMeasurementStore(path)
	if err != nil {
		t.Fatalf("LoadMeasurementStore() error = %v", err)
	}
	if !reflect.DeepEqual(loaded.Profile, profile) {
		t.Errorf("loaded profile = %#v, want %#v", loaded.Profile, profile)
	}
	want := Measurement{Glyph: "😀", Codepoints: []string{"U+1F600"}, ComputedWidth: 2, MeasuredWidth: 1, Comment: "narrow here"}
	if got := loaded.Entries[want.Glyph]; !reflect.DeepEqual(got, want) {
		t.Errorf("loaded entry = %#v, want %#v", got, want)
	}

	loaded.Merge(Measurement{Glyph: "😀", Codepoints: []string{"U+1F600"}, ComputedWidth: 2, MeasuredWidth: 2})
	loaded.Merge(Measurement{Glyph: "🚀", Codepoints: []string{"U+1F680"}, ComputedWidth: 2, MeasuredWidth: 0, Comment: "unsure"})
	if got := loaded.Entries["😀"]; !reflect.DeepEqual(got, want) {
		t.Errorf("merge overwrote saved answer: got %#v, want %#v", got, want)
	}
	if got := loaded.Unmeasured([]string{"😀", "🚀", "🧪"}); !reflect.DeepEqual(got, []string{"🧪"}) {
		t.Errorf("Unmeasured() = %#v, want only 🧪", got)
	}
	if err := loaded.Save(path); err != nil {
		t.Fatalf("Save() after merge error = %v", err)
	}
	loadedAgain, err := LoadMeasurementStore(path)
	if err != nil {
		t.Fatalf("LoadMeasurementStore() after merge error = %v", err)
	}
	if len(loadedAgain.Entries) != 2 {
		t.Errorf("saved entries = %d, want 2", len(loadedAgain.Entries))
	}
}

func TestMeasurementStoreSaveTextReport(t *testing.T) {
	dir := t.TempDir()
	store := NewMeasurementStore(TerminalProfile{Term: "screen", TermProgram: "tmux"})
	store.Merge(Measurement{Glyph: "🇩🇪", Codepoints: []string{"U+1F1E9", "U+1F1EA"}, ComputedWidth: 2, MeasuredWidth: 2, Comment: "flag"})
	jsonPath := filepath.Join(dir, "profile.json")
	if err := store.Save(jsonPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := store.SaveTextReport(filepath.Join(dir, "profile.txt")); err != nil {
		t.Fatalf("SaveTextReport() error = %v", err)
	}
	report, err := os.ReadFile(filepath.Join(dir, "profile.txt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	for _, want := range []string{"TERM=screen", "TERM_PROGRAM=tmux", "🇩🇪", "U+1F1E9 U+1F1EA", "computed=2", "measured=2", "flag"} {
		if !strings.Contains(string(report), want) {
			t.Errorf("text report missing %q: %s", want, report)
		}
	}
}
