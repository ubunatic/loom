// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	_ "embed"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed version.yaml
var versionYAML []byte

// TestVersionConstantMatchesVersionYAML guards against drift between the Go
// fallback Version and the repo-root version.yaml that harnez release writes.
// Version stays a string literal because goreleaser's -X ldflag only replaces
// constant initializers.
func TestVersionConstantMatchesVersionYAML(t *testing.T) {
	var doc struct {
		Version string `yaml:"version"`
	}
	if err := yaml.Unmarshal(versionYAML, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Version != Version {
		t.Fatalf("Version = %q, version.yaml = %q", Version, doc.Version)
	}
}

// TestRespectedSpecDefaultsAreConsumed checks that values moved out of Go
// literals exist in the embedded spec and are returned by their consumers.
func TestRespectedSpecDefaultsAreConsumed(t *testing.T) {
	if SpeccedDefaults.ProgressBar.FallbackInterval <= 0 {
		t.Fatal("progress_bar.fallback_interval missing from spec")
	}
	if got := (&ProgressBar{Indeterminate: true}).TickInterval(); got != SpeccedDefaults.ProgressBar.FallbackInterval {
		t.Fatalf("ProgressBar fallback interval = %v, want spec %v", got, SpeccedDefaults.ProgressBar.FallbackInterval)
	}

	if SpeccedDefaults.Clock.TickInterval <= 0 {
		t.Fatal("clock.tick_interval missing from spec")
	}
	timer := NewTimer(time.Minute)
	timer.Start()
	if got := timer.TickInterval(); got != SpeccedDefaults.Clock.TickInterval {
		t.Fatalf("Timer tick interval = %v, want spec %v", got, SpeccedDefaults.Clock.TickInterval)
	}

	if SpeccedDefaults.Menu.CheckedMarker == "" || SpeccedDefaults.Menu.UncheckedMarker == "" {
		t.Fatal("menu checked/unchecked markers missing from spec")
	}
	checked := true
	if got := menuItemLabel(MenuItem{Label: "x", Checked: &checked}); got != SpeccedDefaults.Menu.CheckedMarker+"x" {
		t.Fatalf("checked menu label = %q, want spec marker", got)
	}
	if got := menuItemLabel(MenuItem{Label: "x"}); got != SpeccedDefaults.Menu.UncheckedMarker+"x" {
		t.Fatalf("unchecked menu label = %q, want spec marker", got)
	}

	if SpeccedDefaults.KeyHelp.Separator == "" {
		t.Fatal("key_help.separator missing from spec")
	}
	km := NewKeyMapWithLabels(
		map[string][]string{"open": {"o"}, "close": {"c"}},
		map[string]string{"open": "Open", "close": "Close"},
	)
	if text := NewKeyHelp(km).Text(); !strings.Contains(text, SpeccedDefaults.KeyHelp.Separator) {
		t.Fatalf("key help %q does not use spec separator %q", text, SpeccedDefaults.KeyHelp.Separator)
	}
}
