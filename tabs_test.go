// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"ubunatic.com/loom"
)

// tabSpy is a minimal Widget that records the events it receives, used to
// verify Tabs delegates only to the active child.
type tabSpy struct {
	drawn      bool
	keys       []loom.KeyEvent
	mice       []loom.MouseEvent
	quitKey    bool
	quitMice   bool
	handleMice bool
}

type focusTabSpy struct {
	tabSpy
	focused bool
}

func (s *focusTabSpy) Focused() bool         { return s.focused }
func (s *focusTabSpy) SetFocus(focused bool) { s.focused = focused }

func (s *tabSpy) Draw(*loom.Canvas, loom.Rect) { s.drawn = true }
func (s *tabSpy) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	s.keys = append(s.keys, e)
	return loom.EventResult{Consumed: s.quitKey, Quit: s.quitKey}
}
func (s *tabSpy) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	s.mice = append(s.mice, e)
	return loom.EventResult{Consumed: s.quitMice || s.handleMice, Quit: s.quitMice}
}

func TestTabsCapturesHandledLeftDrag(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		for _, handled := range []bool{false, true} {
			for _, button := range []loom.MouseButton{loom.MouseLeft, loom.MouseRight} {
				a, b := &tabSpy{handleMice: handled}, &tabSpy{}
				tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: a}, loom.Tab{Title: "B", Widget: b})
				tabs.Vertical = vertical
				tabs.Draw(loom.NewCanvas(50, 20), loom.Rect{X: 5, Y: 3, W: 30, H: 10})
				tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: button, X: 10, Y: 4})
				press := a.mice[0]
				// Capture belongs to the pressed child even if selection changes.
				tabs.Select(1)
				tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MouseDrag, Button: button, X: -2, Y: -3})
				tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MouseRelease, Button: button, X: 40, Y: 15})
				if handled && button == loom.MouseLeft {
					if len(a.mice) != 3 || a.mice[1].X != press.X-12 || a.mice[1].Y != press.Y-7 || a.mice[2].X != press.X+30 || a.mice[2].Y != press.Y+11 {
						t.Fatalf("vertical=%v: captured events lost child-local coordinates: %#v", vertical, a.mice)
					}
				} else if len(a.mice) != 1 {
					t.Fatalf("unhandled/right press captured drag: %#v", a.mice)
				}
				before := len(a.mice)
				tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MouseDrag, Button: button, X: -2, Y: -3})
				if len(a.mice) != before {
					t.Fatal("released capture forwarded another drag to the original child")
				}
				if len(b.mice) != 0 {
					t.Fatal("out-of-panel event reached the newly selected child")
				}
				tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MouseHover, X: 10, Y: 4})
				if len(b.mice) != 1 {
					t.Fatal("release did not restore active-child routing")
				}
			}
		}
	}
}

func TestTabsSwitchWithArrowKeys(t *testing.T) {
	tabs := loom.NewTabs(loom.Tab{Title: "A"}, loom.Tab{Title: "B"}, loom.Tab{Title: "C"})
	if tabs.Focus() != 0 {
		t.Fatalf("initial focus = %d, want 0", tabs.Focus())
	}
	tabs.ConsumeKey(loom.KeyEvent{Key: "right"})
	if tabs.Focus() != 1 {
		t.Fatalf("after right, focus = %d, want 1", tabs.Focus())
	}
	tabs.ConsumeKey(loom.KeyEvent{Key: "right"})
	tabs.ConsumeKey(loom.KeyEvent{Key: "right"}) // wraps back to 0
	if tabs.Focus() != 0 {
		t.Fatalf("after wrap, focus = %d, want 0", tabs.Focus())
	}
	tabs.ConsumeKey(loom.KeyEvent{Key: "left"}) // wraps to last
	if tabs.Focus() != 2 {
		t.Fatalf("after left wrap, focus = %d, want 2", tabs.Focus())
	}
}

