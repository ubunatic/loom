// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func writeFixture(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "asset.ansi")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestMeasureAndEval(t *testing.T) {
	path := writeFixture(t, "\x1b[31m界a\x1b[0m  \ntiny\n")
	var measured bytes.Buffer
	if err := execute([]string{"measure", path}, &measured); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"line 1: 5 columns", "line 2: 4 columns (ragged)", "bounding box: 5 columns x 2 lines", "trailing whitespace"} {
		if !strings.Contains(measured.String(), want) {
			t.Errorf("measure output %q does not contain %q", measured.String(), want)
		}
	}
	if strings.Contains(measured.String(), "boxes:") {
		t.Fatalf("measure unexpectedly includes structural summary: %q", measured.String())
	}

	boxed := writeFixture(t, "┌──┐\n│ x│\n└──┘\n")
	var evaluated bytes.Buffer
	if err := execute([]string{"eval", boxed}, &evaluated); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"boxes: 1", "rows: 3", "columns: 4", "non-blank bounds: x=1..4 y=1..3 (4 x 3)", "ragged rows: 0", "min line width: 4", "max line width: 4"} {
		if !strings.Contains(evaluated.String(), want) {
			t.Errorf("eval output %q does not contain %q", evaluated.String(), want)
		}
	}
}

func TestEvalAnnotatedOutput(t *testing.T) {
	path := writeFixture(t, "┌──┐\n│ x│  \n└──┘\n")
	for _, flag := range []string{"--annotate", "-a"} {
		var out bytes.Buffer
		if err := execute([]string{"eval", flag, path}, &out); err != nil {
			t.Fatalf("eval %s: %v", flag, err)
		}
		got := out.String()
		for _, want := range []string{"┌──┐", "│ x│", "<-- 1", "1: line 1: ragged width", "2: line 2: trailing whitespace"} {
			if !strings.Contains(got, want) {
				t.Errorf("eval %s output %q does not contain %q", flag, got, want)
			}
		}
	}
}

func TestCheckBoxAnnotatedOutput(t *testing.T) {
	path := writeFixture(t, "┌──┐\n│x│\n└─┘\n")
	for _, flag := range []string{"--annotate", "-a"} {
		var out bytes.Buffer
		err := execute([]string{"check-box", flag, path}, &out)
		if err == nil {
			t.Fatalf("check-box %s accepted invalid box", flag)
		}
		got := out.String() + err.Error()
		for _, want := range []string{"└─┘", "<-- 1", "1: line 2: box width mismatch"} {
			if !strings.Contains(got, want) {
				t.Errorf("check-box %s output %q does not contain %q", flag, got, want)
			}
		}
	}
}

func TestEvalCountsRegionalIndicatorPairAsTwoColumns(t *testing.T) {
	path := writeFixture(t, "🇩🇪X\n")
	var out bytes.Buffer
	if err := execute([]string{"eval", path}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "non-blank bounds: x=1..3 y=1..1 (3 x 1)") {
		t.Fatalf("eval output = %q, want a three-column non-blank bound", out.String())
	}
}

func TestCheckBox(t *testing.T) {
	valid := writeFixture(t, "┌──┐\n│hi│\n└──┘\n")
	invalid := writeFixture(t, "┌──┐\n│x│\n└─┘\n")
	var out bytes.Buffer
	if err := execute([]string{"check-box", valid}, &out); err != nil {
		t.Fatalf("valid box rejected: %v", err)
	}
	if !strings.Contains(out.String(), "ok") {
		t.Fatalf("missing success output: %q", out.String())
	}
	if err := execute([]string{"check-box", valid, invalid}, &out); err == nil {
		t.Fatal("misaligned box accepted")
	} else if !strings.Contains(err.Error(), invalid) {
		t.Fatalf("multi-file error %q omits invalid filename %q", err, invalid)
	}
}

func TestCLIArityErrors(t *testing.T) {
	for _, args := range [][]string{{"measure"}, {"measure", "a", "b"}, {"eval"}, {"view"}, {"check-box"}} {
		var out bytes.Buffer
		if err := execute(args, &out); err == nil {
			t.Errorf("execute(%q) succeeded, want arity error", args)
		}
	}
}

func TestMissingFilesHaveCleanFilenameErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.ansi")
	for _, command := range []string{"measure", "eval", "check-box", "view"} {
		var out bytes.Buffer
		err := execute([]string{command, missing}, &out)
		if err == nil {
			t.Errorf("%s unexpectedly read missing file", command)
			continue
		}
		message := err.Error()
		if !strings.Contains(message, missing) || !strings.Contains(message, "no such file") {
			t.Errorf("%s error %q lacks filename or cause", command, message)
		}
		if strings.Count(message, missing) != 1 {
			t.Errorf("%s error repeats filename: %q", command, message)
		}
	}
}

func TestCheckBoxReportsEveryFileError(t *testing.T) {
	missingA := filepath.Join(t.TempDir(), "missing-a.ansi")
	missingB := filepath.Join(t.TempDir(), "missing-b.ansi")
	var out bytes.Buffer
	err := execute([]string{"check-box", missingA, missingB}, &out)
	if err == nil {
		t.Fatal("check-box accepted missing files")
	}
	for _, path := range []string{missingA, missingB} {
		if !strings.Contains(err.Error(), path) {
			t.Errorf("multi-file error %q omits %q", err, path)
		}
	}
}

