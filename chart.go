// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

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
	legend := len(series) > 1
	plotX, plotY, plotW, plotH := r.X+6, r.Y, r.W-7, r.H-2
	if legend {
		plotH--
	}
	if plotW <= 0 || plotH <= 0 {
		return
	}
	labelStyle := Style{FG: ColorIndex(7)}
	axisStyle := Style{FG: ColorIndex(8)}
	for i := 0; i <= 4; i++ {
		y := plotY + plotH - 1 - i*(plotH-1)/4
		v := minV + (maxV-minV)*float64(i)/4
		label := chartNumber(v)
		c.Write(r.X, y, strings.Repeat(" ", maxInt(0, 5-len([]rune(label))))+label, labelStyle)
		c.Write(plotX, y, "│", axisStyle)
	}
	c.Write(plotX, plotY+plotH, "└"+strings.Repeat("─", maxInt(0, plotW-1)), axisStyle)
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
		ch.drawBars(c, series, plotX, plotY, plotW, plotH, minV, maxV, count)
	} else {
		ch.drawLines(c, series, plotX, plotY, plotW, plotH, minV, maxV, count)
	}
	for i := 0; i < count; i++ {
		if i%maxInt(1, count/4) == 0 {
			x := plotX + i*maxInt(0, plotW-1)/maxInt(1, count-1)
			c.Write(x, plotY+plotH+1, strconv.Itoa(i+1), labelStyle)
		}
	}
	if legend {
		x, y := plotX, r.Y+r.H-1
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

func (ch *Chart) valuePos(v, minV, maxV float64, y, h int) int {
	return y + h - 1 - int(math.Round((v-minV)/(maxV-minV)*float64(maxInt(0, h-1))))
}

func (ch *Chart) drawLines(c *Canvas, ss []ChartSeries, x, y, w, h int, lo, hi float64, n int) {
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
			px := x + i*maxInt(0, w-1)/maxInt(1, n-1)
			py := ch.valuePos(v, lo, hi, y, h)
			if px >= x && px < x+w && py >= y && py < y+h {
				c.Write(px, py, "⠿", st)
				if prevX >= 0 {
					steps := maxInt(absInt(px-prevX), absInt(py-prevY))
					for k := 1; k < steps; k++ {
						xx := prevX + (px-prevX)*k/steps
						yy := prevY + (py-prevY)*k/steps
						if xx >= x && xx < x+w && yy >= y && yy < y+h {
							c.Write(xx, yy, "⠿", st)
						}
					}
				}
			}
			prevX, prevY = px, py
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
func (*Chart) HandleKey(KeyEvent) bool     { return false }
func (*Chart) HandleMouse(MouseEvent) bool { return false }
func (*Chart) ContentWidth() int           { return 32 }
func (*Chart) ContentHeight() int          { return 10 }
