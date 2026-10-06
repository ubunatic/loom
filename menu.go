// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"unicode"
)

// MenuItem is one command in a menu. Submenu reserves the model for nested menus.
type MenuItem struct {
	Label    string
	Mnemonic rune
	Shortcut string
	Disabled bool
	Checked  *bool
	Action   func()
	Submenu  []MenuItem
}

// Menu is a titled set of commands shown in a MenuBar dropdown.
type Menu struct {
	Title     string
	Mnemonic  rune
	Items     []MenuItem
	Open      bool
	Selected  int
	lastRect  Rect
	itemRects []Rect
}

type menuSublevel struct {
	items      []MenuItem
	selected   int
	parentItem int
	rect       Rect
	itemRects  []Rect
}

// MenuBar draws and navigates a horizontal row of menus and one-level dropdowns.
type MenuBar struct {
	Menus      []Menu
	ActiveMenu int
	Open       bool
	// Bottom places the bar on the final row of the draw bounds and opens its
	// dropdown upward. It is useful for editor status bars.
	Bottom bool
	Style  MenuStyle

	focused    bool
	lastRect   Rect
	barRect    Rect
	menuRect   Rect
	titleRects []Rect
	itemRects  []Rect
	selected   int
	keys       *KeyMap
	keyActions map[string]func()
	submenus   []menuSublevel
}

// MenuStyle controls the bar and dropdown appearance.
type MenuStyle struct {
	Bar, Active, Normal, Selected, Disabled, Border Style
}

// DefaultMenuStyle uses the plain theme's normal and border colors.
func DefaultMenuStyle() MenuStyle {
	t := Theme("plain")
	normal := Style{FG: t.NormalFG.Color(), BG: t.NormalBG.Color()}
	active := Style{FG: t.SelectedFG.Color(), BG: t.SelectedBG.Color(), Bold: t.SelectedBold}
	border := Style{FG: t.BorderFG.Color(), BG: t.BorderBG.Color()}
	return MenuStyle{Bar: normal, Active: active, Normal: normal, Selected: active, Disabled: Style{Dim: true}, Border: border}
}

// NewMenuBar creates a menu bar with the supplied top-level menus.
func NewMenuBar(menus ...Menu) *MenuBar {
	m := &MenuBar{Menus: append([]Menu(nil), menus...), Style: DefaultMenuStyle()}
	m.rebuildKeys()
	return m
}

func (m *MenuBar) Focused() bool { return m != nil && m.focused }

// SetFocus controls whether the menu bar has keyboard focus.
func (m *MenuBar) SetFocus(focused bool) {
	if m == nil {
		return
	}
	m.focused = focused
	if !focused {
		m.Open = false
		m.submenus = nil
	}
}

func (m *MenuBar) rebuildKeys() {
	bindings := map[string][]string{}
	m.keyActions = map[string]func(){}
	for mi := range m.Menus {
		for ii := range m.Menus[mi].Items {
			item := &m.Menus[mi].Items[ii]
			if item.Shortcut == "" || item.Action == nil || item.Disabled {
				continue
			}
			action := strings.Join([]string{"menu", itoa(mi), itoa(ii)}, ":")
			key := menuKeyName(item.Shortcut)
			if key == "" {
				continue
			}
			bindings[action] = append(bindings[action], key)
			m.keyActions[action] = item.Action
		}
	}
	m.keys = NewKeyMap(bindings)
}

func menuKeyName(shortcut string) string {
	if strings.HasPrefix(shortcut, "^") {
		shortcut = "Ctrl+" + strings.TrimPrefix(shortcut, "^")
	}
	parts := strings.Split(strings.ToLower(strings.ReplaceAll(shortcut, " ", "")), "+")
	for i := range parts {
		switch parts[i] {
		case "control":
			parts[i] = "ctrl"
		case "escape":
			parts[i] = "esc"
		case "return":
			parts[i] = "enter"
		case "space":
			parts[i] = " "
		}
	}
	key := strings.Join(parts, "-")
	if len(parts) == 0 || key == "" {
		return ""
	}
	return key
}

