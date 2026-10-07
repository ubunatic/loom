// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ubunatic.com/loom"
)

func TestLoadEditorConfigDefaults(t *testing.T) {
	// Loading non-existent default path gives spec defaults
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := loom.LoadEditorConfig("")
	if err != nil {
		t.Fatalf("LoadEditorConfig(\"\"): %v", err)
	}
	if cfg.Theme != loom.SpeccedDefaults.Editor.Theme {
		t.Errorf("Theme = %q, want %q", cfg.Theme, loom.SpeccedDefaults.Editor.Theme)
	}
	if cfg.MouseGrab != loom.SpeccedDefaults.Editor.MouseGrab {
		t.Errorf("MouseGrab = %v, want %v", cfg.MouseGrab, loom.SpeccedDefaults.Editor.MouseGrab)
	}
	if cfg.AltScreen != loom.SpeccedDefaults.Editor.AltScreen {
		t.Errorf("AltScreen = %v, want %v", cfg.AltScreen, loom.SpeccedDefaults.Editor.AltScreen)
	}
}

func TestLoadEditorConfigValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "editor.yaml")
	content := "theme: mc-dark\nmousegrab: true\naltscreen: false\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := loom.LoadEditorConfig(path)
	if err != nil {
		t.Fatalf("LoadEditorConfig(%s): %v", path, err)
	}
	if cfg.Theme != "mc-dark" {
		t.Errorf("Theme = %q, want mc-dark", cfg.Theme)
	}
	if !cfg.MouseGrab {
		t.Errorf("MouseGrab = %v, want true", cfg.MouseGrab)
	}
	if cfg.AltScreen {
		t.Errorf("AltScreen = %v, want false", cfg.AltScreen)
	}
}

func TestLoadEditorConfigInvalidFile(t *testing.T) {
	dir := t.TempDir()

	// Missing explicit path
	missing := filepath.Join(dir, "missing.yaml")
	if _, err := loom.LoadEditorConfig(missing); err == nil {
		t.Fatal("LoadEditorConfig on missing explicit path succeeded, want error")
	}

	// Unknown property
	unknownProp := filepath.Join(dir, "unknown.yaml")
	_ = os.WriteFile(unknownProp, []byte("theme: mc-dark\nunknown_field: 123\n"), 0600)
	if _, err := loom.LoadEditorConfig(unknownProp); err == nil || !strings.Contains(err.Error(), "schema validation failed") {
		t.Fatalf("unknown property error = %v, want schema validation failed", err)
	}

	// Invalid type
	invalidType := filepath.Join(dir, "invalid_type.yaml")
	_ = os.WriteFile(invalidType, []byte("mousegrab: \"yes\"\n"), 0600)
	if _, err := loom.LoadEditorConfig(invalidType); err == nil || !strings.Contains(err.Error(), "schema validation failed") {
		t.Fatalf("invalid type error = %v, want schema validation failed", err)
	}

	// Unknown theme
	unknownTheme := filepath.Join(dir, "unknown_theme.yaml")
	_ = os.WriteFile(unknownTheme, []byte("theme: non_existent_theme\n"), 0600)
	if _, err := loom.LoadEditorConfig(unknownTheme); err == nil || !strings.Contains(err.Error(), "unknown theme") {
		t.Fatalf("unknown theme error = %v, want unknown theme error", err)
	}
}
