// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"math"
)

// PaintCanvas is a mouse-driven drawing area. Each terminal cell holds up to
// eight painted dots using the Unicode braille pattern block.
type PaintCanvas struct {
	// Smoothing is the maximum stroke simplification distance in braille dots.
	// Zero preserves the original lines. Closed rectangular strokes retain corners.
	Smoothing int
	// Controls shows a Tune button and enables T to cycle the specced tolerances.
	Controls   bool
	dots       map[[2]int]chartDots
	drawing    bool
	lastX      int
	lastY      int
	pixelW     int
	pixelH     int
	stroke     []paintPoint
	strokeBase map[[2]int]chartDots
}

type paintPoint struct{ x, y int }

// Clear removes all painted dots and ends the current stroke.
func (p *PaintCanvas) Clear() {
	p.dots = nil
	p.drawing = false
	p.stroke, p.strokeBase = nil, nil
}

// Draw renders painted dots into r, clipped to the widget's available area.
func (p *PaintCanvas) Draw(c *Canvas, r Rect) {
	if p == nil || c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	paintH := r.H
	if p.Controls {
		paintH--
	}
	p.pixelW, p.pixelH = r.W*2, paintH*4
	c.PaintSurface(r, Style{})
	for pos, dot := range p.dots {
		x, y := pos[0], pos[1]
		if x < 0 || y < 0 || x >= r.W || y >= paintH {
			continue
		}
		c.Write(r.X+x, r.Y+y, string(rune(0x2800+uint16(dot.mask))), dot.style)
	}
	if p.Controls {
		d := SpeccedDefaults.PaintCanvas
		text := fmt.Sprintf("[%s] %s: %d", d.TuneLabel, d.SmoothingLabel, p.Smoothing)
		c.Write(r.X, r.Y+paintH, TruncateText(text, r.W, ""), Style{})
	}
}

// ConsumeKey clears the drawing when the user presses c or Ctrl-L.
func (p *PaintCanvas) ConsumeKey(e KeyEvent) EventResult {
	if p.Controls && e.Is("t", "T") {
		p.tune()
		return Handled()
	}
	if e.Is("c", "C", "ctrl-l") {
		p.Clear()
		return Handled()
	}
	return Ignored()
}

// ConsumeMouse starts a stroke on left press, connects drag points with
// Bresenham lines, and ends the stroke on left release.
func (p *PaintCanvas) ConsumeMouse(e MouseEvent) EventResult {
	if p == nil {
		return Ignored()
	}
	if p.Controls && e.Action == MousePress && e.Button == MouseLeft && e.Y == p.pixelH/4 && e.X >= 0 && e.X < StringWidth(SpeccedDefaults.PaintCanvas.TuneLabel)+2 {
		p.tune()
		return Handled()
	}
	x, y := e.X*2, e.Y*4
	switch e.Action {
	case MousePress:
		if e.Button != MouseLeft || !p.inBounds(x, y) {
			return Ignored()
		}
		p.drawing = true
		p.lastX, p.lastY = x, y
		p.stroke = []paintPoint{{x, y}}
		p.strokeBase = clonePaintDots(p.dots)
		p.addLine(x, y, x, y)
		return Handled()
	case MouseDrag:
		if !p.drawing || e.Button != MouseLeft || !p.inBounds(x, y) {
			return Ignored()
		}
		if p.Smoothing > 0 {
			p.appendStroke(x, y)
		} else {
			p.addLine(p.lastX, p.lastY, x, y)
		}
		p.lastX, p.lastY = x, y
		return Handled()
	case MouseRelease:
		if !p.drawing || e.Button != MouseLeft {
			return Ignored()
		}
		p.drawing = false
		if p.Smoothing > 0 && p.inBounds(x, y) {
			p.appendStroke(x, y)
		}
		p.stroke, p.strokeBase = nil, nil
		return Handled()
	default:
		return Ignored()
	}
}

