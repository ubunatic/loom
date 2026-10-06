// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"ubunatic.com/loom/measure"
)

// chartPlotHeight is the plot height a Chart prefers when a container asks.
const chartPlotHeight = 5

// ChartMode selects line or grouped-bar rendering.
type ChartMode uint8

const (
	ChartLine ChartMode = iota
	ChartGroupedBar
)

// ChartSeries is one named numeric series. Values are plotted in input order.
type ChartSeries struct {
	Name   string
	Values []float64
	Color  Color
}

// Chart draws multiple numeric series with compact labeled axes.
type Chart struct {
	Series []ChartSeries
	Mode   ChartMode
	Min    float64
	Max    float64
	// RangeSet fixes the vertical range to Min..Max; otherwise data is scaled.
	RangeSet bool
}

var chartColors = [...]Color{ColorIndex(6), ColorIndex(3), ColorIndex(5), ColorIndex(2), ColorIndex(4), ColorIndex(1)}

// Draw renders the chart clipped to r. It is safe for small or empty regions.
func (ch *Chart) Draw(c *Canvas, r Rect) {
	if ch == nil || c == nil || r.W <= 0 || r.H <= 0 || len(ch.Series) == 0 {
		return
	}
	series := ch.Series
	minV, maxV := math.Inf(1), math.Inf(-1)
	for _, s := range series {
		for _, v := range s.Values {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				continue
			}
			minV = math.Min(minV, v)
			maxV = math.Max(maxV, v)
		}
	}
	if ch.RangeSet {
		minV, maxV = ch.Min, ch.Max
	}
	if math.IsInf(minV, 1) {
		minV, maxV = 0, 1
	}
	if minV > maxV {
		minV, maxV = maxV, minV
	}
	if minV == maxV {
		pad := math.Abs(minV) * .1
		if pad == 0 {
			pad = 1
		}
		minV -= pad
		maxV += pad
	}
	minV, maxV, step := niceChartRange(minV, maxV)
	legend := len(series) > 1
	axisX, plotY, plotW, plotH := r.X+6, r.Y, r.W-7, r.H-2
	if legend {
		plotH--
	}
	if plotW <= 0 || plotH <= 0 {
		return
	}
	dataX, dataW := axisX+1, plotW-1
	if dataW <= 0 {
		return
	}
	labelStyle := Style{FG: ColorIndex(7)}
	axisStyle := Style{FG: ColorIndex(8)}
	tickCount := int(math.Round((maxV-minV)/step)) + 1
	for i := 0; i < tickCount; i++ {
		y := plotY + plotH - 1 - i*(plotH-1)/maxInt(1, tickCount-1)
		v := minV + step*float64(i)
		label := chartNumber(v)
		c.Write(r.X, y, strings.Repeat(" ", maxInt(0, 5-len([]rune(label))))+label, labelStyle)
	}
	for y := plotY; y < plotY+plotH; y++ {
		c.Write(axisX, y, "│", axisStyle)
	}
	c.Write(axisX, plotY+plotH, "└"+strings.Repeat("─", maxInt(0, plotW-1)), axisStyle)
	count := 0
	for _, s := range series {
		if len(s.Values) > count {
			count = len(s.Values)
		}
	}
	if count == 0 {
		return
	}
	if ch.Mode == ChartGroupedBar {
		ch.drawBars(c, series, dataX, plotY, dataW, plotH, minV, maxV, count)
	} else {
		ch.drawLines(c, series, dataX, plotY, dataW, plotH, minV, maxV, count)
	}
	for i := 0; i < count; i++ {
		if i%maxInt(1, count/4) == 0 {
			x := dataX + i*maxInt(0, dataW-1)/maxInt(1, count-1)
			c.Write(x, plotY+plotH+1, strconv.Itoa(i+1), labelStyle)
		}
	}
	if legend {
		x, y := dataX, r.Y+r.H-1
		for i, s := range series {
			name := s.Name
			if name == "" {
				name = fmt.Sprintf("Series %d", i+1)
			}
			color := s.Color
			if color == (Color{}) {
				color = chartColors[i%len(chartColors)]
			}
			c.Write(x, y, "● "+name, Style{FG: color})
			x += len([]rune("● "+name)) + 2
			if x >= r.X+r.W {
				break
			}
		}
	}
}

