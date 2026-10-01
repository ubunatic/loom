// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"time"

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

var tabsVerticalGlyph = func() string {
	data, err := frameSpecs.ReadFile("spec/box.yaml")
	if err != nil {
		return "|"
	}
	var border BoxBorder
	if err := yaml.Unmarshal(data, &border); err != nil || border.Vertical == "" {
		return "|"
	}
	return border.Vertical
}()

// Tabs hosts child widgets under a row of tabs, switching which child receives
// Draw/ConsumeKey/ConsumeMouse. Tabs may be added, inserted, removed, or replaced
// at runtime using its lifecycle methods.
type Tabs struct {
	Tabs  []Tab
	Style TabsStyle
	Keys  TabsKeys
	// Vertical places tab titles in a left column. The zero value preserves
	// the historical horizontal tab bar.
	Vertical bool
	// SwitchKey optionally cycles to the next tab in addition to left/right
	// arrow keys (e.g. "tab" or "ctrl-t"). Empty disables it.
	SwitchKey string
	// ArrowSwitch is retained for source compatibility. Arrow bindings now
	// follow the same child-first event bubbling as other tab bindings.
	ArrowSwitch bool
	OnChildQuit func(i int) (quitHost bool)

	focus            int    // index of the active tab
	drawn            bool   // whether Draw has run at least once (for ConsumeMouse hit-testing)
	lastRect         Rect   // the Rect passed to the most recent Draw call
	tabCols          []Rect // per-tab clickable rect on the bar, refreshed each Draw
	mouseCapture     Widget
	mouseCaptureRect Rect // child origin in Tabs-local coordinates at press time
}

// NewTabs creates a Tabs widget hosting the given tabs, analogous to NewStack.
func NewTabs(tabs ...Tab) *Tabs {
	return &Tabs{Tabs: tabs, Style: DefaultTabsStyle(), Keys: DefaultTabsKeys(), ArrowSwitch: true}
}

// Focus returns the index of the currently active tab.
func (t *Tabs) Focus() int { return t.focus }

// TickInterval returns the active child's requested cadence.
func (t *Tabs) TickInterval() time.Duration {
	child, ok := t.active().(Ticker)
	if !ok {
		return 0
	}
	return child.TickInterval()
}

// Tick forwards the update to the active child.
func (t *Tabs) Tick(now time.Time) {
	if child, ok := t.active().(Ticker); ok {
		child.Tick(now)
	}
}

// PaneRequest merges the terminal requirements of all tabs.
func (t *Tabs) PaneRequest() (request PaneRequest) {
	first := true
	for _, tab := range t.Tabs {
		if child, ok := tab.Widget.(PaneRequester); ok {
			childRequest := child.PaneRequest()
			if first {
				request, first = childRequest, false
				continue
			}
			mergePaneRequest(&request, childRequest)
		}
	}
	return request
}

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
	if child, ok := UnwrapWidget(t.active()).(WidgetActivator); ok {
		child.Activate()
	}
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
	if t.Vertical {
		return len(t.Tabs)
	}
	return 2
}

func (t *Tabs) barWidth() int {
	w := 0
	for _, tab := range t.Tabs {
		if width := StringWidth(tab.Title) + 2; width > w {
			w = width
		}
	}
	return w
}

