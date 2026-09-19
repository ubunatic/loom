// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "strings"

// ChoiceStyle controls the visual appearance of a Choice widget.
type ChoiceStyle struct {
	Normal      Style
	Selected    Style
	Prompt      Style
	Placeholder Style
	Scrollbar   ScrollbarStyle
	Border      Style
}

// DefaultChoiceStyle returns a minimal monochrome style.
func DefaultChoiceStyle() ChoiceStyle {
	sel := Style{Bold: true}
	return ChoiceStyle{Normal: Reset, Selected: sel, Prompt: Reset, Placeholder: Style{Dim: true}, Scrollbar: DefaultScrollbarStyle(), Border: Reset}
}

// Choice is a filterable, keyboard-navigable list of Items.
// Arrow keys move the selection; printable characters narrow the filter;
// Enter confirms; Esc or Ctrl-C aborts.
// Typing ':' or '/' activates command mode (see loom.Cmd, loom.Nav).
type Choice struct {
	Items     []Item
	Style     ChoiceStyle
	Prompt    string          // default "> "
	focused   bool            // dims the border when false so focus is visually clear
	PromptTop bool            // place prompt on first row instead of last row
	OnSelect  func(item Item) // called on Enter or a confirming click; if nil, Enter quits
	// SelectOnlyOnClick keeps a click from confirming; Enter still invokes OnSelect.
	SelectOnlyOnClick bool

	// Input customization
	CursorAlign string // "start" or "end", default "start"
	Placeholder string // ghost placeholder text
	Controls    string // right-aligned controls/status label
	MaxWidth    int    // max width limit

	// MultiSelect turns the list into a checkbox picker: Space (or left-click)
	// toggles the item under the caret, Enter confirms the whole set, Esc aborts.
	// Checks are keyed by Item.Name so they survive filtering.
	MultiSelect bool

	cmd        *cmdBar
	cmdNav     Nav
	query      string
	sel        int
	viewOffset int // first visible item index (virtual scrolling)
	itemRows   int // item rows from the last Draw; excludes the prompt
	drawn      bool
	lastRect   Rect
	aborted    bool
	done       bool
	filtered   []Item
	checked    map[string]bool // MultiSelect: set of checked Item.Name
}

// NewChoice creates a ready-to-use Choice with default style.
func NewChoice(items []Item) *Choice {
	c := &Choice{Items: items, Style: DefaultChoiceStyle(), Prompt: "> ", focused: true}
	c.cmd = newCmdBar()
	c.refilter()
	return c
}

// Nav returns the navigation signal set by a prompt command (:home).
// NavNone means the user made a normal selection or cancelled.
func (c *Choice) Nav() Nav { return c.cmdNav }

// AddCmd registers a view-local command accessible via ':name' in this widget.
func (c *Choice) AddCmd(cmd Cmd) { c.cmd.AddLocal(cmd) }

// Selected returns the chosen Item and whether a selection was made (not aborted).
func (c *Choice) Selected() (Item, bool) {
	if c.aborted || len(c.filtered) == 0 {
		return Item{}, false
	}
	if c.sel >= len(c.filtered) {
		return Item{}, false
	}
	return c.filtered[c.sel], true
}

// Aborted reports whether the user dismissed without selecting.
func (c *Choice) Aborted() bool { return c.aborted }

// Checked returns the checked items in original list order (MultiSelect mode).
// It returns nil when the user aborted (Esc) so a cancelled picker yields no set.
func (c *Choice) Checked() []Item {
	if c.aborted {
		return nil
	}
	out := make([]Item, 0, len(c.checked))
	for _, it := range c.Items {
		if c.checked[it.Name] {
			out = append(out, it)
		}
	}
	return out
}

// toggleChecked flips the checked state of the item under the caret.
func (c *Choice) toggleChecked() {
	if c.sel < 0 || c.sel >= len(c.filtered) {
		return
	}
	if c.checked == nil {
		c.checked = make(map[string]bool)
	}
	name := c.filtered[c.sel].Name
	c.checked[name] = !c.checked[name]
}

// FilteredSel returns the index of the current selection within the filtered list.
func (c *Choice) FilteredSel() int { return c.sel }

// FilteredItem returns the item at index i in the current filtered list.
// Returns a zero Item if i is out of range.
func (c *Choice) FilteredItem(i int) Item {
	if i < 0 || i >= len(c.filtered) {
		return Item{}
	}
	return c.filtered[i]
}

func (c *Choice) refilter() {
	q := strings.ToLower(c.query)
	if q == "" {
		c.filtered = c.Items
	} else {
		out := make([]Item, 0, len(c.Items))
		for _, it := range c.Items {
			if strings.Contains(strings.ToLower(it.Name), q) ||
				strings.Contains(strings.ToLower(it.Desc), q) {
				out = append(out, it)
			}
		}
		c.filtered = out
	}
	if c.sel >= len(c.filtered) {
		c.sel = max(0, len(c.filtered)-1)
	}
}

