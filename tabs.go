// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "gopkg.in/yaml.v3"

// Tab pairs a title with the widget shown while that tab is active.
type Tab struct {
	Title  string
	Widget Widget
}

// TabsStyle controls the visual appearance of a Tabs widget's bar.
// It reuses existing theme roles (header_* for the active tab, normal_*
// for inactive tabs, border_* for the bar rule) rather than introducing
// new spec/themes.yaml keys.
type TabsStyle struct {
	Active   Style // active tab title (header_*)
	Inactive Style // inactive tab titles (normal_*)
	Rule     Style // bar separator rule (border_*)
}

// DefaultTabsStyle returns a minimal monochrome style, mirroring
// DefaultChoiceStyle/DefaultTableStyle which are pinned to the "plain"
// theme by theme_test.go rather than hardcoding ad hoc colors here.
func DefaultTabsStyle() TabsStyle {
	return Theme("plain").TabsStyle()
}

// TabsStyle returns a TabsStyle derived from the theme's color roles.
func (t ThemeColors) TabsStyle() TabsStyle {
	return TabsStyle{
		Active:   Style{FG: t.HeaderFG.Color(), BG: t.HeaderBG.Color(), Bold: t.HeaderBold},
		Inactive: Style{FG: t.NormalFG.Color(), BG: t.NormalBG.Color()},
		Rule:     Style{FG: t.BorderFG.Color(), BG: t.BorderBG.Color()},
	}
}

// tabsRuleGlyph is the tab-bar separator glyph, sourced from spec/box.yaml's
// BoxBorder.Horizontal (loaded once from frame.go's embedded frameSpecs)
// instead of a hardcoded box-drawing character.
var tabsRuleGlyph = func() string {
	data, err := frameSpecs.ReadFile("spec/box.yaml")
	if err != nil {
		return "-"
	}
	var border BoxBorder
	if err := yaml.Unmarshal(data, &border); err != nil || border.Horizontal == "" {
		return "-"
	}
	return border.Horizontal
}()

// Tabs hosts a fixed set of child widgets under a row of tabs, switching
// which child receives Draw/HandleKey/HandleMouse. The tab set is fixed at
// construction time (NewTabs) — dynamic add/remove is out of scope, matching
// Stack's fixed Children model.
type Tabs struct {
	Tabs  []Tab
	Style TabsStyle
	// SwitchKey optionally cycles to the next tab in addition to left/right
	// arrow keys (e.g. "tab" or "ctrl-t"). Empty disables it.
	SwitchKey string

	focus    int    // index of the active tab
	drawn    bool   // whether Draw has run at least once (for HandleMouse hit-testing)
	lastRect Rect   // the Rect passed to the most recent Draw call
	tabCols  []Rect // per-tab clickable rect on the bar, refreshed each Draw
}

// NewTabs creates a Tabs widget hosting the given tabs, analogous to NewStack.
func NewTabs(tabs ...Tab) *Tabs {
	return &Tabs{Tabs: tabs, Style: DefaultTabsStyle()}
}

// Focus returns the index of the currently active tab.
func (t *Tabs) Focus() int { return t.focus }

// SetFocusIndex activates the tab at index i, clamping to the valid range.
func (t *Tabs) SetFocusIndex(i int) {
	if len(t.Tabs) == 0 {
		t.focus = 0
		return
	}
	if i < 0 {
		i = 0
	}
	if i >= len(t.Tabs) {
		i = len(t.Tabs) - 1
	}
	t.focus = i
}

// active returns the widget of the currently active tab, or nil if there are none.
func (t *Tabs) active() Widget {
	if t.focus < 0 || t.focus >= len(t.Tabs) {
		return nil
	}
	return t.Tabs[t.focus].Widget
}

// barHeight returns the number of rows the tab bar occupies: one row for
// titles plus one rule row, or zero when there are no tabs to show.
func (t *Tabs) barHeight() int {
	if len(t.Tabs) == 0 {
		return 0
	}
	return 2
}

