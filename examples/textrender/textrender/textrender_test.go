// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package textrender

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

// TestWidths verifies that all test cases have correct display widths
func TestWidths(t *testing.T) {
	for _, tc := range Cases {
		got := loom.StringWidth(tc.Text)
		if got != tc.WantWidth {
			t.Errorf("Case %q: StringWidth(%q) = %d, want %d", tc.Label, tc.Text, got, tc.WantWidth)
		}
	}
}

func TestGenerateM1BordersEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	// Create widget
	widget, err := NewWidget([]string{})
	if err != nil {
		t.Fatalf("NewWidget failed: %v", err)
	}

	// Generate evidence at 80x24
	const cols80, rows24 = 80, 24
	frames := loom.Render(widget, cols80, rows24)

	var buf strings.Builder
	for _, line := range frames {
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	repoRoot := findRepoRoot(t)
	progressDir := filepath.Join(repoRoot, "docs", "progress", "096")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		t.Fatalf("creating progress directory: %v", err)
	}

	outPath := filepath.Join(progressDir, "M1-borders.ansi")
	if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
		t.Fatalf("writing evidence: %v", err)
	}

	t.Logf("M1 evidence (80x24) saved to %s (%d bytes)", outPath, len(buf.String()))

	// Generate evidence at 40x12 (narrow)
	const cols40, rows12 = 40, 12
	framesNarrow := loom.Render(widget, cols40, rows12)

	var bufNarrow strings.Builder
	for _, line := range framesNarrow {
		bufNarrow.WriteString(line)
		bufNarrow.WriteString("\n")
	}

	outPathNarrow := filepath.Join(progressDir, "M1-borders-narrow.ansi")
	if err := os.WriteFile(outPathNarrow, []byte(bufNarrow.String()), 0644); err != nil {
		t.Fatalf("writing narrow evidence: %v", err)
	}

	t.Logf("M1 evidence (40x12) saved to %s (%d bytes)", outPathNarrow, len(bufNarrow.String()))
}

func TestGenerateM2ButtonsEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	// Render the buttons view directly
	buttonsWidget := newButtonsView()
	const cols, rows = 80, 24
	frames := loom.Render(buttonsWidget, cols, rows)

	var buf strings.Builder
	for _, line := range frames {
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	repoRoot := findRepoRoot(t)
	progressDir := filepath.Join(repoRoot, "docs", "progress", "096")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		t.Fatalf("creating progress directory: %v", err)
	}

	outPath := filepath.Join(progressDir, "M2-buttons.ansi")
	if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
		t.Fatalf("writing evidence: %v", err)
	}

	t.Logf("M2 Buttons evidence saved to %s (%d bytes)", outPath, len(buf.String()))
}

func TestGenerateM2ClippingEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	// Render the clipping view directly
	clippingWidget := newClippingView()
	const cols, rows = 80, 24
	frames := loom.Render(clippingWidget, cols, rows)

	var buf strings.Builder
	for _, line := range frames {
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	repoRoot := findRepoRoot(t)
	progressDir := filepath.Join(repoRoot, "docs", "progress", "096")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		t.Fatalf("creating progress directory: %v", err)
	}

	outPath := filepath.Join(progressDir, "M2-clipping.ansi")
	if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
		t.Fatalf("writing evidence: %v", err)
	}

	t.Logf("M2 Clipping evidence saved to %s (%d bytes)", outPath, len(buf.String()))
}

func TestGenerateM2ScrollEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	// Render the scroll view directly
	scrollWidget := newScrollView()
	const cols, rows = 80, 24
	frames := loom.Render(scrollWidget, cols, rows)

	var buf strings.Builder
	for _, line := range frames {
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	repoRoot := findRepoRoot(t)
	progressDir := filepath.Join(repoRoot, "docs", "progress", "096")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		t.Fatalf("creating progress directory: %v", err)
	}

	outPath := filepath.Join(progressDir, "M2-scroll.ansi")
	if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
		t.Fatalf("writing evidence: %v", err)
	}

	t.Logf("M2 Scroll evidence saved to %s (%d bytes)", outPath, len(buf.String()))
}

// findRepoRoot walks up from the current directory to find the repo root (where go.mod is)
func findRepoRoot(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting current directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			t.Fatalf("go.mod not found in any parent directory")
		}
		cwd = parent
	}
}
