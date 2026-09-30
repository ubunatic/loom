// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package gallery

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/internal/ptytest"
	"codeberg.org/ubunatic/loom/spec"
	"gopkg.in/yaml.v3"
)

func TestDemosMatchCatalogAndRender(t *testing.T) {
	var catalog struct {
		Widgets []struct {
			Name string `yaml:"name"`
		} `yaml:"widgets"`
	}
	data, err := spec.WidgetsYAML()
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		t.Fatal(err)
	}
	listed := map[string]bool{}
	for _, entry := range catalog.Widgets {
		listed[entry.Name] = true
	}
	for _, name := range Names() {
		qualified := "loom." + name
		if name == "Media" {
			qualified = "media.Widget"
		}
		if !listed[qualified] {
			t.Errorf("gallery name %q is missing from spec/widgets.yaml", qualified)
		}
		widget, err := New(name)
		if err != nil {
			t.Errorf("New(%q): %v", name, err)
			continue
		}
		rows := loom.Render(widget, 80, 24)
		if strings.TrimSpace(strings.Join(rows, "")) == "" {
			t.Errorf("%q rendered an empty frame", name)
		}
	}
}

func TestNewUnknownDemo(t *testing.T) {
	if _, err := New("Missing"); err == nil {
		t.Fatal("New accepted an unknown demo")
	}
}

func TestGalleryTabsCycleDemos(t *testing.T) {
	tabs := NewAll()
	if len(tabs.Tabs) < 2 {
		t.Fatal("gallery needs at least two demos")
	}
	loom.Render(tabs, 100, 30)
	initial := tabs.Focus()
	tabs.ConsumeKey(loom.KeyEvent{Key: "tab"})
	if got := tabs.Focus(); got != (initial+1)%len(tabs.Tabs) {
		t.Fatalf("Tab focus = %d, want %d", got, (initial+1)%len(tabs.Tabs))
	}
	tabs.ConsumeKey(loom.KeyEvent{Key: "shift-tab"})
	if got := tabs.Focus(); got != initial {
		t.Fatalf("Shift-Tab focus = %d, want %d", got, initial)
	}
}

func TestEveryDemoRespondsToRepresentativeKey(t *testing.T) {
	// These demos intentionally present static content and have no input action.
	displayOnly := map[string]bool{
		"Chart": true, "KeyHelp": true, "PillCluster": true,
		"ProgressBar": true, "Spinner": true, "Stopwatch": true, "Timer": true,
	}
	keys := map[string]loom.KeyEvent{
		"Choice": {Key: "down"}, "Media": {Key: "+", Text: "+"}, "DatePicker": {Key: "right"}, "Dialog": {Key: "tab"},
		"FilePicker": {Key: "down"}, "Form": {Key: "tab"}, "MenuBar": {Key: "down"},
		"NumberInput": {Key: "right"}, "Paginator": {Key: "pgdown"}, "PaintCanvas": {Text: "c"}, "Popup": {Key: "esc"},
		"Table": {Key: "down"}, "Tabs": {Key: "tab"}, "TextArea": {Text: "x"},
		"TextInput": {Text: "x"}, "Toggle": {Key: "enter"}, "Tree": {Key: "down"},
		"Viewport": {Key: "down"},
	}
	for _, name := range Names() {
		name := name
		t.Run(name, func(t *testing.T) {
			widget, err := New(name)
			if err != nil {
				t.Fatal(err)
			}
			if displayOnly[name] {
				return
			}
			event, ok := keys[name]
			if !ok {
				t.Fatalf("demo has no representative key or display-only classification")
			}
			rows := 30
			if name == "Viewport" {
				event = loom.KeyEvent{Key: "pgdown"}
				rows = 10
			}
			before := renderDemoCells(widget, rows)
			loom.DispatchKeyEvent(widget, event)
			after := renderDemoCells(widget, rows)
			if reflect.DeepEqual(before, after) {
				t.Fatalf("key %#v did not change %s render or state", event, name)
			}
		})
	}
}

