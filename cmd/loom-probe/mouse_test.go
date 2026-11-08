// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom"
)

func newTestMouseState() *mouseState {
	return &mouseState{counts: map[loom.MouseAction]int{}}
}

func mouseEvent(action loom.MouseAction, button loom.MouseButton, x, y int) loom.MouseEvent {
	return loom.MouseEvent{Action: action, Button: button, X: x, Y: y}
}

func TestMouseStateGestureCompletes(t *testing.T) {
	s := newTestMouseState()
	now := time.Now()
	s.apply(mouseEvent(loom.MousePress, loom.MouseLeft, 5, 5), now, 80, 24)
	s.apply(mouseEvent(loom.MouseDrag, loom.MouseLeft, 6, 5), now, 80, 24)
	if !s.down[0] {
		t.Fatal("left button not held after press and drag")
	}
	s.apply(mouseEvent(loom.MouseRelease, loom.MouseLeft, 7, 5), now, 80, 24)
	if s.anyDown() || s.noPress != 0 || len(s.history) != 1 {
		t.Fatalf("after release: down=%v noPress=%d history=%v", s.down, s.noPress, s.history)
	}
}

func TestMouseStateHoverEndsDragButKeepsAssumedHeld(t *testing.T) {
	s := newTestMouseState()
	now := time.Now()
	s.apply(mouseEvent(loom.MousePress, loom.MouseLeft, 5, 5), now, 80, 24)
	s.apply(mouseEvent(loom.MouseDrag, loom.MouseLeft, 79, 5), now, 80, 24)
	s.apply(mouseEvent(loom.MouseHover, loom.MouseNone, 85, 5), now, 80, 24)
	if !s.down[0] || !s.stale[0] || s.lostRel != 1 || s.outside != 1 {
		t.Fatalf("down=%v stale=%v lostRel=%d outside=%d", s.down, s.stale, s.lostRel, s.outside)
	}
	if !strings.Contains(s.lostNote, "inside") {
		t.Fatalf("lostNote = %q, want the last drag position", s.lostNote)
	}
	s.apply(mouseEvent(loom.MouseDrag, loom.MouseLeft, 10, 5), now, 80, 24)
	if s.stale[0] {
		t.Fatal("a new drag report should clear the stale mark")
	}
}

func TestMouseStateDragWithoutPressIsInferred(t *testing.T) {
	s := newTestMouseState()
	s.apply(mouseEvent(loom.MouseDrag, loom.MouseRight, 3, 3), time.Now(), 80, 24)
	if !s.down[2] || !s.inferred[2] || s.noPress != 1 {
		t.Fatalf("down=%v inferred=%v noPress=%d", s.down, s.inferred, s.noPress)
	}
}