// Draw renders menu titles and, when open, the active dropdown.
func (m *MenuBar) Draw(c *Canvas, r Rect) {
	if m == nil || c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	m.lastRect = r
	barY := r.Y
	if m.Bottom {
		barY += r.H - 1
	}
	m.barRect = Rect{X: r.X, Y: barY, W: r.W, H: 1}
	c.Fill(m.barRect, Cell{Text: " ", Style: m.Style.Bar})
	m.titleRects = make([]Rect, len(m.Menus))
	x := r.X
	for i, menu := range m.Menus {
		label := " " + menu.Title + " "
		w := StringWidth(label)
		if x+w > r.X+r.W {
			break
		}
		style := m.Style.Bar
		if m.Open && i == m.ActiveMenu {
			style = m.Style.Active
		}
		writeMenuTitle(c, x, barY, menu.Title, menu.Mnemonic, style)
		m.titleRects[i] = Rect{X: x, Y: barY, W: w, H: 1}
		x += w
	}
	if !m.Open || m.ActiveMenu < 0 || m.ActiveMenu >= len(m.Menus) {
		m.menuRect = Rect{}
		m.submenus = nil
		return
	}
	menu := m.Menus[m.ActiveMenu]
	w := 4
	for _, item := range menu.Items {
		labelWidth := StringWidth(menuItemLabel(item))
		if item.Shortcut != "" {
			labelWidth += StringWidth(item.Shortcut) + 2
		}
		if len(item.Submenu) > 0 {
			labelWidth += 2
		}
		if labelWidth+2 > w {
			w = labelWidth + 2
		}
	}
	if w > r.W {
		w = r.W
	}
	h := len(menu.Items) + 2
	if m.Bottom {
		if h > m.barRect.Y-r.Y {
			h = m.barRect.Y - r.Y
		}
	} else if h > r.H-(m.barRect.Y-r.Y+1) {
		h = r.H - (m.barRect.Y - r.Y + 1)
	}
	if h < 1 {
		m.menuRect = Rect{}
		return
	}
	x = r.X
	if m.ActiveMenu < len(m.titleRects) && m.titleRects[m.ActiveMenu].W > 0 {
		x = m.titleRects[m.ActiveMenu].X
	}
	if x+w > r.X+r.W {
		x = r.X + r.W - w
	}
	y := r.Y + 1
	if m.Bottom {
		y = m.barRect.Y - h
	}
	m.menuRect = Rect{X: x, Y: y, W: w, H: h}
	c.Fill(m.menuRect, Cell{Text: " ", Style: m.Style.Normal})
	if w >= 2 && h >= 2 {
		c.DrawBox(m.menuRect, BoxBorderStyleSharp, "", m.Style.Border)
	}
	m.itemRects = make([]Rect, len(menu.Items))
	for i, item := range menu.Items {
		row := y + 1 + i
		if row >= y+h-1 {
			break
		}
		style := m.Style.Normal
		if item.Disabled {
			style = m.Style.Disabled
		}
		if i == m.selected && !item.Disabled {
			style = m.Style.Selected
		}
		c.Fill(Rect{X: x + 1, Y: row, W: max(0, w-2), H: 1}, Cell{Text: " ", Style: style})
		if item.Label == "---" {
			for sx := x + 1; sx < x+w-1; sx++ {
				c.Write(sx, row, "─", style)
			}
		} else {
			lineWidth := w - 2
			if len(item.Submenu) > 0 {
				lineWidth -= 2
			}
			line := menuItemLineWidth(item, lineWidth)
			if len(item.Submenu) > 0 {
				line = TruncateText(line, lineWidth, "")
			}
			c.Write(x+1, row, line, style)
			if len(item.Submenu) > 0 && w >= 3 {
				c.Write(x+w-2, row, "›", style)
			}
		}
		m.itemRects[i] = Rect{X: x + 1, Y: row, W: max(0, w-2), H: 1}
	}
	m.drawSubmenus(c, r)
}

