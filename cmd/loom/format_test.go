package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom/measure"
)

func TestFormatText(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		width       int
		trim, align bool
		want        string
	}{
		{name: "empty", text: "", want: ""},
		{name: "wide padding", text: "界\na", width: 4, want: "界  \na   "},
		{name: "ANSI and reset", text: "\x1b[31ma\x1b[0m\nb", width: 2, want: "\x1b[31ma\x1b[0m \nb "},
		{name: "truncate at cell boundary", text: "界ab", width: 3, want: "界a"},
		{name: "trailing whitespace", text: "a  \nb\t", trim: true, want: "a\nb"},
		{name: "trailing reset", text: "\x1b[31ma\x1b[0m", trim: true, want: "\x1b[31ma\x1b[0m"},
		{name: "ragged boxes", text: "│x│\n│long│", align: true, want: "│x   │\n│long│"},
		{name: "unboxed row untouched", text: "│x│\nloose", align: true, want: "│x│\nloose"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatText(tt.text, tt.width, tt.trim, tt.align)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("formatText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatWidthUsesDisplayColumns(t *testing.T) {
	got, err := formatText("界\nx", 5, false, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(got, "\n") {
		if measure.StringWidth(line) != 5 {
			t.Errorf("width(%q) = %d, want 5", line, measure.StringWidth(line))
		}
	}
}

func TestFormatTextKeepsRegionalIndicatorPairAtWidthBoundary(t *testing.T) {
	got, err := formatText("🇩🇪X", 3, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != "🇩🇪X" {
		t.Fatalf("formatText() = %q, want complete two-column flag and X", got)
	}
}

func TestFormatTextAlignsBoxEdgesAfterFlag(t *testing.T) {
	got, err := formatText("│🇩🇪x│\n│x│", 0, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if want := "│🇩🇪x│\n│x  │"; got != want {
		t.Fatalf("formatText() = %q, want %q", got, want)
	}
}

func TestFormatCommandStdoutAndAtomicWriteMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.ansi")
	if err := os.WriteFile(path, []byte("a\nb\n"), 0640); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := execute([]string{"format", "--width", "3", path}, &out); err != nil {
		t.Fatal(err)
	}
	if out.String() != "a  \nb  \n" {
		t.Fatalf("stdout = %q", out.String())
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(original) != "a\nb\n" {
		t.Fatalf("stdout mode modified file: %q", original)
	}
	out.Reset()
	if err := execute([]string{"format", "-w", "--width", "3", path}, &out); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Fatalf("--write stdout = %q", out.String())
	}
	formatted, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(formatted) != "a  \nb  \n" {
		t.Fatalf("file = %q", formatted)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0640 {
		t.Fatalf("mode = %o, want 640", info.Mode().Perm())
	}
}

func TestFormatInvalidFlags(t *testing.T) {
	path := writeFixture(t, "text")
	for _, args := range [][]string{{"format", "--width", "0", path}, {"format", "--width", "nope", path}, {"format", path, "extra"}} {
		var out bytes.Buffer
		if err := execute(args, &out); err == nil {
			t.Errorf("execute(%q) succeeded", args)
		}
	}
}