// Draw renders the tab bar in the top or left edge of r and the active child
// in the remaining space. The active tab is highlighted with t.Style.Active.
func (t *Tabs) Draw(c *Canvas, r Rect) {
	t.drawn = true
	t.lastRect = r
	n := len(t.Tabs)
	if n == 0 {
		return
	}
	bar := t.barHeight()
	if t.Vertical {
		barWidth := t.barWidth()
		t.tabCols = make([]Rect, n)
		start := max(0, t.focus-max(0, r.H)+1)
		end := min(n, start+r.H)
		for row, i := 0, start; i < end; row, i = row+1, i+1 {
			tab := t.Tabs[i]
			if f, ok := tab.Widget.(Focusable); ok {
				f.SetFocus(i == t.focus)
			}
			style := t.Style.Inactive
			if i == t.focus {
				style = t.Style.Active
			}
			title := " " + tab.Title + " "
			t.tabCols[i] = Rect{X: r.X, Y: r.Y + row, W: barWidth, H: 1}
			c.Write(r.X, r.Y+row, title, style)
			if width := StringWidth(title); width < barWidth {
				c.PaintSurface(Rect{X: r.X + width, Y: r.Y + row, W: barWidth - width, H: 1}, style)
			}
		}
		if r.H > 0 {
			c.Fill(Rect{X: r.X + barWidth, Y: r.Y, W: 1, H: r.H}, Cell{Text: tabsVerticalGlyph, Style: t.Style.Rule})
		}
		child := t.active()
		if child != nil {
			if f, ok := child.(Focusable); ok {
				f.SetFocus(true)
			}
			cr := Rect{X: r.X + barWidth + 1, Y: r.Y, W: max(0, r.W-barWidth-1), H: r.H}
			child.Draw(c, cr)
			if Debug {
				drawDebugBorder(c, cr)
			}
		}
		return
	}
	x := r.X
	t.tabCols = make([]Rect, n)
	for i, tab := range t.Tabs {
		if f, ok := tab.Widget.(Focusable); ok {
			f.SetFocus(i == t.focus)
		}
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

// ConsumeKey applies configured navigation (and SwitchKey, if set), delegating
// everything else to the active child.
func (t *Tabs) ConsumeKey(e KeyEvent) EventResult {
	n := len(t.Tabs)
	if n == 0 {
		return Ignored()
	}
	keys := t.Keys
	if keys.Previous == "" && keys.Next == "" && keys.Cycle == "" && len(keys.Select) == 0 {
		keys = DefaultTabsKeys()
	}
	if child := t.active(); child != nil {
		if res := child.ConsumeKey(e); res.Consumed {
			if res.Quit && t.OnChildQuit != nil {
				if t.OnChildQuit(t.focus) {
					return QuitResult()
				}
				return Handled()
			}
			return res
		}
	}
	switch {
	case matchesTabKey(e, keys.Previous):
		t.Select((t.focus - 1 + n) % n)
		return Handled()
	case matchesTabKey(e, keys.Next):
		t.Select((t.focus + 1) % n)
		return Handled()
	case matchesTabKey(e, keys.Cycle), matchesTabKey(e, t.SwitchKey):
		t.Select((t.focus + 1) % n)
		return Handled()
	}
	for index, binding := range keys.Select {
		if matchesTabKey(e, binding) && t.Select(index) {
			return Handled()
		}
	}
	return Ignored()
}

func (t *Tabs) ConsumePaste(e PasteEvent) EventResult {
	return DispatchPasteEvent(t.active(), e)
}

func matchesTabKey(event KeyEvent, binding string) bool {
	return binding != "" && (event.Key == binding || event.Text == binding)
}

// ConsumeMouse selects a tab on a bar click and dispatches panel events to the
// active child in child-local coordinates.
func (t *Tabs) ConsumeMouse(e MouseEvent) EventResult {
	if child := t.mouseCapture; child != nil && (e.Action == MouseDrag || e.Action == MouseRelease) {
		e.X -= t.mouseCaptureRect.X
		e.Y -= t.mouseCaptureRect.Y
		if e.Action == MouseRelease {
			t.mouseCapture = nil
		}
		return DispatchMouseEvent(child, e)
	}
	if len(t.Tabs) == 0 {
		return Ignored()
	}
	if t.drawn && e.Action == MousePress && e.Button == MouseLeft {
		x, y := e.X+t.lastRect.X, e.Y+t.lastRect.Y
		for i, cr := range t.tabCols {
			if cr.Contains(x, y) {
				t.Select(i)
				return Handled()
			}
		}
	}
	child := t.active()
	if child == nil || !t.drawn {
		return Ignored()
	}
	panel := t.childRect(t.lastRect)
	x, y := e.X+t.lastRect.X, e.Y+t.lastRect.Y
	if !panel.Contains(x, y) {
		return Ignored()
	}
	e.X = x - panel.X
	e.Y = y - panel.Y
	res := DispatchMouseEvent(child, e)
	if res.Consumed && e.Action == MousePress && e.Button == MouseLeft {
		t.mouseCapture = child
		t.mouseCaptureRect = Rect{X: panel.X - t.lastRect.X, Y: panel.Y - t.lastRect.Y, W: panel.W, H: panel.H}
	}
	return res
}

// childRect returns the active panel rectangle from the last draw allocation.
func (t *Tabs) childRect(r Rect) Rect {
	if t.Vertical {
		return Rect{X: r.X + t.barWidth() + 1, Y: r.Y, W: max(0, r.W-t.barWidth()-1), H: r.H}
	}
	bar := t.barHeight()
	return Rect{X: r.X, Y: r.Y + bar, W: r.W, H: max(0, r.H-bar)}
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
	if t.Vertical {
		return max(t.barHeight(), maxH)
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
	if t.Vertical {
		return max(t.barHeight(), maxH)
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
		barW += StringWidth(tab.Title) + 2
		w := 1
		if cw, ok := tab.Widget.(ContentWidther); ok {
			w = cw.ContentWidth()
		}
		if w > maxChildW {
			maxChildW = w
		}
	}
	if t.Vertical {
		return t.barWidth() + 1 + maxChildW
	}
	return max(barW, maxChildW)
}

// ApplyTheme restyles the tab bar and forwards the theme to all children
// that implement Themeable.
func (t *Tabs) ApplyTheme(theme ThemeColors) {
	t.Style = theme.TabsStyle()
	for _, tab := range t.Tabs {
		if themeable, ok := tab.Widget.(Themeable); ok {
			themeable.ApplyTheme(theme)
		}
	}
}
