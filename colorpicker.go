// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"

	"ubunatic.com/loom/measure"
)

// ColorPickerTarget names the channel a ColorPicker assigns colors to.
type ColorPickerTarget uint8

const (
	ColorPickerFG ColorPickerTarget = iota
	ColorPickerBG
)

// ColorPicker selects a foreground or background color and previews the pair.
type ColorPicker struct {
	FG, BG     Color
	Target     ColorPickerTarget
	Selected   int
	Palette    []Color
	Style      Style
	FocusStyle Style
	OnChange   func(ColorPickerTarget, Color)
	OnClose    func()

	focused  bool
	lastRect Rect
}

var _ Widget = (*ColorPicker)(nil)
var _ Focusable = (*ColorPicker)(nil)
var _ Measurer = (*ColorPicker)(nil)
var _ Themeable = (*ColorPicker)(nil)

// NewColorPicker creates a picker initialized with foreground and background colors.
func NewColorPicker(fg, bg Color) *ColorPicker {
	defs := SpeccedDefaults.ColorPicker
	themeStyle := DefaultTheme().TableStyle()
	palette := make([]Color, len(defs.Palette))
	for i, index := range defs.Palette {
		palette[i] = ColorIndex(uint8(index))
	}
	target := ColorPickerFG
	if defs.InitialTarget == "bg" {
		target = ColorPickerBG
	}
	return &ColorPicker{
		FG: fg, BG: bg, Target: target, Palette: palette,
		Style:      themeStyle.Normal,
		FocusStyle: themeStyle.Selected,
		focused:    true,
	}
}

// Focused reports whether this picker owns keyboard focus.
func (p *ColorPicker) Focused() bool { return p != nil && p.focused }

// SetFocus updates keyboard focus rendering.
func (p *ColorPicker) SetFocus(focused bool) {
	if p != nil {
		p.focused = focused
	}
}

// ApplyTheme updates the normal and focused styles from a host theme.
func (p *ColorPicker) ApplyTheme(theme ThemeColors) {
	if p == nil {
		return
	}
	style := theme.TableStyle()
	p.Style, p.FocusStyle = style.Normal, style.Selected
}

// PreviewStyle returns the selected foreground and background as one style.
func (p *ColorPicker) PreviewStyle() Style {
	if p == nil {
		return Style{}
	}
	return Style{FG: p.FG, BG: p.BG}
}

// Measure reports the one-row palette and its optional preview.
func (p *ColorPicker) Measure(width int) measure.Size {
	defs := SpeccedDefaults.ColorPicker
	w := len(p.Palette) * defs.SwatchWidth
	if width > 0 && width < w {
		w = width
	}
	h := 2 // swatches and the mode marker row
	if defs.ShowPreview {
		h++
	}
	return measure.Size{Width: w, Height: h}
}

// Draw renders contiguous color samples, a mode-specific focus marker and preview.
func (p *ColorPicker) Draw(c *Canvas, r Rect) {
	if c == nil || p == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	p.lastRect = r
	c.PaintSurface(r, p.Style)
	defs := SpeccedDefaults.ColorPicker
	for i, color := range p.Palette {
		x := i * defs.SwatchWidth
		if x >= r.W {
			break
		}
		style := p.Style
		if p.Target == ColorPickerBG {
			style.FG, style.BG = p.FG, color
		} else {
			style.FG, style.BG = color, p.BG
		}
		if i == p.Selected && p.Target == ColorPickerFG {
			style.Underline = true
		}
		writeColorPickerText(c, r, x, 0, defs.Glyphs.Swatch, style)
		for pad := measure.StringWidth(defs.Glyphs.Swatch); pad < defs.SwatchWidth && x+pad < r.W; pad++ {
			c.Set(r.X+x+pad, r.Y, Cell{Text: " ", Style: style})
		}
	}
	if p.Target == ColorPickerBG && r.H >= 2 && p.Selected >= 0 && p.Selected < len(p.Palette) {
		markerX := p.Selected * defs.SwatchWidth
		writeColorPickerText(c, r, markerX, 1, defs.Glyphs.ModeMarker, p.FocusStyle)
	}
	if defs.ShowPreview && r.H >= 3 {
		text := defs.Labels.Preview + " " + defs.PreviewText
		writeColorPickerText(c, r, 0, 2, text, p.PreviewStyle())
	}
}

func writeColorPickerText(c *Canvas, r Rect, x, y int, text string, style Style) {
	if y < 0 || y >= r.H || x < 0 || x >= r.W {
		return
	}
	var clipped strings.Builder
	width := 0
	for _, char := range text {
		charWidth := measure.RuneWidth(char)
		if width+charWidth > r.W-x {
			break
		}
		clipped.WriteRune(char)
		width += charWidth
	}
	c.Write(r.X+x, r.Y+y, clipped.String(), style)
}

func (p *ColorPicker) matches(e KeyEvent, action string) bool {
	name := strings.ToLower(e.Name())
	for _, binding := range SpeccedDefaults.ColorPicker.Bindings[action] {
		if strings.EqualFold(name, binding) {
			return true
		}
	}
	return false
}

func (p *ColorPicker) changeTarget(target ColorPickerTarget, assign bool) {
	p.Target = target
	if !assign || p.Selected < 0 || p.Selected >= len(p.Palette) {
		return
	}
	color := p.Palette[p.Selected]
	if target == ColorPickerFG {
		p.FG = color
	} else {
		p.BG = color
	}
	if p.OnChange != nil {
		p.OnChange(target, color)
	}
}

// ConsumeKey handles palette navigation, mode changes and completion.
func (p *ColorPicker) ConsumeKey(e KeyEvent) EventResult {
	if p == nil {
		return Ignored()
	}
	switch {
	case p.matches(e, "close"):
		if p.OnClose != nil {
			p.OnClose()
		}
		return DoneResult()
	case p.matches(e, "mode_bg"):
		p.Target = ColorPickerBG
		return Handled()
	case p.matches(e, "mode_fg"):
		p.Target = ColorPickerFG
		return Handled()
	case p.matches(e, "mode_toggle"):
		p.toggleTarget()
		return Handled()
	case p.matches(e, "select"):
		p.changeTarget(p.Target, true)
		return Handled()
	case p.matches(e, "left"):
		if p.Selected > 0 {
			p.Selected--
		}
		return Handled()
	case p.matches(e, "right"):
		if p.Selected+1 < len(p.Palette) {
			p.Selected++
		}
		return Handled()
	default:
		return Ignored()
	}
}

func (p *ColorPicker) toggleTarget() {
	if p.Target == ColorPickerFG {
		p.Target = ColorPickerBG
	} else {
		p.Target = ColorPickerFG
	}
}

// ConsumeMouse selects a swatch or toggles modes on right-click.
func (p *ColorPicker) ConsumeMouse(e MouseEvent) EventResult {
	if p == nil || e.Action != MousePress || e.X < 0 || e.X >= p.lastRect.W || e.Y < 0 || e.Y >= p.lastRect.H {
		return Ignored()
	}
	if e.Button == MouseRight {
		p.toggleTarget()
		return Handled()
	}
	if e.Button != MouseLeft || e.Y != 0 {
		return Ignored()
	}
	defs := SpeccedDefaults.ColorPicker
	index := e.X / defs.SwatchWidth
	if index < 0 || index >= len(p.Palette) {
		return Ignored()
	}
	p.Selected = index
	p.changeTarget(p.Target, true)
	return Handled()
}
