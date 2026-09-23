// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
	"gopkg.in/yaml.v3"
)

// TestThemePlainChoiceStyleMatchesDefault asserts that the plain theme produces
// exactly the same ChoiceStyle as the hard-coded DefaultChoiceStyle().
func TestThemePlainChoiceStyleMatchesDefault(t *testing.T) {
	got := loom.Theme("plain").ChoiceStyle()
	want := loom.DefaultChoiceStyle()
	if got != want {
		t.Errorf("Theme(\"plain\").ChoiceStyle() = %+v, want %+v", got, want)
	}
}

// TestThemePlainTableStyleMatchesDefault asserts that the plain theme produces
// exactly the same TableStyle as the hard-coded DefaultTableStyle().
func TestThemePlainTableStyleMatchesDefault(t *testing.T) {
	got := loom.Theme("plain").TableStyle()
	want := loom.DefaultTableStyle()
	if got != want {
		t.Errorf("Theme(\"plain\").TableStyle() = %+v, want %+v", got, want)
	}
}

// TestThemeUnknownFallsBackToPlain verifies that an unknown theme name returns
// the plain theme without panicking.
func TestThemeUnknownFallsBackToPlain(t *testing.T) {
	got := loom.Theme("does-not-exist")
	want := loom.Theme("plain")
	if got != want {
		t.Errorf("Theme(\"does-not-exist\") = %+v, want plain %+v", got, want)
	}
}

func TestThemeColorDistinguishesDefaultFromPaletteZero(t *testing.T) {
	defaultColor := loom.DefaultThemeColor()
	indexedZero := loom.ThemeColorIndex(0)
	if !defaultColor.IsDefault() || indexedZero.IsDefault() {
		t.Fatalf("default/indexed distinction lost: default=%+v indexed=%+v", defaultColor, indexedZero)
	}
	if defaultColor.Color() != loom.ColorReset() {
		t.Errorf("default theme color = %+v, want ColorReset", defaultColor.Color())
	}
	if indexedZero.Color() != loom.ColorIndex(0) {
		t.Errorf("indexed theme color = %+v, want ColorIndex(0)", indexedZero.Color())
	}
	if got := (loom.Style{FG: indexedZero.Color()}).ANSI(); !strings.Contains(got, "\x1b[38;5;0m") {
		t.Errorf("palette index 0 ANSI = %q, want indexed foreground", got)
	}
}

func TestThemeColorRGB(t *testing.T) {
	for _, test := range []struct {
		input string
		want  loom.Color
	}{
		{input: "#577cea", want: loom.ColorRGB(0x57, 0x7c, 0xea)},
		{input: "#AABBCC", want: loom.ColorRGB(0xaa, 0xbb, 0xcc)},
	} {
		var got loom.ThemeColor
		if err := yaml.Unmarshal([]byte("color: '"+test.input+"'\n"), &struct {
			Color *loom.ThemeColor `yaml:"color"`
		}{Color: &got}); err != nil {
			t.Fatalf("decode %q: %v", test.input, err)
		}
		if got.Color() != test.want || got.IsDefault() {
			t.Errorf("decode %q = %+v, want RGB %+v", test.input, got.Color(), test.want)
		}
	}
}

func TestThemeColorRejectsMalformedRGB(t *testing.T) {
	for _, input := range []string{"#577ce", "#577ceaa", "#gg77aa", "defaultish", "256"} {
		var got loom.ThemeColor
		if err := yaml.Unmarshal([]byte("color: '"+input+"'\n"), &struct {
			Color *loom.ThemeColor `yaml:"color"`
		}{Color: &got}); err == nil {
			t.Errorf("decode %q succeeded, want error", input)
		}
	}
}