func TestViewRequiresANSI(t *testing.T) {
	var out bytes.Buffer
	if err := execute([]string{"view", "asset.txt"}, &out); err == nil || !strings.Contains(err.Error(), ".ansi") {
		t.Fatalf("view error = %v, want .ansi support message", err)
	}
}

func TestViewPlainTextFlag(t *testing.T) {
	path := writeFixture(t, "\x1b[31mred\x1b[0m  \n┌─┐")
	for _, flag := range []string{"--plain", "-p"} {
		var out bytes.Buffer
		if err := execute([]string{"view", flag, path}, &out); err != nil {
			t.Fatalf("view %s: %v", flag, err)
		}
		if got, want := out.String(), "red\n┌─┐\n"; got != want {
			t.Errorf("view %s output = %q, want %q", flag, got, want)
		}
	}
}

func TestANSIViewScrollAndQuit(t *testing.T) {
	buffer, err := loom.ParseAnsiBuffer("one\ntwo\nthree", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	view := &ansiView{buffer: buffer}
	if view.HandleKey(loom.KeyEvent{Key: "down"}) || view.offsetY != 1 {
		t.Fatalf("down key did not scroll: offsetY=%d", view.offsetY)
	}
	if !view.HandleKey(loom.KeyEvent{Key: "f10"}) {
		t.Fatal("F10 did not quit")
	}
}

func TestANSIViewPansWideBufferAndClampsOffsets(t *testing.T) {
	buffer, err := loom.ParseAnsiBuffer("abcdefghijklmnopqrst\nABCDEFGHIJKLMNOPQRST\n01234567890123456789", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	view := &ansiView{buffer: buffer, offsetX: 2, offsetY: 1}
	canvas := loom.NewCanvas(6, 4)
	view.Draw(canvas, loom.Rect{X: 1, Y: 1, W: 4, H: 2})
	if got := canvas.Get(1, 1).Text; got != "C" {
		t.Fatalf("first rendered cell after 2D pan = %q, want %q", got, "C")
	}
	if got := canvas.Get(4, 2).Text; got != "5" {
		t.Fatalf("last rendered cell after 2D pan = %q, want %q", got, "5")
	}
	view.offsetX = 0
	view.offsetY = 0
	view.Draw(canvas, loom.Rect{W: 4, H: 2})
	view.HandleKey(loom.KeyEvent{Key: "right"})
	view.HandleKey(loom.KeyEvent{Key: "right"})
	view.Draw(canvas, loom.Rect{W: 4, H: 2})
	if got := canvas.Get(0, 0).Text; got != "c" {
		t.Fatalf("first rendered cell after horizontal pan = %q, want %q", got, "c")
	}
	view.HandleKey(loom.KeyEvent{Key: "end"})
	if want := buffer.Cols() - 4; view.offsetX != want {
		t.Fatalf("end horizontal offset = %d, want maximum offset %d", view.offsetX, want)
	}
	view.Draw(canvas, loom.Rect{W: 4, H: 2})
	if got := canvas.Get(0, 0).Text; got != "q" {
		t.Fatalf("first rendered cell at right edge = %q, want %q", got, "q")
	}
	view.HandleKey(loom.KeyEvent{Key: "right"})
	if want := buffer.Cols() - 4; view.offsetX != want {
		t.Fatalf("right edge offset = %d, want %d", view.offsetX, want)
	}
	view.HandleKey(loom.KeyEvent{Key: "home"})
	view.HandleKey(loom.KeyEvent{Key: "]"})
	if view.offsetX != 10 {
		t.Fatalf("] horizontal offset = %d, want 10", view.offsetX)
	}
	view.HandleKey(loom.KeyEvent{Key: "shift-right"})
	if view.offsetX != 16 {
		t.Fatalf("shift-right horizontal offset = %d, want clamped maximum 16", view.offsetX)
	}
	view.HandleKey(loom.KeyEvent{Key: "["})
	if view.offsetX != 6 {
		t.Fatalf("[ horizontal offset = %d, want 6", view.offsetX)
	}
	view.HandleKey(loom.KeyEvent{Key: "left"})
	view.HandleKey(loom.KeyEvent{Key: "home"})
	if view.offsetX != 0 {
		t.Fatalf("home horizontal offset = %d, want 0", view.offsetX)
	}
	view.HandleKey(loom.KeyEvent{Key: "pgdn"})
	if want := buffer.Rows() - 2; view.offsetY != want {
		t.Fatalf("page-down vertical offset = %d, want clamped offset %d", view.offsetY, want)
	}
	view.HandleKey(loom.KeyEvent{Key: "down"})
	if want := buffer.Rows() - 2; view.offsetY != want {
		t.Fatalf("bottom vertical offset = %d, want %d", view.offsetY, want)
	}
}

func TestViewCommandPaneSettings(t *testing.T) {
	pane := &loom.Pane{MaxCols: loom.DefaultMaxCols}
	configureViewPane(pane)
	if pane.MaxCols != 0 {
		t.Errorf("view pane MaxCols = %d, want 0", pane.MaxCols)
	}
	if !pane.Resizeable {
		t.Error("view pane Resizeable = false, want true")
	}
}