func TestTabsVerticalLayoutAndSelection(t *testing.T) {
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: loom.NewView([]string{"first"})}, loom.Tab{Title: "Long", Widget: loom.NewView([]string{"second"})})
	tabs.Vertical = true
	rows := loom.Render(tabs, 16, 5)
	if len(rows) != 5 || !strings.Contains(rows[0], " A ") || !strings.Contains(rows[1], " Long ") || !strings.Contains(rows[0], "first") {
		t.Fatalf("vertical layout = %#v", rows)
	}
	tabs.Select(1)
	rows = loom.Render(tabs, 16, 5)
	if !strings.Contains(rows[0], "second") {
		t.Fatalf("selected vertical tab content missing: %#v", rows)
	}
}

func TestTabsVerticalOverflowKeepsSelectionVisibleAndClickable(t *testing.T) {
	tabs := loom.NewTabs(loom.Tab{Title: "A"}, loom.Tab{Title: "B"}, loom.Tab{Title: "C"}, loom.Tab{Title: "D"}, loom.Tab{Title: "E"})
	tabs.Vertical = true
	for _, tc := range []struct {
		focus, height    int
		visible          []string
		click, wantFocus int
	}{
		{0, 2, []string{" A ", " B "}, 1, 1},
		{4, 2, []string{" D ", " E "}, 0, 3},
		{3, 1, []string{" D "}, 0, 3},
		{0, 2, []string{" A ", " B "}, 0, 0},
	} {
		tabs.Select(tc.focus)
		rows := loom.Render(tabs, 12, tc.height)
		for y, title := range tc.visible {
			if !strings.Contains(rows[y], title) {
				t.Fatalf("focus %d, height %d: row %d = %q, want %q", tc.focus, tc.height, y, rows[y], title)
			}
		}
		if res := tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 1, Y: tc.click}); !res.Consumed || tabs.Focus() != tc.wantFocus {
			t.Fatalf("click row %d: result %+v, focus %d, want %d", tc.click, res, tabs.Focus(), tc.wantFocus)
		}
	}
	canvas := loom.NewCanvas(12, 1)
	tabs.Draw(canvas, loom.Rect{W: 12})
	if res := tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft}); res.Consumed {
		t.Fatal("zero-height tabs consumed a click")
	}
}

func TestTabsSwitchWithConfiguredKey(t *testing.T) {
	tabs := loom.NewTabs(loom.Tab{Title: "A"}, loom.Tab{Title: "B"})
	tabs.SwitchKey = "ctrl-t"
	tabs.ConsumeKey(loom.KeyEvent{Key: "ctrl-t"})
	if tabs.Focus() != 1 {
		t.Fatalf("after ctrl-t, focus = %d, want 1", tabs.Focus())
	}
}

func TestTabsDynamicLifecycle(t *testing.T) {
	a, b, c, inserted := &focusTabSpy{}, &focusTabSpy{}, &focusTabSpy{}, &focusTabSpy{}
	tabs := loom.NewTabs(
		loom.Tab{Title: "A", Widget: a},
		loom.Tab{Title: "B", Widget: b},
		loom.Tab{Title: "C", Widget: c},
	)
	if !tabs.Select(1) {
		t.Fatal("Select(1) = false, want true")
	}
	if got := tabs.Add(loom.Tab{Title: "D"}); got != 3 {
		t.Fatalf("Add index = %d, want 3", got)
	}
	if err := tabs.Insert(1, loom.Tab{Title: "Inserted", Widget: inserted}); err != nil {
		t.Fatalf("Insert(1): %v", err)
	}
	if tabs.Focus() != 2 || tabs.Tabs[tabs.Focus()].Widget != b {
		t.Fatalf("insert changed active tab: focus=%d", tabs.Focus())
	}
	if err := tabs.Remove(0); err != nil {
		t.Fatalf("Remove(0): %v", err)
	}
	if tabs.Focus() != 1 || tabs.Tabs[tabs.Focus()].Widget != b {
		t.Fatalf("remove before selection changed active tab: focus=%d", tabs.Focus())
	}
	if err := tabs.Remove(1); err != nil {
		t.Fatalf("Remove(selected): %v", err)
	}
	if tabs.Focus() != 1 || tabs.Tabs[tabs.Focus()].Widget != c {
		t.Fatalf("remove selected did not choose successor: focus=%d", tabs.Focus())
	}
	if b.Focused() || !c.Focused() {
		t.Fatalf("child focus not transferred: b=%v c=%v", b.Focused(), c.Focused())
	}
}