func (m *MenuBar) drawSubmenus(c *Canvas, bounds Rect) {
	items := m.Menus[m.ActiveMenu].Items
	anchor := m.menuRect
	for i := range m.submenus {
		level := &m.submenus[i]
		if level.parentItem < 0 || level.parentItem >= len(items) {
			m.submenus = m.submenus[:i]
			return
		}
		parentRect := anchor
		var parentItems []Rect
		if i == 0 {
			parentItems = m.itemRects
		} else {
			parentItems = m.submenus[i-1].itemRects
		}
		if level.parentItem >= len(parentItems) {
			m.submenus = m.submenus[:i]
			return
		}
		row := parentItems[level.parentItem].Y
		w := 4
		for _, item := range level.items {
			w = max(w, StringWidth(menuItemLine(item))+2)
			if len(item.Submenu) > 0 {
				w = max(w, StringWidth(menuItemLine(item))+4)
			}
		}
		w = min(w, bounds.W)
		x := parentRect.X + parentRect.W
		if x+w > bounds.X+bounds.W {
			x = parentRect.X - w
		}
		x = max(bounds.X, min(x, bounds.X+bounds.W-w))
		h := min(len(level.items)+2, bounds.H)
		y := max(bounds.Y, min(row, bounds.Y+bounds.H-h))
		level.rect = Rect{X: x, Y: y, W: w, H: h}
		c.Fill(level.rect, Cell{Text: " ", Style: m.Style.Normal})
		if w >= 2 && h >= 2 {
			c.DrawBox(level.rect, BoxBorderStyleSharp, "", m.Style.Border)
		}
		level.itemRects = make([]Rect, len(level.items))
		for j, item := range level.items {
			yi := y + 1 + j
			if yi >= y+h-1 {
				break
			}
			style := m.Style.Normal
			if item.Disabled {
				style = m.Style.Disabled
			} else if j == level.selected {
				style = m.Style.Selected
			}
			c.Fill(Rect{X: x + 1, Y: yi, W: max(0, w-2), H: 1}, Cell{Text: " ", Style: style})
			if item.Label == "---" {
				for sx := x + 1; sx < x+w-1; sx++ {
					c.Write(sx, yi, "─", style)
				}
			} else {
				lineWidth := w - 2
				if len(item.Submenu) > 0 {
					lineWidth -= 2
				}
				line := menuItemLineWidth(item, lineWidth)
				if len(item.Submenu) > 0 {
					line = TruncateText(line, lineWidth, "")
				}
				c.Write(x+1, yi, line, style)
				if len(item.Submenu) > 0 && w >= 3 {
					c.Write(x+w-2, yi, "›", style)
				}
			}
			level.itemRects[j] = Rect{X: x + 1, Y: yi, W: max(0, w-2), H: 1}
		}
		anchor = level.rect
		items = level.items
	}
}

func writeMenuTitle(c *Canvas, x, y int, title string, mnemonic rune, style Style) {
	if mnemonic == 0 {
		c.Write(x, y, " "+title+" ", style)
		return
	}
	runes := []rune(title)
	for i, r := range runes {
		if unicode.ToLower(r) != unicode.ToLower(mnemonic) {
			continue
		}
		before := " " + string(runes[:i])
		after := string(runes[i+1:]) + " "
		px := x + StringWidth(before)
		c.Write(x, y, before, style)
		under := style
		under.Underline = true
		c.Write(px, y, string(r), under)
		c.Write(px+StringWidth(string(r)), y, after, style)
		return
	}
	c.Write(x, y, " "+title+" ", style)
}

func menuItemLine(item MenuItem) string {
	return menuItemLineWidth(item, 0)
}

func menuItemLabel(item MenuItem) string {
	mark := "  "
	if item.Checked != nil && *item.Checked {
		mark = "✓ "
	}
	return mark + item.Label
}
func menuItemLineWidth(item MenuItem, width int) string {
	label := menuItemLabel(item)
	if item.Shortcut == "" || width <= 0 {
		if item.Shortcut != "" {
			return label + "  " + item.Shortcut
		}
		return label
	}
	spaces := max(1, width-StringWidth(label)-StringWidth(item.Shortcut))
	return label + strings.Repeat(" ", spaces) + item.Shortcut
}

