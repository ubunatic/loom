// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package clitest contains helpers shared by command integration tests.
package clitest

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"ubunatic.com/loom"
)

// The test process owns this cache; RunTests removes it after all tests finish.
var loomBinary struct {
	once sync.Once
	dir  string
	path string
	err  error
}

// RunTests runs a package's tests and removes its shared CLI binary.
// Packages using BuildLoomBinary must call os.Exit(clitest.RunTests(m)) in TestMain.
func RunTests(m *testing.M) int {
	code := m.Run()
	if loomBinary.dir != "" {
		if err := os.RemoveAll(loomBinary.dir); err != nil {
			fmt.Fprintf(os.Stderr, "remove test loom binary: %v\n", err)
			return 1
		}
	}
	return code
}

// BuildLoomBinary builds the CLI once per test process for integration tests.
func BuildLoomBinary(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}
	loomBinary.once.Do(func() {
		loomBinary.err = buildLoomBinary()
	})
	if loomBinary.err != nil {
		t.Fatal(loomBinary.err)
	}
	return loomBinary.path
}

func buildLoomBinary() error {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("locate clitest source file")
	}
	var err error
	loomBinary.dir, err = os.MkdirTemp("", "loom-clitest-*")
	if err != nil {
		return fmt.Errorf("create test loom directory: %w", err)
	}
	loomBinary.path = filepath.Join(loomBinary.dir, "loom")
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	cmd := exec.Command("go", "build", "-o", loomBinary.path, "./cmd/loom")
	cmd.Dir = root
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build loom: %w\n%s", err, output)
	}
	return nil
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
