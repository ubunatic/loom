// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// PaintCanvas is a mouse-driven drawing area. Each terminal cell holds up to
// eight painted dots using the Unicode braille pattern block.
type PaintCanvas struct {
	dots    map[[2]int]chartDots
	drawing bool
	lastX   int
	lastY   int
	pixelW  int
	pixelH  int
}

// Clear removes all painted dots and ends the current stroke.
func (p *PaintCanvas) Clear() {
	p.dots = nil
	p.drawing = false
}

// Draw renders painted dots into r, clipped to the widget's available area.
func (p *PaintCanvas) Draw(c *Canvas, r Rect) {
	if p == nil || c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	p.pixelW, p.pixelH = r.W*2, r.H*4
	c.PaintSurface(r, Style{})
	for pos, dot := range p.dots {
		x, y := pos[0], pos[1]
		if x < 0 || y < 0 || x >= r.W || y >= r.H {
			continue
		}
		c.Write(r.X+x, r.Y+y, string(rune(0x2800+uint16(dot.mask))), dot.style)
	}
}

// ConsumeKey clears the drawing when the user presses c or Ctrl-L.
func (p *PaintCanvas) ConsumeKey(e KeyEvent) EventResult {
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
	x, y := e.X*2, e.Y*4
	switch e.Action {
	case MousePress:
		if e.Button != MouseLeft || !p.inBounds(x, y) {
			return Ignored()
		}
		p.drawing = true
		p.lastX, p.lastY = x, y
		p.addLine(x, y, x, y)
		return Handled()
	case MouseDrag:
		if !p.drawing || e.Button != MouseLeft || !p.inBounds(x, y) {
			return Ignored()
		}
		p.addLine(p.lastX, p.lastY, x, y)
		p.lastX, p.lastY = x, y
		return Handled()
	case MouseRelease:
		if !p.drawing || e.Button != MouseLeft {
			return Ignored()
		}
		p.drawing = false
		return Handled()
	default:
		return Ignored()
	}
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