func TestTabsLifecycleBoundsAndEmpty(t *testing.T) {
	tabs := loom.NewTabs()
	if tabs.Select(0) || tabs.Select(-1) {
		t.Error("Select should reject indexes for an empty tab list")
	}
	if err := tabs.Remove(0); err == nil {
		t.Error("Remove(0) on empty tabs = nil, want error")
	}
	if err := tabs.Insert(-1, loom.Tab{}); err == nil {
		t.Error("Insert(-1) = nil, want error")
	}
	if err := tabs.Insert(1, loom.Tab{}); err == nil {
		t.Error("Insert(1) on empty tabs = nil, want error")
	}
	if err := tabs.Insert(0, loom.Tab{Title: "A"}); err != nil {
		t.Fatalf("Insert(0) on empty tabs: %v", err)
	}
	if tabs.Focus() != 0 || !tabs.Select(0) {
		t.Fatalf("first inserted tab not selectable: focus=%d", tabs.Focus())
	}
	tabs.SetFocusIndex(99)
	if tabs.Focus() != 0 {
		t.Fatalf("clamped focus = %d, want 0", tabs.Focus())
	}
	tabs.SetTabs()
	if tabs.Focus() != 0 || len(tabs.Tabs) != 0 {
		t.Fatalf("SetTabs() left invalid state: focus=%d tabs=%d", tabs.Focus(), len(tabs.Tabs))
	}
}

func TestTabsSetTabsClampsSelection(t *testing.T) {
	tabs := loom.NewTabs(loom.Tab{}, loom.Tab{}, loom.Tab{})
	tabs.Select(2)
	tabs.SetTabs(loom.Tab{Title: "Only"})
	if tabs.Focus() != 0 {
		t.Fatalf("focus after shrinking tabs = %d, want 0", tabs.Focus())
	}
}

func TestTabsConfiguredNavigation(t *testing.T) {
	tabs := loom.NewTabs(loom.Tab{Title: "A"}, loom.Tab{Title: "B"}, loom.Tab{Title: "C"})
	tabs.SetKeys(loom.TabsKeys{
		Previous: "shift-tab",
		Next:     "tab",
		Cycle:    "ctrl-t",
		Select:   []string{"1", "2", "alt-3"},
	})

	tests := []struct {
		name  string
		event loom.KeyEvent
		want  int
	}{
		{name: "next", event: loom.KeyEvent{Key: "tab"}, want: 1},
		{name: "previous", event: loom.KeyEvent{Key: "shift-tab"}, want: 0},
		{name: "cycle", event: loom.KeyEvent{Key: "ctrl-t"}, want: 1},
		{name: "text direct jump", event: loom.KeyEvent{Text: "1"}, want: 0},
		{name: "named direct jump", event: loom.KeyEvent{Key: "alt-3"}, want: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tabs.ConsumeKey(test.event)
			if tabs.Focus() != test.want {
				t.Fatalf("focus = %d, want %d", tabs.Focus(), test.want)
			}
		})
	}
}

func TestTabsDelegatesKeyToActiveChildOnly(t *testing.T) {
	a, b := &tabSpy{}, &tabSpy{}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: a}, loom.Tab{Title: "B", Widget: b})

	tabs.ConsumeKey(loom.KeyEvent{Text: "x"})
	if len(a.keys) != 1 || len(b.keys) != 0 {
		t.Fatalf("expected only active child (a) to receive the key; a=%d b=%d", len(a.keys), len(b.keys))
	}

	tabs.ConsumeKey(loom.KeyEvent{Key: "right"}) // switch to b
	tabs.ConsumeKey(loom.KeyEvent{Text: "y"})
	if len(a.keys) != 2 || a.keys[1].Key != "right" || len(b.keys) != 1 || b.keys[0].Text != "y" {
		t.Fatalf("expected only b to receive the key after switch; a=%d b=%d", len(a.keys), len(b.keys))
	}
}