// ConsumeKey processes menu navigation and command accelerators.
// ConsumeKey reports whether the menu bar consumed e.
func (m *MenuBar) ConsumeKey(e KeyEvent) EventResult {
	if m == nil {
		return Ignored()
	}
	if m.keys == nil {
		m.rebuildKeys()
	}
	if action := m.keys.Action(e); action != "" {
		m.keyActions[action]()
		return Consumed()
	}
	if len(m.Menus) == 0 {
		return Ignored()
	}
	if e.Key == "f10" {
		m.focused, m.Open = true, true
		m.ActiveMenu = clampMenu(m.ActiveMenu, len(m.Menus))
		m.selectFirst()
		return Consumed()
	}
	if strings.HasPrefix(e.Key, "alt-") {
		target := []rune(strings.TrimPrefix(e.Key, "alt-"))
		if len(target) == 1 {
			for i, menu := range m.Menus {
				if unicode.ToLower(menu.Mnemonic) == unicode.ToLower(target[0]) {
					if m.Open && m.ActiveMenu == i {
						m.Open = false
						m.submenus = nil
						return Consumed()
					}
					m.focused, m.Open, m.ActiveMenu = true, true, i
					m.submenus = nil
					m.selectFirst()
					return Consumed()
				}
			}
		}
	}
	if !m.focused && !m.Open {
		return Ignored()
	}
	switch e.Key {
	case "esc":
		if len(m.submenus) > 0 {
			m.submenus = m.submenus[:len(m.submenus)-1]
		} else if m.Open {
			m.Open = false
		} else {
			m.focused = false
		}
	case "left":
		if len(m.submenus) > 0 {
			m.submenus = m.submenus[:len(m.submenus)-1]
		} else if m.Open {
			m.ActiveMenu = (clampMenu(m.ActiveMenu, len(m.Menus)) - 1 + len(m.Menus)) % len(m.Menus)
			m.selectFirst()
		}
	case "right":
		if m.Open && m.openSelectedSubmenu() {
			return Consumed()
		}
		if len(m.submenus) > 0 {
			return Consumed()
		}
		if m.Open {
			m.ActiveMenu = (clampMenu(m.ActiveMenu, len(m.Menus)) + 1) % len(m.Menus)
			m.submenus = nil
			m.selectFirst()
		}
	case "down":
		m.focused, m.Open = true, true
		m.moveCurrent(1)
	case "up":
		if m.Open {
			m.moveCurrent(-1)
		}
	case "enter", " ":
		if !m.Open {
			m.focused, m.Open = true, true
			m.selectFirst()
			return Consumed()
		}
		if !m.openSelectedSubmenu() {
			m.activateCurrent()
		}
	default:
		return Ignored()
	}
	return Consumed()
}

func (m *MenuBar) currentItems() []MenuItem {
	if len(m.submenus) > 0 {
		return m.submenus[len(m.submenus)-1].items
	}
	if m.ActiveMenu >= 0 && m.ActiveMenu < len(m.Menus) {
		return m.Menus[m.ActiveMenu].Items
	}
	return nil
}
func (m *MenuBar) currentSelection() int {
	if len(m.submenus) > 0 {
		return m.submenus[len(m.submenus)-1].selected
	}
	return m.selected
}
func (m *MenuBar) setCurrentSelection(i int) {
	if len(m.submenus) > 0 {
		m.submenus[len(m.submenus)-1].selected = i
	} else {
		m.selected = i
	}
}
func (m *MenuBar) moveCurrent(delta int) {
	items := m.currentItems()
	if len(items) == 0 {
		return
	}
	i := m.currentSelection()
	for tries := 0; tries < len(items); tries++ {
		i = (i + delta + len(items)) % len(items)
		if !items[i].Disabled {
			m.setCurrentSelection(i)
			return
		}
		delta = 1
	}
}
func (m *MenuBar) openSelectedSubmenu() bool {
	items := m.currentItems()
	i := m.currentSelection()
	if i < 0 || i >= len(items) || items[i].Disabled || len(items[i].Submenu) == 0 {
		return false
	}
	selected := 0
	for selected < len(items[i].Submenu)-1 && items[i].Submenu[selected].Disabled {
		selected++
	}
	m.submenus = append(m.submenus, menuSublevel{items: items[i].Submenu, parentItem: i, selected: selected})
	return true
}
func (m *MenuBar) activateCurrent() {
	items := m.currentItems()
	i := m.currentSelection()
	if i < 0 || i >= len(items) {
		return
	}
	item := items[i]
	if item.Disabled || len(item.Submenu) > 0 {
		return
	}
	if item.Checked != nil {
		*item.Checked = !*item.Checked
	}
	if item.Action != nil {
		item.Action()
	}
	m.Open = false
	m.focused = false
	m.submenus = nil
}

