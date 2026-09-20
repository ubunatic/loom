// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package winch provides a diagnostic application that exposes Loom's resize
// modes as runtime switches to test and reproduce resize behaviors.
package winch

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"codeberg.org/ubunatic/loom"
)

// App is the Winch resize diagnostics widget.
type App struct {
	pane       *loom.Pane
	themeName  string
	theme      loom.ThemeColors
	themeNames []string
	themeIdx   int
	config     loom.ResizeConfig
	frameCount int
	quitting   bool
}

// NewApp creates a new Winch diagnostic application. pane may be nil in headless tests.
func NewApp(pane *loom.Pane, themeName string, theme loom.ThemeColors) (*App, error) {
	names := availableThemes()
	idx := 0
	for i, name := range names {
		if name == themeName {
			idx = i
			break
		}
	}

	cfg := loom.DefaultResizeConfig()
	if pane != nil {
		cfg = pane.ResizeConfig
	}

	return &App{
		pane:       pane,
		themeName:  themeName,
		theme:      theme,
		themeNames: names,
		themeIdx:   idx,
		config:     cfg,
	}, nil
}

// Config returns the current resize configuration.
func (a *App) Config() loom.ResizeConfig {
	if a.pane != nil {
		return a.pane.ResizeConfig
	}
	return a.config
}

// SetConfig updates the active resize configuration.
func (a *App) SetConfig(cfg loom.ResizeConfig) {
	a.config = cfg
	if a.pane != nil {
		a.pane.ResizeConfig = cfg
	}
}

// Draw renders the Winch diagnostic dashboard into the canvas.
func (a *App) Draw(c *loom.Canvas, r loom.Rect) {
	a.frameCount++
	if r.W <= 0 || r.H <= 0 {
		return
	}

	cfg := a.Config()
	choiceStyle := a.theme.ChoiceStyle()
	frameStyle := a.theme.FrameStyle()
	normalStyle := choiceStyle.Normal
	headerStyle := choiceStyle.Selected
	statusStyle := frameStyle.Status
	contentH := r.H
	if contentH > 1 {
		contentH-- // reserve the final row for the status bar
	}

	// Determine layout: side-by-side if wide (>= 72 cols), stacked if narrow
	if r.W >= 72 {
		modesW := 40
		diagW := r.W - modesW - 1
		a.drawModesPanel(c, loom.Rect{X: r.X, Y: r.Y, W: modesW, H: contentH}, cfg, normalStyle, headerStyle)
		a.drawStressPanel(c, loom.Rect{X: r.X + modesW + 1, Y: r.Y, W: diagW, H: contentH}, cfg, normalStyle, headerStyle)
	} else {
		// Reserve one row per spec-defined mode plus the panel borders.
		modesH := len(loom.SpeccedResizeModeIDs) + 2
		if modesH > contentH-2 {
			modesH = contentH - 2
		}
		if modesH < 3 {
			modesH = contentH
		}
		a.drawModesPanel(c, loom.Rect{X: r.X, Y: r.Y, W: r.W, H: modesH}, cfg, normalStyle, headerStyle)
		if contentH > modesH+2 {
			a.drawStressPanel(c, loom.Rect{X: r.X, Y: r.Y + modesH, W: r.W, H: contentH - modesH}, cfg, normalStyle, headerStyle)
		}
	}

	// Status bar at bottom row
	if r.H > 1 {
		status := fmt.Sprintf(" Winch Diagnostic • Theme: %s • Frames: %d • [Q] Quit", a.themeName, a.frameCount)
		status = loom.TruncateText(status, r.W, "")
		c.Write(r.X, r.Y+r.H-1, status, statusStyle)
	}
}

func drawBoxBorder(c *loom.Canvas, r loom.Rect, border loom.BoxBorder, title string, style, titleStyle loom.Style) {
	if r.W < 2 || r.H < 2 {
		return
	}
	for x := r.X + 1; x < r.X+r.W-1; x++ {
		c.Set(x, r.Y, loom.Cell{Text: border.Horizontal, Style: style})
		c.Set(x, r.Y+r.H-1, loom.Cell{Text: border.Horizontal, Style: style})
	}
	for y := r.Y + 1; y < r.Y+r.H-1; y++ {
		c.Set(r.X, y, loom.Cell{Text: border.Vertical, Style: style})
		c.Set(r.X+r.W-1, y, loom.Cell{Text: border.Vertical, Style: style})
	}
	c.Set(r.X, r.Y, loom.Cell{Text: border.TopLeft, Style: style})
	c.Set(r.X+r.W-1, r.Y, loom.Cell{Text: border.TopRight, Style: style})
	c.Set(r.X, r.Y+r.H-1, loom.Cell{Text: border.BottomLeft, Style: style})
	c.Set(r.X+r.W-1, r.Y+r.H-1, loom.Cell{Text: border.BottomRight, Style: style})
	if title != "" && r.W > 4 {
		c.Write(r.X+2, r.Y, loom.TruncateText(" "+title+" ", r.W-4, ""), titleStyle)
	}
}

