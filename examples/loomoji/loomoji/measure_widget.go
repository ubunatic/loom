// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomoji

import (
	"fmt"
	"strings"

	"codeberg.org/ubunatic/loom"
)

const measurePageSize = 10

// MeasureWidget guides a user through recording terminal-rendered glyph widths.
type MeasureWidget struct {
	store          *MeasurementStore
	glyphs         []string
	page           int
	selected       int
	pending        map[string]Measurement
	jsonPath       string
	textPath       string
	editingComment bool
	comment        []rune
	saveErr        error
}

// NewMeasureWidget creates a keyboard-driven measure widget for unrecorded loomoji glyphs.
func NewMeasureWidget(store *MeasurementStore, jsonPath, textPath string) *MeasureWidget {
	if store == nil {
		store = NewMeasurementStore(TerminalProfile{})
	}
	return &MeasureWidget{
		store:    store,
		glyphs:   store.Unmeasured(uniqueGlyphs()),
		pending:  make(map[string]Measurement),
		jsonPath: jsonPath,
		textPath: textPath,
	}
}

// Err reports the most recent page save failure, if any.
func (w *MeasureWidget) Err() error { return w.saveErr }

// Remaining returns the number of glyphs on the current and following pages.
func (w *MeasureWidget) Remaining() int { return max(0, len(w.glyphs)-w.page*measurePageSize) }

// CurrentGlyph returns the selected glyph on the current page, or an empty string when complete.
func (w *MeasureWidget) CurrentGlyph() string {
	page := w.PageGlyphs()
	if w.selected < 0 || w.selected >= len(page) {
		return ""
	}
	return page[w.selected]
}

// PageGlyphs returns up to ten unmeasured glyphs in the current page.
func (w *MeasureWidget) PageGlyphs() []string {
	start := w.page * measurePageSize
	if start < 0 || start >= len(w.glyphs) {
		return nil
	}
	end := min(len(w.glyphs), start+measurePageSize)
	return w.glyphs[start:end]
}

// PageMeasurement returns a row's staged measurement, seeded with Loom's computed width.
func (w *MeasureWidget) PageMeasurement(row int) Measurement {
	page := w.PageGlyphs()
	if row < 0 || row >= len(page) {
		return Measurement{}
	}
	return w.measurement(page[row])
}

// Draw renders the current page as a numbered glyph measurement table.
func (w *MeasureWidget) Draw(c *loom.Canvas, r loom.Rect) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	bg := loom.ColorRGB(31, 33, 36)
	fg := loom.ColorRGB(225, 228, 230)
	muted := loom.ColorRGB(155, 160, 165)
	accent := loom.ColorRGB(32, 151, 185)
	c.PaintSurface(r, loom.Style{BG: bg})
	w.center(c, r, 1, "Terminal emoji width measurement", loom.Style{FG: accent, Bold: true})
	page := w.PageGlyphs()
	if len(page) == 0 {
		w.center(c, r, r.H/2, "All glyphs are recorded. Press Esc to finish.", loom.Style{FG: fg})
		return
	}
	w.center(c, r, 2, fmt.Sprintf("Page %d   %d glyphs remaining", w.page+1, w.Remaining()), loom.Style{FG: muted})
	w.write(c, r, 3, "Row  Glyph       Codepoints                 Computed  Answer  Comment", loom.Style{FG: muted, Bold: true})
	for row, glyph := range page {
		m := w.measurement(glyph)
		answer := fmt.Sprint(m.MeasuredWidth)
		if m.MeasuredWidth == 0 {
			answer = "?"
		}
		line := fmt.Sprintf("%d    |%s|  %-25s %8d  %6s  %s", row, glyph, strings.Join(m.Codepoints, " "), m.ComputedWidth, answer, m.Comment)
		style := loom.Style{FG: fg}
		if row == w.selected {
			style = loom.Style{FG: fg, BG: loom.ColorRGB(56, 62, 68), Bold: true}
		}
		w.write(c, r, 4+row, line, style)
	}
	comment := "c: edit selected comment"
	if w.editingComment {
		comment = "Comment: " + string(w.comment) + "▏"
	}
	w.center(c, r, r.H-3, comment, loom.Style{FG: fg})
	footer := "0-9: toggle row  ↑/↓: select  ?: unsure  Enter/PgDn: confirm page  PgUp: back  q: quit"
	if w.saveErr != nil {
		footer = "Save error: " + w.saveErr.Error()
	}
	w.center(c, r, r.H-1, footer, loom.Style{FG: muted, Dim: w.saveErr == nil})
}