func (p *PaintCanvas) tune() {
	levels := SpeccedDefaults.PaintCanvas.SmoothingLevels
	for i, value := range levels {
		if value == p.Smoothing {
			p.Smoothing = levels[(i+1)%len(levels)]
			p.drawing = false
			p.stroke, p.strokeBase = nil, nil
			return
		}
	}
	if len(levels) > 0 {
		p.Smoothing = levels[0]
	}
}

func clonePaintDots(dots map[[2]int]chartDots) map[[2]int]chartDots {
	copy := make(map[[2]int]chartDots, len(dots))
	for pos, dot := range dots {
		copy[pos] = dot
	}
	return copy
}

func (p *PaintCanvas) appendStroke(x, y int) {
	point := paintPoint{x, y}
	if len(p.stroke) == 0 || p.stroke[len(p.stroke)-1] != point {
		p.stroke = append(p.stroke, point)
	}
	points := p.stroke
	if !rectangularStroke(points) {
		points = simplifyPaintStroke(points, float64(p.Smoothing))
	}
	p.dots = clonePaintDots(p.strokeBase)
	if len(points) == 1 {
		p.addLine(x, y, x, y)
	}
	for i := 1; i < len(points); i++ {
		a, b := points[i-1], points[i]
		p.addLine(a.x, a.y, b.x, b.y)
	}
}

// Rectangles are deliberate corners, even when their edges are shorter than
// the smoothing tolerance. Axis-aligned runs form four distinct directions.
func rectangularStroke(points []paintPoint) bool {
	if len(points) < 5 || points[0] != points[len(points)-1] {
		return false
	}
	directions := []int{}
	for i := 1; i < len(points); i++ {
		a, b := points[i-1], points[i]
		direction := 0
		switch {
		case a.x < b.x && a.y == b.y:
			direction = 1
		case a.x == b.x && a.y < b.y:
			direction = 2
		case a.x > b.x && a.y == b.y:
			direction = 3
		case a.x == b.x && a.y > b.y:
			direction = 4
		default:
			return false
		}
		if len(directions) == 0 || directions[len(directions)-1] != direction {
			directions = append(directions, direction)
		}
	}
	if len(directions) > 1 && directions[0] == directions[len(directions)-1] {
		directions = directions[:len(directions)-1]
	}
	if len(directions) != 4 {
		return false
	}
	seen := map[int]bool{}
	for _, direction := range directions {
		seen[direction] = true
	}
	return len(seen) == 4
}

func simplifyPaintStroke(points []paintPoint, tolerance float64) []paintPoint {
	if len(points) < 3 {
		return points
	}
	a, b := points[0], points[len(points)-1]
	dx, dy := float64(b.x-a.x), float64(b.y-a.y)
	length := dx*dx + dy*dy
	index, farthest := 0, tolerance*tolerance
	for i := 1; i < len(points)-1; i++ {
		px, py := float64(points[i].x-a.x), float64(points[i].y-a.y)
		fraction := 0.0
		if length > 0 {
			fraction = math.Max(0, math.Min(1, (px*dx+py*dy)/length))
		}
		x, y := px-fraction*dx, py-fraction*dy
		if distance := x*x + y*y; distance > farthest {
			index, farthest = i, distance
		}
	}
	if index == 0 {
		return []paintPoint{a, b}
	}
	left := simplifyPaintStroke(points[:index+1], tolerance)
	right := simplifyPaintStroke(points[index:], tolerance)
	result := make([]paintPoint, 0, len(left)+len(right)-1)
	result = append(result, left[:len(left)-1]...)
	return append(result, right...)
}

func (p *PaintCanvas) inBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < p.pixelW && y < p.pixelH
}

func (p *PaintCanvas) addLine(x0, y0, x1, y1 int) {
	if p.dots == nil {
		p.dots = make(map[[2]int]chartDots)
	}
	rasterChartLine(p.dots, x0, y0, x1, y1, Style{})
}
