// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
	"github.com/spf13/cobra"
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
	for _, args := range [][]string{{"measure"}, {"measure", "a", "b"}, {"eval"}, {"view"}, {"check-box"}, {"widgets", "a", "b"}} {
		var out bytes.Buffer
		if err := execute(args, &out); err == nil {
			t.Errorf("execute(%q) succeeded, want arity error", args)
		}
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
	runWidgetPane = func(widget loom.Widget) error {
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

func TestWidgetsThemeFlagAndF2Cycle(t *testing.T) {
	if len(loom.ThemeNames()) < 2 {
		t.Fatal("theme cycling requires at least two themes")
	}
	previous := runWidgetPane
	var shown loom.Widget
	runWidgetPane = func(widget loom.Widget) error {
		shown = widget
		return nil
	}
	t.Cleanup(func() { runWidgetPane = previous })
	if err := execute([]string{"widgets", "--show", "--theme", "mc", "Chart"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	gallery, ok := shown.(*themedGallery)
	if !ok {
		t.Fatalf("shown widget = %T, want *themedGallery", shown)
	}
	if gallery.themeName != "mc" {
		t.Fatalf("initial theme = %q, want mc", gallery.themeName)
	}
	child := &themeProbeWidget{}
	gallery = newThemedGallery(child, "mc")
	initialCanvas := loom.NewCanvas(40, 4)
	gallery.Draw(initialCanvas, initialCanvas.Bounds())
	initialColor := initialCanvas.Get(0, 0).Style
	gallery.HandleKey(loom.KeyEvent{Key: "f2"})
	names := loom.ThemeNames()
	index := slices.Index(names, "mc")
	want := names[(index+1)%len(names)]
	if gallery.themeName != want || child.theme.NormalFG != loom.Theme(want).NormalFG {
		t.Fatalf("after F2 theme = %q, child theme mismatch; want %q", gallery.themeName, want)
	}
	canvas := loom.NewCanvas(40, 4)
	gallery.Draw(canvas, canvas.Bounds())
	if canvas.Get(0, 0).Style == initialColor {
		t.Fatalf("demo rendered the same color style after switching from mc to %q", want)
	}
	if !strings.Contains(canvas.Row(canvas.Rows()-1), "Theme: "+want) {
		t.Fatalf("gallery chrome = %q, want active theme %q", canvas.Row(canvas.Rows()-1), want)
	}
}

type themeProbeWidget struct{ theme loom.ThemeColors }

func (w *themeProbeWidget) Draw(c *loom.Canvas, r loom.Rect) {
	c.Write(r.X, r.Y, "X", w.theme.ChoiceStyle().Normal)
}
func (*themeProbeWidget) HandleKey(loom.KeyEvent) bool        { return false }
func (*themeProbeWidget) HandleMouse(loom.MouseEvent) bool    { return false }
func (w *themeProbeWidget) ApplyTheme(theme loom.ThemeColors) { w.theme = theme }

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
	if view.HandleKey(loom.KeyEvent{Key: "down"}) || view.offsetY != 1 {
		t.Fatalf("down key did not scroll: offsetY=%d", view.offsetY)
	}
	if !view.HandleKey(loom.KeyEvent{Key: "f10"}) {
		t.Fatal("F10 did not quit")
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
		view.HandleKey(key)
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
	view.HandleKey(loom.KeyEvent{Key: "right"})
	view.HandleKey(loom.KeyEvent{Key: "right"})
	view.Draw(canvas, loom.Rect{W: 4, H: 2})
	if got := canvas.Get(0, 0).Text; got != "c" {
		t.Fatalf("first rendered cell after horizontal pan = %q, want %q", got, "c")
	}
	view.HandleKey(loom.KeyEvent{Key: "end"})
	if want := buffer.Cols() - 4; view.offsetX != want {
		t.Fatalf("end horizontal offset = %d, want maximum offset %d", view.offsetX, want)
	}
	view.Draw(canvas, loom.Rect{W: 4, H: 2})
	if got := canvas.Get(0, 0).Text; got != "q" {
		t.Fatalf("first rendered cell at right edge = %q, want %q", got, "q")
	}
	view.HandleKey(loom.KeyEvent{Key: "right"})
	if want := buffer.Cols() - 4; view.offsetX != want {
		t.Fatalf("right edge offset = %d, want %d", view.offsetX, want)
	}
	view.HandleKey(loom.KeyEvent{Key: "home"})
	view.HandleKey(loom.KeyEvent{Key: "]"})
	if view.offsetX != 10 {
		t.Fatalf("] horizontal offset = %d, want 10", view.offsetX)
	}
	view.HandleKey(loom.KeyEvent{Key: "shift-right"})
	if view.offsetX != 16 {
		t.Fatalf("shift-right horizontal offset = %d, want clamped maximum 16", view.offsetX)
	}
	view.HandleKey(loom.KeyEvent{Key: "["})
	if view.offsetX != 6 {
		t.Fatalf("[ horizontal offset = %d, want 6", view.offsetX)
	}
	view.HandleKey(loom.KeyEvent{Key: "left"})
	view.HandleKey(loom.KeyEvent{Key: "home"})
	if view.offsetX != 0 {
		t.Fatalf("home horizontal offset = %d, want 0", view.offsetX)
	}
	view.HandleKey(loom.KeyEvent{Key: "pgdn"})
	if want := buffer.Rows() - 2; view.offsetY != want {
		t.Fatalf("page-down vertical offset = %d, want clamped offset %d", view.offsetY, want)
	}
	view.HandleKey(loom.KeyEvent{Key: "down"})
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
