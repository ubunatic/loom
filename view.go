// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "strings"

// ScrollbarMode controls when a widget displays its scrollbar.
type ScrollbarMode string

const (
	// ScrollbarAuto displays a scrollbar only when content overflows.
	ScrollbarAuto ScrollbarMode = "auto"
	// ScrollbarAlways displays a scrollbar regardless of overflow.
	ScrollbarAlways ScrollbarMode = "always"
	// ScrollbarNever hides the scrollbar and does not reserve its column.
	ScrollbarNever ScrollbarMode = "never"
)

func scrollbarVisible(mode ScrollbarMode, overflow bool) bool {
	if mode == "" {
		mode = SpeccedDefaults.Scrollbar.Mode
	}
	switch mode {
	case ScrollbarAlways:
		return true
	case ScrollbarNever:
		return false
	default:
		return overflow
	}
}

// ScrollbarStyle controls the visual appearance of scrollbar cells.
type ScrollbarStyle struct {
	Track Style
	Thumb Style
}

// DefaultScrollbarStyle returns terminal-default track and thumb styles.
func DefaultScrollbarStyle() ScrollbarStyle {
	return ScrollbarStyle{Track: Style{Dim: true}, Thumb: Style{Bold: true}}
}

// View is a scrollable list of text lines rendered into a Rect.
// When content exceeds the visible area a specced scroll indicator appears on the
// right edge, proportional to the current scroll position.
type View struct {
	Lines      []string
	Scroll     int   // first visible line index
	OffsetX    int   // first visible display column
	OffsetY    int   // first visible line index (alias for Scroll)
	Height     int   // preferred visible height cap; 0 = len(Lines)
	Style      Style // base style for all lines
	FocusStyle Style // style for all lines when focused (falls back to Style if zero)
	Scrollbar  ScrollbarStyle
	// ScrollbarMode overrides the spec default for this widget.
	ScrollbarMode ScrollbarMode
	focused       bool
	lastH         int // height from last Draw; gates scroll in HandleKey
	lastRect      Rect
	drag          scrollbarDrag
}

// Pan moves the view by display columns and lines, clamping at the origin.
func (v *View) Pan(dx, dy int) {
	v.OffsetX = max(0, v.OffsetX+dx)
	v.Scroll = max(0, v.Scroll+dy)
	v.OffsetY = v.Scroll
}

// SetOffset sets the view's display-column and line offsets.
func (v *View) SetOffset(x, y int) {
	v.OffsetX = max(0, x)
	v.Scroll = max(0, y)
	v.OffsetY = v.Scroll
}

// Offset returns the current display-column and line offsets.
func (v *View) Offset() (x, y int) { return v.OffsetX, v.Scroll }

// NewView creates a View from a slice of pre-formatted lines.
func NewView(lines []string) *View {
	return &View{Lines: lines, Scrollbar: DefaultScrollbarStyle()}
}

// Focused reports whether the view currently has input focus.
func (v *View) Focused() bool { return v.focused }

// SetFocus sets whether the view currently has input focus.
func (v *View) SetFocus(focused bool) {
	if !focused && v.drag.active {
		v.Scroll = v.drag.start
		v.OffsetY = v.Scroll
		v.drag.cancel()
	}
	v.focused = focused
}

// Draw renders visible lines into r. When scrollable, the rightmost column is
// reserved for the scroll indicator and content is truncated one column shorter.
func (v *View) Draw(c *Canvas, r Rect) {
	v.lastH = r.H
	v.lastRect = r
	total := len(v.Lines)
	scrollable := scrollbarVisible(v.ScrollbarMode, total > r.H)

	// Clamp scroll.
	if v.Scroll < 0 {
		v.Scroll = 0
	}
	if total > r.H && v.Scroll > total-r.H {
		v.Scroll = total - r.H
	}
	v.OffsetY = v.Scroll

	// Pre-compute indicator row (proportional to scroll position).
	indicatorRow := 0
	if scrollable && total > r.H {
		indicatorRow = scrollbarThumbStart(r.H, scrollbarThumbLength(r.H, total, r.H), v.Scroll, total-r.H)
	}

	contentW := r.W
	if scrollable {
		contentW = r.W - 1 // reserve rightmost column for indicator
	}

	lineStyle := v.Style
	if v.focused && v.FocusStyle != (Style{}) {
		lineStyle = v.FocusStyle
	}

	for row := 0; row < r.H; row++ {
		y := r.Y + row
		c.PaintSurface(Rect{r.X, y, r.W, 1}, lineStyle)
		lineIdx := v.Scroll + row
		if lineIdx >= 0 && lineIdx < total {
			drawViewLine(c, r.X, y, contentW, v.Lines[lineIdx], v.OffsetX, lineStyle)
		}
		if scrollable {
			thumb := max(1, scrollbarThumbLength(r.H, total, r.H))
			c.Set(r.X+r.W-1, y, scrollbarCell(v.Scrollbar, row >= indicatorRow && row < indicatorRow+thumb))
		}
	}
}

func drawViewLine(c *Canvas, x, y, width int, line string, offset int, base Style) {
	col := 0
	out := x
	for _, cell := range ParseANSI(line) {
		if cell.Continuation {
			continue
		}
		cellWidth := StringWidth(cell.Text)
		if cellWidth <= 0 {
			continue
		}
		if col >= offset && out+cellWidth <= x+width {
			style := cell.Style
			if style == (Style{}) {
				style = base
			}
			c.Set(out, y, Cell{Text: cell.Text, Style: style})
			out += cellWidth
		}
		col += cellWidth
		if out >= x+width {
			break
		}
	}
}

