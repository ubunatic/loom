// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package clitest contains helpers shared by command integration tests.
package clitest

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"ubunatic.com/loom"
)

// BuildLoomBinary builds the CLI from the module root for integration tests.
func BuildLoomBinary(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate clitest source file")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	bin := filepath.Join(t.TempDir(), "loom")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/loom")
	cmd.Dir = root
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	return bin
}

// ExecuteLoom runs the CLI for integration assertions and returns its output.
func ExecuteLoom(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(BuildLoomBinary(t), args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	return output.String(), err
}

// CanvasPlainRow returns a rendered canvas row with spacing normalized.
func CanvasPlainRow(canvas *loom.Canvas, row int) string {
	var text strings.Builder
	for x := 0; x < canvas.Cols(); x++ {
		cell := canvas.Get(x, row)
		if cell.Continuation {
			continue
		}
		if cell.Text == "" {
			text.WriteByte(' ')
		} else {
			text.WriteString(cell.Text)
		}
	}
	return strings.Join(strings.Fields(text.String()), " ")
}

// CanvasScreenText returns all rendered rows with spacing normalized.
func CanvasScreenText(canvas *loom.Canvas) string {
	rows := make([]string, canvas.Rows())
	for row := range rows {
		rows[row] = CanvasPlainRow(canvas, row)
	}
	return strings.Join(rows, "\n")
}
