package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom/measure"
)

func TestFrameTextStyles(t *testing.T) {
	tests := []struct{ style, top, mid, bottom string }{
		{"single", "┌─┐", "│x│", "└─┘"},
		{"double", "╔═╗", "║x║", "╚═╝"},
		{"rounded", "╭─╮", "│x│", "╰─╯"},
		{"heavy", "┏━┓", "┃x┃", "┗━┛"},
	}
	for _, tt := range tests {
		t.Run(tt.style, func(t *testing.T) {
			got, err := frameText("x", tt.style, "", 0, "")
			if err != nil {
				t.Fatal(err)
			}
			if got != strings.Join([]string{tt.top, tt.mid, tt.bottom}, "\n") {
				t.Fatalf("frameText() = %q", got)
			}
		})
	}
}

func TestFrameTextPaddingWideAndEmpty(t *testing.T) {
	tests := []struct {
		name, input, want string
		padding           int
	}{
		{"padding zero", "界", "┌──┐\n│界│\n└──┘", 0},
		{"padding N", "x", "┌─────┐\n│     │\n│     │\n│  x  │\n│     │\n│     │\n└─────┘", 2},
		{"empty input", "", "┌─┐\n│ │\n└─┘", 0},
		{"ragged", "x\nlong", "┌──────┐\n│      │\n│ x    │\n│ long │\n│      │\n└──────┘", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := frameText(tt.input, "single", "", tt.padding, "")
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("frameText() = %q, want %q", got, tt.want)
			}
		})
	}
	if measure.StringWidth("界") != 2 {
		t.Fatal("expected CJK rune to occupy two cells")
	}
}

func TestFrameTitleAndColors(t *testing.T) {
	got, err := frameText("x", "single", "hello", 0, "")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(got, "\n")
	if measure.StringWidth(lines[0]) < measure.StringWidth("hello")+2 {
		t.Fatalf("title did not expand frame: %q", lines[0])
	}
	plain, err := frameText("x", "single", "", 0, "cyan")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(plain, "\x1b[36m┌\x1b[0m") {
		t.Fatalf("color missing from border: %q", plain)
	}
	if _, err := frameText("x", "single", "", 0, "not-a-color"); err == nil {
		t.Fatal("invalid color accepted")
	}
}

func TestFrameCommandStdoutAndWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.ansi")
	if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := execute([]string{"frame", path}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "┌") {
		t.Fatalf("stdout = %q", out.String())
	}
	original, _ := os.ReadFile(path)
	if string(original) != "x" {
		t.Fatalf("default command modified file: %q", original)
	}
	out.Reset()
	if err := execute([]string{"frame", "-w", path}, &out); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Fatalf("--write stdout = %q", out.String())
	}
	written, _ := os.ReadFile(path)
	if string(written) != "┌─┐\n│x│\n└─┘" {
		t.Fatalf("written frame = %q", written)
	}
}

func TestFrameInvalidFlags(t *testing.T) {
	path := writeFixture(t, "x")
	for _, args := range [][]string{{"frame", "--style", "triple", path}, {"frame", "--padding", "-1", path}, {"frame", "--color", "invalid", path}, {"frame"}} {
		var out bytes.Buffer
		if err := execute(args, &out); err == nil {
			t.Errorf("execute(%q) succeeded", args)
		}
	}
}