// Draw renders the tab bar into the top of r (active tab highlighted per
// t.Style.Active, others per t.Style.Inactive, with a rule row beneath
// sourced from spec/box.yaml) and the active child into the remaining space.
func (t *Tabs) Draw(c *Canvas, r Rect) {
	t.drawn = true
	t.lastRect = r
	n := len(t.Tabs)
	if n == 0 {
		return
	}
	bar := t.barHeight()
	x := r.X
	t.tabCols = make([]Rect, n)
	for i, tab := range t.Tabs {
		title := " " + tab.Title + " "
		style := t.Style.Inactive
		if i == t.focus {
			style = t.Style.Active
		}
		w := c.Write(x, r.Y, title, style)
		t.tabCols[i] = Rect{X: x, Y: r.Y, W: w, H: 1}
		x += w
	}
	if x < r.X+r.W {
		c.PaintSurface(Rect{X: x, Y: r.Y, W: r.X + r.W - x, H: 1}, t.Style.Inactive)
	}
	if bar > 1 {
		c.Fill(Rect{X: r.X, Y: r.Y + 1, W: r.W, H: bar - 1}, Cell{Text: tabsRuleGlyph, Style: t.Style.Rule})
	}
	child := t.active()
	if child == nil {
		return
	}
	if f, ok := child.(Focusable); ok {
		f.SetFocus(true)
	}
	cr := Rect{X: r.X, Y: r.Y + bar, W: r.W, H: max(0, r.H-bar)}
	child.Draw(c, cr)
	if Debug {
		drawDebugBorder(c, cr)
	}
}

// HandleKey switches tabs on left/right arrows (and SwitchKey, if set),
// delegating everything else to the active child.
func (t *Tabs) HandleKey(e KeyEvent) (quit bool) {
	n := len(t.Tabs)
	if n == 0 {
		return false
	}
	switch {
	case e.Key == "left":
		t.focus = (t.focus - 1 + n) % n
		return false
	case e.Key == "right":
		t.focus = (t.focus + 1) % n
		return false
	case t.SwitchKey != "" && e.Key == t.SwitchKey:
		t.focus = (t.focus + 1) % n
		return false
	}
	child := t.active()
	if child == nil {
		return false
	}
	return child.HandleKey(e)
}

// HandleMouse switches tabs on a left click within the tab bar; any other
// event is delegated to the active child. e.X/e.Y are 1-based
// terminal/pane-relative coordinates (see Pane.run), so they are converted
// to the 0-based canvas coordinates t.lastRect and t.tabCols use.
func (t *Tabs) HandleMouse(e MouseEvent) (quit bool) {
	if len(t.Tabs) == 0 {
		return false
	}
	if t.drawn && e.Action == MousePress && e.Button == MouseLeft {
		x, y := e.X-1, e.Y-1
		for i, cr := range t.tabCols {
			if cr.Contains(x, y) {
				t.focus = i
				return false
			}
		}
	}
	child := t.active()
	if child == nil {
		return false
	}
	return child.HandleMouse(e)
}

// ContentHeight estimates the required height: the tab bar plus the tallest
// child, so switching tabs does not shift the reserved layout height.
func (t *Tabs) ContentHeight() int {
	if len(t.Tabs) == 0 {
		return 0
	}
	maxH := 0
	for _, tab := range t.Tabs {
		ch := 1
		if chWidget, ok := tab.Widget.(ContentHeighter); ok {
			ch = chWidget.ContentHeight()
		}
		if ch > maxH {
			maxH = ch
		}
	}
	return t.barHeight() + maxH
}

// HeightForWidth estimates preferred height at an allocated width: the tab
// bar plus the tallest child's preferred height at that width.
func (t *Tabs) HeightForWidth(width int) int {
	if len(t.Tabs) == 0 {
		return 0
	}
	maxH := 0
	for _, tab := range t.Tabs {
		h := 1
		if wh, ok := tab.Widget.(WidthHeighter); ok {
			h = wh.HeightForWidth(width)
		} else if chWidget, ok := tab.Widget.(ContentHeighter); ok {
			h = chWidget.ContentHeight()
		}
		if h > maxH {
			maxH = h
		}
	}
	return t.barHeight() + maxH
}

// ContentWidth estimates the widest child width, at least as wide as the
// rendered tab bar.
func (t *Tabs) ContentWidth() int {
	if len(t.Tabs) == 0 {
		return 0
	}
	barW := 0
	maxChildW := 0
	for _, tab := range t.Tabs {
		barW += len(tab.Title) + 2
		w := 1
		if cw, ok := tab.Widget.(ContentWidther); ok {
			w = cw.ContentWidth()
		}
		if w > maxChildW {
			maxChildW = w
		}
	}
	return max(barW, maxChildW)
}
