// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "time"

// Notif renders a single-line status message at the bottom of its Rect.
// After Timeout the message is hidden on the next Draw call (timeout of 0
// means the message persists until explicitly cleared).
type Notif struct {
	Message string
	Style   Style
	Timeout time.Duration
	shown   time.Time
}

// Show sets a new message and starts the timeout clock.
func (n *Notif) Show(msg string) {
	n.Message = msg
	n.shown = time.Now()
}

// Clear removes the current message immediately.
func (n *Notif) Clear() { n.Message = "" }

// visible reports whether the message should be shown right now.
func (n *Notif) visible() bool {
	if n.Message == "" {
		return false
	}
	if n.Timeout > 0 && time.Since(n.shown) > n.Timeout {
		n.Message = "" // auto-expire
		return false
	}
	return true
}

// Draw renders the message on the last row of r if visible.
func (n *Notif) Draw(c *Canvas, r Rect) {
	if r.H < 1 {
		return
	}
	y := r.Y + r.H - 1
	c.PaintSurface(Rect{r.X, y, r.W, 1}, n.Style)
	if n.visible() {
		msg := n.Message
		if len([]rune(msg)) > r.W {
			msg = string([]rune(msg)[:r.W])
		}
		c.Write(r.X, y, msg, n.Style)
	}
}

// HandleKey is a no-op; Notif does not consume keyboard events.
func (n *Notif) HandleKey(KeyEvent) (quit bool) { return false }

// HandleMouse is a no-op; Notif does not consume mouse events.
func (n *Notif) HandleMouse(MouseEvent) (quit bool) { return false }
