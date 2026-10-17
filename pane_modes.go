// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"time"
)

// EnableMouse turns on SGR mouse tracking (button press, release, hover, scroll).
// Call before Run or from an event handler. Mouse events reach the root Widget's
// ConsumeMouse method.
func (p *Pane) EnableMouse() {
	p.baseMouseMode = 1003
	p.reconcileMouseMode()
}

// EnableCursorProximity enables spec-driven cursor hints and background
// brightening. It also enables any-motion reports and must be called before Run.
func (p *Pane) EnableCursorProximity() {
	p.cursorProximity = true
	p.EnableMouse()
}

// EnableCursorStarTrail enables the spec-driven trailing star effect.
// It also enables cursor proximity and any-motion mouse reports.
func (p *Pane) EnableCursorStarTrail() {
	p.cursorStarTrail = true
	p.EnableCursorProximity()
}

// EnableCursorPressPulse enables the one-shot expanding effect for key and
// button presses at the latest known pointer position.
func (p *Pane) EnableCursorPressPulse() {
	p.cursorPressPulse = true
	p.EnableCursorProximity()
}

// EnableCursorEffects enables all currently implemented cursor effects.
func (p *Pane) EnableCursorEffects() {
	p.cursorStarTrail = true
	p.cursorPressPulse = true
	p.EnableCursorProximity()
}

func (p *Pane) triggerCursorPulse(now time.Time) {
	if !p.cursorPressPulse || !p.mouseKnown || !SpeccedCursorPressPulse.Enabled {
		return
	}
	p.cursorPulse = &CursorPulse{X: p.mouseX, Y: p.mouseY, Started: now}
}

func (p *Pane) expireCursorEffects(now time.Time) {
	active := p.cursorTrail[:0]
	for _, point := range p.cursorTrail {
		age := now.Sub(point.At)
		if !point.At.IsZero() && age >= 0 && age < SpeccedCursorStarTrail.Lifetime {
			active = append(active, point)
		}
	}
	p.cursorTrail = active
	if p.cursorPulse != nil && !p.cursorPulse.active(now) {
		p.cursorPulse = nil
	}
}

// EnableMouseClicks tracks clicks and wheel events without any-motion reports.
// Call before Run or from an event handler; Close restores normal mouse behavior.
func (p *Pane) EnableMouseClicks() {
	p.baseMouseMode = 1000
	p.reconcileMouseMode()
}

func (p *Pane) setMouseMode(mode int) {
	if p.mouseMode == mode {
		return
	}
	if p.mouseMode != 0 {
		p.tty.WriteString(fmt.Sprintf("\x1b[?%dl", p.mouseMode)) //nolint:errcheck
	}
	p.mouse, p.mouseMode = true, mode
	p.tty.WriteString(fmt.Sprintf("\x1b[?%dh\x1b[?1006h", mode)) //nolint:errcheck
}

func (p *Pane) disableMouse() {
	if p.mouseMode == 0 {
		return
	}
	p.tty.WriteString(fmt.Sprintf("\x1b[?%dl\x1b[?1006l", p.mouseMode)) //nolint:errcheck
	p.mouse, p.mouseMode = false, 0
}

// DisableMouse sets the base mode to native terminal text selection. Interactive
// overlays temporarily retain click tracking until the last overlay closes.
// Call before Run or from an event handler. EnableMouse re-enables tracking.
func (p *Pane) DisableMouse() {
	p.baseMouseMode = 0
	if p.modalMouseGrab {
		p.reconcileMouseMode()
		return
	}
	p.releaseMouse()
}

func (p *Pane) releaseMouse() {
	if p.tty != nil {
		p.tty.WriteString("\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l") //nolint:errcheck
	}
	p.mouse, p.mouseMode = false, 0
	p.mouseKnown = false
}

// reconcileMouseMode keeps temporary overlay capture independent of the user's
// base mode. Click tracking is sufficient for overlay buttons and scrolling.
func (p *Pane) reconcileMouseMode() {
	mode := p.baseMouseMode
	if mode == 0 && p.modalMouseGrab {
		mode = 1000
	}
	if mode == p.mouseMode {
		return
	}
	if mode == 0 {
		p.releaseMouse()
	} else {
		p.setMouseMode(mode)
	}
}

func (p *Pane) setBracketedPaste(on bool) {
	if p.tty == nil || p.pasteMode == on {
		return
	}
	if on {
		p.tty.WriteString("\x1b[?2004h") //nolint:errcheck
	} else {
		p.tty.WriteString("\x1b[?2004l") //nolint:errcheck
	}
	p.pasteMode = on
}

// Kitty keyboard protocol: push progressive enhancement flag 1 and pop it.
// Each screen (primary, alternate) keeps its own stack, so switchAltScreen
// pops before and pushes after a switch while the protocol is active.
const (
	kittyKeyboardPush = "\x1b[>1u"
	kittyKeyboardPop  = "\x1b[<u"
)

// wantKittyKeyboard reports whether the pane opted in, directly or via spec.
func (p *Pane) wantKittyKeyboard() bool {
	return p.KittyKeyboard || SpeccedDefaults.Pane.KittyKeyboard
}

func (p *Pane) setKittyKeyboard(on bool) {
	if p.tty == nil || p.kittyKeys == on {
		return
	}
	if on {
		p.tty.WriteString(kittyKeyboardPush) //nolint:errcheck
	} else {
		p.tty.WriteString(kittyKeyboardPop) //nolint:errcheck
	}
	p.kittyKeys = on
}
