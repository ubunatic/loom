// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "strings"

// HintEntry pairs a display key and label with the action used by its binding.
type HintEntry struct {
	Key, Label string
	// Binding is the decoded KeyEvent name, such as ctrl-s or f7.
	Binding string
	// Detail is optional context (for example a theme name), dropped before pairs.
	Detail string
	// DropPriority drops higher values first; ties drop from the end.
	DropPriority int
	Action       func() EventResult
}

// HintBarStyle separates the bold key caps from the labels and row surface.
type HintBarStyle struct{ Cap, Label Style }

// HintBar renders one row of atomic key/label pairs and runs actions on clicks.
// Mouse coordinates are zero-based and local to its last draw rectangle.
type HintBar struct {
	Entries []HintEntry
	Style   HintBarStyle
	hits    []hintBarHit
}

type hintBarHit struct {
	index int
	rect  Rect
}

// NewHintBar creates a hint bar with the plain theme's styles.
func NewHintBar(entries ...HintEntry) *HintBar {
	return &HintBar{Entries: append([]HintEntry(nil), entries...), Style: Theme("plain").HintBarStyle()}
}

// ApplyTheme updates the cap and label colors from the theme.
func (b *HintBar) ApplyTheme(theme ThemeColors) { b.Style = theme.HintBarStyle() }

func (b *HintBar) fit(width int) ([]hintBarHit, bool) {
	visible := make([]bool, len(b.Entries))
	for i := range visible {
		visible[i] = true
	}
	details := true
	layout := func() ([]hintBarHit, int) {
		x := 1 // the design's left margin
		var hits []hintBarHit
		for i, entry := range b.Entries {
			if !visible[i] {
				continue
			}
			label := entry.Label
			if details && entry.Detail != "" {
				label += " " + entry.Detail
			}
			w := StringWidth(entry.Key) + 2 + 1 + StringWidth(label)
			hits = append(hits, hintBarHit{index: i, rect: Rect{X: x, W: w, H: 1}})
			x += w + 1
		}
		return hits, x - 1
	}
	for {
		hits, used := layout()
		if used <= width || len(hits) == 0 {
			return hits, details
		}
		if details {
			details = false
			continue
		}
		drop := -1
		for i, entry := range b.Entries {
			if visible[i] && (drop < 0 || entry.DropPriority >= b.Entries[drop].DropPriority) {
				drop = i
			}
		}
		visible[drop] = false
	}
}

// Text returns the fitted pairs without ANSI styles or cap padding.
func (b *HintBar) Text(width int) string {
	hits, details := b.fit(width)
	parts := make([]string, 0, len(hits))
	for _, hit := range hits {
		entry := b.Entries[hit.index]
		text := entry.Key + " " + entry.Label
		if details && entry.Detail != "" {
			text += " " + entry.Detail
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, "  ")
}

// Draw paints only complete caps and labels within the assigned first row.
func (b *HintBar) Draw(c *Canvas, r Rect) {
	b.hits = nil
	if c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	c.Fill(Rect{X: r.X, Y: r.Y, W: r.W, H: 1}, Cell{Text: " ", Style: b.Style.Label})
	hits, details := b.fit(r.W)
	b.hits = hits
	for _, hit := range hits {
		entry := b.Entries[hit.index]
		x := r.X + hit.rect.X
		c.Write(x, r.Y, " "+entry.Key+" ", b.Style.Cap)
		label := " " + entry.Label
		if details && entry.Detail != "" {
			label += " " + entry.Detail
		}
		c.Write(x+StringWidth(entry.Key)+2, r.Y, label, b.Style.Label)
	}
}

// ConsumeKey dispatches a binding to the same action used by mouse clicks.
func (b *HintBar) ConsumeKey(e KeyEvent) EventResult {
	for _, entry := range b.Entries {
		if entry.Binding != "" && e.Is(entry.Binding) && entry.Action != nil {
			return entry.Action()
		}
	}
	return Ignored()
}

// ConsumeMouse runs the cap or label's action on a left press; hover is inert.
func (b *HintBar) ConsumeMouse(e MouseEvent) EventResult {
	if e.Action != MousePress || e.Button != MouseLeft {
		return Ignored()
	}
	for _, hit := range b.hits {
		if hit.rect.Contains(e.X, e.Y) {
			if action := b.Entries[hit.index].Action; action != nil {
				return action()
			}
		}
	}
	return Ignored()
}
