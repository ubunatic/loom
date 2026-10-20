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

// Measure reports the palette width and its three display rows.
func (p *ColorPicker) Measure(width int) measure.Size {
	defs := SpeccedDefaults.ColorPicker
	w := max(measure.StringWidth(defs.Labels.FG)+measure.StringWidth(defs.Labels.BG)+3, len(p.Palette)*defs.SwatchWidth)
	if width > 0 && width < w {
		w = width
	}
	return measure.Size{Width: w, Height: 3}
}

// Draw renders the FG/BG target selector, palette swatches and combined preview.
func (p *ColorPicker) Draw(c *Canvas, r Rect) {
	if c == nil || p == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	p.lastRect = r
	c.PaintSurface(r, p.Style)
	defs := SpeccedDefaults.ColorPicker
	if r.H >= 1 {
		leftStyle, rightStyle := p.Style, p.Style
		if p.Target == ColorPickerFG {
			leftStyle = p.FocusStyle
		} else {
			rightStyle = p.FocusStyle
		}
		leftWidth := measure.StringWidth(defs.Labels.FG)
		separatorX := leftWidth + 1
		rightX := leftWidth + 3
		writeColorPickerText(c, r, 0, 0, defs.Labels.FG, leftStyle)
		writeColorPickerText(c, r, leftWidth, 0, " ", p.Style)
		writeColorPickerText(c, r, separatorX, 0, defs.Glyphs.Separator, p.Style)
		writeColorPickerText(c, r, separatorX+1, 0, " ", p.Style)
		writeColorPickerText(c, r, rightX, 0, defs.Labels.BG, rightStyle)
	}
	if r.H >= 2 {
		for i, color := range p.Palette {
			x := i * defs.SwatchWidth
			if x >= r.W {
				break
			}
			glyph := defs.Glyphs.Swatch
			style := p.Style
			style.BG = color
			if i == p.Selected {
				glyph = defs.Glyphs.Focus
				if p.focused {
					style.Bold = true
				}
			}
			writeColorPickerText(c, r, x, 1, glyph, style)
			for pad := 1; pad < defs.SwatchWidth && x+pad < r.W; pad++ {
				c.Set(r.X+x+pad, r.Y+1, Cell{Text: " ", Style: style})
			}
		}
	}
	if r.H >= 3 {
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

// ConsumeKey handles palette navigation, target changes and completion.
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
	case p.matches(e, "toggle_target"):
		p.Target = ColorPickerFG + ColorPickerBG - p.Target
		return Handled()
	case p.matches(e, "target_fg"):
		p.changeTarget(ColorPickerFG, true)
		return Handled()
	case p.matches(e, "target_bg"):
		p.changeTarget(ColorPickerBG, true)
		return Handled()
	case p.matches(e, "select_bg"):
		p.changeTarget(ColorPickerBG, true)
		return Handled()
	case p.matches(e, "select"):
		p.changeTarget(p.Target, true)
		return Handled()
	case p.matches(e, "left"):
		if p.Selected > 0 {
			p.Selected--
		} else {
			p.Target = ColorPickerFG + ColorPickerBG - p.Target
		}
		return Handled()
	case p.matches(e, "right"):
		if p.Selected+1 < len(p.Palette) {
			p.Selected++
		} else {
			p.Target = ColorPickerFG + ColorPickerBG - p.Target
		}
		return Handled()
	default:
		return Ignored()
	}
}

// ConsumeMouse selects a swatch or changes target through the top-line controls.
func (p *ColorPicker) ConsumeMouse(e MouseEvent) EventResult {
	if p == nil || e.Action != MousePress || e.X < 0 || e.X >= p.lastRect.W || e.Y < 0 || e.Y >= p.lastRect.H {
		return Ignored()
	}
	defs := SpeccedDefaults.ColorPicker
	if e.Y == 0 && e.Button == MouseLeft {
		leftWidth := measure.StringWidth(defs.Labels.FG)
		if e.X >= 0 && e.X < leftWidth {
			p.Target = ColorPickerFG
			return Handled()
		}
		rightX := leftWidth + 3
		rightWidth := measure.StringWidth(defs.Labels.BG)
		if e.X >= rightX && e.X < rightX+rightWidth {
			p.Target = ColorPickerBG
			return Handled()
		}
		return Ignored()
	}
	if e.Y != 1 || e.X < 0 || e.X >= p.lastRect.W || e.Y >= p.lastRect.H {
		return Ignored()
	}
	if e.Button != MouseLeft && e.Button != MouseRight {
		return Ignored()
	}
	index := e.X / defs.SwatchWidth
	if index < 0 || index >= len(p.Palette) {
		return Ignored()
	}
	p.Selected = index
	target := p.Target
	if e.Shift || e.Button == MouseRight {
		target = ColorPickerBG
	}
	p.changeTarget(target, true)
	return Handled()
}
