// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package screens is a demo of the pane layouts and of quasi-fullscreen
// detection. Full screen means the alternate screen, so switching between it
// and inline keeps the scrollback clean; full screen on the primary screen is
// kept as a demo-only layout. An inline pane that is grown, or whose terminal
// is shrunk, to nearly the terminal height switches to the alternate screen by
// itself. The app draws a border so its extent is visible.
package screens

import (
	"fmt"
	"sort"
	"strings"

	"codeberg.org/ubunatic/loom"
	"github.com/spf13/cobra"
)

const (
	minHeight = 3
	minWidth  = 30
	stepWidth = 10
)

// seg is a piece of a row: plain text, or a button when key is set. Clicking
// a button acts like pressing key.
type seg struct {
	text string
	key  string
}

// button is a drawn seg; rect is in canvas coordinates.
type button struct {
	rect loom.Rect
	key  string
}

// App is the demo widget. pane is nil in headless tests, where cfg and height
// stand in for the pane's state.
type App struct {
	pane   *loom.Pane
	cfg    loom.ResizeConfig
	height int // wanted inline height
	width  int // wanted pane width, capped by the terminal width

	themeNames []string
	themeIdx   int
	astra      bool
	buttons    []button
}

// NewApp returns an App wanting height inline rows, starting inline.
func NewApp(pane *loom.Pane, cfg loom.ResizeConfig, height int) *App {
	a := &App{pane: pane, cfg: cfg, height: max(minHeight, height), width: 60, themeNames: themeNames()}
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
	a.pane.MaxCols = a.width
	if a.astra {
		a.pane.Background = loom.NewAstraBackground()
	} else {
		a.pane.Background = nil
	}
	a.pane.Resize(a.height)
}

