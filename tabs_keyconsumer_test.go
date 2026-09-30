// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"codeberg.org/ubunatic/loom"
)

type keyConsumerWidget struct {
	keys   []loom.KeyEvent
	result loom.EventResult
}

func (w *keyConsumerWidget) Draw(*loom.Canvas, loom.Rect)                  {}
func (w *keyConsumerWidget) ConsumeMouse(loom.MouseEvent) loom.EventResult { return loom.Ignored() }
func (w *keyConsumerWidget) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	w.keys = append(w.keys, e)
	return w.result
}

type focusableKeyWidget struct {
	focused bool
}

func (w *focusableKeyWidget) Draw(*loom.Canvas, loom.Rect)                  {}
func (w *focusableKeyWidget) ConsumeKey(loom.KeyEvent) loom.EventResult     { return loom.Ignored() }
func (w *focusableKeyWidget) ConsumeMouse(loom.MouseEvent) loom.EventResult { return loom.Ignored() }
func (w *focusableKeyWidget) Focused() bool                                 { return w.focused }
func (w *focusableKeyWidget) SetFocus(focused bool)                         { w.focused = focused }

func TestEventResultChildFirstRouting(t *testing.T) {
	child := &keyConsumerWidget{result: loom.Handled()}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: child}, loom.Tab{Title: "B"})
	tabs.ArrowSwitch = false

	tabs.ConsumeKey(loom.KeyEvent{Key: "left"})
	tabs.ConsumeKey(loom.KeyEvent{Key: "right"})
	if len(child.keys) != 2 || child.keys[0].Key != "left" || child.keys[1].Key != "right" {
		t.Fatalf("child keys = %#v, want left and right", child.keys)
	}
	if tabs.Focus() != 0 {
		t.Fatalf("focus = %d, want 0", tabs.Focus())
	}
}

func TestArrowSwitchTrue(t *testing.T) {
	child := &keyConsumerWidget{result: loom.Handled()}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: child}, loom.Tab{Title: "B"})

	tabs.ConsumeKey(loom.KeyEvent{Key: "right"})
	if tabs.Focus() != 0 {
		t.Fatalf("focus after consumed right = %d, want 0", tabs.Focus())
	}
	if len(child.keys) != 1 {
		t.Fatalf("child received keys = %#v, want right", child.keys)
	}
}

func TestArrowSwitchRunsWhenChildIgnoresKey(t *testing.T) {
	child := &keyConsumerWidget{result: loom.Ignored()}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: child}, loom.Tab{Title: "B"})
	if result := tabs.ConsumeKey(loom.KeyEvent{Key: "right"}); !result.Consumed || tabs.Focus() != 1 {
		t.Fatalf("right = %+v, focus=%d, want handled at tab 1", result, tabs.Focus())
	}
}

func TestOnChildQuitContainment(t *testing.T) {
	child := &keyConsumerWidget{result: loom.QuitResult()}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: child})
	tabs.OnChildQuit = func(int) bool { return false }
	if result := tabs.ConsumeKey(loom.KeyEvent{Key: "q"}); !result.Consumed || result.Quit {
		t.Fatal("Tabs propagated a contained child quit")
	}
}

func TestOnChildQuitPropagates(t *testing.T) {
	child := &keyConsumerWidget{result: loom.QuitResult()}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: child})
	tabs.OnChildQuit = func(int) bool { return true }
	if result := tabs.ConsumeKey(loom.KeyEvent{Key: "q"}); !result.Quit {
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

func TestEventResultDoesNotTriggerPaneQuitFallback(t *testing.T) {
	child := &keyConsumerWidget{result: loom.Handled()}
	tabs := loom.NewTabs(loom.Tab{Title: "A", Widget: child})
	if result := tabs.ConsumeKey(loom.KeyEvent{Key: "q"}); result.Quit {
		t.Fatal("consumed q caused Tabs to quit")
	}
}
