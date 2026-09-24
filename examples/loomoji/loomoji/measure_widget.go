// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loomoji

import (
	"fmt"
	"strings"

	"codeberg.org/ubunatic/loom"
)

// MeasureWidget guides a user through recording terminal-rendered glyph widths.
type MeasureWidget struct {
	store          *MeasurementStore
	glyphs         []string
	index          int
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
	glyphs := uniqueGlyphs()
	return &MeasureWidget{
		store:    store,
		glyphs:   store.Unmeasured(glyphs),
		jsonPath: jsonPath,
		textPath: textPath,
	}
}

// Err reports the most recent incremental save failure, if any.
func (w *MeasureWidget) Err() error { return w.saveErr }

// Remaining returns the number of glyphs still awaiting a width answer.
func (w *MeasureWidget) Remaining() int { return max(0, len(w.glyphs)-w.index) }

// CurrentGlyph returns the glyph currently shown, or an empty string when complete.
func (w *MeasureWidget) CurrentGlyph() string {
	if w.index < 0 || w.index >= len(w.glyphs) {
		return ""
	}
	return w.glyphs[w.index]
}

// Draw renders the current glyph between cursor markers and shows width references.
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
	glyph := w.CurrentGlyph()
	if glyph == "" {
		w.center(c, r, r.H/2, "All glyphs are recorded. Press Esc to finish.", loom.Style{FG: fg})
		w.center(c, r, r.H-1, "←: review this session  Esc: finish", loom.Style{FG: muted, Dim: true})
		return
	}

	total := len(uniqueGlyphs())
	completed := 0
	for _, measurement := range w.store.Entries {
		if measurement.Answered || measurement.MeasuredWidth != 0 {
			completed++
		}
	}
	position := 1
	for i, allGlyph := range uniqueGlyphs() {
		if allGlyph == glyph {
			position = i + 1
			break
		}
	}
	w.center(c, r, 3, fmt.Sprintf("Glyph %d / %d   Recorded %d / %d", position, total, completed, total), loom.Style{FG: muted})
	w.center(c, r, r.H/2-1, fmt.Sprintf("%d. %s", position, glyph), loom.Style{FG: fg})
	if m, ok := w.store.Entries[glyph]; ok {
		w.center(c, r, r.H/2, fmt.Sprintf("Current answer: %s   Loom computes: %d", widthLabel(m), m.ComputedWidth), loom.Style{FG: muted})
	} else {
		w.center(c, r, r.H/2, fmt.Sprintf("Loom computes: %d", NewMeasurement(glyph).ComputedWidth), loom.Style{FG: muted})
	}
	w.center(c, r, r.H/2+2, "Observed: |"+glyph+"|", loom.Style{FG: loom.ColorRGB(255, 210, 120), Bold: true})
	w.center(c, r, r.H/2+3, "Reference: 1 column |x|     2 columns |xx|", loom.Style{FG: muted})
	comment := "Comment: press c to add or edit"
	if w.editingComment {
		comment = "Comment: " + string(w.comment) + "▏"
	} else if m, ok := w.store.Entries[glyph]; ok && m.Comment != "" {
		comment = "Comment: " + m.Comment
	}
	w.center(c, r, r.H-3, comment, loom.Style{FG: fg})
	footer := "1 / 2: record width   ?: other/unsure   c: comment   ←/→: move   Esc: finish"
	if w.saveErr != nil {
		footer = "Save error: " + w.saveErr.Error()
	}
	w.center(c, r, r.H-1, footer, loom.Style{FG: muted, Dim: w.saveErr == nil})
}

// HandleKey processes width answers, comment editing, and movement between glyphs.
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
	if w.CurrentGlyph() == "" {
		if key == "left" || key == "backspace" {
			w.index = max(0, len(w.glyphs)-1)
		}
		return false
	}
	switch key {
	case "1":
		w.answer(1)
	case "2":
		w.answer(2)
	case "?":
		w.answer(0)
	case "c":
		w.beginComment()
	case "left", "backspace":
		w.index = max(0, w.index-1)
	case "right", "down":
		w.index = min(len(w.glyphs), w.index+1)
	}
	return false
}

// HandleMouse implements the loom.Widget interface; the measure flow is keyboard-only.
func (*MeasureWidget) HandleMouse(loom.MouseEvent) bool { return false }

func (w *MeasureWidget) answer(width int) {
	glyph := w.CurrentGlyph()
	if glyph == "" {
		return
	}
	measurement := NewMeasurement(glyph)
	if previous, ok := w.store.Entries[glyph]; ok {
		measurement.Comment = previous.Comment
	}
	measurement.MeasuredWidth = width
	measurement.Answered = true
	w.store.Set(measurement)
	w.persist()
	w.index = min(len(w.glyphs), w.index+1)
}

func (w *MeasureWidget) beginComment() {
	glyph := w.CurrentGlyph()
	w.comment = nil
	if previous, ok := w.store.Entries[glyph]; ok {
		w.comment = []rune(previous.Comment)
	}
	w.editingComment = true
}

func (w *MeasureWidget) handleCommentKey(key, text string) bool {
	switch key {
	case "esc":
		w.editingComment = false
	case "enter":
		glyph := w.CurrentGlyph()
		measurement := NewMeasurement(glyph)
		if previous, ok := w.store.Entries[glyph]; ok {
			measurement = previous
		}
		measurement.Comment = string(w.comment)
		w.store.Set(measurement)
		w.persist()
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

func widthLabel(m Measurement) string {
	if !m.Answered && m.MeasuredWidth == 0 {
		return "unanswered"
	}
	if m.MeasuredWidth == 0 {
		return "other/unsure"
	}
	return fmt.Sprintf("%d columns", m.MeasuredWidth)
}
