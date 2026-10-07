// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"ubunatic.com/loom"
	"ubunatic.com/loom/gallery"
)

func writeFixture(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "asset.ansi")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestMeasureAndEval(t *testing.T) {
	path := writeFixture(t, "\x1b[31m界a\x1b[0m  \ntiny\n")
	var measured bytes.Buffer
	if err := execute([]string{"measure", path}, &measured); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"line 1: 5 columns", "line 2: 4 columns (ragged)", "bounding box: 5 columns x 2 lines", "trailing whitespace"} {
		if !strings.Contains(measured.String(), want) {
			t.Errorf("measure output %q does not contain %q", measured.String(), want)
		}
	}
	if strings.Contains(measured.String(), "boxes:") {
		t.Fatalf("measure unexpectedly includes structural summary: %q", measured.String())
	}

	boxed := writeFixture(t, "┌──┐\n│ x│\n└──┘\n")
	var evaluated bytes.Buffer
	if err := execute([]string{"eval", boxed}, &evaluated); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"boxes: 1", "rows: 3", "columns: 4", "non-blank bounds: x=1..4 y=1..3 (4 x 3)", "ragged rows: 0", "min line width: 4", "max line width: 4"} {
		if !strings.Contains(evaluated.String(), want) {
			t.Errorf("eval output %q does not contain %q", evaluated.String(), want)
		}
	}
}

func TestEvalAnnotatedOutput(t *testing.T) {
	path := writeFixture(t, "┌──┐\n│ x│  \n└──┘\n")
	for _, flag := range []string{"--annotate", "-a"} {
		var out bytes.Buffer
		if err := execute([]string{"eval", flag, path}, &out); err != nil {
			t.Fatalf("eval %s: %v", flag, err)
		}
		got := out.String()
		for _, want := range []string{"┌──┐    <-- 1", "│ x│    <-- 2", "└──┘    <-- 3", "1: line 1: ragged width", "2: line 2: trailing whitespace"} {
			if !strings.Contains(got, want) {
				t.Errorf("eval %s output %q does not contain %q", flag, got, want)
			}
		}
		for _, line := range strings.Split(got, "\n") {
			if strings.HasPrefix(line, "<-- ") {
				t.Errorf("eval %s output contains a standalone callout row: %q", flag, got)
			}
		}
	}
}

func TestCheckBoxAnnotatedOutput(t *testing.T) {
	path := writeFixture(t, "┌──┐\n│x│\n└─┘\n")
	for _, flag := range []string{"--annotate", "-a"} {
		var out bytes.Buffer
		err := execute([]string{"check-box", flag, path}, &out)
		if err == nil {
			t.Fatalf("check-box %s accepted invalid box", flag)
		}
		got := out.String() + err.Error()
		for _, want := range []string{"┌──┐", "│x│   <-- 1", "└─┘", "1: line 2: box width mismatch"} {
			if !strings.Contains(got, want) {
				t.Errorf("check-box %s output %q does not contain %q", flag, got, want)
			}
		}
		for _, line := range strings.Split(got, "\n") {
			if strings.HasPrefix(line, "<-- ") {
				t.Errorf("check-box %s output contains a standalone callout row: %q", flag, got)
			}
		}
	}
}

func TestEvalCountsRegionalIndicatorPairAsTwoColumns(t *testing.T) {
	path := writeFixture(t, "🇩🇪X\n")
	var out bytes.Buffer
	if err := execute([]string{"eval", path}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "non-blank bounds: x=1..3 y=1..1 (3 x 1)") {
		t.Fatalf("eval output = %q, want a three-column non-blank bound", out.String())
	}
}

func TestCheckBox(t *testing.T) {
	valid := writeFixture(t, "┌──┐\n│hi│\n└──┘\n")
	invalid := writeFixture(t, "┌──┐\n│x│\n└─┘\n")
	var out bytes.Buffer
	if err := execute([]string{"check-box", valid}, &out); err != nil {
		t.Fatalf("valid box rejected: %v", err)
	}
	if !strings.Contains(out.String(), "ok") {
		t.Fatalf("missing success output: %q", out.String())
	}
	if err := execute([]string{"check-box", valid, invalid}, &out); err == nil {
		t.Fatal("misaligned box accepted")
	} else if !strings.Contains(err.Error(), invalid) {
		t.Fatalf("multi-file error %q omits invalid filename %q", err, invalid)
	}
}

func TestCLIArityErrors(t *testing.T) {
	for _, args := range [][]string{{"measure"}, {"measure", "a", "b"}, {"eval"}, {"view"}, {"edit"}, {"check-box"}, {"widgets", "a", "b"}} {
		var out bytes.Buffer
		if err := execute(args, &out); err == nil {
			t.Errorf("execute(%q) succeeded, want arity error", args)
		}
	}
}

func TestInfoReportFields(t *testing.T) {
	t.Setenv("LOOMCOLOR", "truecolor")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("TERM_PROGRAM", "TestTerminal")
	var out bytes.Buffer
	if err := writeInfo(&out, 80, 24); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"loom version: " + loom.Version,
		"terminal size: 80 x 24 cells",
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"TERM_PROGRAM=TestTerminal",
		"colour depth: truecolor",
		"true colour: true",
		"Unicode sample width:",
		"graphics protocol:",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("info report %q does not contain %q", out.String(), want)
		}
	}
}