func themeNames() []string {
	names := make([]string, 0, len(loom.SpeccedThemes))
	for name := range loom.SpeccedThemes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// SetTheme selects a spec theme by name.
func (a *App) SetTheme(name string) error {
	for i, n := range a.themeNames {
		if n == name {
			a.themeIdx = i
			return nil
		}
	}
	return fmt.Errorf("screens: unknown theme %q (available: %s)", name, strings.Join(a.themeNames, ", "))
}

func (a *App) themeName() string {
	if len(a.themeNames) == 0 {
		return ""
	}
	return a.themeNames[a.themeIdx]
}

func (a *App) theme() loom.ThemeColors { return loom.SpeccedThemes[a.themeName()] }

// resizeWidth changes the wanted width by delta, within minWidth and the
// terminal width (maxWidth <= 0 means unknown, uncapped).
func (a *App) resizeWidth(delta, maxWidth int) {
	a.width = max(minWidth, a.width+delta)
	if maxWidth > 0 {
		a.width = min(a.width, max(minWidth, maxWidth))
	}
}

func (a *App) termWidth() int {
	cols, _, _ := loom.TerminalSize()
	return cols
}

func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func screenName(m loom.ScreenMode) string {
	return [...]string{"inline", "primary full screen", "full screen (alt)"}[m]
}

// detected reports what each enabled detector says about the terminal.
func (a *App) detected(termCols, termRows int) (byWidth, byHeight bool) {
	w, h := a.cfg, a.cfg
	w.FullByHeight, h.FullByWidth = false, false
	return w.QuasiFullscreen(a.width, a.height, termCols, termRows), h.QuasiFullscreen(a.width, a.height, termCols, termRows)
}

// rows describes the UI in groups: state, size, look, auto full screen (with
// one row per detector), and the remaining keys.
func (a *App) rows(termCols, termRows int) [][]seg {
	mode := loom.ScreenInline
	if a.pane != nil {
		mode = a.pane.Screen()
	}
	byW, byH := a.detected(termCols, termRows)
	note := ""
	if byW || byH {
		note = " (auto)"
	}
	cfg := a.cfg
	stepper := func(label string, down, up string, value string) []seg {
		return []seg{{text: label}, {"[-]", down}, {text: " " + value + " "}, {"[+]", up}}
	}
	join := func(parts ...[]seg) []seg {
		var out []seg
		for _, p := range parts {
			out = append(out, p...)
		}
		return out
	}
	toggle := func(label, key string, on bool) []seg {
		return []seg{{text: label}, {"[" + onOff(on) + "]", key}}
	}
	return [][]seg{
		{{text: fmt.Sprintf("Screen: %s%s   terminal %dx%d", screenName(mode), note, termCols, termRows)}},
		join(stepper("Width ", "W", "w", fmt.Sprintf("%d/%d", min(a.width, termCols), termCols)),
			stepper("   Height ", "-", "+", fmt.Sprint(a.height))),
		join([]seg{{text: "Theme "}, {"[" + a.themeName() + "]", "t"}}, toggle("   Astra ", "a", a.astra)),
		join(toggle("Auto full screen ", "c", cfg.AutoFullscreen), toggle("   uses alt ", "l", cfg.FullAlt)),
		join(toggle("  by width  ", "x", cfg.FullByWidth), stepper("  margin cols ", "K", "k", fmt.Sprint(cfg.FullMarginCols)),
			[]seg{{text: yesNo(byW)}}),
		join(toggle("  by height ", "y", cfg.FullByHeight), stepper("  margin rows ", "M", "m", fmt.Sprint(cfg.FullMarginRows)),
			[]seg{{text: yesNo(byH)}}),
		join(stepper("  height percent ", "P", "p", fmt.Sprint(cfg.FullMinPercent))),
		{{text: "f full screen (alt)  n primary full (demo)  q quit"}},
	}
}

func yesNo(v bool) string {
	if v {
		return " -> full"
	}
	return ""
}

// Draw renders the state into r.
func (a *App) Draw(c *loom.Canvas, r loom.Rect) {
	termCols, termRows, _ := loom.TerminalSize()
	th := a.theme()
	normal := loom.Style{FG: th.NormalFG.Color(), BG: th.NormalBG.Color()}
	border := loom.Style{FG: th.BorderFG.Color(), BG: th.BorderBG.Color()}
	btnStyle := loom.Style{FG: th.SelectedFG.Color(), BG: th.SelectedBG.Color(), Bold: true}
	// Paint the whole pane in the theme as a surface: Astra stars still shine
	// on it wherever no text is written.
	c.PaintSurface(r, normal)
	inner := drawBorder(c, r, border)

	a.buttons = a.buttons[:0]
	for i, row := range a.rows(termCols, termRows) {
		if i >= inner.H {
			break
		}
		x, y := inner.X, inner.Y+i
		for _, sg := range row {
			if sg.key == "" {
				writeWords(c, x, y, sg.text, normal)
				x += loom.StringWidth(sg.text)
				continue
			}
			w := c.Write(x, y, sg.text, btnStyle)
			a.buttons = append(a.buttons, button{rect: loom.Rect{X: x, Y: y, W: w, H: 1}, key: sg.key})
			x += w
		}
	}
}

// writeWords writes text but leaves its spaces unwritten, so they keep the
// surface and remain open to the background animation.
func writeWords(c *loom.Canvas, x, y int, text string, style loom.Style) {
	col := x
	for _, word := range strings.SplitAfter(text, " ") {
		trimmed := strings.TrimRight(word, " ")
		if trimmed != "" {
			c.Write(col, y, trimmed, style)
		}
		col += loom.StringWidth(word)
	}
}

// drawBorder outlines r and returns the area inside it. Areas too small for a
// border are returned unchanged.
func drawBorder(c *loom.Canvas, r loom.Rect, style loom.Style) loom.Rect {
	if r.W < 3 || r.H < 3 {
		return r
	}
	horiz := strings.Repeat("─", r.W-2)
	c.Write(r.X, r.Y, "┌"+horiz+"┐", style)
	c.Write(r.X, r.Y+r.H-1, "└"+horiz+"┘", style)
	for y := r.Y + 1; y < r.Y+r.H-1; y++ {
		c.Write(r.X, y, "│", style)
		c.Write(r.X+r.W-1, y, "│", style)
	}
	return loom.Rect{X: r.X + 1, Y: r.Y + 1, W: r.W - 2, H: r.H - 2}
}

// HandleKey handles the demo keys.
func (a *App) HandleKey(e loom.KeyEvent) bool {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	switch key {
	case "f":
		a.cfg.FullScreenBuffer, a.cfg.AltScreen = false, !a.cfg.AltScreen
	case "n":
		primaryFull := a.cfg.FullScreenBuffer && !a.cfg.AltScreen
		a.cfg.FullScreenBuffer, a.cfg.AltScreen = !primaryFull, false
	case "c":
		a.cfg.AutoFullscreen = !a.cfg.AutoFullscreen
	case "l":
		a.cfg.FullAlt = !a.cfg.FullAlt
	case "+", "=":
		a.height++
	case "-", "_":
		a.height = max(minHeight, a.height-1)
	case "w":
		a.resizeWidth(stepWidth, a.termWidth())
	case "W":
		a.resizeWidth(-stepWidth, a.termWidth())
	case "t":
		a.themeIdx = (a.themeIdx + 1) % len(a.themeNames)
	case "a":
		a.astra = !a.astra
	case "x":
		a.cfg.FullByWidth = !a.cfg.FullByWidth
	case "y":
		a.cfg.FullByHeight = !a.cfg.FullByHeight
	case "k":
		a.cfg.FullMarginCols++
	case "K":
		a.cfg.FullMarginCols = max(0, a.cfg.FullMarginCols-1)
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

// HandleMouse triggers the clicked button. Coordinates are pane-relative and
// 1-based, so canvas coordinates are one less.
func (a *App) HandleMouse(e loom.MouseEvent) bool {
	if e.Action != loom.MousePress || e.Button != loom.MouseLeft {
		return false
	}
	x, y := e.X-1, e.Y-1
	for _, b := range a.buttons {
		if y != b.rect.Y || x < b.rect.X || x >= b.rect.X+b.rect.W {
			continue
		}
		return a.HandleKey(loom.KeyEvent{Key: b.key})
	}
	return false
}

type options struct {
	height, width, margin, marginCols, percent int
	auto, autoAlt, astra, byWidth, byHeight    bool
	start, theme                               string
}

// Run parses args with cobra and runs the demo.
func Run(args []string) error {
	var o options
	cmd := &cobra.Command{
		Use:           "screens",
		Short:         "Inline TUI that switches to full screen (alternate screen) and back",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          func(*cobra.Command, []string) error { return o.run() },
	}
	f := cmd.Flags()
	f.IntVar(&o.height, "height", 12, "inline height in rows")
	f.IntVar(&o.width, "width", 60, "pane width in columns, capped by the terminal width")
	spec := loom.DefaultResizeConfig()
	f.BoolVar(&o.auto, "auto", true, "switch to full screen when the pane is nearly as big as the terminal")
	f.BoolVar(&o.byWidth, "by-width", spec.FullByWidth, "detect full screen by width")
	f.BoolVar(&o.byHeight, "by-height", spec.FullByHeight, "detect full screen by height")
	f.IntVar(&o.marginCols, "margin-cols", -1, "columns short of the terminal width that still count as full (-1: spec default)")
	f.IntVar(&o.margin, "margin", -1, "rows short of the terminal height that still count as full (-1: spec default)")
	f.IntVar(&o.percent, "percent", -1, "percent of the terminal height that counts as full, 0 = off (-1: spec default)")
	f.BoolVar(&o.autoAlt, "auto-alt", true, "automatic full screen uses the alternate screen (keeps scrollback clean)")
	f.BoolVar(&o.astra, "astra", true, "draw the Astra star field background")
	f.StringVar(&o.theme, "theme", "plain", "color theme ("+strings.Join(themeNames(), ", ")+")")
	f.StringVar(&o.start, "screen", "inline", "start layout: inline, alt (full screen) or full (primary screen, demo)")
	cmd.SetArgs(args)
	return cmd.Execute()
}

func (o options) run() error {
	pane, err := loom.New(o.height)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	pane.DisableDefaultQuit = true
	pane.EnableMouseClicks()

	cfg := pane.ResizeConfig
	cfg.AutoFullscreen, cfg.FullAlt = o.auto, o.autoAlt
	cfg.FullByWidth, cfg.FullByHeight = o.byWidth, o.byHeight
	if o.marginCols >= 0 {
		cfg.FullMarginCols = o.marginCols
	}
	if o.margin >= 0 {
		cfg.FullMarginRows = o.margin
	}
	if o.percent >= 0 {
		cfg.FullMinPercent = o.percent
	}
	app := NewApp(pane, cfg, o.height)
	app.width, app.astra = max(minWidth, o.width), o.astra
	if err := app.SetTheme(o.theme); err != nil {
		return err
	}
	switch o.start {
	case "inline":
	case "alt":
		app.cfg.AltScreen = true
	case "full":
		app.cfg.FullScreenBuffer = true
	default:
		return fmt.Errorf("screens: unknown --screen %q (inline, alt, full)", o.start)
	}
	app.apply()
	return pane.Run(app)
}