func renderDemoCells(widget loom.Widget, rows int) [][]loom.Cell {
	canvas := loom.NewCanvas(100, rows)
	widget.Draw(canvas, loom.Rect{W: 100, H: rows})
	cells := make([][]loom.Cell, rows)
	for y := range cells {
		cells[y] = make([]loom.Cell, 100)
		for x := range cells[y] {
			cells[y][x] = canvas.Get(x, y)
		}
	}
	return cells
}

func TestMouseDrivenDemosRespondToClick(t *testing.T) {
	type clickCase struct {
		locate func([]string) (int, int)
		state  func(loom.Widget) any
	}
	runeColumn := func(line, text string) int {
		index := strings.Index(line, text)
		if index < 0 {
			return -1
		}
		return utf8.RuneCountInString(line[:index])
	}
	clicks := map[string]clickCase{
		"Choice":     {locate: func([]string) (int, int) { return 3, 1 }, state: func(w loom.Widget) any { return w.(*loom.Choice).FilteredSel() }},
		"DatePicker": {locate: func([]string) (int, int) { return 18, 4 }, state: func(w loom.Widget) any { return *w.(*loom.DatePicker).Value }},
		// Row 1 is the parent entry; row 2 is already selected on construction.
		"FilePicker": {locate: func([]string) (int, int) { return 3, 1 }, state: func(w loom.Widget) any { entry, _ := w.(*loom.FilePicker).Selected(); return entry.Name }},
		"Form":       {locate: func([]string) (int, int) { return 2, 2 }, state: func(w loom.Widget) any { return w.(*loom.Form).FocusIndex() }},
		"MenuBar":    {locate: func([]string) (int, int) { return 1, 0 }, state: func(w loom.Widget) any { return w.(*loom.MenuBar).Open }},
		"Paginator":  {locate: func([]string) (int, int) { return 6, 0 }, state: func(w loom.Widget) any { return w.(*loom.Paginator).Page }},
		"Tabs":       {locate: func(rows []string) (int, int) { return runeColumn(rows[1], "Details"), 1 }, state: func(w loom.Widget) any { return w.(*loom.Tabs).Focus() }},
		"Tree":       {locate: func([]string) (int, int) { return 0, 0 }, state: func(w loom.Widget) any { return len(w.(*loom.Tree).VisibleNodes()) }},
		"Toggle":     {locate: func([]string) (int, int) { return 1, 0 }, state: func(w loom.Widget) any { return w.(*loom.Toggle).String() }},
		"Dialog": {locate: func(rows []string) (int, int) {
			for y, line := range rows {
				if !strings.Contains(line, "Discard") {
					continue
				}
				if x := runeColumn(line, "Save"); x >= 0 {
					return x, y
				}
			}
			return -1, -1
		}, state: func(w loom.Widget) any { return w.(*dialogDemo).dialog.Open }},
	}
	for name, test := range clicks {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			widget, err := New(name)
			if err != nil {
				t.Fatal(err)
			}
			// Render returns ANSI rows; hit-testing needs visible cell columns.
			cells := renderDemoCells(widget, 30)
			beforeRows := make([]string, len(cells))
			for y, row := range cells {
				for _, cell := range row {
					beforeRows[y] += cell.Text
				}
			}
			x, y := test.locate(beforeRows)
			if x < 0 {
				t.Fatalf("click target for %s is not visible", name)
			}
			before := test.state(widget)
			loom.DispatchMouseEvent(widget, loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: x, Y: y})
			after := test.state(widget)
			if before == after {
				t.Fatalf("click at (%d,%d) did not change %s state", x, y, name)
			}
		})
	}
}

