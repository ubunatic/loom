// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package winch

import (
	"errors"
	"flag"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestWinchAppHeadlessConfiguration(t *testing.T) {
	pane := &loom.Pane{ResizeConfig: loom.DefaultResizeConfig()}
	theme := loom.SpeccedThemes["plain"]

	app, err := NewApp(pane, "plain", theme)
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}

	// Initial defaults
	cfg := app.Config()
	if !cfg.Coalesce || !cfg.AtomicFlush || !cfg.RowClear || !cfg.SynchronizedOutput || !cfg.AutoWrap || cfg.OutOfBandClear || !cfg.WidthGuard || cfg.WidthGuardN != 1 {
		t.Fatalf("unexpected initial config: %+v", cfg)
	}

	// Toggle key '1' (coalesce)
	if app.HandleKey(loom.KeyEvent{Text: "1"}) {
		t.Fatal("HandleKey(1) should not quit")
	}
	if app.Config().Coalesce {
		t.Fatal("expected coalesce to be toggled off")
	}

	// Toggle key '9' (width_guard)
	app.HandleKey(loom.KeyEvent{Text: "9"})
	if app.Config().WidthGuard {
		t.Fatal("expected width_guard to be toggled off")
	}

	// Adjust guard n with '+' and '-'
	app.HandleKey(loom.KeyEvent{Text: "+"})
	if app.Config().WidthGuardN != 2 {
		t.Fatalf("expected WidthGuardN = 2 after '+', got %d", app.Config().WidthGuardN)
	}
	app.HandleKey(loom.KeyEvent{Text: "+"})
	if app.Config().WidthGuardN != 3 {
		t.Fatalf("expected WidthGuardN = 3 after second '+', got %d", app.Config().WidthGuardN)
	}
	app.HandleKey(loom.KeyEvent{Text: "-"})
	if app.Config().WidthGuardN != 2 {
		t.Fatalf("expected WidthGuardN = 2 after '-', got %d", app.Config().WidthGuardN)
	}
	// Decrease past 1, should clamp to 1
	app.HandleKey(loom.KeyEvent{Text: "-"})
	app.HandleKey(loom.KeyEvent{Text: "-"})
	if app.Config().WidthGuardN != 1 {
		t.Fatalf("expected WidthGuardN clamped to 1, got %d", app.Config().WidthGuardN)
	}

	// Toggle key '6' (out_of_band_clear)
	app.HandleKey(loom.KeyEvent{Text: "6"})
	if !app.Config().OutOfBandClear {
		t.Fatal("expected out_of_band_clear to be toggled on")
	}

	// Reset key 'r'
	app.HandleKey(loom.KeyEvent{Text: "r"})
	if !app.Config().Coalesce || app.Config().OutOfBandClear || !app.Config().WidthGuard || app.Config().WidthGuardN != 1 {
		t.Fatal("expected reset to restore defaults")
	}

	// Motion key 'm'
	if pane.ReduceMotion {
		t.Fatal("expected initial ReduceMotion false")
	}
	app.HandleKey(loom.KeyEvent{Text: "m"})
	if !pane.ReduceMotion {
		t.Fatal("expected ReduceMotion true after 'm'")
	}

	// Theme cycle 't'
	initialTheme := app.themeName
	app.HandleKey(loom.KeyEvent{Text: "t"})
	if app.themeName == initialTheme {
		t.Fatal("expected theme to change on 't'")
	}

	// Quit key 'q'
	if !app.HandleKey(loom.KeyEvent{Text: "q"}) {
		t.Fatal("expected HandleKey('q') to return true (quit)")
	}
	if !app.quitting {
		t.Fatal("expected app.quitting to be true")
	}
}

func TestWinchAppDrawLayouts(t *testing.T) {
	theme := loom.SpeccedThemes["plain"]
	app, err := NewApp(nil, "plain", theme)
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}

	for _, tc := range []struct {
		name string
		w, h int
	}{
		{"wide", 100, 25},
		{"medium", 72, 20},
		{"narrow", 50, 18},
		{"tiny", 20, 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			canvas := loom.NewCanvas(tc.w, tc.h)
			canvas.Clear()
			app.Draw(canvas, canvas.Bounds())

			var b strings.Builder
			canvas.Flush(&b, 1)
			out := b.String()
			if out == "" {
				t.Fatal("empty canvas flush")
			}
		})
	}
}

func TestWinchRunHelp(t *testing.T) {
	err := Run([]string{"--help"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("Run(--help) = %v, want flag.ErrHelp", err)
	}
}

func TestWinchAppAdaptiveGuardSwitch(t *testing.T) {
	pane := &loom.Pane{ResizeConfig: loom.DefaultResizeConfig()}
	app, err := NewApp(pane, "plain", loom.SpeccedThemes["plain"])
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	if app.Config().AdaptiveGuard {
		t.Fatal("adaptive guard must default to the spec value (off)")
	}
	app.HandleKey(loom.KeyEvent{Text: "a"})
	if !app.Config().AdaptiveGuard {
		t.Fatal("key 'a' should enable adaptive guard")
	}

	canvas := loom.NewCanvas(100, 30)
	app.Draw(canvas, canvas.Bounds())
	var rows []string
	for y := 0; y < canvas.Rows(); y++ {
		rows = append(rows, canvas.Row(y))
	}
	text := strings.Join(rows, "\n")
	for _, want := range []string{"Use WINCH speed", "WINCH:", "manual n=1"} {
		if !strings.Contains(text, want) {
			t.Fatalf("draw lacks %q:\n%s", want, text)
		}
	}

	app.HandleKey(loom.KeyEvent{Text: "a"})
	if app.Config().AdaptiveGuard {
		t.Fatal("second 'a' should disable adaptive guard")
	}
}
