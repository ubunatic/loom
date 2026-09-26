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

func TestANSIViewScrollAndQuit(t *testing.T) {
	buffer, err := loom.ParseAnsiBuffer("one\ntwo\nthree", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	view := &ansiView{buffer: buffer}
	if view.HandleKey(loom.KeyEvent{Key: "down"}) || view.offset != 1 {
		t.Fatalf("down key did not scroll: offset=%d", view.offset)
	}
	if !view.HandleKey(loom.KeyEvent{Key: "f10"}) {
		t.Fatal("F10 did not quit")
	}
}
