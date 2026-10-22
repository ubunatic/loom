// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func runDiff(t *testing.T, want, got string, flags ...string) (string, error) {
	t.Helper()
	wantPath, gotPath := writeFixture(t, want), writeFixture(t, got)
	args := append([]string{"diff"}, flags...)
	args = append(args, wantPath, gotPath)
	var out bytes.Buffer
	err := execute(args, &out)
	return out.String(), err
}

func TestDiffEqualAndTmuxCapture(t *testing.T) {
	want := "\x1b[31mA界\x1b[0m"
	// Representative `tmux capture-pane -pe` output includes SGR and CRLF.
	got := "\x1b[31mA界\x1b[0m\r\n"
	out, err := runDiff(t, want, got)
	if err != nil || out != "screens match\n" {
		t.Fatalf("diff output=%q err=%v", out, err)
	}
}

func TestDiffGlyphAndColourWithCellCoordinates(t *testing.T) {
	want := "界\x1b[38;5;196;48;2;1;2;3mx\x1b[0m"
	got := "界\x1b[38;2;255;0;0;48;5;4my\x1b[0m"
	out, err := runDiff(t, want, got)
	if errorExitCode(err) != 1 {
		t.Fatalf("error=%v, want exit status 1", err)
	}
	for _, expected := range []string{"row=1 col=3", `glyph="x"`, `glyph="y"`, "fg=index:196", "fg=#ff0000", "bg=#010203", "bg=index:4", "differences: 1"} {
		if !strings.Contains(out, expected) {
			t.Errorf("output %q missing %q", out, expected)
		}
	}
}

func TestDiffWideRuneColourCoversBothDisplayCells(t *testing.T) {
	out, err := runDiff(t, "\x1b[31m界\x1b[0m", "\x1b[32m界\x1b[0m")
	if errorExitCode(err) != 1 {
		t.Fatalf("error=%v, want exit status 1", err)
	}
	if !strings.Contains(out, "row=1 col=2") {
		t.Fatalf("wide rune continuation comparison output=%q, want second display column", out)
	}
}

func TestDiffResetColoursAndIgnoreFlags(t *testing.T) {
	want := "\x1b[31;44mx\x1b[39;49m"
	got := "\x1b[31;44mx\x1b[0m"
	out, err := runDiff(t, want, got)
	if err != nil || out != "screens match\n" {
		t.Fatalf("reset diff output=%q err=%v", out, err)
	}
	out, err = runDiff(t, "a", "b", "--ignore-glyph")
	if err != nil || out != "screens match\n" {
		t.Fatalf("ignore glyph output=%q err=%v", out, err)
	}
	out, err = runDiff(t, "\x1b[31mx\x1b[0m", "\x1b[32mx\x1b[0m", "--ignore-colour")
	if err != nil || out != "screens match\n" {
		t.Fatalf("ignore colour output=%q err=%v", out, err)
	}
}

func TestDiffRegionAndDifferentSizes(t *testing.T) {
	out, err := runDiff(t, "abcd\nefgh", "abcd\nefgX", "--region", "2,4,1,1")
	if errorExitCode(err) != 1 || !strings.Contains(out, "row=2 col=4") {
		t.Fatalf("region output=%q err=%v", out, err)
	}
	out, err = runDiff(t, "a", "ab")
	if errorExitCode(err) != 1 || !strings.Contains(out, "col=2") {
		t.Fatalf("size output=%q err=%v", out, err)
	}
	if errorExitCode(runDiffError(t, "a", "b", "--region", "0,1,2,1")) != 2 {
		t.Fatal("invalid region should use exit status 2")
	}
}

func runDiffError(t *testing.T, want, got string, flags ...string) error {
	t.Helper()
	_, err := runDiff(t, want, got, flags...)
	return err
}

func TestDiffErrorStatusAndMissingInput(t *testing.T) {
	if errorExitCode(execute([]string{"diff"}, &bytes.Buffer{})) != 2 {
		t.Fatal("diff usage error should use status 2")
	}
	if errorExitCode(fmt.Errorf("ordinary command error")) != 1 {
		t.Fatal("ordinary command errors should retain status 1")
	}
	if errorExitCode(runDiffError(t, "x", "x", "--region", "1,1,0,1")) != 2 {
		t.Fatal("invalid region should use exit status 2")
	}
	if errorExitCode(fmt.Errorf("wrapped: %w", diffExitError{2, fmt.Errorf("bad input")})) != 2 {
		t.Fatal("wrapped input error should use status 2")
	}
}

func TestCropWritesSelectedRegion(t *testing.T) {
	input := writeFixture(t, "ab\ncd")
	var out bytes.Buffer
	if err := execute([]string{"crop", input, "--region", "2,2,1,1"}, &out); err != nil {
		t.Fatalf("crop: %v", err)
	}
	if got := parseDiffText(out.String()); got != "d" {
		t.Fatalf("crop output=%q, want d", got)
	}
}

func TestDiffFindCropsGotCapture(t *testing.T) {
	wantPath := writeFixture(t, "abc\ndef")
	gotPath := writeFixture(t, "header\n  abc\n  def")
	var out bytes.Buffer
	err := execute([]string{"diff", wantPath, gotPath, "--find", "abc"}, &out)
	if err != nil || out.String() != "screens match\n" {
		t.Fatalf("diff output=%q err=%v", out.String(), err)
	}
}

func TestDiffGotRegionCropsOnlyCapture(t *testing.T) {
	wantPath := writeFixture(t, "abc\ndef")
	gotPath := writeFixture(t, "header\n  abc\n  def")
	var out bytes.Buffer
	err := execute([]string{"diff", wantPath, gotPath, "--got-region", "2,3,3,2"}, &out)
	if err != nil || out.String() != "screens match\n" {
		t.Fatalf("diff output=%q err=%v", out.String(), err)
	}
}

func parseDiffText(text string) string {
	buf, err := parseDiffScreen([]byte(text))
	if err != nil {
		return "<parse error>"
	}
	return buf.PlainText()
}