func TestThemeFullRGBYAML(t *testing.T) {
	spec := `
normal_fg: "#ffffff"
normal_bg: "#1a1b26"
selected_fg: "#000000"
selected_bg: "#7aa2f7"
selected_bold: true
header_fg: "#e0af68"
header_bg: "#1a1b26"
header_bold: true
prompt_fg: "#7dcfff"
prompt_bg: "#1a1b26"
placeholder_fg: "#565f89"
placeholder_bg: "#1a1b26"
placeholder_dim: false
scrollbar_track_fg: "#1a1b26"
scrollbar_track_bg: "#1a1b26"
scrollbar_track_dim: true
scrollbar_thumb_fg: "#7aa2f7"
scrollbar_thumb_bg: "#1a1b26"
scrollbar_thumb_bold: false
status_fg: "#c0caf5"
status_bg: "#1a1b26"
status_bold: false
status_dim: false
border_fg: "#7aa2f7"
border_bg: "#1a1b26"
focus_bg: "#2ac3de"
`
	var colors loom.ThemeColors
	if err := yaml.Unmarshal([]byte(spec), &colors); err != nil {
		t.Fatalf("unmarshal theme with RGB: %v", err)
	}
	cs := colors.ChoiceStyle()
	if cs.Normal.FG != loom.ColorRGB(0xff, 0xff, 0xff) {
		t.Errorf("normal_fg = %+v, want RGB(255, 255, 255)", cs.Normal.FG)
	}
	if cs.Normal.BG != loom.ColorRGB(0x1a, 0x1b, 0x26) {
		t.Errorf("normal_bg = %+v, want RGB(26, 27, 38)", cs.Normal.BG)
	}
	ansi := cs.Normal.ANSI()
	if !strings.Contains(ansi, "38;2;255;255;255") || !strings.Contains(ansi, "48;2;26;27;38") {
		t.Errorf("normal ANSI = %q, want 24-bit RGB escape sequences", ansi)
	}
}

// TestThemeMCHasExpectedColors spot-checks key color roles in the mc theme.
func TestThemeMCHasExpectedColors(t *testing.T) {
	mc := loom.Theme("mc")
	cs := mc.ChoiceStyle()
	if cs.Normal.BG != loom.ColorIndex(69) {
		t.Errorf("mc normal_bg: got %+v, want ColorIndex(69)", cs.Normal.BG)
	}
	if cs.Selected.BG != loom.ColorIndex(75) {
		t.Errorf("mc selected_bg: got %+v, want ColorIndex(75)", cs.Selected.BG)
	}
	if cs.Selected.FG != loom.ColorIndex(16) {
		t.Errorf("mc selected_fg: got %+v, want ColorIndex(16)", cs.Selected.FG)
	}
	if cs.Prompt.FG != loom.ColorIndex(16) {
		t.Errorf("mc prompt_fg: got %+v, want ColorIndex(16)", cs.Prompt.FG)
	}
	if cs.Placeholder != (loom.Style{FG: loom.ColorIndex(243), BG: loom.ColorIndex(75)}) {
		t.Errorf("mc placeholder: got %+v, want fixed medium grey on cyan", cs.Placeholder)
	}
	if cs.Scrollbar.Track != (loom.Style{FG: loom.ColorIndex(69), BG: loom.ColorIndex(69), Dim: true}) ||
		cs.Scrollbar.Thumb != (loom.Style{FG: loom.ColorIndex(75), BG: loom.ColorIndex(69)}) {
		t.Errorf("mc scrollbar: got %+v", cs.Scrollbar)
	}
	if got := mc.FrameStyle().Status; got != (loom.Style{FG: loom.ColorIndex(252), BG: loom.ColorIndex(69)}) {
		t.Errorf("mc frame status: got %+v, want white on blue", got)
	}
	if got := mc.BoxStyle().Border; got != (loom.Style{FG: loom.ColorIndex(252), BG: loom.ColorIndex(69)}) {
		t.Errorf("mc box border: got %+v, want white on blue", got)
	}
	ts := mc.TableStyle()
	if ts.Header.FG != loom.ColorIndex(186) {
		t.Errorf("mc header_fg: got %+v, want ColorIndex(186)", ts.Header.FG)
	}
	if mc.FocusBGColor() != loom.ColorIndex(75) {
		t.Errorf("mc focus_bg: got %+v, want ColorIndex(75)", mc.FocusBGColor())
	}
}

func TestThemeMCVariantsHaveDistinctPalettes(t *testing.T) {
	tests := []struct {
		name                 string
		normalBG, selectedBG loom.Color
	}{
		{name: "mc", normalBG: loom.ColorIndex(69), selectedBG: loom.ColorIndex(75)},
		{name: "mc-classic", normalBG: loom.ColorIndex(27), selectedBG: loom.ColorIndex(51)},
		{name: "mc-dark", normalBG: loom.ColorIndex(19), selectedBG: loom.ColorIndex(37)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			style := loom.Theme(test.name).ChoiceStyle()
			if style.Normal.BG != test.normalBG || style.Selected.BG != test.selectedBG {
				t.Errorf("%s palette: normal=%+v selected=%+v", test.name, style.Normal.BG, style.Selected.BG)
			}
		})
	}
}

