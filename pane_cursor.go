// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
)

type cursorShapeState struct {
	shape CursorShape
	set   bool
}

func (p *Pane) applyCanvasCursorShape(w interface{ WriteString(string) (int, error) }, canvas *Canvas) {
	if canvas != nil && canvas.CursorShapeSet {
		_ = p.cursorShape.apply(w, canvas.CursorShape)
		return
	}
	_ = p.cursorShape.restore(w)
}

func (s *cursorShapeState) apply(w interface{ WriteString(string) (int, error) }, shape CursorShape) error {
	if s.set && s.shape == shape {
		return nil
	}
	sequence := ""
	switch shape {
	case CursorShapeBar:
		sequence = "\x1b[6 q"
	case CursorShapeBlock:
		sequence = "\x1b[2 q"
	default:
		return fmt.Errorf("loom: unknown cursor shape %d", shape)
	}
	if _, err := w.WriteString(sequence); err != nil {
		return err
	}
	s.shape, s.set = shape, true
	return nil
}

func (s *cursorShapeState) restore(w interface{ WriteString(string) (int, error) }) error {
	if !s.set {
		return nil
	}
	if _, err := w.WriteString("\x1b[0 q"); err != nil {
		return err
	}
	s.set = false
	return nil
}