func TestWidgetsPTYClickTabAndTreeDisclosure(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	build := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/cmd/loom")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build loom binary: %v\n%s", err, output)
	}
	const cols, rows = 100, 30
	s := ptytest.Start(t, cols, rows, bin, "widgets", "--show")
	clickText := func(text string) {
		t.Helper()
		for y, line := range s.Screen() {
			if i := strings.Index(line, text); i >= 0 {
				x := utf8.RuneCountInString(line[:i])
				s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM", x+1, y+1)))
				s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dm", x+1, y+1)))
				return
			}
		}
		t.Fatalf("%q not visible on PTY screen:\n%s", text, strings.Join(s.Screen(), "\n"))
	}
	s.WaitFor("Tree", 5*time.Second)
	clickText("Tree")
	s.WaitFor("app.go", 5*time.Second)
	// Click the disclosure glyph immediately before src; collapsing the node
	// removes app.go from the rendered child panel.
	for y, line := range s.Screen() {
		if i := strings.Index(line, "src"); i >= 0 {
			x := utf8.RuneCountInString(line[:i]) - 1
			s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM", x+1, y+1)))
			s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dm", x+1, y+1)))
			break
		}
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if !strings.Contains(strings.Join(s.Screen(), "\n"), "app.go") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("click inside Tree did not collapse src:\n%s", strings.Join(s.Screen(), "\n"))
}

func TestWidgetsPTYSizeAndF2ThemePropagation(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	build := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/cmd/loom")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build loom binary: %v\n%s", err, output)
	}
	themes := loom.ThemeNames()
	if len(themes) < 2 {
		t.Fatal("PTY theme cycle requires at least two themes")
	}
	const cols, rows, width, height = 100, 35, 48, 14
	initial, next := themes[0], themes[1]
	s := ptytest.Start(t, cols, rows, bin, "widgets", "--show", "--theme", initial, "-W", "48", "-H", "14", "Dialog", "Choice")
	s.WaitFor("Save changes", 5*time.Second)
	s.WaitFor("Theme: "+initial, 5*time.Second)

	findText := func(text string) (int, int) {
		t.Helper()
		for y, line := range s.Screen() {
			if i := strings.Index(line, text); i >= 0 {
				return utf8.RuneCountInString(line[:i]), y
			}
		}
		t.Fatalf("%q missing from PTY screen:\n%s", text, strings.Join(s.Screen(), "\n"))
		return 0, 0
	}
	findRune := func(want rune) (int, int) {
		t.Helper()
		for y, line := range s.Screen() {
			for x, got := range []rune(line) {
				if got == want {
					return x, y
				}
			}
		}
		t.Fatalf("%q missing from PTY screen:\n%s", want, strings.Join(s.Screen(), "\n"))
		return 0, 0
	}
	_, footerY := findText("Theme: " + initial)
	if wantY := rows - 24 + height - 1; footerY != wantY {
		t.Fatalf("gallery footer row = %d, want %d for -H %d:\n%s", footerY, wantY, height, strings.Join(s.Screen(), "\n"))
	}
	if got := s.Cell(width-1, footerY-2).Style.BG; !ptyColorMatches(got, loom.Theme(initial).NormalBG.Color()) {
		t.Fatalf("-W %d did not paint its last column with the gallery background: got %v", width, got)
	}
	if got := s.Cell(width, footerY-2).Style.BG; ptyColorMatches(got, loom.Theme(initial).NormalBG.Color()) {
		t.Fatalf("gallery background extended past -W %d into column %d", width, width)
	}
	if ptyColorMatches(s.Cell(width-1, footerY+1).Style.BG, loom.Theme(initial).NormalBG.Color()) {
		t.Fatalf("gallery background extended past -H %d into row %d", height, footerY+1)
	}

	// Capture the active Choice demo before cycling, then return to the modal
	// demo so one F2 redraw can be checked across the tab bar, modal frame, and
	// gallery background.
	s.Send("\t")
	s.WaitFor("filebrowser-widget", 5*time.Second)
	demoX, demoY := findText("filebrowser-widget")
	oldDemoStyle := s.Cell(demoX, demoY).Style
	s.Send("\x1b[Z")
	s.WaitFor("Save changes", 5*time.Second)
	tabX, tabY := findText("Dialog")
	frameX, frameY := findRune('┌')
	oldTabStyle := s.Cell(tabX, tabY).Style
	oldFrameStyle := s.Cell(frameX, frameY).Style
	oldBackgroundStyle := s.Cell(width-1, footerY-2).Style

	s.SendRaw([]byte("\x1b[12~"))
	s.WaitFor("Theme: "+next, 5*time.Second)
	if got := s.Cell(tabX, tabY).Style; got == oldTabStyle || !ptyColorMatches(got.FG, loom.Theme(next).HeaderFG.Color()) {
		t.Fatalf("tab bar cell did not recolor after F2: %v", got)
	}
	if got := s.Cell(frameX, frameY).Style; got == oldFrameStyle || !ptyColorMatches(got.BG, loom.Theme(next).NormalBG.Color()) {
		t.Fatalf("modal frame cell did not recolor after F2: %v", got)
	}
	if got := s.Cell(width-1, footerY-2).Style; got == oldBackgroundStyle || !ptyColorMatches(got.BG, loom.Theme(next).NormalBG.Color()) {
		t.Fatalf("gallery background did not recolor after F2: got %v", got)
	}
	s.Send("\t")
	s.WaitFor("filebrowser-widget", 5*time.Second)
	demoX, demoY = findText("filebrowser-widget")
	if got := s.Cell(demoX, demoY).Style; got == oldDemoStyle || !ptyColorMatches(got.FG, loom.Theme(next).NormalFG.Color()) {
		t.Fatalf("active demo cell did not use the new theme after F2: got %v", got)
	}
}

func TestRicherDemosPTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	build := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/cmd/loom")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build loom binary: %v\n%s", err, output)
	}
	checks := []struct {
		name, visible string
		animated      bool
	}{
		{"ProgressBar", "16/24 files", true}, {"Spinner", "Syncing workspace", true},
		{"Stopwatch", "00:", true}, {"Timer", "04:", true}, {"NumberInput", "7.5", false},
		{"TextInput", "Ada Lovelace", false}, {"Toggle", "Notifications", false},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", check.name)
			s.WaitFor(check.visible, 5*time.Second)
			if check.name == "TextInput" {
				s.WaitFor("Placeholder:", 5*time.Second)
				s.WaitFor("Masked:", 5*time.Second)
			}
			before := strings.Join(s.Screen(), "\n")
			if check.animated {
				deadline := time.Now().Add(3 * time.Second)
				for time.Now().Before(deadline) {
					time.Sleep(30 * time.Millisecond)
					if after := strings.Join(s.Screen(), "\n"); after != before {
						return
					}
				}
				t.Fatalf("%s did not render a changed frame over time:\n%s", check.name, before)
			}
		})
	}
	t.Run("NumberInputFixedWidth", func(t *testing.T) {
		s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "NumberInput")
		s.WaitFor("◂ 7.5 ▸", 5*time.Second)
		before := strings.Join(s.Screen(), "\n")
		start := strings.Index(before, "◂ 7.5 ▸")
		if start < 0 {
			t.Fatal("initial number input missing")
		}
		s.Send("\x1b[C")
		s.WaitFor("◂ 8", 5*time.Second)
		after := strings.Join(s.Screen(), "\n")
		left, right := strings.Index(after, "◂ 8"), strings.Index(after, "▸")
		if left < 0 || right < 0 || utf8.RuneCountInString(after[left:right]) != utf8.RuneCountInString("◂ 7.5 ") {
			t.Fatalf("fixed-width NumberInput changed its footprint after stepping:\n%s", after)
		}
	})
	t.Run("StopwatchControls", func(t *testing.T) {
		s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "Stopwatch")
		s.WaitFor("00:", 5*time.Second)
		s.Send(" ")
		paused := strings.Join(s.Screen(), "\n")
		time.Sleep(1100 * time.Millisecond)
		if strings.Join(s.Screen(), "\n") != paused {
			t.Fatal("Space did not pause the stopwatch")
		}
		s.Send("r")
		if !strings.Contains(strings.Join(s.Screen(), "\n"), "00:00") {
			t.Fatal("R did not reset the stopwatch")
		}
		s.Send(" ")
		time.Sleep(1100 * time.Millisecond)
		if strings.Contains(strings.Join(s.Screen(), "\n"), "00:00") {
			t.Fatal("Space did not restart the reset stopwatch")
		}
	})
}