func TestTabsQuitPropagatesFromActiveChild(t *testing.T) {
	a := &tabSpy{quitKey: true}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: a})
	if quit := tabs.ConsumeKey(loom.KeyEvent{Text: "z"}).Quit; !quit {
		t.Error("Tabs should propagate the active child's quit=true")
	}
}

func TestTabsDelegatesMouseToActiveChildOnly(t *testing.T) {
	a, b := &tabSpy{}, &tabSpy{}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: a}, loom.Tab{Title: "B", Widget: b})
	c := loom.NewCanvas(40, 10)
	tabs.Draw(c, loom.Rect{X: 0, Y: 0, W: 40, H: 10})

	// A hover event well below the tab bar should reach the active child (a).
	tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MouseHover, X: 5, Y: 5})
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
	// Click at canvas column 4, inside " B ", on the bar's row 0.
	quit := tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 4, Y: 0}).Quit
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

func TestTabsVerticalMouseRoutingUsesChildLocalCoordinates(t *testing.T) {
	a, b := &tabSpy{}, &tabSpy{}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: a}, loom.Tab{Title: "B", Widget: b})
	tabs.Vertical = true
	tabs.ArrowSwitch = false
	c := loom.NewCanvas(40, 10)
	tabs.Draw(c, loom.Rect{X: 5, Y: 3, W: 30, H: 6})

	tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 10, Y: 2})
	if len(a.mice) != 1 || a.mice[0].X != 6 || a.mice[0].Y != 2 {
		t.Fatalf("child click = %#v, want child-local (6,2)", a.mice)
	}
	tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollDown, X: 10, Y: 3})
	if len(a.mice) != 2 || a.mice[1].X != 6 || a.mice[1].Y != 3 || a.mice[1].Action != loom.MouseScrollDown {
		t.Fatalf("child wheel = %#v, want child-local (6,3)", a.mice)
	}

	tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 1, Y: 1})
	if tabs.Focus() != 1 {
		t.Fatalf("focus after vertical tab click = %d, want 1", tabs.Focus())
	}
}

func TestTabsCanForwardArrowsToChildAndUseTabBindings(t *testing.T) {
	child := &tabSpy{}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: child}, loom.Tab{Title: "B"})
	tabs.ArrowSwitch = false
	tabs.SetKeys(loom.TabsKeys{Previous: "shift-tab", Next: "tab"})
	tabs.ConsumeKey(loom.KeyEvent{Key: "left"})
	if len(child.keys) != 1 || child.keys[0].Key != "left" || tabs.Focus() != 0 {
		t.Fatalf("arrow routing: keys=%#v focus=%d", child.keys, tabs.Focus())
	}
	tabs.ConsumeKey(loom.KeyEvent{Key: "tab"})
	if tabs.Focus() != 1 {
		t.Fatalf("focus after Tab = %d, want 1", tabs.Focus())
	}
	tabs.ConsumeKey(loom.KeyEvent{Key: "shift-tab"})
	if tabs.Focus() != 0 {
		t.Fatalf("focus after Shift-Tab = %d, want 0", tabs.Focus())
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
	tabs.ConsumeKey(loom.KeyEvent{Key: "right"}) // no-op, stays on the only tab
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
	if quit := tabs.ConsumeKey(loom.KeyEvent{Key: "right"}).Quit; quit {
		t.Error("zero-tab ConsumeKey should not quit")
	}
	if quit := tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 1, Y: 1}).Quit; quit {
		t.Error("zero-tab ConsumeMouse should not quit")
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
