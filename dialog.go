// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "strings"

// Dialog is a centered or positioned modal prompt rendered with Popup's border
// and background treatment. Enter selects the highlighted button and closes;
// without buttons it does nothing. Escape closes without calling OnSelect.
// Closing consumes the event without quitting or completing the hosting pane.
type Dialog struct {
	Title    string
	Body     string
	Buttons  []string
	Rect     Rect // explicit top-left and optional size; zero W/H are content-sized
	Width    int  // used when Rect is unset; zero sizes from content
	Height   int
	Style    Style
	Open     bool
	OnSelect func(string)
	selected int
	popup    *Popup
	drawRect Rect
}

// ApplyTheme updates dialog chrome and its modal popup.
func (d *Dialog) ApplyTheme(theme ThemeColors) {
	d.Style = theme.BoxStyle().Background
	if d.popup != nil {
		d.popup.ApplyTheme(theme)
	}
}

// NewDialog creates an open modal dialog. The first button is highlighted.
func NewDialog(title, body string, buttons ...string) *Dialog {
	return &Dialog{Title: title, Body: body, Buttons: buttons, Open: true}
}

// Activate reopens the dialog when its hosting tab is selected.
func (d *Dialog) Activate() { d.Open = true }

// SelectedButton reports the highlighted button label, or an empty string.
func (d *Dialog) SelectedButton() string {
	if len(d.Buttons) == 0 {
		return ""
	}
	return d.Buttons[d.selected%len(d.Buttons)]
}

func (d *Dialog) Draw(c *Canvas, r Rect) {
	d.drawRect = r
	if !d.Open {
		return
	}
	lines := strings.Split(d.Body, "\n")
	content := &dialogContent{dialog: d, lines: lines}
	w, h := d.Width, d.Height
	if d.Rect.W > 0 {
		w = d.Rect.W
	}
	if d.Rect.H > 0 {
		h = d.Rect.H
	}
	if w <= 0 {
		w = maxDialogWidth(d.Title, lines, d.Buttons) + 4
	}
	if h <= 0 {
		h = len(lines) + 2
		if len(d.Buttons) > 0 {
			h++
		}
	}
	if w < 4 {
		w = 4
	}
	if h < 3 {
		h = 3
	}
	area := r
	if d.Rect.X != 0 || d.Rect.Y != 0 || d.Rect.W > 0 || d.Rect.H > 0 {
		area = Rect{X: d.Rect.X, Y: d.Rect.Y, W: w, H: h}
	}
	d.popup = NewPopup(d.Title, content)
	d.popup.Open, d.popup.Width, d.popup.Height, d.popup.Style = true, w, h, d.Style
	d.popup.Draw(c, area)
}

func maxDialogWidth(title string, lines, buttons []string) int {
	w := StringWidth(title)
	for _, line := range lines {
		if n := StringWidth(line); n > w {
			w = n
		}
	}
	buttonsWidth := 0
	for _, b := range buttons {
		buttonsWidth += StringWidth(b) + 4
	}
	if buttonsWidth > w {
		w = buttonsWidth
	}
	return w
}

// ConsumeKey consumes closing actions and changes to the button selection.
// Unused keys and navigation with fewer than two buttons bubble to the parent.
// Tab is left to the container, which uses it to move focus between widgets.
func (d *Dialog) ConsumeKey(e KeyEvent) EventResult {
	if !d.Open {
		return Ignored()
	}
	if e.Key == "esc" {
		d.Open = false
		return Handled()
	}
	if len(d.Buttons) == 0 {
		return Ignored()
	}
	previous := d.selected
	switch e.Key {
	case "left":
		d.selected = (d.selected + len(d.Buttons) - 1) % len(d.Buttons)
	case "right":
		d.selected = (d.selected + 1) % len(d.Buttons)
	case "enter":
		if d.OnSelect != nil {
			d.OnSelect(d.SelectedButton())
		}
		d.Open = false
		return Handled()
	}
	if d.selected != previous {
		return Handled()
	}
	return Ignored()
}

func (d *Dialog) ConsumeMouse(e MouseEvent) EventResult {
	if d.popup == nil || !d.Open {
		return Ignored()
	}
	// Popup receives coordinates local to the area it was drawn in. A placed
	// dialog uses a different area from the allocation receiving this event.
	e.X += d.drawRect.X - d.popup.lastRect.X
	e.Y += d.drawRect.Y - d.popup.lastRect.Y
	return d.popup.ConsumeMouse(e)
}

type dialogContent struct {
	dialog  *Dialog
	lines   []string
	buttonY int
}

func (w *dialogContent) Draw(c *Canvas, r Rect) {
	w.buttonY = r.H - 1
	for i, line := range w.lines {
		if i >= r.H {
			break
		}
		c.Write(r.X, r.Y+i, line, Style{})
	}
	if len(w.dialog.Buttons) == 0 || r.H == 0 {
		return
	}
	y := r.Y + r.H - 1
	x := r.X
	for i, label := range w.dialog.Buttons {
		text, style := "  "+label+"  ", Style{Dim: true}
		if i == w.dialog.selected {
			text, style = "▶ "+label+"  ", Style{Bold: true}
		}
		x += c.Write(x, y, text, style)
	}
}
func (*dialogContent) ConsumeKey(KeyEvent) EventResult { return Ignored() }
func (w *dialogContent) ConsumeMouse(e MouseEvent) EventResult {
	if w == nil || w.dialog == nil || e.Action != MousePress || e.Button != MouseLeft || len(w.dialog.Buttons) == 0 || e.Y != w.buttonY {
		return Ignored()
	}
	x := 0
	for i, label := range w.dialog.Buttons {
		text := "  " + label + "  "
		if i == w.dialog.selected {
			text = "▶ " + label + "  "
		}
		width := StringWidth(text)
		if e.X >= x && e.X < x+width {
			w.dialog.selected = i
			if w.dialog.OnSelect != nil {
				w.dialog.OnSelect(label)
			}
			w.dialog.Open = false
			return Handled()
		}
		x += width
	}
	return Ignored()
}
func (w *dialogContent) ContentHeight() int { return len(w.lines) + 1 }