// clampView keeps viewOffset so that sel is always visible within itemRows.
func (c *Choice) clampView(itemRows int) {
	if c.viewOffset > c.sel {
		c.viewOffset = c.sel
	}
	if c.sel >= c.viewOffset+itemRows {
		c.viewOffset = c.sel - itemRows + 1
	}
	if c.viewOffset < 0 {
		c.viewOffset = 0
	}
}

// Draw renders the choice list into r.
// PromptTop=false (default): items fill the top, prompt occupies the last row.
// PromptTop=true:            prompt occupies the first row, items fill the rest.
// When items exceed the visible area a scroll indicator appears on the right edge.
func (c *Choice) Draw(cv *Canvas, r Rect) {
	itemRows := r.H - 1
	if itemRows < 0 {
		itemRows = 0
	}
	c.itemRows = itemRows
	c.drawn = true
	c.lastRect = r

	c.clampView(itemRows)

	scrollable := len(c.filtered) > itemRows
	contentW := r.W
	if scrollable {
		contentW = r.W - 1 // reserve right column for indicator
	}

	// Pre-compute scroll indicator row.
	indicatorRow := -1
	if scrollable {
		total := len(c.filtered)
		maxOffset := total - itemRows
		if maxOffset > 0 {
			ratio := float64(c.viewOffset) / float64(maxOffset)
			indicatorRow = int(ratio * float64(itemRows-1))
		}
	}

	// Determine first item row and prompt row.
	firstItemY := r.Y
	promptY := r.Y + r.H - 1
	if c.PromptTop {
		firstItemY = r.Y + 1
		promptY = r.Y
	}

	// Draw item rows.
	for row := 0; row < itemRows; row++ {
		y := firstItemY + row
		cv.PaintSurface(Rect{r.X, y, r.W, 1}, c.Style.Normal)
		fi := c.viewOffset + row
		if fi >= 0 && fi < len(c.filtered) {
			item := c.filtered[fi]
			style := c.Style.Normal
			marker := "  "
			if fi == c.sel {
				style = c.Style.Selected
				marker = "▶ "
			}
			if c.MultiSelect {
				// Caret marker + checkbox: "▶ [✓] name" / "  [ ] name".
				if c.checked[item.Name] {
					marker += "[✓] "
				} else {
					marker += "[ ] "
				}
			}
			line := marker + item.Name
			if item.Desc != "" {
				line += "  " + item.Desc
			}
			runes := []rune(line)
			if len(runes) > contentW {
				line = string(runes[:contentW])
			}
			cv.Write(r.X, y, line, style)
		}
		if scrollable {
			cv.Set(r.X+r.W-1, y, scrollbarCell(c.Style.Scrollbar, row == indicatorRow))
		}
	}

	// Apply MaxWidth constraint if defined
	drawW := r.W
	if c.MaxWidth > 0 && drawW > c.MaxWidth {
		drawW = c.MaxWidth
	}

	// Prompt row.
	cv.PaintSurface(Rect{r.X, promptY, drawW, 1}, c.Style.Prompt)
	if prefix, hint := c.cmd.PromptParts(); prefix != "" {
		// Command mode: ":typed[completion]  dim title"
		n := cv.Write(r.X, promptY, prefix, c.Style.Prompt)
		if hint != "" {
			cv.Write(r.X+n, promptY, hint, Style{Dim: true})
		}
		if c.focused {
			cv.CursorX = r.X + n
			cv.CursorY = promptY
		}
	} else {
		// Normal mode: base prompt + filter query (or placeholder)
		if c.query == "" && c.Placeholder != "" {
			n := cv.Write(r.X, promptY, c.Prompt, c.Style.Prompt)
			cv.Write(r.X+n, promptY, c.Placeholder, c.Style.Placeholder)
		} else {
			cv.Write(r.X, promptY, c.Prompt+c.query, c.Style.Prompt)
		}

		if c.Controls != "" {
			ctrlW := len([]rune(c.Controls))
			if drawW > ctrlW+len([]rune(c.Prompt)) {
				cv.Write(r.X+drawW-ctrlW, promptY, c.Controls, Style{Dim: true})
			}
		}

		if c.focused {
			cursorX := r.X + StringWidth(c.Prompt) + StringWidth(c.query)
			if c.CursorAlign == "end" {
				cursorX = r.X + drawW - 1
			}
			if cursorX >= r.X && cursorX < r.X+drawW {
				cv.CursorX = cursorX
				cv.CursorY = promptY
			}
		}
	}
}

