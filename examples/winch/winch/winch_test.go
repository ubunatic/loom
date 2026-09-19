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
	if !cfg.Coalesce || !cfg.AtomicFlush || !cfg.RowClear || !cfg.SynchronizedOutput || !cfg.AutoWrap || cfg.OutOfBandClear {
		t.Fatalf("unexpected initial config: %+v", cfg)
	}

	// Toggle key '1' (coalesce)
	if app.HandleKey(loom.KeyEvent{Text: "1"}) {
		t.Fatal("HandleKey(1) should not quit")
	}
	if app.Config().Coalesce {
		t.Fatal("expected coalesce to be toggled off")
	}

	// Toggle key '6' (out_of_band_clear)
	app.HandleKey(loom.KeyEvent{Text: "6"})
	if !app.Config().OutOfBandClear {
		t.Fatal("expected out_of_band_clear to be toggled on")
	}

	// Reset key 'r'
	app.HandleKey(loom.KeyEvent{Text: "r"})
	if !app.Config().Coalesce || app.Config().OutOfBandClear {
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