func clampMenu(i, n int) int {
	if i < 0 {
		return 0
	}
	if i >= n {
		return n - 1
	}
	return i
}
func (m *MenuBar) selectFirst() { m.selected = 0; m.moveSelection(0) }
func (m *MenuBar) moveSelection(delta int) {
	if m.ActiveMenu < 0 || m.ActiveMenu >= len(m.Menus) {
		return
	}
	items := m.Menus[m.ActiveMenu].Items
	if len(items) == 0 {
		m.selected = 0
		return
	}
	for tries := 0; tries < len(items); tries++ {
		m.selected = (m.selected + delta + len(items)) % len(items)
		if !items[m.selected].Disabled {
			return
		}
		delta = 1
	}
}
func (m *MenuBar) activateSelected() {
	if m.ActiveMenu < 0 || m.ActiveMenu >= len(m.Menus) {
		return
	}
	items := m.Menus[m.ActiveMenu].Items
	if m.selected < 0 || m.selected >= len(items) {
		return
	}
	item := items[m.selected]
	if item.Disabled {
		return
	}
	if item.Checked != nil {
		*item.Checked = !*item.Checked
	}
	if item.Action != nil {
		item.Action()
	}
	m.Open = false
	m.focused = false
	m.submenus = nil
}

// ConsumeMouse handles title clicks, dropdown selection, hover switching, and outside dismissal.
// ConsumeMouse reports whether the menu bar consumed e.
func (m *MenuBar) ConsumeMouse(e MouseEvent) EventResult {
	if m == nil {
		return Ignored()
	}
	// Hit rectangles are recorded in the Draw canvas coordinate space, while
	// events are local to the widget boundary.
	e.X += m.lastRect.X
	e.Y += m.lastRect.Y
	for i := len(m.submenus) - 1; i >= 0; i-- {
		level := &m.submenus[i]
		if !level.rect.Contains(e.X, e.Y) {
			continue
		}
		for j, rect := range level.itemRects {
			if !rect.Contains(e.X, e.Y) {
				continue
			}
			level.selected = j
			if e.Action == MouseHover {
				m.submenus = m.submenus[:i+1]
				if len(level.items[j].Submenu) > 0 {
					m.openSelectedSubmenu()
				}
				return Consumed()
			}
			if e.Action == MousePress && e.Button == MouseLeft {
				if len(level.items[j].Submenu) > 0 {
					m.submenus = m.submenus[:i+1]
					m.openSelectedSubmenu()
				} else {
					m.activateCurrent()
				}
				return Consumed()
			}
		}
		return Consumed()
	}
	for i, rect := range m.titleRects {
		if rect.W > 0 && rect.Contains(e.X, e.Y) {
			if e.Action == MouseHover && m.Open {
				m.ActiveMenu = i
				m.submenus = nil
				m.selectFirst()
				return Consumed()
			}
			if e.Action == MousePress && e.Button == MouseLeft {
				if m.Open && m.ActiveMenu == i {
					m.Open = false
					m.focused = false
					m.submenus = nil
				} else {
					m.ActiveMenu = i
					m.Open = true
					m.focused = true
					m.submenus = nil
					m.selectFirst()
				}
				return Consumed()
			}
		}
	}
	if m.Open && m.menuRect.Contains(e.X, e.Y) {
		for i, rect := range m.itemRects {
			if rect.H > 0 && rect.Contains(e.X, e.Y) {
				if e.Action == MouseHover {
					m.selected = i
					m.submenus = nil
					if len(m.Menus[m.ActiveMenu].Items[i].Submenu) > 0 {
						m.openSelectedSubmenu()
					}
					return Consumed()
				}
				if e.Action == MousePress && e.Button == MouseLeft {
					m.selected = i
					if len(m.Menus[m.ActiveMenu].Items[i].Submenu) > 0 {
						m.submenus = nil
						m.openSelectedSubmenu()
					} else {
						m.activateSelected()
					}
					return Consumed()
				}
			}
		}
		return Consumed()
	}
	if m.Open && e.Action == MousePress && e.Button == MouseLeft {
		m.Open = false
		m.focused = false
		m.submenus = nil
		return Consumed()
	}
	return Ignored()
}