func ptyColorMatches(got ptytest.Color, want loom.Color) bool {
	gr, gg, gb, gok := got.RGB()
	wr, wg, wb, wok := want.RGB()
	return gok && wok && gr == wr && gg == wg && gb == wb
}

func TestPaintCanvasPTYMouseDragDrawsBrailleLine(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	build := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/cmd/loom")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build loom binary: %v\n%s", err, output)
	}
	s := ptytest.Start(t, 80, 24, bin, "widgets", "--show", "PaintCanvas")
	s.WaitFor("Theme:", 5*time.Second)
	s.Send("c")
	clearDeadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(clearDeadline) {
		if !galleryHasBrailleGlyph(strings.Join(s.Screen(), "\n")) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if galleryHasBrailleGlyph(strings.Join(s.Screen(), "\n")) {
		t.Fatalf("clear key did not erase the gallery sample stroke:\n%s", strings.Join(s.Screen(), "\n"))
	}
	// SGR mouse coordinates are 1-based on the wire; normal dispatch converts
	// them to 0-based, child-local widget coordinates.
	s.SendRaw([]byte("\x1b[<0;12;10M"))
	s.SendRaw([]byte("\x1b[<32;28;17M"))
	s.SendRaw([]byte("\x1b[<0;28;17m"))
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		rows := s.Screen()
		if galleryBrailleAt(rows, 11, 9) && galleryBrailleAt(rows, 19, 12) && galleryBrailleAt(rows, 27, 16) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("mouse drag did not draw braille on the PaintCanvas PTY:\n%s", strings.Join(s.Screen(), "\n"))
}

func galleryBrailleAt(rows []string, x, y int) bool {
	r := galleryRuneAt(rows, x, y)
	return r >= 0x2800 && r <= 0x28ff
}

func galleryRuneAt(rows []string, x, y int) rune {
	if y < 0 || y >= len(rows) {
		return 0
	}
	for column, r := range []rune(rows[y]) {
		if column == x {
			return r
		}
	}
	return 0
}

func galleryHasBrailleGlyph(text string) bool {
	for _, r := range text {
		if r >= 0x2800 && r <= 0x28ff {
			return true
		}
	}
	return false
}

func TestWidgetKeyRoutingPTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	build := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/cmd/loom")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build loom binary: %v\n%s", err, output)
	}
	start := func(t *testing.T, name string) *ptytest.Session {
		t.Helper()
		s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", name)
		s.WaitFor("Theme: plain", 5*time.Second)
		return s
	}
	waitForAbsent := func(t *testing.T, s *ptytest.Session, text string) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if !strings.Contains(strings.Join(s.Screen(), "\n"), text) {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatalf("%q remained visible after close:\n%s", text, strings.Join(s.Screen(), "\n"))
	}

	t.Run("TextArea accepts typing", func(t *testing.T) {
		s := start(t, "TextArea")
		s.WaitFor("A multi-line editor", 5*time.Second)
		s.Send("Z")
		// TextArea places its initial caret at the end of the sample value.
		s.WaitFor("Use the arrow keys to move.Z", 5*time.Second)
	})

	t.Run("TextInput accepts typing", func(t *testing.T) {
		s := start(t, "TextInput")
		s.WaitFor("Name:", 5*time.Second)
		countBullets := func() int { return strings.Count(strings.Join(s.Screen(), "\n"), "•") }
		before := countBullets()
		if before == 0 {
			t.Fatal("TextInput did not render its masked sample value")
		}
		s.Send("Z")
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) && countBullets() <= before {
			time.Sleep(10 * time.Millisecond)
		}
		if after := countBullets(); after <= before {
			t.Fatalf("typing did not extend the masked value: before=%d after=%d", before, after)
		}
	})

	t.Run("Popup closes and reopens", func(t *testing.T) {
		s := start(t, "Popup")
		s.WaitFor("Gallery popup", 5*time.Second)
		s.Send("\x1b")
		waitForAbsent(t, s, "Gallery popup")
		s.Send("\r")
		s.WaitFor("Gallery popup", 5*time.Second)
	})

	t.Run("Dialog closes and reopens", func(t *testing.T) {
		s := start(t, "Dialog")
		s.WaitFor("Save changes", 5*time.Second)
		s.Send("\r")
		waitForAbsent(t, s, "Save changes")
		s.Send("\r")
		s.WaitFor("Save changes", 5*time.Second)
	})
}

func TestGalleryMouseRoutingPTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	if output, err := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	start := func(t *testing.T, name string) *ptytest.Session {
		t.Helper()
		// Nest demos in outer tabs so their draw origins are nonzero.
		s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", name, "Choice")
		s.WaitFor("Theme: plain", 5*time.Second)
		return s
	}
	point := func(t *testing.T, s *ptytest.Session, text string) (int, int) {
		t.Helper()
		for y, line := range s.Screen() {
			if i := strings.Index(line, text); i >= 0 {
				return utf8.RuneCountInString(line[:i]), y
			}
		}
		t.Fatalf("%q missing:\n%s", text, strings.Join(s.Screen(), "\n"))
		return -1, -1
	}
	click := func(t *testing.T, s *ptytest.Session, text string) {
		t.Helper()
		x, y := point(t, s, text)
		s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", x+1, y+1, x+1, y+1)))
	}
	waitAbsent := func(t *testing.T, s *ptytest.Session, text string) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if !strings.Contains(strings.Join(s.Screen(), "\n"), text) {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatalf("%q remained visible:\n%s", text, strings.Join(s.Screen(), "\n"))
	}
	t.Run("MenuBar title columns", func(t *testing.T) {
		s := start(t, "MenuBar")
		s.WaitFor("Autosave", 5*time.Second)
		for _, menu := range []struct{ title, item string }{{"Edit", "Undo"}, {"Help", "Keyboard shortcuts"}, {"File", "Autosave"}} {
			// Click every letter, including the formerly misrouted e in Help.
			for offset := range []rune(menu.title) {
				s.Send("\x1b")
				waitAbsent(t, s, "Autosave")
				waitAbsent(t, s, "Undo")
				waitAbsent(t, s, "Keyboard shortcuts")
				x, y := point(t, s, menu.title)
				s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", x+offset+1, y+1, x+offset+1, y+1)))
				s.WaitFor(menu.item, 5*time.Second)
			}
		}
	})
	t.Run("Tabs demo switches both ways", func(t *testing.T) {
		s := start(t, "Tabs")
		s.WaitFor("Loom widget gallery", 5*time.Second)
		click(t, s, "Details")
		s.WaitFor("Tabs host any Loom widgets.", 5*time.Second)
		click(t, s, "Overview")
		s.WaitFor("Loom widget gallery", 5*time.Second)
	})
	t.Run("Form hover preserves focus and click changes it", func(t *testing.T) {
		s := start(t, "Form")
		s.WaitFor("Engineer", 5*time.Second)
		x, y := point(t, s, "Engineer")
		// A terminal read can contain a mouse report followed by a key.
		s.SendRaw([]byte(fmt.Sprintf("\x1b[<35;%d;%dMZ", x+1, y+1)))
		s.WaitFor("Ada LovelaceZ", 5*time.Second)
		click(t, s, "Engineer")
		s.Send("Y")
		s.WaitFor("EngineerY", 5*time.Second)
	})
	t.Run("Dialog clicks button rather than title", func(t *testing.T) {
		s := start(t, "Dialog")
		s.WaitFor("Discard", 5*time.Second)
		// Discard uniquely identifies the button row; Save also occurs in title.
		for y, line := range s.Screen() {
			if !strings.Contains(line, "Discard") {
				continue
			}
			i := strings.Index(line, "Save")
			if i < 0 {
				t.Fatal("Save button missing")
			}
			x := utf8.RuneCountInString(line[:i])
			s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", x+1, y+1, x+1, y+1)))
			waitAbsent(t, s, "Discard")
			s.Send("\r")
			s.WaitFor("Discard", 5*time.Second)
			return
		}
		t.Fatal("button row missing")
	})
}