// Measure returns the preferred size at the given width: a plot of
// chartPlotHeight lines plus the axis and label lines, and a legend line when
// the chart has more than one series. Draw still fills any height it gets.
func (ch *Chart) Measure(width int) measure.Size {
	height := chartPlotHeight + 2
	if ch != nil && len(ch.Series) > 1 {
		height++
	}
	return measure.Size{Width: width, Height: height}
}

func (ch *Chart) valuePos(v, minV, maxV float64, y, h int) int {
	return y + h - 1 - int(math.Round((v-minV)/(maxV-minV)*float64(maxInt(0, h-1))))
}

type chartDots struct {
	mask  uint8
	style Style
}

func (ch *Chart) drawLines(c *Canvas, ss []ChartSeries, x, y, w, h int, lo, hi float64, n int) {
	pixels := make(map[[2]int]chartDots)
	pixelW, pixelH := w*2, h*4
	for si, s := range ss {
		color := s.Color
		if color == (Color{}) {
			color = chartColors[si%len(chartColors)]
		}
		st := Style{FG: color}
		prevX, prevY := -1, -1
		for i, v := range s.Values {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				prevX = -1
				continue
			}
			px := i * maxInt(0, pixelW-1) / maxInt(1, n-1)
			py := pixelH - 1 - int(math.Round((v-lo)/(hi-lo)*float64(maxInt(0, pixelH-1))))
			if px >= 0 && px < pixelW && py >= 0 && py < pixelH {
				if prevX >= 0 {
					rasterChartLine(pixels, prevX, prevY, px, py, st)
				} else {
					setChartDot(pixels, px, py, st)
				}
			}
			prevX, prevY = px, py
		}
	}
	for pos, dot := range pixels {
		c.Write(x+pos[0], y+pos[1], string(rune(0x2800+uint16(dot.mask))), dot.style)
	}
}

func setChartDot(pixels map[[2]int]chartDots, px, py int, style Style) {
	bits := [4][2]uint8{{0, 3}, {1, 4}, {2, 5}, {6, 7}}
	key := [2]int{px / 2, py / 4}
	dot := pixels[key]
	dot.mask |= 1 << bits[py%4][px%2]
	dot.style = style
	pixels[key] = dot
}

func rasterChartLine(pixels map[[2]int]chartDots, x0, y0, x1, y1 int, style Style) {
	dx, dy := absInt(x1-x0), -absInt(y1-y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		setChartDot(pixels, x0, y0, style)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func (ch *Chart) drawBars(c *Canvas, ss []ChartSeries, x, y, w, h int, lo, hi float64, n int) {
	group := maxInt(1, w/n)
	bw := maxInt(1, (group-1)/len(ss))
	for si, s := range ss {
		color := s.Color
		if color == (Color{}) {
			color = chartColors[si%len(chartColors)]
		}
		for i, v := range s.Values {
			bx := x + i*group + si*bw
			top := ch.valuePos(v, lo, hi, y, h)
			base := ch.valuePos(0, lo, hi, y, h)
			if 0 < lo {
				base = y + h - 1
			}
			if top > base {
				top, base = base, top
			}
			for yy := maxInt(y, top); yy <= minInt(y+h-1, base); yy++ {
				for xx := bx; xx < minInt(x+w, bx+bw); xx++ {
					c.Write(xx, yy, "█", Style{FG: color})
				}
			}
		}
	}
}

func chartNumber(v float64) string {
	a := math.Abs(v)
	if a >= 1000 || (a > 0 && a < .01) {
		return fmt.Sprintf("%.1g", v)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}
func niceChartStep(raw float64) float64 {
	if raw <= 0 || math.IsNaN(raw) || math.IsInf(raw, 0) {
		return 1
	}
	exponent := math.Floor(math.Log10(raw))
	unit := math.Pow(10, exponent)
	fraction := raw / unit
	switch {
	case fraction <= 1:
		return unit
	case fraction <= 2:
		return 2 * unit
	case fraction <= 5:
		return 5 * unit
	default:
		return 10 * unit
	}
}
func niceChartRange(minV, maxV float64) (float64, float64, float64) {
	step := niceChartStep((maxV - minV) / 4)
	minV = math.Floor(minV/step) * step
	maxV = math.Ceil(maxV/step) * step
	if minV == maxV {
		maxV = minV + step
	}
	return minV, maxV, step
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func (*Chart) ConsumeKey(KeyEvent) EventResult     { return Ignored() }
func (*Chart) ConsumeMouse(MouseEvent) EventResult { return Ignored() }
func (*Chart) ContentWidth() int                   { return 32 }
func (*Chart) ContentHeight() int                  { return 10 }
