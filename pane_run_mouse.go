// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "time"

// dispatchMouseRead drains the leading SGR mouse reports from one input read.
// It returns any trailing key bytes to the caller and redraws after each
// report so pointer movement remains visible during a burst.
func (p *Pane) dispatchMouseRead(root Widget, canvas *Canvas, raw []byte, redraw func(bool)) ([]byte, bool) {
	for len(raw) > 0 {
		me, used, ok := scanMouse(raw)
		if !ok {
			break
		}
		// Translate terminal 1-based coordinates to pane/canvas-absolute
		// 0-based coordinates.
		me.X--
		me.Y -= p.startRow
		if p.DebugMode {
			me.X--
			me.Y--
		}
		if p.cursorProximity {
			if p.cursorStarTrail && p.mouseKnown && (me.X != p.mouseX || me.Y != p.mouseY) {
				p.cursorTrail = append(p.cursorTrail, CursorTrailPoint{X: p.mouseX, Y: p.mouseY, At: time.Now()})
				if len(p.cursorTrail) > SpeccedCursorStarTrail.MaxPoints {
					p.cursorTrail = append([]CursorTrailPoint(nil), p.cursorTrail[len(p.cursorTrail)-SpeccedCursorStarTrail.MaxPoints:]...)
				}
			}
			p.mouseX, p.mouseY = me.X, me.Y
			p.mouseKnown = me.X >= 0 && me.X < canvas.Cols() && me.Y >= 0 && me.Y < canvas.Rows()
			if me.Action == MousePress {
				p.triggerCursorPulse(time.Now())
			}
		}
		if p.dispatchMouse(root, me).Quit {
			return raw, true
		}
		raw = raw[used:]
		redraw(false)
	}
	return raw, false
}