func TestInfoReportWithoutTerminalSize(t *testing.T) {
	var out bytes.Buffer
	if err := writeInfo(&out, 0, 0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "terminal size: unknown (pixels: unknown)") {
		t.Fatalf("info report without terminal size = %q", out.String())
	}
}

func TestWidgetsCommandListsAndSelectsCatalogEntries(t *testing.T) {
	var all bytes.Buffer
	if err := execute([]string{"widgets"}, &all); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(all.String()), "\n")
	if len(lines) == 0 {
		t.Fatal("widgets output is empty")
	}
	for _, want := range []string{"loom.Choice [input]", "loom.Gauge [display]", "loom.AlignBox [layout]", "loom.Router [infra]"} {
		if !strings.Contains(all.String(), want) {
			t.Errorf("widgets output %q does not contain %q", all.String(), want)
		}
	}
	for _, unwanted := range []string{"Example:", "Source:", "Capabilities:", "Ticket:"} {
		if strings.Contains(all.String(), unwanted) {
			t.Errorf("widgets list output unexpectedly contains %q", unwanted)
		}
	}

	// Verify category grouping order (input -> display -> layout -> infra) and alphabetical sorting within groups
	currentCategoryIndex := 0
	categories := []string{"[input]", "[display]", "[layout]", "[infra]"}
	var prevNameInCat string
	for _, line := range lines {
		cat := -1
		for i, c := range categories {
			if strings.Contains(line, c) {
				cat = i
				break
			}
		}
		if cat == -1 {
			t.Errorf("line %q does not contain known category", line)
			continue
		}
		if cat < currentCategoryIndex {
			t.Errorf("category group regression: line %q appeared after category %s", line, categories[currentCategoryIndex])
		}
		name := strings.Split(line, " ")[0]
		if cat > currentCategoryIndex {
			currentCategoryIndex = cat
			prevNameInCat = name
		} else {
			if prevNameInCat != "" && name < prevNameInCat {
				t.Errorf("entries not sorted within category %s: %s after %s", categories[cat], name, prevNameInCat)
			}
			prevNameInCat = name
		}
	}

	var one bytes.Buffer
	if err := execute([]string{"widgets", "Gauge"}, &one); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(one.String(), "loom.Gauge [display]") {
		t.Fatalf("single widget output = %q", one.String())
	}
	if !strings.Contains(one.String(), "Example:") || !strings.Contains(one.String(), "Source: widget_graph.go") || !strings.Contains(one.String(), "Docs: docs/Widgets.md §3") {
		t.Errorf("single widget output %q missing example/source/docs", one.String())
	}

	var settings bytes.Buffer
	if err := execute([]string{"widgets", "Settings"}, &settings); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(settings.String(), "Capabilities: KindBool toggle rows; KindString text rows; KindChoice selection rows; KindNumber bounded numeric rows") {
		t.Errorf("Settings output %q missing capabilities", settings.String())
	}
	if strings.Contains(settings.String(), "Docs:") {
		t.Errorf("Settings output %q unexpectedly contains Docs:", settings.String())
	}

	var textInput bytes.Buffer
	if err := execute([]string{"widgets", "TextInput"}, &textInput); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(textInput.String(), "Docs:") {
		t.Errorf("TextInput output %q unexpectedly has Docs:", textInput.String())
	}
	if !strings.Contains(textInput.String(), "Ticket: issues/173") {
		t.Errorf("TextInput output %q missing ticket issues/173", textInput.String())
	}

	if err := execute([]string{"widgets", "FileOpener"}, &one); err == nil {
		t.Fatal("unknown widget was accepted")
	}
}

func TestWidgetsShowFlagsKeepCatalogAndRunGallery(t *testing.T) {
	var list bytes.Buffer
	if err := execute([]string{"widgets", "--list"}, &list); err != nil {
		t.Fatal(err)
	}
	var legacy bytes.Buffer
	if err := execute([]string{"widgets"}, &legacy); err != nil {
		t.Fatal(err)
	}
	if list.String() != legacy.String() {
		t.Fatalf("--list output differs from existing listing")
	}
	var shown bytes.Buffer
	previous := runWidgetPane
	runWidgetPane = func(widget loom.Widget, _, _ int, altScreen bool) error {
		if altScreen {
			t.Fatal("non-editor demos requested alternate screen")
		}
		return loom.RenderTo(&shown, widget, 80, 24)
	}
	t.Cleanup(func() { runWidgetPane = previous })
	if err := execute([]string{"widgets", "--show", "TextInput", "ProgressBar"}, &shown); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"TextInput", "ProgressBar"} {
		if !strings.Contains(shown.String(), want) {
			t.Errorf("--show output %q missing %q", shown.String(), want)
		}
	}
	if err := execute([]string{"widgets", "--show", "NoSuchWidget"}, &shown); err == nil {
		t.Fatal("--show accepted unknown demo")
	}
}