// Menu is also a Widget so callers can host its dropdown independently.
func (m *Menu) Draw(c *Canvas, r Rect) {
	if m == nil || c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	m.Open = true
	w := 4
	for _, item := range m.Items {
		if n := StringWidth(menuItemLine(item)); n > w {
			w = n
		}
	}
	w = min(w, r.W)
	h := min(len(m.Items)+2, r.H)
	r.W, r.H = w, h
	m.lastRect = r
	c.Fill(r, Cell{Text: " ", Style: DefaultMenuStyle().Normal})
	if r.W >= 2 && r.H >= 2 {
		c.DrawBox(r, BoxBorderStyleSharp, "", DefaultMenuStyle().Border)
	}
	m.itemRects = make([]Rect, len(m.Items))
	for i, item := range m.Items {
		y := r.Y + 1 + i
		if y >= r.Y+r.H-1 {
			break
		}
		style := DefaultMenuStyle().Normal
		if item.Disabled {
			style = DefaultMenuStyle().Disabled
		}
		if i == m.Selected && !item.Disabled {
			style = DefaultMenuStyle().Selected
		}
		c.Fill(Rect{X: r.X + 1, Y: y, W: max(0, r.W-2), H: 1}, Cell{Text: " ", Style: style})
		if item.Label == "---" {
			for sx := r.X + 1; sx < r.X+r.W-1; sx++ {
				c.Write(sx, y, "─", style)
			}
		} else {
			c.Write(r.X+1, y, menuItemLineWidth(item, r.W-2), style)
		}
		m.itemRects[i] = Rect{X: r.X + 1, Y: y, W: max(0, r.W-2), H: 1}
	}
}
func (m *Menu) ConsumeKey(e KeyEvent) EventResult {
	if m == nil {
		return Ignored()
	}
	if !m.Open {
		return Ignored()
	}
	switch e.Key {
	case "esc":
		m.Open = false
	case "up":
		m.move(-1)
	case "down":
		m.move(1)
	case "enter", " ":
		m.activate()
	default:
		return Ignored()
	}
	return Consumed()
}
func (m *Menu) ConsumeMouse(e MouseEvent) EventResult {
	if m == nil || !m.Open {
		return Ignored()
	}
	if !m.lastRect.Contains(e.X, e.Y) {
		if e.Action == MousePress && e.Button == MouseLeft {
			m.Open = false
			return Consumed()
		}
		return Ignored()
	}
	for i, r := range m.itemRects {
		if r.Contains(e.X, e.Y) {
			if e.Action == MouseHover {
				m.Selected = i
				return Consumed()
			}
			if e.Action == MousePress && e.Button == MouseLeft {
				m.Selected = i
				m.activate()
				return Consumed()
			}
		}
	}
	return Consumed()
}

func (m *Menu) move(delta int) {
	if len(m.Items) == 0 {
		return
	}
	for tries := 0; tries < len(m.Items); tries++ {
		m.Selected = (m.Selected + delta + len(m.Items)) % len(m.Items)
		if !m.Items[m.Selected].Disabled {
			return
		}
		delta = 1
	}
}
func (m *Menu) activate() {
	if m.Selected < 0 || m.Selected >= len(m.Items) {
		return
	}
	item := m.Items[m.Selected]
	if item.Disabled {
		return
	}
	if item.Checked != nil {
		*item.Checked = !*item.Checked
	}
	if item.Action != nil {
		item.Action()
	}
	m.Open = false
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	p := len(b)
	for i > 0 {
		p--
		b[p] = byte('0' + i%10)
		i /= 10
	}
	return string(b[p:])
}
