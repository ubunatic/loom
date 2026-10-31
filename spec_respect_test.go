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

// withSpecDefaults overrides spec values for the duration of a test. Consumers
// read SpeccedDefaults at call time, so a sentinel that comes back out proves the
// value is consumed, not copied into Go.
func withSpecDefaults(t *testing.T, mutate func(*LibDefaults)) {
	t.Helper()
	saved := SpeccedDefaults
	t.Cleanup(func() { SpeccedDefaults = saved })
	mutate(&SpeccedDefaults)
}

// TestRespectedSpecDefaultsAreConsumed checks that values moved out of Go
// literals are read by their consumers. The splash and media keys are covered in
// their own packages.
func TestRespectedSpecDefaultsAreConsumed(t *testing.T) {
	const (
		fallback  = 37 * time.Millisecond
		tick      = 3 * time.Second
		separator = " | "
	)
	withSpecDefaults(t, func(d *LibDefaults) {
		d.ProgressBar.FallbackInterval = fallback
		d.Clock.TickInterval = tick
		d.Menu.CheckedMarker = "[x] "
		d.Menu.UncheckedMarker = "[ ] "
		d.KeyHelp.Separator = separator
	})

	if got := (&ProgressBar{Indeterminate: true}).TickInterval(); got != fallback {
		t.Fatalf("ProgressBar fallback interval = %v, want spec %v", got, fallback)
	}

	timer := NewTimer(time.Minute)
	timer.Start()
	if got := timer.TickInterval(); got != tick {
		t.Fatalf("Timer tick interval = %v, want spec %v", got, tick)
	}

	stopwatch := NewStopwatch()
	stopwatch.Start()
	if got := stopwatch.TickInterval(); got != tick {
		t.Fatalf("Stopwatch tick interval = %v, want spec %v", got, tick)
	}

	checked := true
	if got := menuItemLabel(MenuItem{Label: "x", Checked: &checked}); got != "[x] x" {
		t.Fatalf("checked menu label = %q, want spec marker", got)
	}
	if got := menuItemLabel(MenuItem{Label: "x"}); got != "[ ] x" {
		t.Fatalf("unchecked menu label = %q, want spec marker", got)
	}

	km := NewKeyMapWithLabels(
		map[string][]string{"open": {"o"}, "close": {"c"}},
		map[string]string{"open": "Open", "close": "Close"},
	)
	if text := NewKeyHelp(km).Text(); !strings.Contains(text, separator) {
		t.Fatalf("key help %q does not use spec separator %q", text, separator)
	}
}