func TestWidgetsThemeFlagAndF9Cycle(t *testing.T) {
	if len(loom.ThemeNames()) < 2 {
		t.Fatal("theme cycling requires at least two themes")
	}
	previous := runWidgetPane
	var shown loom.Widget
	runWidgetPane = func(widget loom.Widget, width, height int, altScreen bool) error {
		shown = widget
		if altScreen {
			t.Fatal("Chart demo requested alternate screen")
		}
		if width != 0 || height != 0 {
			t.Fatalf("default gallery size = %dx%d, want terminal size", width, height)
		}
		return nil
	}
	t.Cleanup(func() { runWidgetPane = previous })
	if err := execute([]string{"widgets", "--show", "Chart"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	gallery, ok := shown.(*themedGallery)
	if !ok {
		t.Fatalf("shown widget = %T, want *themedGallery", shown)
	}
	if gallery.themeName != "julia256" {
		t.Fatalf("default theme = %q, want julia256", gallery.themeName)
	}
	names := loom.ThemeNames()
	index := slices.Index(names, "julia256")
	if index < 0 {
		t.Fatal("julia256 is missing from the theme registry")
	}
	want := names[(index+1)%len(names)]
	gallery.ConsumeKey(loom.KeyEvent{Key: "f9"})
	if gallery.themeName != want {
		t.Fatalf("F9 from default theme = %q, want %q", gallery.themeName, want)
	}
	if err := execute([]string{"widgets", "--show", "--theme", "plain", "Chart"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	gallery, ok = shown.(*themedGallery)
	if !ok {
		t.Fatalf("shown widget = %T, want *themedGallery", shown)
	}
	if gallery.themeName != "plain" {
		t.Fatalf("explicit theme = %q, want plain", gallery.themeName)
	}
	if err := execute([]string{"widgets", "--show", "--theme", "mc", "Chart"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	gallery, ok = shown.(*themedGallery)
	if !ok {
		t.Fatalf("shown widget = %T, want *themedGallery", shown)
	}
	if gallery.themeName != "mc" {
		t.Fatalf("initial theme = %q, want mc", gallery.themeName)
	}
	child := &themeProbeWidget{}
	gallery = newThemedGallery(child, "mc")
	initialCanvas := loom.NewCanvas(80, 4)
	gallery.Draw(initialCanvas, initialCanvas.Bounds())
	initialColor := initialCanvas.Get(0, 0).Style
	gallery.ConsumeKey(loom.KeyEvent{Key: "f9"})
	names = loom.ThemeNames()
	index = slices.Index(names, "mc")
	want = names[(index+1)%len(names)]
	if gallery.themeName != want || child.theme.NormalFG != loom.Theme(want).NormalFG {
		t.Fatalf("after F9 theme = %q, child theme mismatch; want %q", gallery.themeName, want)
	}
	canvas := loom.NewCanvas(80, 4)
	gallery.Draw(canvas, canvas.Bounds())
	if canvas.Get(0, 0).Style == initialColor {
		t.Fatalf("demo rendered the same color style after switching from mc to %q", want)
	}
	status := canvasPlainRow(canvas, canvas.Rows()-1)
	if !strings.Contains(status, "F9 Theme "+want) {
		t.Fatalf("gallery chrome = %q, want active theme %q", canvasPlainRow(canvas, canvas.Rows()-1), want)
	}
}

func TestWidgetsSelectAlternateScreenFromGalleryMetadata(t *testing.T) {
	previous := runWidgetPane
	t.Cleanup(func() { runWidgetPane = previous })
	for _, tc := range []struct {
		name string
		args []string
		want bool
	}{
		{"editor", []string{"widgets", "--show", "RichTextEdit"}, true},
		{"combined-gallery", []string{"widgets", "--show"}, true},
		{"non-editor", []string{"widgets", "--show", "Chart"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			runWidgetPane = func(_ loom.Widget, _, _ int, altScreen bool) error {
				called = true
				if altScreen != tc.want {
					t.Errorf("altScreen = %v, want %v", altScreen, tc.want)
				}
				return nil
			}
			if err := execute(tc.args, &bytes.Buffer{}); err != nil {
				t.Fatal(err)
			}
			if !called {
				t.Fatal("gallery runner was not called")
			}
		})
	}
}

func TestWidgetsDebugFlag(t *testing.T) {
	previousDebug := loom.Debug
	loom.Debug = false
	t.Cleanup(func() { loom.Debug = previousDebug })
	previousRun := runWidgetPane
	runWidgetPane = func(loom.Widget, int, int, bool) error {
		if !loom.Debug {
			t.Error("--debug did not enable debug outlines while running the gallery")
		}
		return nil
	}
	t.Cleanup(func() { runWidgetPane = previousRun })

	if err := execute([]string{"widgets", "--show", "--debug", "Chart"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if loom.Debug {
		t.Fatal("--debug remained enabled after the gallery exited")
	}
}

func TestWidgetsShowDemoArgs(t *testing.T) {
	previousRun := runWidgetPane
	var shownWidget loom.Widget
	runWidgetPane = func(w loom.Widget, _, _ int, _ bool) error {
		shownWidget = w
		return nil
	}
	t.Cleanup(func() { runWidgetPane = previousRun })

	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.ansi")

	// Valid single widget with demo args after --
	if err := execute([]string{"widgets", "--show", "RichTextEdit", "--", filePath}, &bytes.Buffer{}); err != nil {
		t.Fatalf("widgets --show RichTextEdit -- file: %v", err)
	}
	if shownWidget == nil {
		t.Fatal("gallery runner was not called")
	}

	// Demo args with multiple widgets -> error
	if err := execute([]string{"widgets", "--show", "RichTextEdit", "Chart", "--", filePath}, &bytes.Buffer{}); err == nil {
		t.Fatal("widgets --show with multiple widgets accepted -- args")
	}

	// Demo args with demo that takes none -> error
	if err := execute([]string{"widgets", "--show", "Chart", "--", "arg"}, &bytes.Buffer{}); err == nil {
		t.Fatal("widgets --show Chart accepted -- args")
	}
}

func TestGalleryAppKeys(t *testing.T) {
	child := &keyProbeWidget{}
	g := newThemedGallery(child, "mc")
	type appliedBackground struct {
		background loom.Background
		onRedraw   bool
	}
	var applied []appliedBackground
	g.setBackground = func(b loom.Background, onRedraw bool) {
		applied = append(applied, appliedBackground{background: b, onRedraw: onRedraw})
	}

	if r := g.ConsumeKey(loom.KeyEvent{Key: "f8"}); !r.Consumed || len(applied) != 1 || applied[0].background == nil || applied[0].onRedraw {
		t.Fatalf("first F8 should enable a background, applied=%v", applied)
	}
	canvas := loom.NewCanvas(60, 4)
	g.Draw(canvas, canvas.Bounds())
	if row := canvasPlainRow(canvas, canvas.Rows()-1); !strings.Contains(row, "F8 BG astra") || !strings.Contains(row, "F10 Quit") {
		t.Fatalf("status bar = %q", row)
	}
	for i, want := range []string{"astra (on redraw)", "plain"} {
		g.ConsumeKey(loom.KeyEvent{Key: "f8"})
		canvas = loom.NewCanvas(60, 4)
		g.Draw(canvas, canvas.Bounds())
		if row := canvasPlainRow(canvas, canvas.Rows()-1); !strings.Contains(row, "F8 BG "+want) {
			t.Fatalf("status bar = %q, want mode %q", row, want)
		}
		entry := applied[i+1]
		if want == "astra (on redraw)" && (entry.background == nil || !entry.onRedraw) {
			t.Fatalf("on-redraw state = %+v", entry)
		}
		if want == "plain" && (entry.background != nil || entry.onRedraw) {
			t.Fatalf("plain state = %+v", entry)
		}
	}
	if r := g.ConsumeKey(loom.KeyEvent{Key: "f10"}); !r.Quit {
		t.Fatal("F10 must quit")
	}
	if r := g.ConsumeKey(loom.KeyEvent{Key: "ctrl-q"}); !r.Quit {
		t.Fatal("Ctrl+Q must quit")
	}

	child.consume = true
	if r := g.ConsumeKey(loom.KeyEvent{Key: "q"}); r.Quit || child.keys != 1 {
		t.Fatalf("q consumed by child must not quit (quit=%v keys=%d)", r.Quit, child.keys)
	}
	child.consume = false
	if r := g.ConsumeKey(loom.KeyEvent{Key: "q"}); r.Quit {
		t.Fatal("plain q must remain available to the widget")
	}
	if r := g.ConsumeKey(loom.KeyEvent{Key: "esc"}); r.Quit {
		t.Fatal("Esc must remain available to the widget")
	}
	if r := g.ConsumeKey(loom.KeyEvent{Key: "ctrl-q"}); !r.Quit {
		t.Fatal("Ctrl+Q must quit")
	}
}

func TestRichTextEditGalleryHintsUseLowerRowsAndFitNarrowWidth(t *testing.T) {
	widget, err := gallery.New("RichTextEdit")
	if err != nil {
		t.Fatal(err)
	}
	for _, width := range []int{80, 40} {
		t.Run(fmt.Sprintf("%d columns", width), func(t *testing.T) {
			g := newThemedGallery(widget, "plain")
			canvas := loom.NewCanvas(width, 12)
			g.Draw(canvas, canvas.Bounds())
			hintRow := canvasPlainRow(canvas, 10)
			if !strings.Contains(hintRow, "F1 Help") || !strings.Contains(hintRow, "⌃S Save") {
				t.Fatalf("lower hint row = %q", hintRow)
			}
			if width == 80 && (!strings.Contains(hintRow, "F7 View") || !strings.Contains(hintRow, "⌃⌥S Save as")) {
				t.Fatalf("normal-width hint row hides full shortcuts: %q", hintRow)
			}
			if strings.Contains(canvasPlainRow(canvas, 9), "⌃S") || strings.Contains(canvasPlainRow(canvas, 9), "⌃⌥S") {
				t.Fatalf("file bar repeats save shortcuts: %q", canvasPlainRow(canvas, 9))
			}
			controls := canvasPlainRow(canvas, 11)
			for _, key := range []string{"F8", "F9", "F10"} {
				if !strings.Contains(controls, key) {
					t.Errorf("gallery controls %q omit %s", controls, key)
				}
			}
			for row := 0; row < canvas.Rows(); row++ {
				if got := loom.StringWidth(canvasPlainRow(canvas, row)); got > width {
					t.Errorf("row %d width = %d, exceeds %d: %q", row, got, width, canvasPlainRow(canvas, row))
				}
			}
		})
	}
}

func TestRichTextEditGalleryBoxModeHintAndFallback(t *testing.T) {
	widget, err := gallery.New("RichTextEdit")
	if err != nil {
		t.Fatal(err)
	}
	g := newThemedGallery(widget, "plain")
	canvas := loom.NewCanvas(80, 12)
	if result := g.ConsumeKey(loom.KeyEvent{Key: "f5"}); !result.Consumed {
		t.Fatalf("F5 box toggle result = %+v", result)
	}
	g.Draw(canvas, canvas.Bounds())
	if fileRow := canvasPlainRow(canvas, 9); !strings.Contains(fileRow, "[Box]") || !strings.Contains(fileRow, "Esc") {
		t.Fatalf("box mode file-row hint = %q", fileRow)
	}
	if hintRow := canvasPlainRow(canvas, 10); !strings.Contains(hintRow, "⌃S") || !strings.Contains(hintRow, "⌃⌥S") {
		t.Fatalf("lower hotkey row omits save shortcuts: %q", hintRow)
	}
	g.ConsumeKey(loom.KeyEvent{Key: "esc"})
	g.Draw(canvas, canvas.Bounds())
	if fileRow := canvasPlainRow(canvas, 9); strings.Contains(fileRow, "[Box]") || strings.Contains(fileRow, "F7") {
		t.Fatalf("file row after box mode shows a hint: %q", fileRow)
	}
	if hintRow := canvasPlainRow(canvas, 10); !strings.Contains(hintRow, "F7 View") {
		t.Fatalf("lower hotkey row omits F7 View: %q", hintRow)
	}
}

func canvasPlainRow(canvas *loom.Canvas, row int) string {
	var text strings.Builder
	for x := 0; x < canvas.Cols(); x++ {
		cell := canvas.Get(x, row)
		if cell.Continuation {
			continue
		}
		if cell.Text == "" {
			text.WriteByte(' ')
		} else {
			text.WriteString(cell.Text)
		}
	}
	return strings.Join(strings.Fields(text.String()), " ")
}

func canvasScreenText(canvas *loom.Canvas) string {
	rows := make([]string, canvas.Rows())
	for row := range rows {
		rows[row] = canvasPlainRow(canvas, row)
	}
	return strings.Join(rows, "\n")
}

func TestRichTextEditGalleryF1AndOutsideClickRouting(t *testing.T) {
	widget, err := gallery.New("RichTextEdit")
	if err != nil {
		t.Fatal(err)
	}
	g := newThemedGallery(widget, "plain")
	if result := g.ConsumeKey(loom.KeyEvent{Key: "f1"}); !result.Consumed {
		t.Fatalf("F1 result = %+v, want consumed", result)
	}
	canvas := loom.NewCanvas(80, 14)
	g.Draw(canvas, canvas.Bounds())
	if !strings.Contains(canvasScreenText(canvas), "RichTextEdit Help") {
		t.Fatal("F1 help popup was not rendered by the gallery")
	}
	if result := g.ConsumeKey(loom.KeyEvent{Key: "esc"}); !result.Consumed {
		t.Fatalf("close F1 popup result = %+v", result)
	}
	if result := g.ConsumeKey(loom.KeyEvent{Key: "ctrl-s"}); !result.Consumed {
		t.Fatalf("open Save as result = %+v", result)
	}
	canvas = loom.NewCanvas(80, 12)
	g.Draw(canvas, canvas.Bounds())
	if result := g.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 4, Y: 0}); !result.Consumed {
		t.Fatalf("preview click result = %+v, want consumed by Save as popup", result)
	}
	if result := g.ConsumeKey(loom.KeyEvent{Key: "f1"}); !result.Consumed {
		t.Fatalf("F1 after outside click result = %+v", result)
	}
	canvas = loom.NewCanvas(80, 12)
	g.Draw(canvas, canvas.Bounds())
	if !strings.Contains(canvasScreenText(canvas), "RichTextEdit Help") {
		t.Fatal("preview click did not dismiss Save as before routing F1")
	}
}

type keyProbeWidget struct {
	consume bool
	keys    int
}

func (*keyProbeWidget) Draw(*loom.Canvas, loom.Rect) {}
func (w *keyProbeWidget) ConsumeKey(loom.KeyEvent) loom.EventResult {
	w.keys++
	if w.consume {
		return loom.Handled()
	}
	return loom.Ignored()
}
func (*keyProbeWidget) ConsumeMouse(loom.MouseEvent) loom.EventResult { return loom.Ignored() }

func TestWidgetsSizeFlags(t *testing.T) {
	previous := runWidgetPane
	var gotWidth, gotHeight int
	runWidgetPane = func(_ loom.Widget, width, height int, altScreen bool) error {
		gotWidth, gotHeight = width, height
		if altScreen {
			t.Fatal("Chart demo requested alternate screen")
		}
		return nil
	}
	t.Cleanup(func() { runWidgetPane = previous })
	if err := execute([]string{"widgets", "--show", "-W", "42", "-H", "11", "Chart"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if gotWidth != 42 || gotHeight != 11 {
		t.Fatalf("gallery size = %dx%d, want 42x11", gotWidth, gotHeight)
	}
}

func TestWidgetsGalleryPaneWidthDefaultsToTerminal(t *testing.T) {
	if got := galleryPaneMaxCols(0); got != 0 {
		t.Fatalf("default gallery max columns = %d, want terminal width (0)", got)
	}
	if got := galleryPaneMaxCols(72); got != 72 {
		t.Fatalf("explicit gallery max columns = %d, want 72", got)
	}
}

func TestThemedGalleryQuitKeysLeaveEscToWidget(t *testing.T) {
	g := newThemedGallery(loom.NewRichTextEdit(&loom.RichDocument{Lines: []loom.RichLine{{}}}), "julia256")
	if got := g.ConsumeKey(loom.KeyEvent{Key: "esc"}); got.Quit {
		t.Fatal("Esc quit the gallery instead of reaching its widget")
	}
	for _, key := range []string{"ctrl-q", "f10"} {
		if got := g.ConsumeKey(loom.KeyEvent{Key: key}); !got.Quit {
			t.Errorf("%s did not quit the gallery: %+v", key, got)
		}
	}
	if got := g.ConsumeKey(loom.KeyEvent{Key: "q"}); got.Quit {
		t.Fatal("plain q quit the gallery; only Ctrl+Q and F10 should quit")
	}
}

type themeProbeWidget struct{ theme loom.ThemeColors }

func (w *themeProbeWidget) Draw(c *loom.Canvas, r loom.Rect) {
	c.Write(r.X, r.Y, "X", w.theme.ChoiceStyle().Normal)
}
func (*themeProbeWidget) ConsumeKey(loom.KeyEvent) loom.EventResult     { return loom.Ignored() }
func (*themeProbeWidget) ConsumeMouse(loom.MouseEvent) loom.EventResult { return loom.Ignored() }
func (w *themeProbeWidget) ApplyTheme(theme loom.ThemeColors)           { w.theme = theme }

func TestWidgetsThemeFlagRejectsUnknownTheme(t *testing.T) {
	if err := execute([]string{"widgets", "--show", "--theme", "missing"}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "unknown theme") {
		t.Fatalf("unknown theme error = %v", err)
	}
}

func TestWidgetsCompletions(t *testing.T) {
	command := widgetsCommand()
	want := []string{"Chart", "loom.Chart", "TextInput", "loom.TextInput"}
	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "positional", args: nil},
		{name: "show flag", args: []string{"--show"}},
	} {
		got, directive := command.ValidArgsFunction(command, tc.args, "")
		if directive != cobra.ShellCompDirectiveNoFileComp {
			t.Errorf("%s directive = %v, want no file completion", tc.name, directive)
		}
		for _, name := range want {
			if !slices.Contains(got, name) {
				t.Errorf("%s completions %v omit %q", tc.name, got, name)
			}
		}
		if slices.Contains(got, "FileOpener") || slices.Contains(got, "loom.FileOpener") {
			t.Errorf("%s completions include non-demo widget: %v", tc.name, got)
		}
	}

	flagCompletion, ok := command.GetFlagCompletionFunc("show")
	if !ok {
		t.Fatal("--show has no flag completion function")
	}
	got, directive := flagCompletion(command, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp || !slices.Contains(got, "Chart") || !slices.Contains(got, "loom.Chart") {
		t.Errorf("--show completions = %v, directive %v", got, directive)
	}

	var out bytes.Buffer
	if err := execute([]string{"__complete", "widgets", "--show", ""}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Chart") || !strings.Contains(out.String(), "loom.Chart") {
		t.Errorf("shell completion output = %q", out.String())
	}
}

func TestMissingFilesHaveCleanFilenameErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.ansi")
	for _, command := range []string{"measure", "eval", "check-box", "view"} {
		var out bytes.Buffer
		err := execute([]string{command, missing}, &out)
		if err == nil {
			t.Errorf("%s unexpectedly read missing file", command)
			continue
		}
		message := err.Error()
		if !strings.Contains(message, missing) || !strings.Contains(message, "no such file") {
			t.Errorf("%s error %q lacks filename or cause", command, message)
		}
		if strings.Count(message, missing) != 1 {
			t.Errorf("%s error repeats filename: %q", command, message)
		}
	}
}

func TestCheckBoxReportsEveryFileError(t *testing.T) {
	missingA := filepath.Join(t.TempDir(), "missing-a.ansi")
	missingB := filepath.Join(t.TempDir(), "missing-b.ansi")
	var out bytes.Buffer
	err := execute([]string{"check-box", missingA, missingB}, &out)
	if err == nil {
		t.Fatal("check-box accepted missing files")
	}
	for _, path := range []string{missingA, missingB} {
		if !strings.Contains(err.Error(), path) {
			t.Errorf("multi-file error %q omits %q", err, path)
		}
	}
}

func TestViewRequiresANSI(t *testing.T) {
	var out bytes.Buffer
	if err := execute([]string{"view", "asset.txt"}, &out); err == nil || !strings.Contains(err.Error(), ".ansi") {
		t.Fatalf("view error = %v, want .ansi support message", err)
	}
}

func TestViewPlainTextFlag(t *testing.T) {
	path := writeFixture(t, "\x1b[31mred\x1b[0m  \n┌─┐")
	for _, flag := range []string{"--plain", "-p"} {
		var out bytes.Buffer
		if err := execute([]string{"view", flag, path}, &out); err != nil {
			t.Fatalf("view %s: %v", flag, err)
		}
		if got, want := out.String(), "red\n┌─┐\n"; got != want {
			t.Errorf("view %s output = %q, want %q", flag, got, want)
		}
	}
}

func TestANSIViewScrollAndQuit(t *testing.T) {
	buffer, err := loom.ParseAnsiBuffer("one\ntwo\nthree", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	view := &ansiView{buffer: buffer}
	if view.ConsumeKey(loom.KeyEvent{Key: "down"}).Quit || view.offsetY != 1 {
		t.Fatalf("down key did not scroll: offsetY=%d", view.offsetY)
	}
	if !view.ConsumeKey(loom.KeyEvent{Key: "f10"}).Quit {
		t.Fatal("F10 did not quit the ANSI viewer")
	}
}

func TestANSIViewAcceptsPageDownAliases(t *testing.T) {
	buffer, err := loom.ParseAnsiBuffer("one\ntwo\nthree\nfour\nfive\nsix\nseven\neight", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	keys := []loom.KeyEvent{{Key: "pgdown"}, {Key: "pgdn"}, {Key: "pagedown"}, loom.DecodeKey([]byte("\x1b[6~"))}
	for _, key := range keys {
		view := &ansiView{buffer: buffer}
		view.ConsumeKey(key)
		if view.offsetY == 0 {
			t.Errorf("key %q did not pan down", key.Key)
		}
	}
}

func TestANSIViewPansWideBufferAndClampsOffsets(t *testing.T) {
	buffer, err := loom.ParseAnsiBuffer("abcdefghijklmnopqrst\nABCDEFGHIJKLMNOPQRST\n01234567890123456789", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	view := &ansiView{buffer: buffer, offsetX: 2, offsetY: 1}
	canvas := loom.NewCanvas(6, 4)
	view.Draw(canvas, loom.Rect{X: 1, Y: 1, W: 4, H: 2})
	if got := canvas.Get(1, 1).Text; got != "C" {
		t.Fatalf("first rendered cell after 2D pan = %q, want %q", got, "C")
	}
	if got := canvas.Get(4, 2).Text; got != "5" {
		t.Fatalf("last rendered cell after 2D pan = %q, want %q", got, "5")
	}
	view.offsetX = 0
	view.offsetY = 0
	view.Draw(canvas, loom.Rect{W: 4, H: 2})
	view.ConsumeKey(loom.KeyEvent{Key: "right"})
	view.ConsumeKey(loom.KeyEvent{Key: "right"})
	view.Draw(canvas, loom.Rect{W: 4, H: 2})
	if got := canvas.Get(0, 0).Text; got != "c" {
		t.Fatalf("first rendered cell after horizontal pan = %q, want %q", got, "c")
	}
	view.ConsumeKey(loom.KeyEvent{Key: "end"})
	if want := buffer.Cols() - 4; view.offsetX != want {
		t.Fatalf("end horizontal offset = %d, want maximum offset %d", view.offsetX, want)
	}
	view.Draw(canvas, loom.Rect{W: 4, H: 2})
	if got := canvas.Get(0, 0).Text; got != "q" {
		t.Fatalf("first rendered cell at right edge = %q, want %q", got, "q")
	}
	view.ConsumeKey(loom.KeyEvent{Key: "right"})
	if want := buffer.Cols() - 4; view.offsetX != want {
		t.Fatalf("right edge offset = %d, want %d", view.offsetX, want)
	}
	view.ConsumeKey(loom.KeyEvent{Key: "home"})
	view.ConsumeKey(loom.KeyEvent{Key: "]"})
	if view.offsetX != 10 {
		t.Fatalf("] horizontal offset = %d, want 10", view.offsetX)
	}
	view.ConsumeKey(loom.KeyEvent{Key: "shift-right"})
	if view.offsetX != 16 {
		t.Fatalf("shift-right horizontal offset = %d, want clamped maximum 16", view.offsetX)
	}
	view.ConsumeKey(loom.KeyEvent{Key: "["})
	if view.offsetX != 6 {
		t.Fatalf("[ horizontal offset = %d, want 6", view.offsetX)
	}
	view.ConsumeKey(loom.KeyEvent{Key: "left"})
	view.ConsumeKey(loom.KeyEvent{Key: "home"})
	if view.offsetX != 0 {
		t.Fatalf("home horizontal offset = %d, want 0", view.offsetX)
	}
	view.ConsumeKey(loom.KeyEvent{Key: "pgdn"})
	if want := buffer.Rows() - 2; view.offsetY != want {
		t.Fatalf("page-down vertical offset = %d, want clamped offset %d", view.offsetY, want)
	}
	view.ConsumeKey(loom.KeyEvent{Key: "down"})
	if want := buffer.Rows() - 2; view.offsetY != want {
		t.Fatalf("bottom vertical offset = %d, want %d", view.offsetY, want)
	}
}

func TestViewCommandPaneSettings(t *testing.T) {
	pane := &loom.Pane{MaxCols: loom.DefaultMaxCols}
	configureViewPane(pane)
	if pane.MaxCols != 0 {
		t.Errorf("view pane MaxCols = %d, want 0", pane.MaxCols)
	}
	if !pane.Resizeable {
		t.Error("view pane Resizeable = false, want true")
	}
}

func TestEditCommandArityAndFileErrors(t *testing.T) {
	var out bytes.Buffer
	if err := execute([]string{"edit"}, &out); err == nil {
		t.Fatal("execute(edit) without args succeeded, want arity error")
	}
	dir := t.TempDir()
	if err := execute([]string{"edit", dir}, &out); err == nil {
		t.Fatal("execute(edit) on directory succeeded, want error")
	}
}

func TestEditViewUnsavedChangesDialog(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.ansi")
	if err := os.WriteFile(path, []byte("Initial content\n"), 0600); err != nil {
		t.Fatal(err)
	}

	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile: %v", err)
	}

	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain", AltScreen: true})
	if err != nil {
		t.Fatalf("newEditView: %v", err)
	}
	canvas := loom.NewCanvas(80, 10)

	// 1. Draw and verify bottom bar includes F10 Quit
	view.Draw(canvas, canvas.Bounds())
	screenText := canvasScreenText(canvas)
	if !strings.Contains(screenText, "Initial content") {
		t.Fatalf("editor did not render initial content: %q", screenText)
	}
	if !strings.Contains(screenText, "F10 Quit") {
		t.Fatalf("editor bottom bar omits F10 Quit: %q", screenText)
	}

	// 2. The Pane owns F10: a clean view allows the close request.
	quits := countQuits(view)
	if got := view.closeRequest(loom.CloseReasonQuitKey); got != loom.CloseAllow {
		t.Fatalf("close request on unmodified view = %v, want allow", got)
	}

	// 3. Modify document: the close request is vetoed behind the Save changes? dialog
	view.edit.ConsumeKey(loom.KeyEvent{Text: "X"})
	if !view.edit.IsModified() {
		t.Fatal("document expected to be modified after typing")
	}
	if got := view.closeRequest(loom.CloseReasonQuitKey); got != loom.CloseVeto {
		t.Fatalf("close request on modified document = %v, want veto", got)
	}
	if view.unsavedDialog == nil || !view.unsavedDialog.Open {
		t.Fatal("close request on modified view did not open unsaved changes dialog")
	}

	// Draw and verify dialog is shown
	canvas.Clear()
	view.Draw(canvas, canvas.Bounds())
	screenText = canvasScreenText(canvas)
	if !strings.Contains(screenText, "Save changes?") {
		t.Fatalf("dialog title missing in screen text: %q", screenText)
	}

	// 4. Cancel dialog with Esc
	view.ConsumeKey(loom.KeyEvent{Key: "esc"})
	if view.unsavedDialog != nil {
		t.Fatal("esc did not dismiss unsaved changes dialog")
	}
	if *quits != 0 {
		t.Fatal("cancelling dialog must not quit")
	}

	// 5. Open dialog again and select Discard
	view.closeRequest(loom.CloseReasonQuitKey)
	if view.unsavedDialog == nil {
		t.Fatal("dialog failed to reopen")
	}
	view.ConsumeKey(loom.KeyEvent{Key: "right"})
	view.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if *quits != 1 {
		t.Fatalf("Discard quits = %d, want 1", *quits)
	}

	// Test Save
	view.edit.ConsumeKey(loom.KeyEvent{Text: "Y"})
	view.closeRequest(loom.CloseReasonQuitKey)
	view.ConsumeKey(loom.KeyEvent{Key: "enter"}) // default button is "Save"
	if *quits != 2 {
		t.Fatalf("Save quits = %d, want 2", *quits)
	}

	// Verify saved content on disk
	savedData, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(savedData), "Initial content") || !strings.Contains(string(savedData), "XY") {
		t.Fatalf("saved disk content = %q, want modified content", string(savedData))
	}
}

func TestEditCommandPaneSettings(t *testing.T) {
	pane := &loom.Pane{MaxCols: loom.DefaultMaxCols}
	cfg := loom.EditorConfig{MouseGrab: true, AltScreen: true}
	configureEditPane(pane, cfg)
	if pane.MaxCols != 0 {
		t.Errorf("edit pane MaxCols = %d, want 0", pane.MaxCols)
	}
	if !pane.Resizeable {
		t.Error("edit pane Resizeable = false, want true")
	}
	if pane.DisableGlobalF10Quit {
		t.Error("edit pane DisableGlobalF10Quit = true, want false: the Pane owns F10")
	}
	if !pane.ResizeConfig.AltScreen {
		t.Errorf("pane AltScreen = false, want true")
	}
}

func TestEditCLIFlagsAndConfigPrecedence(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "doc.txt")
	_ = os.WriteFile(filePath, []byte("hello\n"), 0600)

	cfgPath := filepath.Join(dir, "editor.yaml")
	cfgContent := "theme: mc-dark\nmousegrab: true\naltscreen: false\n"
	_ = os.WriteFile(cfgPath, []byte(cfgContent), 0600)

	cmd := editCommand()
	cmd.SetArgs([]string{"--config", cfgPath, "--mousegrab=false", "--altscreen", filePath})

	// Parse flags
	if err := cmd.ParseFlags([]string{"--config", cfgPath, "--mousegrab=false", "--altscreen", filePath}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}

	cfg, err := loom.LoadEditorConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadEditorConfig: %v", err)
	}
	if cmd.Flags().Changed("mousegrab") {
		m, _ := cmd.Flags().GetBool("mousegrab")
		cfg.MouseGrab = m
	}
	if cmd.Flags().Changed("altscreen") {
		a, _ := cmd.Flags().GetBool("altscreen")
		cfg.AltScreen = a
	}

	if cfg.Theme != "mc-dark" {
		t.Errorf("Theme = %q, want mc-dark", cfg.Theme)
	}
	if cfg.MouseGrab {
		t.Errorf("MouseGrab = %v, want false (overridden by CLI flag)", cfg.MouseGrab)
	}
	if !cfg.AltScreen {
		t.Errorf("AltScreen = %v, want true (overridden by CLI flag)", cfg.AltScreen)
	}
}
