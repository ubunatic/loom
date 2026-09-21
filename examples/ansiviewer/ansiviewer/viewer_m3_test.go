// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansiviewer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestGenerateM3Evidence generates evidence that ansiviewer renders ANSI colors correctly.
// This demonstrates that the M1 and M2 implementations (ParseANSI and Canvas.WriteANSI)
// work for the ansiviewer use case.
// Only generates evidence when LOOM_EVIDENCE=1 environment variable is set.
func TestGenerateM3Evidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	dir := t.TempDir()

	// Create a colored ANSI file with various colors and styles
	coloredContent := "" +
		"\x1b[1;31mBold Red\x1b[0m\n" +
		"\x1b[32mGreen\x1b[0m \x1b[34mBlue\x1b[0m \x1b[33mYellow\x1b[0m\n" +
		"\x1b[38;5;196mColor196\x1b[0m \x1b[38;5;226mColor226\x1b[0m\n" +
		"\x1b[38;2;255;100;50mOrange RGB\x1b[0m\n" +
		"\x1b[41m Red BG \x1b[0m \x1b[42m Green BG \x1b[0m\n" +
		"\x1b[1;4;38;2;128;200;255mBold Underline Cyan\x1b[0m\n"

	colorPath := filepath.Join(dir, "colors.ansi")
	if err := os.WriteFile(colorPath, []byte(coloredContent), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Render the ansiviewer with the colored file
	var out bytes.Buffer
	if err := Render(&out, dir, 80, 20); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Save evidence
	baseDir := filepath.Join("..", "..", "docs", "progress", "034")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	outPath := filepath.Join(baseDir, "M3-ansiviewer.ansi")
	if err := os.WriteFile(outPath, out.Bytes(), 0o644); err != nil {
		t.Fatalf("WriteFile %s failed: %v", outPath, err)
	}

	t.Logf("Evidence saved to %s (%d bytes)", outPath, out.Len())
}