// TestSpeccedThemesContainsRequiredThemes asserts that all required themes are loaded.
func TestSpeccedThemesContainsRequiredThemes(t *testing.T) {
	for _, name := range []string{"plain", "mc", "mc-classic", "mc-dark", "julia256"} {
		if _, ok := loom.SpeccedThemes[name]; !ok {
			t.Errorf("SpeccedThemes missing key %q", name)
		}
	}
}

func TestThemeJulia256(t *testing.T) {
	julia := loom.Theme("julia256")
	cs := julia.ChoiceStyle()
	if cs.Normal != (loom.Style{FG: loom.ColorIndex(250), BG: loom.ColorIndex(237)}) {
		t.Errorf("julia256 normal: got %+v, want lightgray on color237", cs.Normal)
	}
	if cs.Selected != (loom.Style{FG: loom.ColorIndex(16), BG: loom.ColorIndex(51)}) {
		t.Errorf("julia256 selected: got %+v, want black on cyan", cs.Selected)
	}
	if cs.Prompt != (loom.Style{FG: loom.ColorIndex(16), BG: loom.ColorIndex(51)}) {
		t.Errorf("julia256 prompt: got %+v, want black on cyan", cs.Prompt)
	}
	ts := julia.TableStyle()
	if ts.Header.FG != loom.ColorIndex(226) || ts.Header.BG != loom.ColorIndex(237) {
		t.Errorf("julia256 header: got %+v, want yellow on color237", ts.Header)
	}
	if got := julia.BoxStyle().Border; got != (loom.Style{FG: loom.ColorIndex(250), BG: loom.ColorIndex(237)}) {
		t.Errorf("julia256 box border: got %+v, want lightgray on color237", got)
	}
	if julia.FocusBGColor() != loom.ColorIndex(240) {
		t.Errorf("julia256 focus_bg: got %+v, want ColorIndex(240)", julia.FocusBGColor())
	}
}

// TestThemeableNestedWidgets tests that ApplyTheme correctly propagates through
// a nested Tabs{Frame{Choice}} widget tree and generates M1 evidence.
func TestThemeableNestedWidgets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	// Create a nested widget structure: Tabs containing Frame containing Choice
	items := []loom.Item{
		{Name: "Option A", Desc: "First option"},
		{Name: "Option B", Desc: "Second option"},
		{Name: "Option C", Desc: "Third option"},
	}
	choice := loom.NewChoice(items)

	// Create a Frame containing the Choice
	border := loom.BoxBorder{
		TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘",
		Horizontal: "─", Vertical: "│",
	}
	frame := &loom.Frame{
		Title:  "File Browser",
		Status: "Themed Frame",
		Boxes: []loom.Box{
			{
				ID: "files", Title: "Files", Dynamic: true, FillHeight: true,
				MinWidth: 20, Height: 10, Border: border, Child: choice,
			},
		},
	}

	// Create Tabs with two different frames
	tabs := loom.NewTabs(
		loom.Tab{Title: "Tab 1", Widget: frame},
		loom.Tab{Title: "Tab 2", Widget: loom.NewChoice([]loom.Item{
			{Name: "Item 1"},
			{Name: "Item 2"},
		})},
	)

	// Render with two different themes
	const cols, rows = 80, 25
	themes := []string{"plain", "mc"}

	var buf bytes.Buffer
	for _, themeName := range themes {
		theme := loom.Theme(themeName)
		tabs.ApplyTheme(theme)

		// Render the tabs
		canvas := loom.NewCanvas(cols, rows)
		tabs.Draw(canvas, canvas.Bounds())

		// Add a label line before each theme rendering
		label := fmt.Sprintf("--- Theme: %s ---", themeName)
		buf.WriteString(label)
		buf.WriteString("\n\n")

		// Render the canvas to ANSI
		renderBuf := &bytes.Buffer{}
		err := loom.RenderTo(renderBuf, &simpleCanvasWidget{canvas: canvas}, cols, rows)
		if err != nil {
			t.Fatalf("RenderTo failed: %v", err)
		}
		buf.Write(renderBuf.Bytes())
		buf.WriteString("\n\n")
	}

	// Write evidence file
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("finding repo root: %v", err)
	}
	progressDir := filepath.Join(repoRoot, "docs", "progress", "061")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		t.Fatalf("creating progress directory: %v", err)
	}

	outPath := filepath.Join(progressDir, "M1-nested-themes.ansi")
	if err := os.WriteFile(outPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("writing evidence: %v", err)
	}

	t.Logf("M1 evidence saved to %s (%d bytes)", outPath, buf.Len())
}

