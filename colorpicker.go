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

// ColorPicker selects foreground and background colors and previews their pair.
type ColorPicker struct {
	FG, BG   Color
	Target   ColorPickerTarget
	Selected int
	// ResetSelected focuses the terminal-default/no-color cell. Selected remains
	// a valid index into Palette for callers that use it to inspect the palette.
	ResetSelected bool
	Palette       []Color
	Style         Style
	FocusStyle    Style
	OnChange      func(ColorPickerTarget, Color)
	OnClose       func()

	focused  bool
	lastRect Rect
	fgIndex  int
	bgIndex  int
	fgReset  bool
	bgReset  bool
}

var _ Widget = (*ColorPicker)(nil)
var _ Focusable = (*ColorPicker)(nil)
var _ Measurer = (*ColorPicker)(nil)
var _ Themeable = (*ColorPicker)(nil)

// NewColorPicker creates a picker initialized with foreground and background colors.
func NewColorPicker(fg, bg Color) *ColorPicker {
	defs := SpeccedDefaults.ColorPicker
	palette := make([]Color, len(defs.Palette))
	for i, index := range defs.Palette {
		palette[i] = ColorIndex(uint8(index))
	}
	themeStyle := DefaultTheme().TableStyle()
	p := &ColorPicker{FG: fg, BG: bg, Palette: palette, Style: themeStyle.Normal, FocusStyle: themeStyle.Selected, focused: true}
	p.fgIndex, p.fgReset = colorPickerIndex(palette, fg)
	p.bgIndex, p.bgReset = colorPickerIndex(palette, bg)
	p.Target = ColorPickerFG
	if defs.InitialTarget == "bg" {
		p.Target = ColorPickerBG
	}
	p.syncCursor()
	return p
}

func colorPickerIndex(palette []Color, color Color) (int, bool) {
	if color == ColorReset() {
		return 0, true
	}
	for i, candidate := range palette {
		if candidate == color {
			return i, false
		}
	}
	return 0, false
}

func (p *ColorPicker) syncCursor() {
	if p.Target == ColorPickerBG {
		p.Selected, p.ResetSelected = p.bgIndex, p.bgReset
		return
	}
	p.Selected, p.ResetSelected = p.fgIndex, p.fgReset
}