func (a *App) drawModesPanel(c *loom.Canvas, r loom.Rect, cfg loom.ResizeConfig, normal, header loom.Style) {
	if r.W < 4 || r.H < 2 {
		return
	}
	border := loom.BoxBorder{
		TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘",
		Horizontal: "─", Vertical: "│",
	}
	drawBoxBorder(c, r, border, "Resize Modes [1-9,A,B Toggle, +/- Guard, R Reset]", normal, header)

	row := r.Y + 1
	for _, id := range loom.SpeccedResizeModeIDs {
		if row >= r.Y+r.H-1 {
			break
		}
		mode, ok := loom.SpeccedResizeModes.Modes[id]
		if !ok {
			continue
		}

		enabled, _ := cfg.Get(id)
		stateTag := "[OFF]"
		stateStyle := loom.Style{Dim: true}
		if enabled {
			stateTag = "[ON] "
			stateStyle = loom.Style{Bold: true}
		}

		keyBadge := fmt.Sprintf("[%s]", mode.Key)
		tag := ""
		if id == "width_guard" {
			tag = fmt.Sprintf(" (n=%d)", cfg.WidthGuardN)
		} else if id == "adaptive_guard" {
			tag = fmt.Sprintf(" (%d-%d)", mode.MinN, mode.MaxN)
		} else if mode.DiagnosticOnly {
			tag = " (diag)"
		}

		line := fmt.Sprintf(" %s %s %s%s", keyBadge, stateTag, mode.Title, tag)
		line = loom.TruncateText(line, r.W-2, "")
		c.Write(r.X+1, row, line, normal)
		c.Write(r.X+6, row, stateTag, stateStyle)
		row++
	}
}

func (a *App) drawStressPanel(c *loom.Canvas, r loom.Rect, cfg loom.ResizeConfig, normal, header loom.Style) {
	if r.W < 4 || r.H < 2 {
		return
	}
	border := loom.BoxBorder{
		TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘",
		Horizontal: "─", Vertical: "│",
	}
	drawBoxBorder(c, r, border, "Stress Surface & Geometry", normal, header)

	row := r.Y + 1
	if row < r.Y+r.H-1 {
		guardState := "idle"
		if a.pane != nil && a.pane.WidthGuardActive() {
			guardState = "ACTIVE (burst)"
		} else if !cfg.WidthGuard {
			guardState = "OFF"
		}
		dimText := fmt.Sprintf(" Term: %dx%d | Canvas: %dx%d | Guard: %s", c.Cols(), c.Rows(), r.W, r.H, guardState)
		c.Write(r.X+1, row, loom.TruncateText(dimText, r.W-2, ""), normal)
		row++
	}
	if row < r.Y+r.H-1 && a.pane != nil {
		// The [B] mode only shows the explicit alt-screen request; auto full
		// screen may have switched to it as well.
		screen := [...]string{"inline", "primary full screen", "alternate screen"}[a.pane.Screen()]
		if a.pane.Screen() == loom.ScreenAlt && !cfg.AltScreen {
			screen += " (auto)"
		}
		c.Write(r.X+1, row, loom.TruncateText(" Screen: "+screen+" | Auto full screen: "+onOff(cfg.AutoFullscreen), r.W-2, ""), normal)
		row++
	}
	if row < r.Y+r.H-1 {
		c.Write(r.X+1, row, loom.TruncateText(" "+a.guardSummary(cfg), r.W-2, ""), normal)
		row++
	}

	if row < r.Y+r.H-1 {
		motion := "ON"
		if a.pane != nil && a.pane.ReduceMotion {
			motion = "REDUCED"
		}
		info := fmt.Sprintf(" Motion [M]: %s | Theme [T]: %s | Guard [+/-]: %d", motion, a.themeName, cfg.WidthGuardN)
		c.Write(r.X+1, row, loom.TruncateText(info, r.W-2, ""), normal)
		row++
	}

	// Alignment ruler
	if row < r.Y+r.H-1 {
		var ruler strings.Builder
		ruler.WriteString(" L|")
		for x := 0; x < r.W-8; x++ {
			if x%10 == 0 {
				ruler.WriteString(fmt.Sprintf("%d", (x/10)%10))
			} else if x%5 == 0 {
				ruler.WriteString("+")
			} else {
				ruler.WriteString("·")
			}
		}
		ruler.WriteString("|R")
		c.Write(r.X+1, row, loom.TruncateText(ruler.String(), r.W-2, ""), normal)
		row++
	}

	// Checkerboard / stripe pattern to verify clearing on resize
	for ; row < r.Y+r.H-1; row++ {
		var pattern strings.Builder
		pattern.WriteString(" ")
		for x := 0; x < r.W-4; x++ {
			if (x+row)%2 == 0 {
				pattern.WriteString("░")
			} else {
				pattern.WriteString("▒")
			}
		}
		c.Write(r.X+1, row, loom.TruncateText(pattern.String(), r.W-2, ""), loom.Style{Dim: true})
	}
}