// HandleKey drives navigation and filtering.
// ':' or '/' activates command mode; all other keys behave normally when inactive.
func (c *Choice) HandleKey(e KeyEvent) (quit bool) {
	if consumed, result := c.cmd.HandleKey(e); consumed {
		switch result {
		case cmdBack:
			c.aborted = true
			c.done = true
			return true
		case cmdHome:
			c.cmdNav = NavHome
			c.aborted = true
			c.done = true
			return true
		}
		return false
	}
	switch e.Key {
	case "esc", "ctrl-c", "ctrl-d", "ctrl-q":
		c.aborted = true
		c.done = true
		return true
	case "enter":
		if c.MultiSelect {
			// Confirm the whole checked set; OnSelect/single-select rules don't apply.
			c.done = true
			return true
		}
		if c.OnSelect != nil && len(c.filtered) > 0 {
			c.OnSelect(c.filtered[c.sel])
			return false
		}
		c.done = true
		return true
	case "up":
		if c.sel > 0 {
			c.sel--
		} else {
			c.sel = max(0, len(c.filtered)-1)
		}
	case "down":
		if c.sel < len(c.filtered)-1 {
			c.sel++
		} else {
			c.sel = 0
		}
	case "pgup", "pageup":
		step := max(1, c.itemRows-1)
		c.sel -= step
		if c.sel < 0 {
			c.sel = 0
		}
	case "pgdown", "pgdn", "pagedown":
		step := max(1, c.itemRows-1)
		c.sel += step
		if c.sel >= len(c.filtered) {
			c.sel = max(0, len(c.filtered)-1)
		}
	case "home":
		c.sel = 0
	case "end":
		c.sel = max(0, len(c.filtered)-1)
	case "backspace":
		if len(c.query) > 0 {
			runes := []rune(c.query)
			c.query = string(runes[:len(runes)-1])
			c.refilter()
		} else {
			// Backspace past an empty filter leaves the view, mirroring Esc/`:back`.
			c.aborted = true
			c.done = true
			return true
		}
	default:
		if c.MultiSelect && e.Text == " " {
			c.toggleChecked() // Space toggles the checkbox instead of filtering
			return false
		}
		if e.Text != "" {
			c.query += e.Text
			c.refilter()
		}
	}
	return false
}

// HandleMouse updates selection on hover or wheel, confirms on left-click,
// and jumps the viewport when the scrollbar track is clicked.
// e.Y is the 1-based widget-relative row, so e.Y-1 is the visible item row;
// viewOffset maps that back to a filtered index (mirrors Draw's fi mapping)
// so hit-tests stay correct once the list has been scrolled.
func (c *Choice) HandleMouse(e MouseEvent) (quit bool) {
	if e.Action == MousePress && e.Button == MouseLeft && c.drawn &&
		c.lastRect.W > 0 && c.itemRows > 0 &&
		e.X == c.lastRect.X+c.lastRect.W && len(c.filtered) > c.itemRows {
		row := e.Y - c.lastRect.Y - 1
		if c.PromptTop {
			row--
		}
		if row >= 0 && row < c.itemRows {
			c.viewOffset = scrollTrackPosition(row, c.itemRows, len(c.filtered)-c.itemRows)
			c.sel = max(c.viewOffset, min(c.sel, c.viewOffset+c.itemRows-1))
		}
		return false
	}
	switch e.Action {
	case MouseScrollUp:
		if c.sel > 0 {
			c.sel--
		}
		return false
	case MouseScrollDown:
		if c.sel < len(c.filtered)-1 {
			c.sel++
		}
		return false
	}
	if c.drawn {
		e.Y -= c.lastRect.Y
	}
	if c.PromptTop {
		e.Y--
	}
	if e.Y <= 0 {
		return false
	}
	if c.drawn && e.Y > c.itemRows {
		return false
	}
	fi := c.viewOffset + e.Y - 1
	if fi < 0 || fi >= len(c.filtered) {
		return false
	}
	switch e.Action {
	case MousePress:
		if e.Button == MouseLeft {
			c.sel = fi
			if c.MultiSelect {
				c.toggleChecked() // click toggles the checkbox, never confirms
				return false
			}
			if c.SelectOnlyOnClick {
				return false
			}
			if c.OnSelect != nil {
				c.OnSelect(c.filtered[c.sel])
				return false
			}
			c.done = true
			return true
		}
	case MouseHover, MouseDrag:
		c.sel = fi
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Focused reports whether this widget is focused.
func (c *Choice) Focused() bool {
	return c.focused
}

// SetFocus sets the focus state of this widget.
func (c *Choice) SetFocus(f bool) {
	c.focused = f
}

// ContentHeight estimates the required height for this choice widget.
func (c *Choice) ContentHeight() int {
	h := len(c.Items)
	if h > 0 {
		return h + 1
	}
	return 1
}
