// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"codeberg.org/ubunatic/loom"
)

type keyConsumerWidget struct {
	keys    []loom.KeyEvent
	quit    bool
	consume bool
}

func (w *keyConsumerWidget) Draw(*loom.Canvas, loom.Rect)     {}
func (w *keyConsumerWidget) HandleKey(loom.KeyEvent) bool     { return w.quit }
func (w *keyConsumerWidget) HandleMouse(loom.MouseEvent) bool { return false }
func (w *keyConsumerWidget) ConsumeKey(e loom.KeyEvent) (bool, bool) {
	w.keys = append(w.keys, e)
	return w.quit, w.consume
}

type focusableKeyWidget struct {
	focused bool
}

func (w *focusableKeyWidget) Draw(*loom.Canvas, loom.Rect)     {}
func (w *focusableKeyWidget) HandleKey(loom.KeyEvent) bool     { return false }
func (w *focusableKeyWidget) HandleMouse(loom.MouseEvent) bool { return false }
func (w *focusableKeyWidget) Focused() bool                    { return w.focused }
func (w *focusableKeyWidget) SetFocus(focused bool)            { w.focused = focused }

func TestKeyConsumerChildFirstRouting(t *testing.T) {
	child := &keyConsumerWidget{consume: true}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: child}, loom.Tab{Title: "B"})
	tabs.ArrowSwitch = false

	tabs.HandleKey(loom.KeyEvent{Key: "left"})
	tabs.HandleKey(loom.KeyEvent{Key: "right"})
	if len(child.keys) != 2 || child.keys[0].Key != "left" || child.keys[1].Key != "right" {
		t.Fatalf("child keys = %#v, want left and right", child.keys)
	}
	if tabs.Focus() != 0 {
		t.Fatalf("focus = %d, want 0", tabs.Focus())
	}
}

func TestArrowSwitchTrue(t *testing.T) {
	child := &keyConsumerWidget{consume: true}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: child}, loom.Tab{Title: "B"})

	tabs.HandleKey(loom.KeyEvent{Key: "right"})
	if tabs.Focus() != 1 {
		t.Fatalf("focus after right = %d, want 1", tabs.Focus())
	}
	if len(child.keys) != 0 {
		t.Fatalf("child received keys = %#v, want none", child.keys)
	}
}

func TestOnChildQuitContainment(t *testing.T) {
	child := &keyConsumerWidget{consume: true, quit: true}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: child})
	tabs.OnChildQuit = func(int) bool { return false }
	if tabs.HandleKey(loom.KeyEvent{Key: "q"}) {
		t.Fatal("Tabs propagated a contained child quit")
	}
}

func TestOnChildQuitPropagates(t *testing.T) {
	child := &keyConsumerWidget{consume: true, quit: true}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: child})
	tabs.OnChildQuit = func(int) bool { return true }
	if !tabs.HandleKey(loom.KeyEvent{Key: "q"}) {
		t.Fatal("Tabs suppressed a propagated child quit")
	}
}

func TestTabsFocusClearOnInactive(t *testing.T) {
	a := &focusableKeyWidget{}
	b := &focusableKeyWidget{}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: a}, loom.Tab{Title: "B", Widget: b})
	tabs.Select(1)
	tabs.Draw(loom.NewCanvas(20, 5), loom.Rect{W: 20, H: 5})
	if a.Focused() || !b.Focused() {
		t.Fatalf("focus states = a:%v b:%v, want a:false b:true", a.Focused(), b.Focused())
	}
}

func TestKeyConsumerDoesNotTriggerPaneQuitFallback(t *testing.T) {
	child := &keyConsumerWidget{consume: true}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: child})
	if tabs.HandleKey(loom.KeyEvent{Key: "q"}) {
		t.Fatal("consumed q caused Tabs to quit")
	}
}
