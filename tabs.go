// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

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

// TabsKeys configures keyboard navigation for a Tabs widget. Previous, Next,
// and Cycle each name one key. Select maps keys to tab indexes by position.
// Bindings match either KeyEvent.Key or KeyEvent.Text.
type TabsKeys struct {
	Previous string
	Next     string
	Cycle    string
	Select   []string
}

// DefaultTabsKeys returns the historical left/right arrow navigation.
func DefaultTabsKeys() TabsKeys {
	return TabsKeys{Previous: "left", Next: "right"}
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

// Tabs hosts child widgets under a row of tabs, switching which child receives
// Draw/HandleKey/HandleMouse. Tabs may be added, inserted, removed, or replaced
// at runtime using its lifecycle methods.
type Tabs struct {
	Tabs  []Tab
	Style TabsStyle
	Keys  TabsKeys
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
	return &Tabs{Tabs: tabs, Style: DefaultTabsStyle(), Keys: DefaultTabsKeys()}
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
	t.Select(i)
}

// Add appends tab and returns its index.
func (t *Tabs) Add(tab Tab) int {
	index := len(t.Tabs)
	t.Tabs = append(t.Tabs, tab)
	t.invalidateLayout()
	return index
}

// Insert adds tab at index. Index may equal the current tab count to append.
// The currently selected tab remains selected.
func (t *Tabs) Insert(index int, tab Tab) error {
	if index < 0 || index > len(t.Tabs) {
		return fmt.Errorf("tabs: insert index %d out of range [0,%d]", index, len(t.Tabs))
	}
	t.Tabs = append(t.Tabs, Tab{})
	copy(t.Tabs[index+1:], t.Tabs[index:])
	t.Tabs[index] = tab
	if len(t.Tabs) > 1 && index <= t.focus {
		t.focus++
	}
	t.invalidateLayout()
	return nil
}

// Remove deletes the tab at index. Removing a tab before the selection keeps
// the same tab selected; removing the selected tab chooses its successor, or
// the preceding tab when the removed tab was last.
func (t *Tabs) Remove(index int) error {
	if index < 0 || index >= len(t.Tabs) {
		return fmt.Errorf("tabs: remove index %d out of range [0,%d)", index, len(t.Tabs))
	}
	old := t.active()
	copy(t.Tabs[index:], t.Tabs[index+1:])
	t.Tabs[len(t.Tabs)-1] = Tab{}
	t.Tabs = t.Tabs[:len(t.Tabs)-1]
	if len(t.Tabs) == 0 {
		t.focus = 0
	} else if index < t.focus {
		t.focus--
	} else if t.focus >= len(t.Tabs) {
		t.focus = len(t.Tabs) - 1
	}
	t.updateChildFocus(old, t.active())
	t.invalidateLayout()
	return nil
}

// SetTabs replaces all tabs and clamps the selected index to the new bounds.
func (t *Tabs) SetTabs(tabs ...Tab) {
	old := t.active()
	t.Tabs = tabs
	if len(t.Tabs) == 0 {
		t.focus = 0
	} else if t.focus >= len(t.Tabs) {
		t.focus = len(t.Tabs) - 1
	} else if t.focus < 0 {
		t.focus = 0
	}
	t.updateChildFocus(old, t.active())
	t.invalidateLayout()
}

// Select activates index and reports whether it was valid.
func (t *Tabs) Select(index int) bool {
	if index < 0 || index >= len(t.Tabs) {
		return false
	}
	old := t.active()
	t.focus = index
	t.updateChildFocus(old, t.active())
	return true
}

// SetKeys replaces the declarative navigation key configuration.
func (t *Tabs) SetKeys(keys TabsKeys) { t.Keys = keys }

func (t *Tabs) invalidateLayout() {
	t.drawn = false
	t.tabCols = nil
}

func (t *Tabs) updateChildFocus(old, current Widget) {
	if child, ok := old.(Focusable); ok {
		child.SetFocus(false)
	}
	if child, ok := current.(Focusable); ok {
		child.SetFocus(true)
	}
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

// HandleKey applies configured navigation (and SwitchKey, if set), delegating
// everything else to the active child.
func (t *Tabs) HandleKey(e KeyEvent) (quit bool) {
	n := len(t.Tabs)
	if n == 0 {
		return false
	}
	keys := t.Keys
	if keys.Previous == "" && keys.Next == "" && keys.Cycle == "" && len(keys.Select) == 0 {
		keys = DefaultTabsKeys()
	}
	switch {
	case matchesTabKey(e, keys.Previous):
		t.Select((t.focus - 1 + n) % n)
		return false
	case matchesTabKey(e, keys.Next):
		t.Select((t.focus + 1) % n)
		return false
	case matchesTabKey(e, keys.Cycle), matchesTabKey(e, t.SwitchKey):
		t.Select((t.focus + 1) % n)
		return false
	}
	for index, binding := range keys.Select {
		if matchesTabKey(e, binding) && t.Select(index) {
			return false
		}
	}
	child := t.active()
	if child == nil {
		return false
	}
	return child.HandleKey(e)
}

func matchesTabKey(event KeyEvent, binding string) bool {
	return binding != "" && (event.Key == binding || event.Text == binding)
}

// HandleMouse switches tabs on a left click within the tab bar; any other
// event is delegated to the active child. Coordinates are canvas-absolute
// and 0-based.
func (t *Tabs) HandleMouse(e MouseEvent) (quit bool) {
	if len(t.Tabs) == 0 {
		return false
	}
	if t.drawn && e.Action == MousePress && e.Button == MouseLeft {
		x, y := e.X, e.Y
		for i, cr := range t.tabCols {
			if cr.Contains(x, y) {
				t.Select(i)
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
