// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package gallery

import (
	"ubunatic.com/loom"
	"ubunatic.com/loom/measure"
)

const (
	colorPickerSampleColumnWidth = 20
	colorPickerSampleGap         = 2
	colorPickerSampleHeight      = 2
)

var colorPickerSamples = []struct {
	label string
	style func(loom.ThemeColors) loom.Style
}{
	{"Default", func(loom.ThemeColors) loom.Style { return loom.Style{FG: loom.ColorReset(), BG: loom.ColorReset()} }},
	{"Faint text", func(loom.ThemeColors) loom.Style {
		return loom.Style{FG: loom.ColorReset(), BG: loom.ColorReset(), Dim: true}
	}},
	{"Subdued BG (theme)", func(theme loom.ThemeColors) loom.Style {
		return loom.Style{FG: loom.ColorReset(), BG: theme.NormalBG.Color()}
	}},
}

type colorPickerDemo struct {
	picker     *loom.ColorPicker
	theme      loom.ThemeColors
	pickerRect loom.Rect
}

func newColorPickerDemo() *colorPickerDemo {
	demo := &colorPickerDemo{picker: loom.NewColorPicker(loom.ColorIndex(4), loom.ColorIndex(15))}
	demo.ApplyTheme(loom.DefaultTheme())
	return demo
}

func (w *colorPickerDemo) Draw(c *loom.Canvas, r loom.Rect) {
	if c == nil || w == nil || w.picker == nil || r.W <= 0 || r.H <= 0 {
		if w != nil {
			w.pickerRect = loom.Rect{}
		}
		return
	}
	pickerHeight := min(r.H, w.picker.Measure(r.W).Height)
	w.pickerRect = loom.Rect{W: r.W, H: pickerHeight}
	w.picker.Draw(c, loom.Rect{X: r.X, Y: r.Y, W: r.W, H: pickerHeight})

	sampleY := r.Y + pickerHeight
	if sampleY < r.Y+r.H {
		sampleY++
	}
	if sampleY >= r.Y+r.H {
		return
	}
	if r.W >= len(colorPickerSamples)*colorPickerSampleColumnWidth+(len(colorPickerSamples)-1)*colorPickerSampleGap {
		w.drawWideSamples(c, r, sampleY)
		return
	}
	w.drawStackedSamples(c, r, sampleY)
}

func (w *colorPickerDemo) drawWideSamples(c *loom.Canvas, r loom.Rect, y int) {
	columnWidth := (r.W - (len(colorPickerSamples)-1)*colorPickerSampleGap) / len(colorPickerSamples)
	for i, sample := range colorPickerSamples {
		x := r.X + i*(columnWidth+colorPickerSampleGap)
		w.drawSample(c, loom.Rect{X: x, Y: y, W: min(columnWidth, r.X+r.W-x), H: min(colorPickerSampleHeight, r.Y+r.H-y)}, sample)
	}
}

func (w *colorPickerDemo) drawStackedSamples(c *loom.Canvas, r loom.Rect, y int) {
	for i, sample := range colorPickerSamples {
		sampleY := y + i*colorPickerSampleHeight
		if sampleY >= r.Y+r.H {
			return
		}
		w.drawSample(c, loom.Rect{X: r.X, Y: sampleY, W: r.W, H: min(colorPickerSampleHeight, r.Y+r.H-sampleY)}, sample)
	}
}

func (w *colorPickerDemo) drawSample(c *loom.Canvas, r loom.Rect, sample struct {
	label string
	style func(loom.ThemeColors) loom.Style
}) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	labelStyle := loom.Style{FG: w.theme.NormalFG.Color(), BG: w.theme.NormalBG.Color()}
	w.writeSampleText(c, r, 0, sample.label, labelStyle)
	if r.H > 1 {
		w.writeSampleText(c, r, 1, "Aa sample text", sample.style(w.theme))
	}
}

func (w *colorPickerDemo) writeSampleText(c *loom.Canvas, r loom.Rect, row int, text string, style loom.Style) {
	x := 0
	for _, char := range text {
		cluster := string(char)
		width := loom.StringWidth(cluster)
		if x+width > r.W {
			return
		}
		c.Set(r.X+x, r.Y+row, loom.Cell{Text: cluster, Style: style, Claim: true})
		x += width
	}
}

func (w *colorPickerDemo) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if w == nil || w.picker == nil {
		return loom.Ignored()
	}
	return w.picker.ConsumeKey(e)
}

func (w *colorPickerDemo) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	if w == nil || w.picker == nil || !w.pickerRect.Contains(e.X, e.Y) {
		return loom.Ignored()
	}
	return w.picker.ConsumeMouse(e)
}

func (w *colorPickerDemo) Focused() bool { return w != nil && w.picker != nil && w.picker.Focused() }

func (w *colorPickerDemo) SetFocus(focused bool) {
	if w != nil && w.picker != nil {
		w.picker.SetFocus(focused)
	}
}

func (w *colorPickerDemo) Measure(width int) measure.Size {
	if w == nil || w.picker == nil {
		return measure.Size{}
	}
	pickerSize := w.picker.Measure(width)
	height := pickerSize.Height + 1
	if width >= len(colorPickerSamples)*colorPickerSampleColumnWidth+(len(colorPickerSamples)-1)*colorPickerSampleGap {
		height += colorPickerSampleHeight
	} else {
		height += len(colorPickerSamples) * colorPickerSampleHeight
	}
	return measure.Size{Width: pickerSize.Width, Height: height}
}

func (w *colorPickerDemo) ApplyTheme(theme loom.ThemeColors) {
	if w == nil {
		return
	}
	w.theme = theme
	if w.picker != nil {
		w.picker.ApplyTheme(theme)
	}
}
