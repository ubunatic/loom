// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "codeberg.org/ubunatic/loom/measure"

// AnsiEditor is an interactive 2D ANSI graphic cell editor widget.
// It embeds an AnsiBuffer and provides viewport scrolling, cursor navigation,
// overtype/insert text editing, and styled cell mutations.
type AnsiEditor struct {
	Buffer *AnsiBuffer

	// Cursor position within the buffer
	CursorX int
	CursorY int

	// Viewport scroll offset
	ScrollX int
	ScrollY int

	// Active styling for new characters
	ActiveFG        Color
	ActiveBG        Color
	ActiveBold      bool
	ActiveDim       bool
	ActiveUnderline bool
	ActiveInvert    bool

	// Active editing mode
	EditMode AnsiEditMode

	// Focus and presentation
	focused    bool
	ShowCursor bool

	// Cached canvas rect from last Draw
	lastRect Rect

	// Optional callbacks
	OnChange     func(buf *AnsiBuffer)
	OnCursorMove func(x, y int)
}

// NewAnsiEditor creates a new AnsiEditor for the given buffer.
// If buf is nil, an empty default buffer (80x24) is created.
func NewAnsiEditor(buf *AnsiBuffer) *AnsiEditor {
	if buf == nil {
		buf = NewAnsiBuffer(80, 24)
	}
	return &AnsiEditor{
		Buffer:     buf,
		ActiveFG:   ColorReset(),
		ActiveBG:   ColorReset(),
		EditMode:   AnsiModeOvertype,
		focused:    true,
		ShowCursor: true,
	}
}

// Focused reports whether the editor has keyboard focus.
func (e *AnsiEditor) Focused() bool {
	return e.focused
}

// SetFocus sets the editor focus state.
func (e *AnsiEditor) SetFocus(f bool) {
	e.focused = f
}

// Cursor returns the current cursor coordinates (x, y).
func (e *AnsiEditor) Cursor() (x, y int) {
	return e.CursorX, e.CursorY
}

