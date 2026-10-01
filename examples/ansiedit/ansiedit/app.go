// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansiedit

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"ubunatic.com/loom"
	"ubunatic.com/loom/measure"
)

// FocusArea denotes which panel currently possesses user keyboard focus.
type FocusArea int

const (
	FocusCanvas FocusArea = iota
	FocusSidePanel
)

func (f FocusArea) String() string {
	switch f {
	case FocusSidePanel:
		return "SidePanel"
	default:
		return "Editor"
	}
}

// SidePanelMode determines the active view in the left side panel.
type SidePanelMode int

const (
	PanelInfo SidePanelMode = iota
	PanelPalette
	PanelKeys
	PanelTheme
)

func (m SidePanelMode) String() string {
	switch m {
	case PanelPalette:
		return "Palette"
	case PanelKeys:
		return "Keybindings"
	case PanelTheme:
		return "Themes"
	default:
		return "Info"
	}
}

var paletteGrid = [7][6]uint8{
	{52, 88, 124, 160, 196, 203},   // Reds
	{58, 94, 130, 166, 208, 215},   // Oranges / Browns / Amber
	{22, 28, 34, 40, 46, 120},      // Greens
	{23, 29, 36, 44, 51, 123},      // Cyans / Teals
	{17, 18, 19, 20, 21, 75},       // Blues
	{53, 89, 126, 163, 201, 207},   // Purples / Magentas
	{232, 236, 240, 244, 248, 254}, // Grayscale
}

// AnsiEditApp is the root ANSI editor application widget.
type AnsiEditApp struct {
	buffer      *Buffer
	activeFocus FocusArea
	panelMode   SidePanelMode
	editMode    EditMode

	// Cursor position in buffer
	cursorX int
	cursorY int

	// Viewport scrolling for canvas
	scrollX int
	scrollY int

	// Active styling for new characters
	activeFG        loom.Color
	activeBG        loom.Color
	activeBold      bool
	activeDim       bool
	activeUnderline bool
	activeInvert    bool

	// Palette cursor
	palCol int
	palRow int

	// Theme selection
	themeNames   []string
	themeIndex   int
	currentTheme string
	theme        loom.ThemeColors

	// Status and state
	statusMsg string
	quit      bool
}

// NewAnsiEditApp creates a new AnsiEditApp instance for a given buffer.
func NewAnsiEditApp(buf *Buffer) *AnsiEditApp {
	if buf == nil {
		buf = NewBuffer(54, 22)
	}

	var themeNames []string
	for k := range loom.SpeccedThemes {
		themeNames = append(themeNames, k)
	}
	sort.Strings(themeNames)

	curTheme := "julia256"
	tIdx := 0
	for i, name := range themeNames {
		if name == curTheme {
			tIdx = i
			break
		}
	}

	app := &AnsiEditApp{
		buffer:       buf,
		activeFocus:  FocusCanvas,
		panelMode:    PanelPalette,
		editMode:     ModeOvertype,
		activeFG:     loom.ColorIndex(130),
		activeBG:     loom.ColorIndex(232),
		themeNames:   themeNames,
		themeIndex:   tIdx,
		currentTheme: curTheme,
		theme:        loom.Theme(curTheme),
		palRow:       1, // #130 row
		palCol:       2, // #130 col
	}

	return app
}

// Buffer returns the underlying buffer.
func (app *AnsiEditApp) Buffer() *Buffer {
	return app.buffer
}

// Quit reports whether the application has received a quit signal.
func (app *AnsiEditApp) Quit() bool {
	return app.quit
}

