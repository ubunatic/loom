// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"

	"codeberg.org/ubunatic/loom/layout"
	"codeberg.org/ubunatic/loom/measure"
)

// ProviderState represents the lifecycle status of an external provider.
type ProviderState int

const (
	// ProviderPending indicates a provider awaiting initialization.
	ProviderPending ProviderState = iota
	// ProviderFetching indicates a provider currently being queried.
	ProviderFetching
	// ProviderDone indicates successful provider initialization.
	ProviderDone
	// ProviderFailed indicates provider error or timeout.
	ProviderFailed
	// ProviderSkipped indicates provider initialization was bypassed.
	ProviderSkipped
)

// ProviderPill represents a single provider status indicator badge.
type ProviderPill struct {
	Name   string
	Symbol string
	State  ProviderState
}

var (
	colorYellow = ColorIndex(3)
	colorGreen  = ColorIndex(2)
	colorRed    = ColorIndex(1)
)

// StyleForProviderState returns the visual style corresponding to a provider state.
func StyleForProviderState(state ProviderState) Style {
	switch state {
	case ProviderPending:
		return Style{Dim: true}
	case ProviderFetching:
		return Style{FG: colorYellow}
	case ProviderDone:
		return Style{FG: colorGreen}
	case ProviderFailed:
		return Style{FG: colorRed}
	case ProviderSkipped:
		return Style{Dim: true}
	default:
		return Style{}
	}
}

// PlainText returns the unstyled string representation: "<Symbol> <Name>".
func (p ProviderPill) PlainText() string {
	if p.Symbol == "" {
		return p.Name
	}
	return p.Symbol + " " + p.Name
}

// Width returns the display column width of the pill.
func (p ProviderPill) Width() int {
	return measure.StringWidth(p.PlainText())
}

// PillCluster is a widget that renders a horizontal row of ProviderPills.
type PillCluster struct {
	Pills []ProviderPill
	Gap   int          // spacing between pills; defaults to 2
	Align layout.Align // horizontal alignment within available width; default AlignCenter
}

// NewPillCluster creates a new PillCluster with default gap and centered alignment.
func NewPillCluster(pills ...ProviderPill) *PillCluster {
	return &PillCluster{
		Pills: pills,
		Gap:   2,
		Align: layout.AlignCenter,
	}
}

// ContentWidth computes the total horizontal column width of the pill cluster.
func (pc *PillCluster) ContentWidth() int {
	if len(pc.Pills) == 0 {
		return 0
	}
	gap := pc.Gap
	if gap < 0 {
		gap = 2
	}
	total := 0
	for i, pill := range pc.Pills {
		total += pill.Width()
		if i > 0 {
			total += gap
		}
	}
	return total
}

// ContentHeight is 1 for a single-row cluster.
func (pc *PillCluster) ContentHeight() int {
	if len(pc.Pills) == 0 {
		return 0
	}
	return 1
}

// Measure implements Measurer.
func (pc *PillCluster) Measure(width int) measure.Size {
	return measure.Size{
		Width:  pc.ContentWidth(),
		Height: pc.ContentHeight(),
	}
}

// Draw renders the pill cluster into the provided Rect on Canvas.
func (pc *PillCluster) Draw(c *Canvas, r Rect) {
	if len(pc.Pills) == 0 || r.W <= 0 || r.H <= 0 {
		return
	}
	gap := pc.Gap
	if gap < 0 {
		gap = 2
	}
	contentW := pc.ContentWidth()
	offsetX := layout.AlignOffset(r.W, contentW, pc.Align)

	x := r.X + offsetX
	y := r.Y

	for i, pill := range pc.Pills {
		if i > 0 {
			x += gap
		}
		if x >= r.X+r.W {
			break
		}
		style := StyleForProviderState(pill.State)
		text := pill.PlainText()
		c.Write(x, y, text, style)
		x += pill.Width()
	}
}

// HandleKey implements Widget.
func (pc *PillCluster) HandleKey(e KeyEvent) bool {
	return false
}

// HandleMouse implements Widget.
func (pc *PillCluster) HandleMouse(e MouseEvent) bool {
	return false
}

// FormatPillCluster returns an ANSI-formatted string of the pill cluster.
func FormatPillCluster(pills []ProviderPill, gap int, useANSI bool) string {
	if len(pills) == 0 {
		return ""
	}
	if gap < 0 {
		gap = 2
	}
	gapStr := strings.Repeat(" ", gap)
	var b strings.Builder
	for i, pill := range pills {
		if i > 0 {
			b.WriteString(gapStr)
		}
		text := pill.PlainText()
		if useANSI {
			style := StyleForProviderState(pill.State)
			b.WriteString(style.ANSI())
			b.WriteString(text)
			b.WriteString("\x1b[0m")
		} else {
			b.WriteString(text)
		}
	}
	return b.String()
}
