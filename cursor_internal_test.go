package loom

import (
	"testing"
	"time"
)

func TestCursorAnimationFramesOnlyWhileEffectsAreActive(t *testing.T) {
	now := time.Unix(400, 0)
	if got := cursorAnimationFrameInterval(nil, nil, now); got != 0 {
		t.Fatalf("empty animation interval = %v; want zero", got)
	}
	trail := []CursorTrailPoint{{X: 1, Y: 1, At: now.Add(-SpeccedCursorStarTrail.Lifetime)}}
	if got := cursorAnimationFrameInterval(trail, nil, now); got != 0 {
		t.Fatalf("expired trail interval = %v; want zero", got)
	}
	trail[0].At = now
	if got := cursorAnimationFrameInterval(trail, nil, now); got != SpeccedCursorStarTrail.Frame {
		t.Fatalf("active trail interval = %v; want %v", got, SpeccedCursorStarTrail.Frame)
	}
	pulse := &CursorPulse{X: 2, Y: 2, Started: now}
	if got := cursorAnimationFrameInterval(nil, pulse, now); got != SpeccedCursorPressPulse.Frame {
		t.Fatalf("active pulse interval = %v; want %v", got, SpeccedCursorPressPulse.Frame)
	}
	pulse.Started = now.Add(-SpeccedCursorPressPulse.Lifetime)
	if got := cursorAnimationFrameInterval(nil, pulse, now); got != 0 {
		t.Fatalf("expired pulse interval = %v; want zero", got)
	}
}

func TestCursorPressPulseStartsAtKnownCursor(t *testing.T) {
	now := time.Unix(500, 0)
	p := &Pane{cursorPressPulse: true, mouseKnown: true, mouseX: 4, mouseY: 3}
	p.triggerCursorPulse(now)
	if p.cursorPulse == nil || p.cursorPulse.X != 4 || p.cursorPulse.Y != 3 || p.cursorPulse.Started != now {
		t.Fatalf("triggered cursor pulse = %+v", p.cursorPulse)
	}
	p.mouseKnown = false
	p.triggerCursorPulse(now.Add(time.Second))
	if p.cursorPulse.Started != now {
		t.Fatalf("pulse triggered without known cursor: %+v", p.cursorPulse)
	}
}