// Draw renders the full 80x24 AnsiEdit frame and all subcomponents.
func (app *AnsiEditApp) Draw(c *loom.Canvas, r loom.Rect) {
	if r.W < 10 || r.H < 5 {
		return
	}

	boxStyle := app.theme.BoxStyle()
	borderStyle := boxStyle.Border
	if borderStyle.FG == loom.ColorReset() {
		borderStyle.FG = loom.ColorIndex(36) // cyan border default
	}

	sidePanelWidth := 22
	if r.W < 40 {
		sidePanelWidth = max(10, r.W/3)
	}

	splitCol := r.X + 1 + sidePanelWidth // X position of vertical split bar
	rightWidth := r.W - sidePanelWidth - 3

	contentH := r.H - 2 // space between top border and bottom border/footer
	if contentH < 1 {
		contentH = 1
	}

	// 1. Draw frame borders
	app.drawFrameBorders(c, r, splitCol, borderStyle)

	// 2. Draw Side Panel interior
	sideRect := loom.Rect{
		X: r.X + 1,
		Y: r.Y + 1,
		W: sidePanelWidth,
		H: contentH,
	}
	app.drawSidePanel(c, sideRect)

	// 3. Draw Canvas interior
	canvasRect := loom.Rect{
		X: splitCol + 1,
		Y: r.Y + 1,
		W: max(1, rightWidth),
		H: contentH,
	}
	app.drawCanvas(c, canvasRect)

	// 4. Draw Footer Status Bar (on bottom row)
	footerY := r.Y + r.H - 1
	if footerY < r.Y+r.H {
		app.drawFooter(c, loom.Rect{X: r.X, Y: footerY, W: r.W, H: 1})
	}
}

// drawFrameBorders draws the top, split, and bottom borders with header title & status badges.
func (app *AnsiEditApp) drawFrameBorders(c *loom.Canvas, r loom.Rect, splitCol int, borderStyle loom.Style) {
	// Top border
	c.Write(r.X, r.Y, "┌─", borderStyle)
	c.Write(r.X+2, r.Y, " Loom AnsiEdit ", loom.Style{Bold: true, FG: loom.ColorIndex(15)})

	for x := r.X + 17; x < splitCol; x++ {
		c.Write(x, r.Y, "─", borderStyle)
	}
	c.Write(splitCol, r.Y, "┬─", borderStyle)

	// Header badges: [filename ●] ─── [Focus: Editor] ─────── [24x54] ───┐
	filename := filepath.Base(app.buffer.Path())
	if filename == "" || filename == "." {
		filename = "untitled.ansi"
	}
	dirtyMark := ""
	if app.buffer.Modified() {
		dirtyMark = " ●"
	}

	badgeCol := splitCol + 2
	c.Write(badgeCol, r.Y, "[", loom.Style{FG: loom.ColorIndex(240)})
	c.Write(badgeCol+1, r.Y, filename, loom.Style{Bold: true, FG: loom.ColorIndex(36)})
	if app.buffer.Modified() {
		c.Write(badgeCol+1+len(filename), r.Y, " ●", loom.Style{Bold: true, FG: loom.ColorIndex(3)})
	}
	badgeCol += len(filename) + len(dirtyMark) + 1
	c.Write(badgeCol, r.Y, "] ─── [", loom.Style{FG: loom.ColorIndex(240)})
	badgeCol += 7
	c.Write(badgeCol, r.Y, "Focus: ", loom.Style{FG: loom.ColorIndex(7)})
	badgeCol += 7
	c.Write(badgeCol, r.Y, app.activeFocus.String(), loom.Style{Bold: true, FG: loom.ColorIndex(11)})
	badgeCol += len(app.activeFocus.String())
	c.Write(badgeCol, r.Y, "] ─── [", loom.Style{FG: loom.ColorIndex(240)})
	badgeCol += 7
	dims := fmt.Sprintf("%dx%d", app.buffer.Rows(), app.buffer.Cols())
	c.Write(badgeCol, r.Y, dims, loom.Style{Dim: true, FG: loom.ColorIndex(7)})
	badgeCol += len(dims)
	c.Write(badgeCol, r.Y, "] ", loom.Style{FG: loom.ColorIndex(240)})
	badgeCol += 2

	for x := badgeCol; x < r.X+r.W-1; x++ {
		c.Write(x, r.Y, "─", borderStyle)
	}
	c.Write(r.X+r.W-1, r.Y, "┐", borderStyle)

	// Side and split vertical borders
	contentH := r.H - 2
	for y := r.Y + 1; y <= r.Y+contentH; y++ {
		c.Write(r.X, y, "│", borderStyle)
		c.Write(splitCol, y, "│", borderStyle)
		c.Write(r.X+r.W-1, y, "│", borderStyle)
	}

	// Bottom border
	bottomY := r.Y + contentH + 1
	c.Write(r.X, bottomY, "└", borderStyle)
	for x := r.X + 1; x < splitCol; x++ {
		c.Write(x, bottomY, "─", borderStyle)
	}
	c.Write(splitCol, bottomY, "┴", borderStyle)
	for x := splitCol + 1; x < r.X+r.W-1; x++ {
		c.Write(x, bottomY, "─", borderStyle)
	}
	c.Write(r.X+r.W-1, bottomY, "┘", borderStyle)
}

