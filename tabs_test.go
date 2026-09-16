// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"codeberg.org/ubunatic/loom"
)

// tabSpy is a minimal Widget that records the events it receives, used to
// verify Tabs delegates only to the active child.
type tabSpy struct {
	drawn    bool
	keys     []loom.KeyEvent
	mice     []loom.MouseEvent
	quitKey  bool
	quitMice bool
}

func (s *tabSpy) Draw(*loom.Canvas, loom.Rect) { s.drawn = true }
func (s *tabSpy) HandleKey(e loom.KeyEvent) bool {
	s.keys = append(s.keys, e)
	return s.quitKey
}
func (s *tabSpy) HandleMouse(e loom.MouseEvent) bool {
	s.mice = append(s.mice, e)
	return s.quitMice
}

func TestTabsSwitchWithArrowKeys(t *testing.T) {
	tabs := loom.NewTabs(loom.Tab{Title: "A"}, loom.Tab{Title: "B"}, loom.Tab{Title: "C"})
	if tabs.Focus() != 0 {
		t.Fatalf("initial focus = %d, want 0", tabs.Focus())
	}
	tabs.HandleKey(loom.KeyEvent{Key: "right"})
	if tabs.Focus() != 1 {
		t.Fatalf("after right, focus = %d, want 1", tabs.Focus())
	}
	tabs.HandleKey(loom.KeyEvent{Key: "right"})
	tabs.HandleKey(loom.KeyEvent{Key: "right"}) // wraps back to 0
	if tabs.Focus() != 0 {
		t.Fatalf("after wrap, focus = %d, want 0", tabs.Focus())
	}
	tabs.HandleKey(loom.KeyEvent{Key: "left"}) // wraps to last
	if tabs.Focus() != 2 {
		t.Fatalf("after left wrap, focus = %d, want 2", tabs.Focus())
	}
}

func TestTabsSwitchWithConfiguredKey(t *testing.T) {
	tabs := loom.NewTabs(loom.Tab{Title: "A"}, loom.Tab{Title: "B"})
	tabs.SwitchKey = "ctrl-t"
	tabs.HandleKey(loom.KeyEvent{Key: "ctrl-t"})
	if tabs.Focus() != 1 {
		t.Fatalf("after ctrl-t, focus = %d, want 1", tabs.Focus())
	}
}

func TestTabsDelegatesKeyToActiveChildOnly(t *testing.T) {
	a, b := &tabSpy{}, &tabSpy{}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: a}, loom.Tab{Title: "B", Widget: b})

	tabs.HandleKey(loom.KeyEvent{Text: "x"})
	if len(a.keys) != 1 || len(b.keys) != 0 {
		t.Fatalf("expected only active child (a) to receive the key; a=%d b=%d", len(a.keys), len(b.keys))
	}

	tabs.HandleKey(loom.KeyEvent{Key: "right"}) // switch to b
	tabs.HandleKey(loom.KeyEvent{Text: "y"})
	if len(a.keys) != 1 || len(b.keys) != 1 {
		t.Fatalf("expected only b to receive the key after switch; a=%d b=%d", len(a.keys), len(b.keys))
	}
}

func TestTabsQuitPropagatesFromActiveChild(t *testing.T) {
	a := &tabSpy{quitKey: true}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: a})
	if quit := tabs.HandleKey(loom.KeyEvent{Text: "z"}); !quit {
		t.Error("Tabs should propagate the active child's quit=true")
	}
}

func TestTabsDelegatesMouseToActiveChildOnly(t *testing.T) {
	a, b := &tabSpy{}, &tabSpy{}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: a}, loom.Tab{Title: "B", Widget: b})
	c := loom.NewCanvas(40, 10)
	tabs.Draw(c, loom.Rect{X: 0, Y: 0, W: 40, H: 10})

	// A hover event well below the tab bar should reach the active child (a).
	tabs.HandleMouse(loom.MouseEvent{Action: loom.MouseHover, X: 5, Y: 5})
	if len(a.mice) != 1 || len(b.mice) != 0 {
		t.Fatalf("expected only active child (a) to receive the mouse event; a=%d b=%d", len(a.mice), len(b.mice))
	}
}

func TestTabsMouseClickSwitchesTab(t *testing.T) {
	a, b := &tabSpy{}, &tabSpy{}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: a}, loom.Tab{Title: "B", Widget: b})
	c := loom.NewCanvas(40, 10)
	tabs.Draw(c, loom.Rect{X: 0, Y: 0, W: 40, H: 10})

	// " A " occupies columns 0-2 (0-based canvas); " B " starts at column 3.
	// Terminal/pane-relative mouse coords are 1-based, so click at X=5 (canvas
	// col 4, inside " B ") on the bar row (canvas row 0 -> terminal Y=1).
	quit := tabs.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 5, Y: 1})
	if quit {
		t.Error("switching tabs via mouse must not quit")
	}
	if tabs.Focus() != 1 {
		t.Fatalf("focus after tab-bar click = %d, want 1", tabs.Focus())
	}
	if len(a.mice) != 0 || len(b.mice) != 0 {
		t.Error("a tab-bar click should switch tabs, not forward the event to a child")
	}
}

func TestTabsRendersActiveVsInactiveStyle(t *testing.T) {
	tabs := loom.NewTabs(loom.Tab{Title: "One"}, loom.Tab{Title: "Two"})
	c := loom.NewCanvas(40, 10)
	tabs.Draw(c, loom.Rect{X: 0, Y: 0, W: 40, H: 10})

	active := c.Get(1, 0)   // inside " One "
	inactive := c.Get(6, 0) // inside " Two "
	if active.Style == inactive.Style {
		t.Error("active and inactive tab titles should render with different styles")
	}
	if active.Style != tabs.Style.Active {
		t.Errorf("active tab style = %+v, want %+v", active.Style, tabs.Style.Active)
	}
}

func TestTabsSingleTab(t *testing.T) {
	a := &tabSpy{}
	tabs := loom.NewTabs(loom.Tab{Title: "Only", Widget: a})
	tabs.HandleKey(loom.KeyEvent{Key: "right"}) // no-op, stays on the only tab
	if tabs.Focus() != 0 {
		t.Fatalf("single-tab focus = %d, want 0", tabs.Focus())
	}
	c := loom.NewCanvas(40, 10)
	tabs.Draw(c, loom.Rect{X: 0, Y: 0, W: 40, H: 10})
	if !a.drawn {
		t.Error("the single child should still be drawn")
	}
}

func TestTabsZeroTabsIsNoop(t *testing.T) {
	tabs := loom.NewTabs()
	c := loom.NewCanvas(40, 10)

	// None of these should panic on an empty tab set.
	tabs.Draw(c, loom.Rect{X: 0, Y: 0, W: 40, H: 10})
	if quit := tabs.HandleKey(loom.KeyEvent{Key: "right"}); quit {
		t.Error("zero-tab HandleKey should not quit")
	}
	if quit := tabs.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 1, Y: 1}); quit {
		t.Error("zero-tab HandleMouse should not quit")
	}
	if h := tabs.ContentHeight(); h != 0 {
		t.Errorf("zero-tab ContentHeight = %d, want 0", h)
	}
	if w := tabs.ContentWidth(); w != 0 {
		t.Errorf("zero-tab ContentWidth = %d, want 0", w)
	}
}

func TestTabsContentHeightReservesBar(t *testing.T) {
	tabs := loom.NewTabs(loom.Tab{Title: "A"}, loom.Tab{Title: "B"})
	if h := tabs.ContentHeight(); h < 2 {
		t.Errorf("ContentHeight = %d, want at least 2 (bar + rule)", h)
	}
}
