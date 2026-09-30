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
		"NumberInput": {Key: "right"}, "Paginator": {Key: "pgdown"}, "Popup": {Key: "esc"},
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
		"DatePicker": {locate: func([]string) (int, int) { return 6, 4 }, state: func(w loom.Widget) any { return *w.(*loom.DatePicker).Value }},
		"FilePicker": {locate: func([]string) (int, int) { return 3, 2 }, state: func(w loom.Widget) any { entry, _ := w.(*loom.FilePicker).Selected(); return entry.Name }},
		"Form":       {locate: func([]string) (int, int) { return 2, 2 }, state: func(w loom.Widget) any { return w.(*loom.Form).FocusIndex() }},
		"MenuBar":    {locate: func([]string) (int, int) { return 1, 0 }, state: func(w loom.Widget) any { return w.(*loom.MenuBar).Open }},
		"Paginator":  {locate: func([]string) (int, int) { return 6, 0 }, state: func(w loom.Widget) any { return w.(*loom.Paginator).Page }},
		"Tabs":       {locate: func(rows []string) (int, int) { return runeColumn(rows[1], "Details"), 1 }, state: func(w loom.Widget) any { return w.(*loom.Tabs).Focus() }},
		"Tree":       {locate: func([]string) (int, int) { return 0, 0 }, state: func(w loom.Widget) any { return len(w.(*loom.Tree).VisibleNodes()) }},
	}
	for name, test := range clicks {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			widget, err := New(name)
			if err != nil {
				t.Fatal(err)
			}
			beforeRows := loom.Render(widget, 100, 30)
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