// SetCursor moves the cursor to (x, y), clamping to valid buffer bounds.
func (e *AnsiEditor) SetCursor(x, y int) {
	if e.Buffer == nil {
		e.CursorX = 0
		e.CursorY = 0
		return
	}
	if x < 0 {
		x = 0
	}
	if x >= e.Buffer.Cols() {
		x = e.Buffer.Cols() - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= e.Buffer.Rows() {
		y = e.Buffer.Rows() - 1
	}
	e.CursorX = x
	e.CursorY = y
	if e.OnCursorMove != nil {
		e.OnCursorMove(x, y)
	}
}

// Draw renders the visible viewport of the ANSI buffer and the cursor into canvas region r.
func (e *AnsiEditor) Draw(c *Canvas, r Rect) {
	e.lastRect = r
	if e.Buffer == nil || r.W < 1 || r.H < 1 {
		return
	}

	// Clamp cursor to buffer dimensions
	if e.CursorX < 0 {
		e.CursorX = 0
	}
	if e.CursorX >= e.Buffer.Cols() {
		e.CursorX = e.Buffer.Cols() - 1
	}
	if e.CursorY < 0 {
		e.CursorY = 0
	}
	if e.CursorY >= e.Buffer.Rows() {
		e.CursorY = e.Buffer.Rows() - 1
	}

	// Adjust viewport scroll to keep cursor in view
	if e.CursorX < e.ScrollX {
		e.ScrollX = e.CursorX
	}
	if e.CursorX >= e.ScrollX+r.W {
		e.ScrollX = e.CursorX - r.W + 1
	}
	if e.CursorY < e.ScrollY {
		e.ScrollY = e.CursorY
	}
	if e.CursorY >= e.ScrollY+r.H {
		e.ScrollY = e.CursorY - r.H + 1
	}

	for vy := 0; vy < r.H; vy++ {
		by := e.ScrollY + vy
		for vx := 0; vx < r.W; vx++ {
			bx := e.ScrollX + vx
			screenX := r.X + vx
			screenY := r.Y + vy

			cell := e.Buffer.Get(bx, by)
			lCell := cell.ToCell()

			if bx == e.CursorX && by == e.CursorY && e.ShowCursor {
				if e.focused {
					lCell.Style.BG = ColorIndex(15)
					lCell.Style.FG = ColorIndex(0)
					lCell.Style.Bold = true
					c.CursorX = screenX
					c.CursorY = screenY
				} else {
					lCell.Style.Underline = true
				}
			}

			c.Set(screenX, screenY, lCell)
		}
	}
}

// ConsumeKey processes keyboard input, returning an EventResult value struct.
func (e *AnsiEditor) ConsumeKey(ke KeyEvent) EventResult {
	if e.Buffer == nil {
		return Ignored()
	}

	switch {
	case ke.Is("ctrl-left"):
		e.SetCursor(e.Buffer.PrevWord(e.CursorX, e.CursorY), e.CursorY)
		return Handled()
	case ke.Is("ctrl-right"):
		e.SetCursor(e.Buffer.NextWord(e.CursorX, e.CursorY), e.CursorY)
		return Handled()
	case ke.Is("ctrl-up"):
		e.SetCursor(e.CursorX, e.Buffer.PrevObjectRow(e.CursorY))
		return Handled()
	case ke.Is("ctrl-down"):
		e.SetCursor(e.CursorX, e.Buffer.NextObjectRow(e.CursorY))
		return Handled()
	case ke.Is("up"):
		if e.CursorY > 0 {
			e.SetCursor(e.CursorX, e.CursorY-1)
		}
		return Handled()
	case ke.Is("down"):
		if e.CursorY < e.Buffer.Rows()-1 {
			e.SetCursor(e.CursorX, e.CursorY+1)
		}
		return Handled()
	case ke.Is("left"):
		if e.CursorX > 0 {
			e.SetCursor(e.CursorX-1, e.CursorY)
		}
		return Handled()
	case ke.Is("right"):
		if e.CursorX < e.Buffer.Cols()-1 {
			e.SetCursor(e.CursorX+1, e.CursorY)
		}
		return Handled()
	case ke.Is("home"):
		e.SetCursor(0, e.CursorY)
		return Handled()
	case ke.Is("end"):
		e.SetCursor(e.Buffer.Cols()-1, e.CursorY)
		return Handled()
	case ke.Is("pgup", "pageup"):
		e.SetCursor(e.CursorX, max(0, e.CursorY-10))
		return Handled()
	case ke.Is("pgdown", "pgdn", "pagedown"):
		e.SetCursor(e.CursorX, min(e.Buffer.Rows()-1, e.CursorY+10))
		return Handled()
	case ke.Is("insert"):
		if e.EditMode == AnsiModeOvertype {
			e.EditMode = AnsiModeInsert
		} else {
			e.EditMode = AnsiModeOvertype
		}
		return Handled()
	case ke.Is("delete"):
		e.Buffer.Delete(e.CursorX, e.CursorY, e.EditMode)
		if e.OnChange != nil {
			e.OnChange(e.Buffer)
		}
		return Handled()
	case ke.Is("backspace"):
		newX := e.Buffer.Backspace(e.CursorX, e.CursorY, e.EditMode)
		e.SetCursor(newX, e.CursorY)
		if e.OnChange != nil {
			e.OnChange(e.Buffer)
		}
		return Handled()
	case ke.Is("ctrl-c"):
		e.Buffer.Copy(e.CursorX, e.CursorY)
		return Handled()
	case ke.Is("ctrl-x"):
		e.Buffer.Cut(e.CursorX, e.CursorY)
		if e.OnChange != nil {
			e.OnChange(e.Buffer)
		}
		return Handled()
	case ke.Is("ctrl-v"):
		e.Buffer.Paste(e.CursorX, e.CursorY)
		if e.OnChange != nil {
			e.OnChange(e.Buffer)
		}
		return Handled()
	case ke.Is("ctrl-s"):
		if e.Buffer.Path() != "" {
			_ = e.Buffer.SaveFile(e.Buffer.Path())
		}
		return Handled()
	case ke.Is("enter", "return"):
		if e.CursorY < e.Buffer.Rows()-1 {
			e.SetCursor(0, e.CursorY+1)
		}
		return Handled()
	default:
		if ke.Text != "" {
			rs := []rune(ke.Text)
			for _, r := range rs {
				if r >= 32 {
					e.Buffer.PutChar(
						e.CursorX, e.CursorY,
						r,
						e.ActiveFG, e.ActiveBG,
						e.ActiveBold, e.ActiveDim, e.ActiveUnderline, e.ActiveInvert,
						e.EditMode,
					)
					w := measure.RuneWidth(r)
					if w < 1 {
						w = 1
					}
					if e.CursorX+w < e.Buffer.Cols() {
						e.SetCursor(e.CursorX+w, e.CursorY)
					}
				}
			}
			if e.OnChange != nil {
				e.OnChange(e.Buffer)
			}
			return Handled()
		}
	}

	return Ignored()
}

// HandleKey implements the Widget interface.
func (e *AnsiEditor) HandleKey(ke KeyEvent) bool {
	res := e.ConsumeKey(ke)
	return res.Quit
}

// ConsumeMouse processes mouse input, returning an EventResult value struct.
func (e *AnsiEditor) ConsumeMouse(me MouseEvent) EventResult {
	if e.Buffer == nil {
		return Ignored()
	}

	switch me.Action {
	case MouseScrollUp:
		if e.ScrollY > 0 {
			e.ScrollY--
		}
		return Handled()
	case MouseScrollDown:
		if e.ScrollY < e.Buffer.Rows()-1 {
			e.ScrollY++
		}
		return Handled()
	case MousePress, MouseRelease:
		if me.Button == MouseLeft && e.lastRect.W > 0 && e.lastRect.H > 0 {
			if me.X >= e.lastRect.X && me.X < e.lastRect.X+e.lastRect.W &&
				me.Y >= e.lastRect.Y && me.Y < e.lastRect.Y+e.lastRect.H {
				bx := e.ScrollX + (me.X - e.lastRect.X)
				by := e.ScrollY + (me.Y - e.lastRect.Y)
				e.SetCursor(bx, by)
				return Handled()
			}
		}
	}

	return Ignored()
}

// HandleMouse implements the Widget interface.
func (e *AnsiEditor) HandleMouse(me MouseEvent) bool {
	res := e.ConsumeMouse(me)
	return res.Quit
}

// PaneRequest declares terminal capabilities needed by AnsiEditor.
func (e *AnsiEditor) PaneRequest() PaneRequest {
	return PaneRequest{
		Mouse:      1003,
		Resizeable: true,
		OwnsQuit:   false,
	}
}