// TestChoiceAndTableThemeable tests that Choice and Table implement Themeable
// and generate M2 evidence showing filebrowser-like structure with themes.
func TestChoiceAndTableThemeable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	// Create a frame structure similar to filebrowser: Frame containing Choice
	fileItems := []loom.Item{
		{Name: "file1.txt", Desc: "text file"},
		{Name: "file2.go", Desc: "go source"},
		{Name: "dir1", Desc: "<dir>"},
		{Name: "file3.md", Desc: "markdown"},
		{Name: "file4.json", Desc: "data file"},
	}
	fileChoice := loom.NewChoice(fileItems)

	border := loom.BoxBorder{
		TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘",
		Horizontal: "─", Vertical: "│",
	}

	// Create a Frame similar to filebrowser
	frame := &loom.Frame{
		Title:  "File Browser",
		Status: "↑↓ select  •  Enter open  •  F9 theme  •  F10 quit",
		Boxes: []loom.Box{
			{
				ID: "files", Title: "Files", Dynamic: true, FillHeight: true,
				MinWidth: 20, Height: 15, Border: border, Child: fileChoice,
			},
		},
	}

	// Also test a Table with Themeable
	tableCols := []loom.Column{
		{Header: "Name", Width: 12},
		{Header: "Size", Width: 8},
		{Header: "Modified", Width: 12},
	}
	tableRows := []loom.Row{
		{Cells: []string{"file1.txt", "1.2 KB", "2025-01-15"}, Key: "file1"},
		{Cells: []string{"file2.go", "3.5 KB", "2025-01-14"}, Key: "file2"},
		{Cells: []string{"dir1", "-", "2025-01-13"}, Key: "dir1"},
	}
	table := loom.NewTable(tableCols, tableRows)

	// Render with two different themes
	const cols, rows = 80, 30
	themes := []string{"plain", "mc"}

	var buf bytes.Buffer
	for _, themeName := range themes {
		theme := loom.Theme(themeName)

		// Apply theme to widgets
		fileChoice.ApplyTheme(theme)
		frame.ApplyTheme(theme)
		table.ApplyTheme(theme)

		// Render the frame
		canvas := loom.NewCanvas(cols, rows)
		frame.Draw(canvas, canvas.Bounds())

		// Add a label line before each theme rendering
		label := fmt.Sprintf("=== Theme: %s ===", themeName)
		buf.WriteString(label)
		buf.WriteString("\n\n")

		// Render the canvas to ANSI
		renderBuf := &bytes.Buffer{}
		err := loom.RenderTo(renderBuf, &simpleCanvasWidget{canvas: canvas}, cols, rows)
		if err != nil {
			t.Fatalf("RenderTo failed: %v", err)
		}
		buf.Write(renderBuf.Bytes())
		buf.WriteString("\n\n")

		// Also render the table separately
		tableBuf := loom.NewCanvas(cols, 8)
		table.Draw(tableBuf, tableBuf.Bounds())

		tableLabel := fmt.Sprintf("--- Table with %s theme ---", themeName)
		buf.WriteString(tableLabel)
		buf.WriteString("\n\n")

		tableRenderBuf := &bytes.Buffer{}
		err = loom.RenderTo(tableRenderBuf, &simpleCanvasWidget{canvas: tableBuf}, cols, 8)
		if err != nil {
			t.Fatalf("RenderTo table failed: %v", err)
		}
		buf.Write(tableRenderBuf.Bytes())
		buf.WriteString("\n\n")
	}

	// Write evidence file
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("finding repo root: %v", err)
	}
	progressDir := filepath.Join(repoRoot, "docs", "progress", "061")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		t.Fatalf("creating progress directory: %v", err)
	}

	outPath := filepath.Join(progressDir, "M2-filebrowser-themes.ansi")
	if err := os.WriteFile(outPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("writing evidence: %v", err)
	}

	t.Logf("M2 evidence saved to %s (%d bytes)", outPath, buf.Len())
}

// findRepoRoot walks up the directory tree until it finds a go.mod file.
func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd, nil
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}
		cwd = parent
	}
}
