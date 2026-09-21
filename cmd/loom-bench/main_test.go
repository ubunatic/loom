// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/internal/examplesreg"
)

func TestHeadlessSmokeNewWidget(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"split"},
		{"tabs"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			example, ok := examplesreg.Find(tt.name)
			if !ok {
				t.Fatalf("example %q not found", tt.name)
			}
			if example.NewWidget == nil {
				t.Fatalf("example %q has no NewWidget", tt.name)
			}
			
			// Create widget
			widget, err := example.NewWidget([]string{})
			if err != nil {
				t.Fatalf("NewWidget failed: %v", err)
			}
			
			// Cast to loom.Widget
			w, ok := widget.(loom.Widget)
			if !ok {
				t.Fatal("NewWidget did not return a loom.Widget")
			}
			
			// Test 80x24 render
			frames80 := loom.Render(w, 80, 24)
			if len(frames80) != 24 {
				t.Errorf("80x24 render produced %d lines, expected 24", len(frames80))
			}
			if frames80[0] == "" {
				t.Error("80x24 first line is empty")
			}
			
			// Test 20x5 render
			frames20 := loom.Render(w, 20, 5)
			if len(frames20) != 5 {
				t.Errorf("20x5 render produced %d lines, expected 5", len(frames20))
			}
			if frames20[0] == "" {
				t.Error("20x5 first line is empty")
			}
		})
	}
}

func TestGenerateM2BenchEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}
	
	examples := []string{"split", "tabs"}
	for _, exampleName := range examples {
		example, ok := examplesreg.Find(exampleName)
		if !ok || example.NewWidget == nil {
			continue
		}
		
		// Create widget
		widget, err := example.NewWidget([]string{})
		if err != nil {
			t.Fatalf("NewWidget failed for %q: %v", exampleName, err)
		}
		
		// Cast to loom.Widget
		w, ok := widget.(loom.Widget)
		if !ok {
			t.Fatalf("%q NewWidget did not return a loom.Widget", exampleName)
		}
		
		// Render at tiny size (20x5)
		frames := loom.Render(w, 20, 5)
		
		// Join frames into a single string
		var buf strings.Builder
		for _, line := range frames {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
		
		// Write evidence file
		repoRoot := findBenchRepoRoot(t)
		progressDir := filepath.Join(repoRoot, "docs", "progress", "062")
		if err := os.MkdirAll(progressDir, 0755); err != nil {
			t.Fatalf("creating progress directory: %v", err)
		}
		
		outPath := filepath.Join(progressDir, "M2-bench-tiny-"+exampleName+".ansi")
		if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
			t.Fatalf("writing evidence: %v", err)
		}
		
		t.Logf("M2 evidence saved to %s (%d bytes)", outPath, len(buf.String()))
	}
}

// findBenchRepoRoot walks up from the current directory to find the repo root (where go.mod is)
func findBenchRepoRoot(t *testing.T) string {
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
