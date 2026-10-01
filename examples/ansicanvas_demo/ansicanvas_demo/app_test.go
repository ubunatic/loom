// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansicanvas_demo

import (
	"testing"

	"ubunatic.com/loom"
)

func TestDemoAppCreationAndDraw(t *testing.T) {
	app := NewDemoApp(nil)
	if app.editor == nil || app.frame == nil {
		t.Fatalf("expected non-nil editor and frame")
	}

	c := loom.NewCanvas(80, 24)
	app.Draw(c, loom.Rect{X: 0, Y: 0, W: 80, H: 24})

	// Check that canvas was painted
	if c.Get(0, 0).Text == "" {
		t.Fatalf("expected canvas content drawn")
	}
}

func TestDemoAppKeyHandling(t *testing.T) {
	app := NewDemoApp(nil)

	// Arrows should be handled and not quit
	res := app.ConsumeKey(loom.KeyEvent{Key: "down"})
	if !res.Consumed || res.Quit {
		t.Fatalf("expected down arrow handled without quit: %+v", res)
	}

	// F10 should quit
	res = app.ConsumeKey(loom.KeyEvent{Key: "f10"})
	if !res.Consumed || !res.Quit {
		t.Fatalf("expected f10 to quit: %+v", res)
	}

	// 'q' should quit
	app2 := NewDemoApp(nil)
	res = app2.ConsumeKey(loom.KeyEvent{Key: "q"})
	if !res.Consumed || !res.Quit {
		t.Fatalf("expected 'q' to quit: %+v", res)
	}
}

func TestDemoAppMouseStillScrollsEditor(t *testing.T) {
	app := NewDemoApp(nil)
	res := app.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollDown})
	if !res.Consumed || app.editor.ScrollY != 1 {
		t.Fatalf("mouse scroll result=%+v ScrollY=%d; want consumed and 1", res, app.editor.ScrollY)
	}
}
