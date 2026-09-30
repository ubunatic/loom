// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"sort"
	"strconv"
	"strings"
	"unicode"

	"codeberg.org/ubunatic/loom/measure"
)

// ChoiceStyle controls the visual appearance of a Choice widget.
type ChoiceStyle struct {
	Normal      Style
	Selected    Style
	Match       Style // style for matched runes; zero uses the row style with bold and underline enabled
	Prompt      Style
	Placeholder Style
	Scrollbar   ScrollbarStyle
	Border      Style
}

// DefaultChoiceStyle returns a minimal monochrome style, derived from the plain theme.
func DefaultChoiceStyle() ChoiceStyle {
	return Theme("plain").ChoiceStyle()
}

// Choice is a filterable, keyboard-navigable list of Items.
// Arrow keys move the selection; printable characters narrow the filter;
// Enter confirms; Esc or Ctrl-C aborts.
// Typing ':' or '/' activates command mode (see loom.Cmd, loom.Nav).
type Choice struct {
	Items []Item
	Style ChoiceStyle
	// Fuzzy enables subsequence matching and score-ranked results.
	Fuzzy bool
	// ScrollbarMode overrides the spec default for this widget.
	ScrollbarMode ScrollbarMode
	Prompt        string          // default "> "
	focused       bool            // dims the border when false so focus is visually clear
	PromptTop     bool            // place prompt on first row instead of last row
	OnSelect      func(item Item) // called on Enter or an activating click; if nil, Enter quits
	// SelectOnlyOnClick keeps a single click from confirming; Enter still invokes OnSelect.
	SelectOnlyOnClick bool
	// DoubleClickToActivate makes a matching second click activate the selected
	// item. The first click only selects; a double-click can activate even when
	// SelectOnlyOnClick is set.
	DoubleClickToActivate bool
	// MouseTextOnly restricts mouse selection and activation to rendered item content.
	MouseTextOnly bool

	// Input customization
	CursorAlign string // "start" or "end", default "start"
	Placeholder string // ghost placeholder text
	Controls    string // right-aligned controls/status label
	MaxWidth    int    // max width limit

	// MultiSelect turns the list into a checkbox picker: Space (or left-click)
	// toggles the item under the caret, Enter confirms the whole set, Esc aborts.
	// Checks are keyed by Item.Name so they survive filtering.
	MultiSelect bool

	cmd         *cmdBar
	cmdNav      Nav
	query       string
	sel         int
	viewOffset  int // first visible item index (virtual scrolling)
	itemRows    int // item rows from the last Draw; excludes the prompt
	drawn       bool
	lastRect    Rect
	drag        scrollbarDrag
	doubleClick *DoubleClickRecognizer
	aborted     bool
	done        bool
	filtered    []Item
	checked     map[string]bool // MultiSelect: set of checked Item.Name
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

// Query returns the current filter query string.
func (c *Choice) Query() string { return c.query }

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

// SetItems replaces the item list, clears the filter query, and resets selection.
func (c *Choice) SetItems(items []Item) {
	c.Items = items
	c.query = ""
	c.sel = 0
	c.viewOffset = 0
	c.aborted = false
	c.done = false
	c.refilter()
}

// SelectIndex sets the selection to index i if in range.
func (c *Choice) SelectIndex(i int) {
	if len(c.filtered) == 0 {
		c.sel = 0
		return
	}
	c.sel = max(0, min(i, len(c.filtered)-1))
}

func (c *Choice) refilter() {
	q := strings.ToLower(c.query)
	if q == "" {
		c.filtered = c.Items
	} else if c.Fuzzy {
		type rankedItem struct {
			item  Item
			score int
		}
		ranked := make([]rankedItem, 0, len(c.Items))
		for _, it := range c.Items {
			nameScore, nameOK := fuzzyScore(it.Name, q)
			descScore, descOK := fuzzyScore(it.Desc, q)
			if nameOK || descOK {
				score := descScore
				if nameOK && (!descOK || nameScore >= descScore) {
					score = nameScore
				}
				ranked = append(ranked, rankedItem{item: it, score: score})
			}
		}
		sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })
		out := make([]Item, len(ranked))
		for i := range ranked {
			out[i] = ranked[i].item
		}
		c.filtered = out
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

// fuzzyScore scores a rune subsequence, rewarding adjacent matches and
// matches at the beginning of a word. The returned score is zero on no match.
func fuzzyScore(text, query string) (score int, matched bool) {
	textRunes := []rune(strings.ToLower(text))
	queryRunes := []rune(query)
	if len(queryRunes) == 0 {
		return 0, true
	}
	qi, previous := 0, -2
	for i, r := range textRunes {
		if r != queryRunes[qi] {
			continue
		}
		score += 1
		if i == 0 || !unicode.IsLetter(textRunes[i-1]) && !unicode.IsNumber(textRunes[i-1]) {
			score += 4
		}
		if i == previous+1 {
			score += 5
		}
		previous = i
		qi++
		if qi == len(queryRunes) {
			return score, true
		}
	}
	return 0, false
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

	scrollable := scrollbarVisible(c.ScrollbarMode, len(c.filtered) > itemRows)
	contentW := r.W
	if scrollable {
		contentW = r.W - 1 // reserve right column for indicator
	}

	// Pre-compute scroll indicator row.
	indicatorRow := 0
	if scrollable {
		total := len(c.filtered)
		maxOffset := total - itemRows
		if maxOffset > 0 {
			indicatorRow = scrollbarThumbStart(itemRows, scrollbarThumbLength(itemRows, total, itemRows), c.viewOffset, maxOffset)
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
			style := c.Style.Normal
			if fi == c.sel {
				style = c.Style.Selected
			}
			line := c.choiceRowText(fi, contentW)
			cv.Write(r.X, y, line, style)
			if c.Fuzzy && c.query != "" {
				item := c.filtered[fi]
				markerWidth := measure.StringWidth(c.choiceMarker(fi, item))
				matchStyle := c.Style.Match
				if matchStyle == (Style{}) {
					matchStyle = style
					matchStyle.Bold = true
					matchStyle.Underline = true
				}
				nameOffset := markerWidth
				c.highlightMatch(cv, r.X+nameOffset, y, item.Name, matchStyle, contentW-nameOffset)
				if item.Desc != "" {
					descOffset := nameOffset + measure.StringWidth(item.Name) + 2
					c.highlightMatch(cv, r.X+descOffset, y, item.Desc, matchStyle, contentW-descOffset)
				}
			}
		}
		if scrollable {
			thumb := max(1, scrollbarThumbLength(itemRows, len(c.filtered), itemRows))
			cv.Set(r.X+r.W-1, y, scrollbarCell(c.Style.Scrollbar, row >= indicatorRow && row < indicatorRow+thumb))
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
			ctrlW := StringWidth(c.Controls)
			if drawW > ctrlW+StringWidth(c.Prompt) {
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
	c.cmd.drawHelp(cv, r)
}

// ConsumeKey drives navigation and filtering.
// ':' or '/' activates command mode; all other keys behave normally when inactive.
func (c *Choice) ConsumeKey(e KeyEvent) EventResult {
	if e.Is("esc") && c.drag.active {
		c.viewOffset = c.drag.start
		c.drag.cancel()
	}
	if c.cmd.handleHelp(e) {
		return Handled()
	}
	if consumed, result := c.cmd.ConsumeKey(e); consumed {
		switch result {
		case cmdBack:
			c.aborted = true
			c.done = true
			return QuitResult()
		case cmdHome:
			c.cmdNav = NavHome
			c.aborted = true
			c.done = true
			return QuitResult()
		}
		return Handled()
	}
	switch e.Key {
	case "esc", "ctrl-c", "ctrl-d", "ctrl-q":
		c.aborted = true
		c.done = true
		return QuitResult()
	case "enter":
		if c.MultiSelect {
			// Confirm the whole checked set; OnSelect/single-select rules don't apply.
			c.done = true
			return DoneResult()
		}
		if c.OnSelect != nil && len(c.filtered) > 0 {
			c.OnSelect(c.filtered[c.sel])
			return Handled()
		}
		c.done = true
		return DoneResult()
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
		}
	default:
		if c.MultiSelect && e.Text == " " {
			c.toggleChecked() // Space toggles the checkbox instead of filtering
			return Handled()
		}
		if e.Text != "" {
			c.query += e.Text
			c.refilter()
			return Handled()
		}
		return Ignored()
	}
	return Handled()
}

// ConsumeMouse updates selection on hover or wheel, confirms on left-click,
// and jumps the viewport when the scrollbar track is clicked.
// e.Y is a widget-local 0-based row;
// viewOffset maps that back to a filtered index (mirrors Draw's fi mapping)
// so hit-tests stay correct once the list has been scrolled.
func (c *Choice) ConsumeMouse(e MouseEvent) EventResult {
	if c.cmd.handleHelpMouse(e) {
		c.observeDoubleClick(e, "")
		return Handled()
	}
	if e.Action == MousePress && e.Button == MouseLeft && c.drawn &&
		c.lastRect.W > 0 && c.itemRows > 0 &&
		e.X == c.lastRect.W-1 && len(c.filtered) > c.itemRows && scrollbarVisible(c.ScrollbarMode, true) {
		c.observeDoubleClick(e, "")
		row := e.Y
		if c.PromptTop {
			row--
		}
		if row >= 0 && row < c.itemRows {
			maxOffset := len(c.filtered) - c.itemRows
			thumb := max(1, scrollbarThumbLength(c.itemRows, len(c.filtered), c.itemRows))
			start := scrollbarThumbStart(c.itemRows, thumb, c.viewOffset, maxOffset)
			if row >= start && row < start+thumb {
				c.drag.press(row, start, c.viewOffset)
			} else {
				c.viewOffset = scrollTrackPosition(row, c.itemRows, maxOffset)
			}
			c.sel = max(c.viewOffset, min(c.sel, c.viewOffset+c.itemRows-1))
		}
		return Handled()
	}
	if c.drag.active {
		c.observeDoubleClick(e, "")
		switch e.Action {
		case MouseDrag:
			c.viewOffset = scrollbarOffset(c.itemRows, scrollbarThumbLength(c.itemRows, len(c.filtered), c.itemRows), e.Y, c.drag.grab, len(c.filtered)-c.itemRows)
			c.sel = max(c.viewOffset, min(c.sel, c.viewOffset+c.itemRows-1))
			return Handled()
		case MouseRelease:
			c.drag.cancel()
			return Handled()
		}
	}
	switch e.Action {
	case MouseScrollUp:
		c.observeDoubleClick(e, "")
		if c.sel > 0 {
			c.sel--
		}
		return Handled()
	case MouseScrollDown:
		c.observeDoubleClick(e, "")
		if c.sel < len(c.filtered)-1 {
			c.sel++
		}
		return Handled()
	}
	if c.PromptTop {
		e.Y--
	}
	if e.Y < 0 {
		c.observeDoubleClick(e, "")
		return Ignored()
	}
	if c.drawn && e.Y >= c.itemRows {
		c.observeDoubleClick(e, "")
		return Ignored()
	}
	fi := c.viewOffset + e.Y
	if fi < 0 || fi >= len(c.filtered) {
		c.observeDoubleClick(e, "")
		return Ignored()
	}
	if c.MouseTextOnly && (e.Action == MousePress || e.Action == MouseHover || e.Action == MouseDrag) {
		contentW := c.lastRect.W
		if len(c.filtered) > c.itemRows && scrollbarVisible(c.ScrollbarMode, true) {
			contentW--
		}
		_, end, ok := choiceMouseHitRegion(c.choiceRowText(fi, contentW), contentW)
		if !ok || e.X < 0 || e.X >= end {
			c.observeDoubleClick(e, "")
			return Handled()
		}
	}
	doubleClick := c.observeDoubleClick(e, strconv.Itoa(fi)+"\x00"+c.filtered[fi].Name)
	switch e.Action {
	case MousePress:
		if e.Button == MouseLeft {
			c.sel = fi
			if c.MultiSelect {
				c.toggleChecked() // click toggles the checkbox, never confirms
				return Handled()
			}
			if c.DoubleClickToActivate {
				if doubleClick {
					if c.OnSelect != nil {
						c.OnSelect(c.filtered[c.sel])
						return Handled()
					}
					c.done = true
					return DoneResult()
				}
				return Handled()
			}
			if c.SelectOnlyOnClick {
				return Handled()
			}
			if c.OnSelect != nil {
				c.OnSelect(c.filtered[c.sel])
				return Handled()
			}
			c.done = true
			return DoneResult()
		}
	case MouseHover, MouseDrag:
		c.sel = fi
	}
	return Ignored()
}

func (c *Choice) observeDoubleClick(e MouseEvent, target string) bool {
	if !c.DoubleClickToActivate || c.MultiSelect {
		if c.doubleClick != nil {
			c.doubleClick.Handle(e, "")
		}
		return false
	}
	if c.doubleClick == nil {
		c.doubleClick = NewDoubleClickRecognizer(nil)
	}
	return c.doubleClick.Handle(e, target)
}

func (c *Choice) choiceRowText(fi, width int) string {
	if fi < 0 || fi >= len(c.filtered) {
		return ""
	}
	item := c.filtered[fi]
	marker := c.choiceMarker(fi, item)
	line := marker + item.Name
	if item.Desc != "" {
		line += "  " + item.Desc
	}
	return measure.Truncate(line, width, "…")
}

func (c *Choice) choiceMarker(fi int, item Item) string {
	marker := "  "
	if fi == c.sel {
		marker = "▶ "
	}
	if c.MultiSelect {
		if c.checked[item.Name] {
			marker += "[✓] "
		} else {
			marker += "[ ] "
		}
	}
	return marker
}

func (c *Choice) highlightMatch(cv *Canvas, x, y int, text string, style Style, width int) {
	_, query := fuzzyScore(text, strings.ToLower(c.query))
	if !query || width <= 0 {
		return
	}
	runes := []rune(strings.ToLower(text))
	queryRunes := []rune(strings.ToLower(c.query))
	qi := 0
	for i, r := range runes {
		if r != queryRunes[qi] {
			continue
		}
		cellX := x + measure.StringWidth(string(runes[:i]))
		for dx := 0; dx < measure.StringWidth(string(r)); dx++ {
			if dx+cellX-x >= width {
				break
			}
			cell := cv.Get(cellX+dx, y)
			cell.Style = style
			cv.Set(cellX+dx, y, cell)
		}
		qi++
		if qi == len(queryRunes) {
			return
		}
	}
}

// choiceMouseHitRegion returns the inclusive content span in terminal cells.
func choiceMouseHitRegion(line string, width int) (start, end int, ok bool) {
	if width <= 0 {
		return 0, 0, false
	}
	pos := 0
	for _, cluster := range measure.Clusters(line) {
		w := measure.StringWidth(cluster)
		if pos >= width {
			break
		}
		if strings.TrimSpace(cluster) != "" {
			if !ok {
				start = pos
			}
			end = min(width, pos+w)
			ok = true
		}
		pos += w
	}
	return start, end, ok
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
	if !f && c.drag.active {
		c.viewOffset = c.drag.start
		c.drag.cancel()
	}
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

// ApplyTheme updates the Choice style from the theme.
func (c *Choice) ApplyTheme(theme ThemeColors) {
	c.Style = theme.ChoiceStyle()
}