// HandleKey handles keyboard inputs for mode toggling, themes, and exiting.
func (a *App) HandleKey(e loom.KeyEvent) bool {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	keyLower := strings.ToLower(key)

	switch keyLower {
	case "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b":
		for _, id := range loom.SpeccedResizeModeIDs {
			if mode, ok := loom.SpeccedResizeModes.Modes[id]; ok && mode.Key == keyLower {
				a.toggleMode(id)
				return false
			}
		}
	case "+", "=", "]":
		a.adjustGuardN(1)
		return false
	case "-", "_", "[":
		a.adjustGuardN(-1)
		return false
	case "r":
		a.resetDefaults()
		return false
	case "t":
		a.cycleTheme()
		return false
	case "m":
		a.toggleReduceMotion()
		return false
	case "q", "f10":
		a.quitting = true
		return true
	}
	return false
}

// guardSummary shows the effective guard width next to the manual n and the
// measured WINCH speed, so adaptive and manual behavior can be told apart.
func (a *App) guardSummary(cfg loom.ResizeConfig) string {
	source, eff, rate := "manual", cfg.WidthGuardN, 0.0
	if a.pane != nil {
		eff, rate = a.pane.EffectiveGuardN(), a.pane.WinchRate()
		if cfg.AdaptiveGuard && eff != cfg.WidthGuardN {
			source = "adaptive"
		}
	}
	return fmt.Sprintf("n=%d %s | manual n=%d | WINCH: %.1f/s", eff, source, cfg.WidthGuardN, rate)
}

func (a *App) adjustGuardN(delta int) {
	cfg := a.Config()
	n := cfg.WidthGuardN + delta
	if n < 1 {
		n = 1
	}
	if n > 20 {
		n = 20
	}
	cfg.WidthGuardN = n
	a.SetConfig(cfg)
}

// HandleMouse handles click interactions.
func (a *App) HandleMouse(_ loom.MouseEvent) bool {
	return false
}

func (a *App) toggleMode(id string) {
	if a.pane != nil {
		a.pane.ToggleResizeMode(id)
		a.config = a.pane.ResizeConfig
	} else {
		a.config.Toggle(id)
	}
}

func (a *App) resetDefaults() {
	if a.pane != nil {
		a.pane.ResetResizeModes()
		a.config = a.pane.ResizeConfig
	} else {
		a.config.Reset()
	}
}

func (a *App) toggleReduceMotion() {
	if a.pane != nil {
		a.pane.ReduceMotion = !a.pane.ReduceMotion
	}
}

func (a *App) cycleTheme() {
	if len(a.themeNames) == 0 {
		return
	}
	a.themeIdx = (a.themeIdx + 1) % len(a.themeNames)
	a.themeName = a.themeNames[a.themeIdx]
	if th, ok := loom.SpeccedThemes[a.themeName]; ok {
		a.theme = th
	}
}

func availableThemes() []string {
	names := make([]string, 0, len(loom.SpeccedThemes))
	for name := range loom.SpeccedThemes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func resolveTheme(name string) (loom.ThemeColors, error) {
	if theme, ok := loom.SpeccedThemes[name]; ok {
		return theme, nil
	}
	names := availableThemes()
	return loom.ThemeColors{}, fmt.Errorf("winch: unknown theme %q (available: %s)", name, strings.Join(names, ", "))
}

// Run parses arguments and runs the Winch resize diagnostics application.
func Run(args []string) error {
	flags := flag.NewFlagSet("winch", flag.ContinueOnError)
	themeName := flags.String("theme", "plain", "color theme")
	if err := flags.Parse(args); err != nil {
		return err
	}
	theme, err := resolveTheme(*themeName)
	if err != nil {
		return err
	}

	pane, err := loom.New(1 << 16)
	if err != nil {
		return err
	}
	defer pane.Close()

	pane.Resizeable = true
	pane.MaxCols = 0
	pane.DisableDefaultQuit = true
	pane.Background = loom.NewAstraBackground()
	pane.EnableMouseClicks()

	app, err := NewApp(pane, *themeName, theme)
	if err != nil {
		return err
	}
	return pane.Run(app)
}

func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}