// drawSidePanel renders the currently active left panel view.
func (app *AnsiEditApp) drawSidePanel(c *loom.Canvas, r loom.Rect) {
	switch app.panelMode {
	case PanelInfo:
		app.drawInfoPanel(c, r)
	case PanelPalette:
		app.drawPalettePanel(c, r)
	case PanelKeys:
		app.drawKeysPanel(c, r)
	case PanelTheme:
		app.drawThemePanel(c, r)
	}
}

// drawInfoPanel renders the F1 File Info and Cursor Inspector view.
func (app *AnsiEditApp) drawInfoPanel(c *loom.Canvas, r loom.Rect) {
	y := r.Y

	// Header
	c.Write(r.X+1, y, "ℹ️  File Info", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+15, y, "[F1]", loom.Style{Dim: true, FG: loom.ColorIndex(244)})
	y++

	c.Write(r.X+1, y, "────────────────────", loom.Style{Dim: true, FG: loom.ColorIndex(240)})
	y++

	// File info
	filename := filepath.Base(app.buffer.Path())
	if filename == "" || filename == "." {
		filename = "untitled.ansi"
	}
	c.Write(r.X+1, y, "Name: ", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+7, y, filename, loom.Style{FG: loom.ColorIndex(7)})
	y++

	size := len(app.buffer.Serialize())
	c.Write(r.X+1, y, "Size: ", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+7, y, fmt.Sprintf("%d bytes", size), loom.Style{FG: loom.ColorIndex(7)})
	y++

	c.Write(r.X+1, y, "Grid: ", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+7, y, fmt.Sprintf("%d x %d", app.buffer.Rows(), app.buffer.Cols()), loom.Style{FG: loom.ColorIndex(7)})
	y++

	c.Write(r.X+1, y, "Mode: ", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+7, y, app.editMode.String(), loom.Style{FG: loom.ColorIndex(7)})
	y += 2

	// Cursor inspector
	c.Write(r.X+1, y, "Cursor Inspector:", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	y++

	c.Write(r.X+1, y, "Pos:  ", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+7, y, fmt.Sprintf("X: %02d   Y: %02d", app.cursorX, app.cursorY), loom.Style{FG: loom.ColorIndex(7)})
	y++

	cell := app.buffer.Get(app.cursorX, app.cursorY)
	rDisplay := cell.Rune
	if rDisplay == 0 || rDisplay == ' ' {
		rDisplay = ' '
	}
	c.Write(r.X+1, y, "Cell: ", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+7, y, fmt.Sprintf("'%c' (U+%04X)", rDisplay, rDisplay), loom.Style{FG: loom.ColorIndex(7)})
	y++

	c.Write(r.X+1, y, "FG:   ", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+7, y, formatColorDesc(cell.FG), loom.Style{FG: cell.FG})
	y++

	c.Write(r.X+1, y, "BG:   ", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+7, y, formatColorDesc(cell.BG), loom.Style{FG: loom.ColorIndex(7)})
	y++

	attrStr := formatAttributes(cell)
	c.Write(r.X+1, y, "Attr: ", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+7, y, attrStr, loom.Style{FG: loom.ColorIndex(7)})
	y += 2

	// Panel Switcher
	c.Write(r.X+1, y, "Panel Switcher:", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	y++

	items := []struct {
		key  string
		name string
		mode SidePanelMode
	}{
		{"F1", "Info (Active)", PanelInfo},
		{"F2", "Palette", PanelPalette},
		{"F8", "Keybindings", PanelKeys},
		{"F9", "Themes", PanelTheme},
	}
	for _, it := range items {
		c.Write(r.X+1, y, it.key, loom.Style{Bold: true, FG: loom.ColorIndex(15)})
		style := loom.Style{FG: loom.ColorIndex(7)}
		if it.mode == app.panelMode {
			style = loom.Style{Dim: true, FG: loom.ColorIndex(244)}
		}
		c.Write(r.X+6, y, it.name, style)
		y++
	}
	c.Write(r.X+1, y, "F10", loom.Style{Bold: true, FG: loom.ColorIndex(11)})
	c.Write(r.X+6, y, "Exit Editor", loom.Style{FG: loom.ColorIndex(7)})
}

// drawPalettePanel renders the F2 2D Color Palette matrix view.
func (app *AnsiEditApp) drawPalettePanel(c *loom.Canvas, r loom.Rect) {
	y := r.Y

	// Header
	c.Write(r.X+1, y, "🎨 Palette", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+13, y, "[F2]", loom.Style{Dim: true, FG: loom.ColorIndex(244)})
	y++

	c.Write(r.X+1, y, "Dark ────────► Bright", loom.Style{Dim: true, FG: loom.ColorIndex(244)})
	y++

	// 2D Palette Grid (7 rows x 6 cols)
	for pr := 0; pr < 7; pr++ {
		x := r.X + 2
		for pc := 0; pc < 6; pc++ {
			colorIdx := paletteGrid[pr][pc]
			chipStyle := loom.Style{BG: loom.ColorIndex(colorIdx)}
			text := "   "
			if pr == app.palRow && pc == app.palCol && app.activeFocus == FocusSidePanel {
				text = " • "
				chipStyle.FG = loom.ColorIndex(15)
				chipStyle.Bold = true
			}
			c.Write(x, y, text, chipStyle)
			x += 3
		}
		y++
	}

	c.Write(r.X+1, y, "▲ Hue ▼   ◀ Tone ▶", loom.Style{Dim: true, FG: loom.ColorIndex(244)})
	y += 2

	// Active Color preview
	c.Write(r.X+1, y, "Active Color:", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	y++

	// FG
	c.Write(r.X+1, y, "FG: ", loom.Style{FG: loom.ColorIndex(7)})
	c.Write(r.X+5, y, "   ", loom.Style{BG: app.activeFG})
	c.Write(r.X+9, y, formatColorDesc(app.activeFG), loom.Style{FG: app.activeFG})
	y++

	// BG
	c.Write(r.X+1, y, "BG: ", loom.Style{FG: loom.ColorIndex(7)})
	c.Write(r.X+5, y, "   ", loom.Style{BG: app.activeBG})
	c.Write(r.X+9, y, formatColorDesc(app.activeBG), loom.Style{FG: loom.ColorIndex(244)})
	y += 2

	c.Write(r.X+1, y, "Enter:FG  S-Enter:BG", loom.Style{Dim: true, FG: loom.ColorIndex(244)})
	y += 3

	c.Write(r.X+1, y, "F1:Info  F2:Palette", loom.Style{Dim: true, FG: loom.ColorIndex(244)})
	y++
	c.Write(r.X+1, y, "F8:Keys  F9:Theme", loom.Style{Dim: true, FG: loom.ColorIndex(244)})
	y++
	c.Write(r.X+1, y, "F10 Quit  Tab Focus", loom.Style{Dim: true, FG: loom.ColorIndex(244)})
}

// drawKeysPanel renders the F8 Keybindings reference sheet.
func (app *AnsiEditApp) drawKeysPanel(c *loom.Canvas, r loom.Rect) {
	y := r.Y
	c.Write(r.X+1, y, "⌨️  Keybindings", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+15, y, "[F8]", loom.Style{Dim: true, FG: loom.ColorIndex(244)})
	y++
	c.Write(r.X+1, y, "────────────────────", loom.Style{Dim: true, FG: loom.ColorIndex(240)})
	y++

	c.Write(r.X+1, y, "Editing:", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	y++

	keys := []struct {
		k, desc string
	}{
		{"C-s", "Save buffer"},
		{"C-c", "Copy cell"},
		{"C-x", "Cut cell"},
		{"C-v", "Paste cell"},
		{"Del/BS", "Delete / Erase"},
		{"Ins", "Toggle mode"},
	}
	for _, it := range keys {
		c.Write(r.X+1, y, it.k, loom.Style{Bold: true, FG: loom.ColorIndex(15)})
		c.Write(r.X+8, y, it.desc, loom.Style{FG: loom.ColorIndex(7)})
		y++
	}
	y++

	c.Write(r.X+1, y, "Navigation:", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	y++
	navKeys := []struct {
		k, desc string
	}{
		{"Arrows", "Navigate"},
		{"C-Left", "Word back"},
		{"C-Rgt", "Word forward"},
		{"C-Up", "Object up"},
		{"C-Down", "Object down"},
	}
	for _, it := range navKeys {
		c.Write(r.X+1, y, it.k, loom.Style{Bold: true, FG: loom.ColorIndex(15)})
		c.Write(r.X+8, y, it.desc, loom.Style{FG: loom.ColorIndex(7)})
		y++
	}
	y++

	c.Write(r.X+1, y, "Tab", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+8, y, "Switch focus", loom.Style{FG: loom.ColorIndex(7)})
	y++
	c.Write(r.X+1, y, "F10", loom.Style{Bold: true, FG: loom.ColorIndex(11)})
	c.Write(r.X+8, y, "Universal quit", loom.Style{FG: loom.ColorIndex(7)})
}

// drawThemePanel renders the F9 Theme switcher view.
func (app *AnsiEditApp) drawThemePanel(c *loom.Canvas, r loom.Rect) {
	y := r.Y
	c.Write(r.X+1, y, "🎨 Themes", loom.Style{Bold: true, FG: loom.ColorIndex(15)})
	c.Write(r.X+13, y, "[F9]", loom.Style{Dim: true, FG: loom.ColorIndex(244)})
	y++
	c.Write(r.X+1, y, "────────────────────", loom.Style{Dim: true, FG: loom.ColorIndex(240)})
	y++

	maxVisible := r.H - 4
	startIdx := 0
	if app.themeIndex >= maxVisible {
		startIdx = app.themeIndex - maxVisible + 1
	}

	for i := startIdx; i < len(app.themeNames) && y < r.Y+r.H; i++ {
		name := app.themeNames[i]
		prefix := "  "
		style := loom.Style{FG: loom.ColorIndex(7)}
		if i == app.themeIndex {
			if app.activeFocus == FocusSidePanel {
				prefix = "▸ "
				style = loom.Style{Bold: true, FG: loom.ColorIndex(15), BG: loom.ColorIndex(238)}
			} else {
				prefix = "● "
				style = loom.Style{Bold: true, FG: loom.ColorIndex(36)}
			}
		}
		c.Write(r.X+1, y, prefix+name, style)
		y++
	}
}

// drawCanvas renders the 2D ANSI buffer and the cursor.
func (app *AnsiEditApp) drawCanvas(c *loom.Canvas, r loom.Rect) {
	if app.cursorX < app.scrollX {
		app.scrollX = app.cursorX
	}
	if app.cursorX >= app.scrollX+r.W {
		app.scrollX = app.cursorX - r.W + 1
	}
	if app.cursorY < app.scrollY {
		app.scrollY = app.cursorY
	}
	if app.cursorY >= app.scrollY+r.H {
		app.scrollY = app.cursorY - r.H + 1
	}

	for vy := 0; vy < r.H; vy++ {
		by := app.scrollY + vy
		for vx := 0; vx < r.W; vx++ {
			bx := app.scrollX + vx
			screenX := r.X + vx
			screenY := r.Y + vy

			cell := app.buffer.Get(bx, by)
			lCell := cell.ToLoomCell()

			// Highlight cursor cell on canvas
			if bx == app.cursorX && by == app.cursorY {
				if app.activeFocus == FocusCanvas {
					lCell.Style.BG = loom.ColorIndex(15)
					lCell.Style.FG = loom.ColorIndex(0)
					lCell.Style.Bold = true
					c.CursorX = screenX
					c.CursorY = screenY
				} else {
					lCell.Style.Underline = true
				}
			}

			c.Set(screenX, screenY, lCell)
		}
	}
}

// drawFooter renders the function key shortcuts bar.
func (app *AnsiEditApp) drawFooter(c *loom.Canvas, r loom.Rect) {
	barStyle := loom.Style{BG: loom.ColorIndex(236)}
	c.Fill(r, loom.Cell{Text: " ", Style: barStyle})

	buttons := []struct {
		key    string
		label  string
		active bool
		warn   bool
	}{
		{" F1 ", "Info", app.panelMode == PanelInfo, false},
		{" F2 ", "Pal", app.panelMode == PanelPalette, false},
		{" F8 ", "Keys", app.panelMode == PanelKeys, false},
		{" F9 ", "Theme", app.panelMode == PanelTheme, false},
		{" Tab ", "Focus", false, false},
		{" C-s ", "Save", false, false},
		{" F10 ", "Quit", false, true},
	}

	x := r.X + 2
	for _, b := range buttons {
		btnStyle := loom.Style{Bold: true, FG: loom.ColorIndex(15), BG: loom.ColorIndex(239)}
		if b.active {
			btnStyle.BG = loom.ColorIndex(24) // Blue highlighted
		}
		if b.warn {
			btnStyle.FG = loom.ColorIndex(11) // Yellow/orange quit
		}

		c.Write(x, r.Y, b.key, btnStyle)
		x += len(b.key)

		lblStyle := loom.Style{FG: loom.ColorIndex(7), BG: loom.ColorIndex(236)}
		if b.active {
			lblStyle.FG = loom.ColorIndex(15)
		}
		c.Write(x+1, r.Y, b.label+"  ", lblStyle)
		x += len(b.label) + 3
	}
}

// ConsumeKey processes keyboard events, returning an EventResult value struct.
// Navigation and editing keys are consumed without quitting.
// Explicit exit keys (F10, Ctrl-Q) signal quit.
func (app *AnsiEditApp) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	// Global quit keys
	if e.Is("f10", "ctrl-q") {
		app.quit = true
		return loom.QuitResult()
	}

	// Global keybindings
	switch {
	case e.Is("f1"):
		app.panelMode = PanelInfo
		return loom.Handled()
	case e.Is("f2"):
		app.panelMode = PanelPalette
		return loom.Handled()
	case e.Is("f8"):
		app.panelMode = PanelKeys
		return loom.Handled()
	case e.Is("f9"):
		app.panelMode = PanelTheme
		return loom.Handled()
	case e.Is("tab"):
		if app.activeFocus == FocusCanvas {
			app.activeFocus = FocusSidePanel
		} else {
			app.activeFocus = FocusCanvas
		}
		return loom.Handled()
	case e.Is("ctrl-s"):
		if app.buffer.Path() != "" {
			_ = app.buffer.SaveFile(app.buffer.Path())
			app.statusMsg = "Saved file successfully"
		}
		return loom.Handled()
	}

	// Focus-specific key handling
	if app.activeFocus == FocusSidePanel {
		if app.handleSidePanelKey(e) {
			return loom.Handled()
		}
		return loom.Ignored()
	}
	if app.handleCanvasKey(e) {
		return loom.Handled()
	}
	return loom.Ignored()
}

func (app *AnsiEditApp) handleSidePanelKey(e loom.KeyEvent) bool {
	switch app.panelMode {
	case PanelPalette:
		switch {
		case e.Is("up"):
			if app.palRow > 0 {
				app.palRow--
			}
			return true
		case e.Is("down"):
			if app.palRow < 6 {
				app.palRow++
			}
			return true
		case e.Is("left"):
			if app.palCol > 0 {
				app.palCol--
			}
			return true
		case e.Is("right"):
			if app.palCol < 5 {
				app.palCol++
			}
			return true
		case e.Is("shift-enter", "ctrl-enter", "ctrl-m", "shift-return"):
			selectedColor := paletteGrid[app.palRow][app.palCol]
			app.activeBG = loom.ColorIndex(selectedColor)
			return true
		case e.Is("enter", "return"):
			selectedColor := paletteGrid[app.palRow][app.palCol]
			app.activeFG = loom.ColorIndex(selectedColor)
			return true
		}
	case PanelTheme:
		switch {
		case e.Is("up"):
			if app.themeIndex > 0 {
				app.themeIndex--
				app.applyTheme(app.themeNames[app.themeIndex])
			}
			return true
		case e.Is("down"):
			if app.themeIndex < len(app.themeNames)-1 {
				app.themeIndex++
				app.applyTheme(app.themeNames[app.themeIndex])
			}
			return true
		case e.Is("enter", "return"):
			app.applyTheme(app.themeNames[app.themeIndex])
			return true
		}
	}
	return false
}

func (app *AnsiEditApp) handleCanvasKey(e loom.KeyEvent) bool {
	switch {
	case e.Is("ctrl-left"):
		app.cursorX = app.buffer.PrevWord(app.cursorX, app.cursorY)
		return true
	case e.Is("ctrl-right"):
		app.cursorX = app.buffer.NextWord(app.cursorX, app.cursorY)
		return true
	case e.Is("ctrl-up"):
		app.cursorY = app.buffer.PrevObjectRow(app.cursorY)
		return true
	case e.Is("ctrl-down"):
		app.cursorY = app.buffer.NextObjectRow(app.cursorY)
		return true
	case e.Is("up"):
		if app.cursorY > 0 {
			app.cursorY--
		}
		return true
	case e.Is("down"):
		if app.cursorY < app.buffer.Rows()-1 {
			app.cursorY++
		}
		return true
	case e.Is("left"):
		if app.cursorX > 0 {
			app.cursorX--
		}
		return true
	case e.Is("right"):
		if app.cursorX < app.buffer.Cols()-1 {
			app.cursorX++
		}
		return true
	case e.Is("insert"):
		if app.editMode == ModeOvertype {
			app.editMode = ModeInsert
		} else {
			app.editMode = ModeOvertype
		}
		return true
	case e.Is("delete"):
		app.buffer.Delete(app.cursorX, app.cursorY, app.editMode)
		return true
	case e.Is("backspace"):
		app.cursorX = app.buffer.Backspace(app.cursorX, app.cursorY, app.editMode)
		return true
	case e.Is("ctrl-c"):
		app.buffer.Copy(app.cursorX, app.cursorY)
		return true
	case e.Is("ctrl-x"):
		app.buffer.Cut(app.cursorX, app.cursorY)
		return true
	case e.Is("ctrl-v"):
		app.buffer.Paste(app.cursorX, app.cursorY)
		return true
	case e.Is("enter", "return"):
		if app.cursorY < app.buffer.Rows()-1 {
			app.cursorY++
			app.cursorX = 0
		}
		return true
	default:
		if e.Text != "" {
			rs := []rune(e.Text)
			for _, r := range rs {
				if r >= 32 {
					app.buffer.PutChar(
						app.cursorX, app.cursorY,
						r,
						app.activeFG, app.activeBG,
						app.activeBold, app.activeDim, app.activeUnderline, app.activeInvert,
						app.editMode,
					)
					w := measure.RuneWidth(r)
					if w < 1 {
						w = 1
					}
					if app.cursorX+w < app.buffer.Cols() {
						app.cursorX += w
					}
				}
			}
			return true
		}
	}
	return false
}

// ConsumeMouse handles mouse clicks.
func (app *AnsiEditApp) ConsumeMouse(e loom.MouseEvent) loom.EventResult { return loom.Ignored() }

// PaneRequest declares terminal requirements.
func (app *AnsiEditApp) PaneRequest() loom.PaneRequest {
	return loom.PaneRequest{
		Mouse:      1003,
		Resizeable: true,
		OwnsQuit:   true,
	}
}

func (app *AnsiEditApp) applyTheme(name string) {
	app.currentTheme = name
	app.theme = loom.Theme(name)
}

func formatColorDesc(c loom.Color) string {
	r, g, b, ok := c.RGB()
	if !ok {
		return "None (Default)"
	}
	return fmt.Sprintf("#%d,%d,%d", r, g, b)
}

func formatAttributes(c BufferCell) string {
	var attrs []string
	if c.Bold {
		attrs = append(attrs, "Bold")
	}
	if c.Dim {
		attrs = append(attrs, "Dim")
	}
	if c.Underline {
		attrs = append(attrs, "Underline")
	}
	if c.Invert {
		attrs = append(attrs, "Invert")
	}
	if len(attrs) == 0 {
		return "None"
	}
	return strings.Join(attrs, ", ")
}

// NewWidget builds the ansiedit root widget for in-process hosting.
func NewWidget(args []string) (loom.Widget, error) {
	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	buf, err := LoadBuffer(path, 54, 22)
	if err != nil {
		return nil, err
	}
	return NewAnsiEditApp(buf), nil
}

// Run executes the ansiedit interactive editor command.
func Run(args []string) error {
	cmd := &cobra.Command{
		Use:           "ansiedit [file.ansi]",
		Short:         "Standalone ANSI file and graphic text art editor",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, cmdArgs []string) error {
			path := ""
			if len(cmdArgs) > 0 {
				path = cmdArgs[0]
			}
			buf, err := LoadBuffer(path, 54, 22)
			if err != nil {
				return err
			}
			return runWithBuffer(buf)
		},
	}
	cmd.SetArgs(args)
	return cmd.Execute()
}

func runWithBuffer(buf *Buffer) error {
	app := NewAnsiEditApp(buf)
	pane, err := loom.New(1 << 16)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	pane.DisableDefaultQuit = true
	pane.EnableMouse()
	return pane.Run(app)
}