// HandleKey processes page answers, comments, confirmation, and navigation.
func (w *MeasureWidget) HandleKey(e loom.KeyEvent) bool {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	if w.editingComment {
		return w.handleCommentKey(key, e.Text)
	}
	if key == "esc" || key == "ctrl-c" || key == "q" {
		return true
	}
	page := w.PageGlyphs()
	if len(page) == 0 {
		if key == "pgup" || key == "pageup" {
			w.previousPage()
		}
		return false
	}
	switch key {
	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		w.toggleRow(int(key[0] - '0'))
	case "?":
		w.setSelectedWidth(0)
	case "c":
		w.beginComment()
	case "up":
		w.selected = max(0, w.selected-1)
	case "down":
		w.selected = min(len(page)-1, w.selected+1)
	case "enter", "pgdown", "pgdn", "pagedown":
		w.confirmPage()
	case "pgup", "pageup":
		w.previousPage()
	}
	return false
}

// HandleMouse implements the loom.Widget interface; the measure flow is keyboard-only.
func (*MeasureWidget) HandleMouse(loom.MouseEvent) bool { return false }

func (w *MeasureWidget) measurement(glyph string) Measurement {
	if m, ok := w.pending[glyph]; ok {
		return m
	}
	if m, ok := w.store.Entries[glyph]; ok {
		m.Answered = false
		return m
	}
	m := NewMeasurement(glyph)
	m.MeasuredWidth = m.ComputedWidth
	return m
}

func (w *MeasureWidget) toggleRow(row int) {
	page := w.PageGlyphs()
	if row < 0 || row >= len(page) {
		return
	}
	m := w.measurement(page[row])
	if m.MeasuredWidth == 1 {
		m.MeasuredWidth = 2
	} else {
		m.MeasuredWidth = 1
	}
	m.Answered = false
	w.pending[m.Glyph] = m
}

func (w *MeasureWidget) setSelectedWidth(width int) {
	glyph := w.CurrentGlyph()
	if glyph == "" {
		return
	}
	m := w.measurement(glyph)
	m.MeasuredWidth = width
	m.Answered = false
	w.pending[glyph] = m
}

func (w *MeasureWidget) beginComment() {
	glyph := w.CurrentGlyph()
	if glyph == "" {
		return
	}
	w.comment = []rune(w.measurement(glyph).Comment)
	w.editingComment = true
}

func (w *MeasureWidget) handleCommentKey(key, text string) bool {
	switch key {
	case "esc":
		w.editingComment = false
	case "enter":
		glyph := w.CurrentGlyph()
		m := w.measurement(glyph)
		m.Comment = string(w.comment)
		m.Answered = false
		w.pending[glyph] = m
		w.editingComment = false
	case "backspace":
		if len(w.comment) > 0 {
			w.comment = w.comment[:len(w.comment)-1]
		}
	default:
		w.comment = append(w.comment, []rune(text)...)
	}
	return false
}

func (w *MeasureWidget) confirmPage() {
	page := w.PageGlyphs()
	if len(page) == 0 {
		return
	}
	for _, glyph := range page {
		m := w.measurement(glyph)
		m.Answered = true
		w.store.Set(m)
		delete(w.pending, glyph)
	}
	w.persist()
	w.page++
	w.selected = 0
}

func (w *MeasureWidget) previousPage() {
	if w.page > 0 {
		w.page--
		w.selected = 0
	}
}

func (w *MeasureWidget) persist() {
	w.saveErr = nil
	if w.jsonPath != "" {
		w.saveErr = w.store.Save(w.jsonPath)
	}
	if w.saveErr == nil && w.textPath != "" {
		w.saveErr = w.store.SaveTextReport(w.textPath)
	}
}

func (w *MeasureWidget) center(c *loom.Canvas, r loom.Rect, row int, text string, style loom.Style) {
	y := r.Y + min(max(0, row), r.H-1)
	x := r.X + max(0, (r.W-loom.StringWidth(text))/2)
	c.Write(x, y, strings.TrimSpace(text), style)
}

func (w *MeasureWidget) write(c *loom.Canvas, r loom.Rect, row int, text string, style loom.Style) {
	y := r.Y + min(max(0, row), r.H-1)
	c.Write(r.X, y, truncate(text, r.W), style)
}

func truncate(text string, width int) string {
	if loom.StringWidth(text) <= width {
		return text
	}
	var b strings.Builder
	for _, cluster := range strings.Split(text, "") {
		if loom.StringWidth(b.String()+cluster+"…") > width {
			break
		}
		b.WriteString(cluster)
	}
	return b.String() + "…"
}

func uniqueGlyphs() []string {
	seen := make(map[string]bool, len(entryList))
	glyphs := make([]string, 0, len(entryList))
	for _, e := range entryList {
		if !seen[e.icon] {
			seen[e.icon] = true
			glyphs = append(glyphs, e.icon)
		}
	}
	return glyphs
}
