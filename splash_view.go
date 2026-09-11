// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"codeberg.org/ubunatic/loom/graph"
	"codeberg.org/ubunatic/loom/layout"
	"codeberg.org/ubunatic/loom/measure"
)

// SplashView is a declarative widget that renders a complete centered splash screen.
type SplashView struct {
	Title        string
	SpinnerFrame int
	Progress     float64 // 0.0 to 100.0
	Pattern      string  // if set, overrides progress bar with repeating pattern (e.g. ":")
	StepText     string
	Pills        []ProviderPill
	FooterText   string // keyboard hint, e.g. "Esc to skip"
	BracketWidth int    // inner bar width, default 24 or 32
	PillGap      int    // spacing between pills, default 2
	Controller   *SplashController
}


// NewSplashView creates a SplashView with default settings matching HarnezSplashTarget.
func NewSplashView(title string, pills ...ProviderPill) *SplashView {
	return &SplashView{
		Title:        title,
		BracketWidth: SpeccedDefaults.Splash.BracketWidth,
		PillGap:      SpeccedDefaults.Splash.PillGap,
		FooterText:   SpeccedDefaults.Splash.FooterText,
		Pills:        pills,
		StepText:     SpeccedDefaults.Splash.StepText,
	}
}


// ApplySnapshot updates the view properties from a SplashSnapshot.
func (sv *SplashView) ApplySnapshot(snap SplashSnapshot) {
	sv.SpinnerFrame = snap.SpinnerFrame
	sv.Progress = snap.Progress
	sv.StepText = snap.StepText
	sv.Pills = snap.Pills
	if snap.Completed {
		sv.Pattern = ":"
	}
}

// ContentHeight returns the fixed height required for the splash content block.
func (sv *SplashView) ContentHeight() int {
	return 8
}

// ContentWidth returns the preferred width (the widest sub-component).
func (sv *SplashView) ContentWidth() int {
	w := 0
	// Title line: spinner (1) + 2 spaces + title
	titleW := 3 + measure.StringWidth(sv.Title)
	if titleW > w {
		w = titleW
	}
	// Bar width
	barW := sv.BracketWidth + 2
	if barW > w {
		w = barW
	}
	// Step text
	stepW := measure.StringWidth(sv.StepText)
	if stepW > w {
		w = stepW
	}
	// Pills
	cluster := NewPillCluster(sv.Pills...)
	cluster.Gap = sv.PillGap
	if pillW := cluster.ContentWidth(); pillW > w {
		w = pillW
	}
	// Footer
	footW := measure.StringWidth(sv.FooterText)
	if footW > w {
		w = footW
	}
	return w
}

// Measure implements Measurer.
func (sv *SplashView) Measure(width int) measure.Size {
	return measure.Size{
		Width:  sv.ContentWidth(),
		Height: sv.ContentHeight(),
	}
}

// Draw renders the splash screen centered horizontally and vertically inside r.
func (sv *SplashView) Draw(c *Canvas, r Rect) {
	if r.W <= 0 || r.H <= 0 {
		return
	}

	totalH := sv.ContentHeight()
	offsetY := layout.AlignOffset(r.H, totalH, layout.AlignCenter)
	startY := r.Y + offsetY

	// Line 0: Spinner + "  " + Title
	if startY < r.Y+r.H {
		spinner := string(graph.SpinnerGlyph(sv.SpinnerFrame))
		titleLine := spinner + "  " + sv.Title
		titleW := measure.StringWidth(titleLine)
		titleX := r.X + layout.AlignOffset(r.W, titleW, layout.AlignCenter)
		c.Write(titleX, startY, titleLine, Style{Bold: true})
	}

	// Line 2: Bracketed Progress Bar
	barY := startY + 2
	if barY < r.Y+r.H {
		barOpts := graph.BracketedBarOptions{
			Width:   sv.BracketWidth,
			Pattern: sv.Pattern,
			SubChar: true,
		}
		barStr := graph.RenderBracketedBar(sv.Progress, barOpts)
		barW := graph.BracketedBarWidth(barOpts)
		barX := r.X + layout.AlignOffset(r.W, barW, layout.AlignCenter)
		c.Write(barX, barY, barStr, Style{})
	}

	// Line 4: Step text
	stepY := startY + 4
	if stepY < r.Y+r.H {
		stepW := measure.StringWidth(sv.StepText)
		stepX := r.X + layout.AlignOffset(r.W, stepW, layout.AlignCenter)
		c.Write(stepX, stepY, sv.StepText, Style{})
	}

	// Line 5: Provider pills
	pillY := startY + 5
	if pillY < r.Y+r.H {
		cluster := NewPillCluster(sv.Pills...)
		cluster.Gap = sv.PillGap
		cluster.Align = layout.AlignCenter
		cluster.Draw(c, Rect{X: r.X, Y: pillY, W: r.W, H: 1})
	}

	// Line 7: Footer hint
	footY := startY + 7
	if footY < r.Y+r.H && sv.FooterText != "" {
		footW := measure.StringWidth(sv.FooterText)
		footX := r.X + layout.AlignOffset(r.W, footW, layout.AlignCenter)
		c.Write(footX, footY, sv.FooterText, Style{Dim: true})
	}
}

// HandleKey implements Widget.
func (sv *SplashView) HandleKey(e KeyEvent) bool {
	if sv.Controller != nil {
		return sv.Controller.HandleKey(e)
	}
	key := e.Key
	if key == "" {
		key = e.Text
	}
	switch key {
	case "esc", "q", "enter", "ctrl-c", "ctrl-q":
		return true
	}
	return false
}


// HandleMouse implements Widget.
func (sv *SplashView) HandleMouse(e MouseEvent) bool {
	return false
}