func (p *ColorPicker) saveCursor() {
	if p.Target == ColorPickerBG {
		p.bgIndex, p.bgReset = p.Selected, p.ResetSelected
		return
	}
	p.fgIndex, p.fgReset = p.Selected, p.ResetSelected
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

// Measure reports the two palette rows, slim edges and optional preview rug.
func (p *ColorPicker) Measure(width int) measure.Size {
	defs := SpeccedDefaults.ColorPicker
	w := defs.LeadingPadding + (len(p.Palette)+1)*defs.SwatchWidth
	if width > 0 && width < w {
		w = width
	}
	h := 4
	if defs.ShowPreview {
		h += 2
	}
	return measure.Size{Width: w, Height: h}
}

// Draw renders independent foreground and background palettes and their preview.
func (p *ColorPicker) Draw(c *Canvas, r Rect) {
	if c == nil || p == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	p.lastRect = r
	panelHeight := min(r.H, 4)
	defs := SpeccedDefaults.ColorPicker
	panelBG := ColorIndex(uint8(defs.PanelBG))
	panelStyle := Style{FG: p.Style.FG, BG: panelBG}
	c.PaintSurface(Rect{X: r.X, Y: r.Y, W: r.W, H: panelHeight}, panelStyle)
	if defs.ShowPreview && r.H >= 6 {
		c.PaintSurface(Rect{X: r.X, Y: r.Y + 4, W: min(r.W, defs.PreviewWidth), H: 2}, panelStyle)
	}
	paletteX := defs.LeadingPadding
	for row, target := range []ColorPickerTarget{ColorPickerFG, ColorPickerBG} {
		y := row + 1
		for cell := 0; cell < len(p.Palette)+1; cell++ {
			x := paletteX + cell*defs.SwatchWidth
			if x >= r.W {
				break
			}
			style := panelStyle
			text := defs.Glyphs.Reset
			if cell == 0 {
				if target == ColorPickerFG {
					style.FG = ColorReset()
					style.BG = panelBG
				} else {
					style.FG = panelStyle.FG
					style.BG = ColorReset()
				}
			} else if target == ColorPickerFG {
				style.FG = p.Palette[cell-1]
				text = defs.Glyphs.Swatch
			} else {
				style.BG = p.Palette[cell-1]
				text = defs.Glyphs.Swatch
			}
			writeColorPickerText(c, r, x, y, text, style)
			for pad := measure.StringWidth(text); pad < defs.SwatchWidth && x+pad < r.W; pad++ {
				c.Set(r.X+x+pad, r.Y+y, Cell{Text: " ", Style: style})
			}
		}
	}
	markerFG := ColorIndex(uint8(defs.MarkerFG))
	fgX := paletteX + p.fgIndex*defs.SwatchWidth
	if !p.fgReset {
		fgX += defs.SwatchWidth
	}
	for x := 0; x < r.W; x++ {
		style := Style{FG: panelBG, BG: p.Style.BG}
		if x >= fgX && x < fgX+defs.SwatchWidth {
			style.FG = markerFG
		}
		writeColorPickerText(c, r, x, 0, defs.Glyphs.EdgeLower, style)
	}
	bgX := paletteX + p.bgIndex*defs.SwatchWidth
	if !p.bgReset {
		bgX += defs.SwatchWidth
	}
	if defs.ShowPreview && r.H >= 6 {
		for x := 0; x < r.W; x++ {
			if x < 9 && (x < bgX || x >= bgX+defs.SwatchWidth) {
				continue
			}
			style := Style{FG: panelBG, BG: p.Style.BG}
			if x >= bgX && x < bgX+defs.SwatchWidth {
				style.FG = markerFG
			}
			writeColorPickerText(c, r, x, 3, defs.Glyphs.EdgeUpper, style)
		}
		previewX := defs.PreviewPadding
		writeColorPickerText(c, r, previewX, 4, defs.Labels.Preview, p.PreviewStyle())
		for x := 0; x < min(r.W, defs.PreviewWidth); x++ {
			writeColorPickerText(c, r, x, 5, defs.Glyphs.EdgeUpper, Style{FG: panelBG, BG: p.Style.BG})
		}
	} else {
		for x := 0; x < r.W; x++ {
			if x < 9 && (x < bgX || x >= bgX+defs.SwatchWidth) {
				continue
			}
			style := Style{FG: panelBG, BG: p.Style.BG}
			if x >= bgX && x < bgX+defs.SwatchWidth {
				style.FG = markerFG
			}
			writeColorPickerText(c, r, x, 3, defs.Glyphs.EdgeUpper, style)
		}
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
	if target != p.Target {
		p.saveCursor()
		p.Target = target
		p.syncCursor()
	}
	p.saveCursor()
	if !assign || p.Selected < 0 || p.Selected >= len(p.Palette) {
		return
	}
	color := p.Palette[p.Selected]
	if p.ResetSelected {
		color = ColorReset()
	}
	if target == ColorPickerFG {
		p.FG = color
	} else {
		p.BG = color
	}
	if p.OnChange != nil {
		p.OnChange(target, color)
	}
}

func (p *ColorPicker) applyCursor() {
	p.saveCursor()
	p.changeTarget(ColorPickerFG, true)
	p.changeTarget(ColorPickerBG, true)
	p.changeTarget(p.Target, false)
}

// ConsumeKey handles row navigation, selection and completion.
func (p *ColorPicker) ConsumeKey(e KeyEvent) EventResult {
	if p == nil {
		return Ignored()
	}
	switch {
	case p.matches(e, "apply"):
		p.applyCursor()
		if p.OnClose != nil {
			p.OnClose()
		}
		return DoneResult()
	case p.matches(e, "close"):
		if p.OnClose != nil {
			p.OnClose()
		}
		return DoneResult()
	case p.matches(e, "row_bg"):
		p.saveCursor()
		p.Target = ColorPickerBG
		p.syncCursor()
		return Handled()
	case p.matches(e, "row_fg"):
		p.saveCursor()
		p.Target = ColorPickerFG
		p.syncCursor()
		return Handled()
	case p.matches(e, "select"):
		p.changeTarget(p.Target, true)
		return Handled()
	case p.matches(e, "home"):
		p.ResetSelected = true
		p.saveCursor()
		return Handled()
	case p.matches(e, "left"):
		if p.ResetSelected {
			return Handled()
		}
		if p.Selected == 0 {
			p.ResetSelected = true
		} else if p.Selected > 0 {
			p.Selected--
		}
		p.saveCursor()
		return Handled()
	case p.matches(e, "right"):
		if p.ResetSelected {
			p.ResetSelected = false
		} else if p.Selected+1 < len(p.Palette) {
			p.Selected++
		}
		p.saveCursor()
		return Handled()
	default:
		return Ignored()
	}
}

// ConsumeMouse applies a left-clicked swatch in its row.
func (p *ColorPicker) ConsumeMouse(e MouseEvent) EventResult {
	if p == nil || e.Action != MousePress || e.X < 0 || e.X >= p.lastRect.W || e.Y < 0 || e.Y >= p.lastRect.H {
		return Ignored()
	}
	if e.Button != MouseLeft || e.Y < 1 || e.Y > 2 {
		return Ignored()
	}
	defs := SpeccedDefaults.ColorPicker
	x := e.X - defs.LeadingPadding
	if x < 0 {
		return Ignored()
	}
	cell := x / defs.SwatchWidth
	if cell == 0 {
		p.Selected, p.ResetSelected = 0, true
	} else if cell <= len(p.Palette) {
		p.Selected, p.ResetSelected = cell-1, false
	} else {
		return Ignored()
	}
	if e.Y == 1 {
		p.Target = ColorPickerFG
	} else {
		p.Target = ColorPickerBG
	}
	p.saveCursor()
	p.changeTarget(p.Target, true)
	return Handled()
}