// scrollbarCell uses a light track and a solid thumb in the clickable column.
func scrollbarCell(style ScrollbarStyle, thumb bool) Cell {
	if thumb {
		return Cell{Text: SpeccedDefaults.Scrollbar.ForegroundChar, Style: style.Thumb}
	}
	return Cell{Text: SpeccedDefaults.Scrollbar.BackgroundChar, Style: style.Track}
}

// HandleKey supports line, half-page, page, and boundary navigation.
func (v *View) HandleKey(e KeyEvent) (quit bool) {
	maxScroll := len(v.Lines) - v.lastH
	if maxScroll < 0 {
		maxScroll = 0
	}
	key := e.Key
	if key == "" {
		key = e.Text
	}
	page := max(1, v.lastH)
	halfPage := max(1, page/2)
	switch key {
	case "up", "k":
		if v.Scroll > 0 {
			v.Scroll--
			v.OffsetY = v.Scroll
		}
	case "down", "j":
		if v.Scroll < maxScroll {
			v.Scroll++
			v.OffsetY = v.Scroll
		}
	case "left", "h":
		v.OffsetX = max(0, v.OffsetX-1)
	case "right", "l":
		v.OffsetX++
	case "pgdown", "ctrl-f", " ":
		v.Scroll = min(maxScroll, v.Scroll+page)
		v.OffsetY = v.Scroll
	case "pgup", "ctrl-b", "b":
		v.Scroll = max(0, v.Scroll-page)
		v.OffsetY = v.Scroll
	case "ctrl-d":
		v.Scroll = min(maxScroll, v.Scroll+halfPage)
		v.OffsetY = v.Scroll
	case "ctrl-u":
		v.Scroll = max(0, v.Scroll-halfPage)
		v.OffsetY = v.Scroll
	case "home", "g":
		v.Scroll = 0
		v.OffsetY = v.Scroll
	case "end", "G":
		v.Scroll = maxScroll
		v.OffsetY = v.Scroll
	case "esc", "ctrl-c", "q":
		if v.drag.active {
			v.Scroll = v.drag.start
			v.OffsetY = v.Scroll
			v.drag.cancel()
		}
		return true
	}
	return false
}

// HandleMouse supports wheel navigation and clicks in the scrollbar track.
func (v *View) HandleMouse(e MouseEvent) (quit bool) {
	maxScroll := len(v.Lines) - v.lastH
	if maxScroll < 0 {
		maxScroll = 0
	}
	switch e.Action {
	case MouseScrollUp:
		if v.Scroll > 0 {
			v.Scroll--
			v.OffsetY = v.Scroll
		}
	case MouseScrollDown:
		if v.Scroll < maxScroll {
			v.Scroll++
			v.OffsetY = v.Scroll
		}
	case MousePress:
		if e.Button == MouseLeft && maxScroll > 0 && scrollbarVisible(v.ScrollbarMode, true) && v.lastRect.W > 0 &&
			e.X == v.lastRect.X+v.lastRect.W-1 &&
			e.Y >= v.lastRect.Y && e.Y < v.lastRect.Y+v.lastRect.H {
			row := e.Y - v.lastRect.Y
			thumb := scrollbarThumbLength(v.lastRect.H, len(v.Lines), v.lastH)
			start := scrollbarThumbStart(v.lastRect.H, thumb, v.Scroll, maxScroll)
			if row >= start && row < start+thumb {
				v.drag.press(row, start, v.Scroll)
			} else {
				v.Scroll = scrollTrackPosition(row, v.lastRect.H, maxScroll)
				v.OffsetY = v.Scroll
			}
		}
	case MouseDrag:
		if v.drag.active {
			v.Scroll = scrollbarOffset(v.lastRect.H, scrollbarThumbLength(v.lastRect.H, len(v.Lines), v.lastH), e.Y-v.lastRect.Y, v.drag.grab, maxScroll)
			v.OffsetY = v.Scroll
		}
	case MouseRelease:
		if v.drag.active {
			v.drag.cancel()
		}
	}
	return false
}

// scrollTrackPosition maps a clicked track row to the viewport's first item.
func scrollTrackPosition(row, rows, maxOffset int) int {
	if rows < 2 || maxOffset <= 0 {
		return 0
	}
	return row * maxOffset / (rows - 1)
}

// stripANSI removes escape sequences from s, returning plain text.
func stripANSI(s string) string {
	var b strings.Builder
	esc := false
	csi := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			esc = true
		case esc && r == '[':
			csi = true
			esc = false
		case csi && r >= 0x40 && r <= 0x7e:
			csi = false
		case esc:
			esc = false
		case !csi:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ContentHeight estimates the required height for this view.
func (v *View) ContentHeight() int {
	if v.Height > 0 && v.Height < len(v.Lines) {
		return v.Height
	}
	return len(v.Lines)
}

// Selected satisfies Paneable; View is read-only and yields no selection.
func (v *View) Selected() (Item, bool) {
	return Item{}, false
}

// Nav satisfies Paneable; View generates no navigation signals.
func (v *View) Nav() Nav {
	return NavNone
}
