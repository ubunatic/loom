// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package screens is a demo of the three pane layouts (inline, full screen,
// alternate screen) and of quasi-fullscreen detection: an inline pane that is
// grown, or whose terminal is shrunk, to nearly the terminal height switches
// to full screen by itself.
package screens

import (
	"flag"
	"fmt"

	"codeberg.org/ubunatic/loom"
)

const minHeight = 3

// App is the demo widget. pane is nil in headless tests, where cfg and height
// stand in for the pane's state.
type App struct {
	pane   *loom.Pane
	cfg    loom.ResizeConfig
	height int // wanted inline height
}

// NewApp returns an App wanting height inline rows, starting inline.
func NewApp(pane *loom.Pane, cfg loom.ResizeConfig, height int) *App {
	a := &App{pane: pane, cfg: cfg, height: max(minHeight, height)}
	a.cfg.FullScreenBuffer, a.cfg.AltScreen = false, false
	a.apply()
	return a
}

// Config returns the resize configuration the app drives.
func (a *App) Config() loom.ResizeConfig { return a.cfg }

// apply pushes cfg and the wanted height to the pane.
func (a *App) apply() {
	if a.pane == nil {
		return
	}
	a.pane.ResizeConfig = a.cfg
	a.pane.Resize(a.height)
}

func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func screenName(m loom.ScreenMode) string {
	return [...]string{"inline", "full screen", "alternate screen"}[m]
}

// lines describes the current state; termRows is 0 when unknown.
func (a *App) lines(termCols, termRows int) []string {
	mode := loom.ScreenInline
	if a.pane != nil {
		mode = a.pane.Screen()
	}
	quasi := a.cfg.QuasiFullscreen(a.height, termRows)
	note := ""
	if quasi {
		note = " (auto)"
	}
	return []string{
		"Screens demo: inline -> full screen -> alternate screen",
		fmt.Sprintf("Screen: %s%s   terminal %dx%d   wanted height %d", screenName(mode), note, termCols, termRows, a.height),
		fmt.Sprintf("Auto full screen: %s   margin_rows=%d min_percent=%d alt=%s   quasi-full now: %v",
			onOff(a.cfg.AutoFullscreen), a.cfg.FullMarginRows, a.cfg.FullMinPercent, onOff(a.cfg.FullAlt), quasi),
		"",
		"f full screen   b alternate screen   c auto full screen   l auto uses alt",
		"+/- inline height   m/M margin rows   p/P min percent   q quit",
	}
}

// Draw renders the state into r.
func (a *App) Draw(c *loom.Canvas, r loom.Rect) {
	termCols, termRows, _ := loom.TerminalSize()
	for i, line := range a.lines(termCols, termRows) {
		if i >= r.H {
			break
		}
		c.Write(r.X, r.Y+i, line, loom.Style{})
	}
}

// HandleKey handles the demo keys.
func (a *App) HandleKey(e loom.KeyEvent) bool {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	switch key {
	case "f":
		full := a.cfg.FullScreenBuffer && !a.cfg.AltScreen
		a.cfg.FullScreenBuffer, a.cfg.AltScreen = !full, false
	case "b":
		a.cfg.AltScreen = !a.cfg.AltScreen
	case "c":
		a.cfg.AutoFullscreen = !a.cfg.AutoFullscreen
	case "l":
		a.cfg.FullAlt = !a.cfg.FullAlt
	case "+", "=":
		a.height++
	case "-", "_":
		a.height = max(minHeight, a.height-1)
	case "m":
		a.cfg.FullMarginRows++
	case "M":
		a.cfg.FullMarginRows = max(0, a.cfg.FullMarginRows-1)
	case "p":
		a.cfg.FullMinPercent = min(100, a.cfg.FullMinPercent+10)
	case "P":
		a.cfg.FullMinPercent = max(0, a.cfg.FullMinPercent-10)
	case "q", "Q":
		return true
	default:
		return false
	}
	a.apply()
	return false
}

// HandleMouse ignores the mouse.
func (a *App) HandleMouse(loom.MouseEvent) bool { return false }

// Run parses flags and runs the demo.
func Run(args []string) error {
	flags := flag.NewFlagSet("screens", flag.ContinueOnError)
	height := flags.Int("height", 8, "inline height in rows")
	auto := flags.Bool("auto", true, "switch to full screen when the pane is nearly full height")
	margin := flags.Int("margin", -1, "rows short of the terminal height that still count as full (-1: spec default)")
	percent := flags.Int("percent", -1, "percent of the terminal height that counts as full, 0 = off (-1: spec default)")
	autoAlt := flags.Bool("auto-alt", true, "use the alternate screen for the automatic full screen")
	start := flags.String("screen", "inline", "start layout: inline, full or alt")
	if err := flags.Parse(args); err != nil {
		return err
	}
	pane, err := loom.New(*height)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	pane.DisableDefaultQuit = true
	pane.MaxCols = 0

	cfg := pane.ResizeConfig
	cfg.AutoFullscreen, cfg.FullAlt = *auto, *autoAlt
	if *margin >= 0 {
		cfg.FullMarginRows = *margin
	}
	if *percent >= 0 {
		cfg.FullMinPercent = *percent
	}
	app := NewApp(pane, cfg, *height)
	switch *start {
	case "inline":
	case "full":
		app.cfg.FullScreenBuffer = true
	case "alt":
		app.cfg.AltScreen = true
	default:
		return fmt.Errorf("screens: unknown -screen %q (inline, full, alt)", *start)
	}
	app.apply()
	return pane.Run(app)
}
