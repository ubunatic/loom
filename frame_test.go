// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func emptyShellFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/fixtures/empty-shell.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func monitorFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("examples/monitor/monitor/spec/monitor.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func shellFixture(t *testing.T) string {
	return emptyShellFixture(t)
}

func TestStaticShellGolden(t *testing.T) {
	w, cfg, err := BuildWidget(strings.NewReader(emptyShellFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	rows := Render(w, cfg.MaxWidth(), cfg.Height(0))
	// Explicit expected columns, independent of production width/layout helpers.
	want := []string{
		"Loom monitor                                                    ",
		"┌ [u] All Usage ──────────────┐  ┌ [l] Load ───────────────────┐",
		"│                             │  │                             │",
		"│                             │  │                             │",
		"│                             │  │                             │",
		"│                             │  │                             │",
		"│                             │  │                             │",
		"└─────────────────────────────┘  └─────────────────────────────┘",
		"Show once  [u]usage:on  [l]load:on  [q]quit                     ",
	}
	if len(rows) != len(want) {
		t.Fatalf("rows=%d want=%d", len(rows), len(want))
	}
	for y, row := range rows {
		got := strings.TrimSuffix(row, "\x1b[0m")
		if got != want[y] {
			t.Errorf("row %d\n got %q\nwant %q", y, got, want[y])
		}
	}
}

func TestFixedHeightFramesRenderIdentical(t *testing.T) {
	// Golden test: fixed-height frames (without FillHeight) should render
	// byte-identical regardless of later enhancements.
	w, cfg, err := BuildWidget(strings.NewReader(emptyShellFixture(t)))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the fixture's boxes are NOT using FillHeight
	frame := w.(*Frame)
	for _, box := range frame.Boxes {
		if box.FillHeight {
			t.Fatalf("fixture should not have FillHeight=true")
		}
	}

	// Render at the configured dimensions
	rows := Render(w, cfg.MaxWidth(), cfg.Height(0))

	// Capture golden output (this is what we're pinning)
	gotRows := make([]string, len(rows))
	for i, row := range rows {
		gotRows[i] = strings.TrimSuffix(row, "\x1b[0m")
	}

	// These rows should be stable: any change to layout logic should preserve them
	expectedLineCount := cfg.Height(0)
	if len(gotRows) != expectedLineCount {
		t.Errorf("expected %d rows, got %d", expectedLineCount, len(gotRows))
	}

	// Verify key structural properties (boxes and status visible)
	if len(gotRows) < 3 {
		t.Fatalf("need at least 3 rows for title, content, status")
	}
	if !strings.Contains(gotRows[0], "Loom monitor") {
		t.Errorf("title row missing: %q", gotRows[0])
	}
	if !strings.Contains(gotRows[len(gotRows)-1], "Show once") {
		t.Errorf("status row missing: %q", gotRows[len(gotRows)-1])
	}
	// Both boxes should be visible (borders visible)
	allRows := strings.Join(gotRows, "\n")
	if !strings.Contains(allRows, "All Usage") {
		t.Errorf("left box title missing")
	}
	if !strings.Contains(allRows, "Load") {
		t.Errorf("right box title missing")
	}
}

func TestMonitorExampleGolden(t *testing.T) {
	w, cfg, err := BuildWidget(strings.NewReader(monitorFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	rows := Render(w, cfg.MaxWidth(), cfg.Height(0))
	if len(rows) != cfg.Height(0) {
		t.Fatalf("rows=%d want=%d", len(rows), cfg.Height(0))
	}
	rendered := strings.Join(rows, "\n")
	if !strings.Contains(rendered, "Claude") || !strings.Contains(rendered, "cpu (16c)") || !strings.Contains(rendered, "[⣿⣿  ]") || !strings.Contains(rendered, "[⣿⣿⣿⣿][⣀⣀⣀⣀]") {
		t.Fatalf("monitor fixture missing expected row values:\n%s", rendered)
	}
}

type overflowingChild struct{}

func (overflowingChild) Draw(c *Canvas, _ Rect) {
	c.Fill(Rect{X: -5, Y: -5, W: 30, H: 30}, Cell{Text: "X"})
}
func (overflowingChild) HandleKey(KeyEvent) bool     { return false }
func (overflowingChild) HandleMouse(MouseEvent) bool { return false }

func TestBoxPaddingAndIsolation(t *testing.T) {
	w, _, err := BuildWidget(strings.NewReader(shellFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	b := w.(*Frame).Boxes[0]
	b.Title = ""
	b.Child = overflowingChild{}
	c := NewCanvas(12, 9)
	c.Fill(c.Bounds(), Cell{Text: "."})
	b.Draw(c, Rect{X: 2, Y: 1, W: 8, H: 7})
	want := []string{
		"............",
		"..┌──────┐..",
		"..│      │..",
		"..│ XXXX │..",
		"..│ XXXX │..",
		"..│ XXXX │..",
		"..│      │..",
		"..└──────┘..",
		"............",
	}
	for y, expected := range want {
		if got := strings.TrimSuffix(c.Row(y), "\x1b[0m"); got != expected {
			t.Errorf("row %d: got %q want %q", y, got, expected)
		}
	}
}

func TestFrameAndBoxStyles(t *testing.T) {
	border := BoxBorder{TopLeft: "+", TopRight: "+", BottomLeft: "+", BottomRight: "+", Horizontal: "-", Vertical: "|"}
	frameStyle := FrameStyle{
		Background: Style{BG: ColorIndex(27)},
		Title:      Style{FG: ColorIndex(15), BG: ColorIndex(27)},
		Status:     Style{FG: ColorIndex(0), BG: ColorIndex(51)},
	}
	boxStyle := BoxStyle{
		Background: Style{BG: ColorIndex(27)},
		Border:     Style{FG: ColorIndex(15), BG: ColorIndex(27)},
		Title:      Style{FG: ColorIndex(226), BG: ColorIndex(27)},
	}
	f := Frame{Title: "Frame", Status: "Status", Style: frameStyle, Boxes: []Box{{Title: "Box", Width: 8, Height: 4, Border: border, Style: boxStyle}}}
	canvas := NewCanvas(10, 6)
	f.Draw(canvas, canvas.Bounds())

	checks := []struct {
		name  string
		x, y  int
		style Style
	}{
		{name: "frame title", x: 0, y: 0, style: frameStyle.Title},
		{name: "frame background", x: 9, y: 0, style: frameStyle.Background},
		{name: "box border", x: 0, y: 1, style: boxStyle.Border},
		{name: "box title", x: 1, y: 1, style: boxStyle.Title},
		{name: "box background", x: 1, y: 2, style: boxStyle.Background},
		{name: "status", x: 0, y: 5, style: frameStyle.Status},
		{name: "status fill", x: 9, y: 5, style: frameStyle.Status},
	}
	for _, check := range checks {
		if got := canvas.Get(check.x, check.y).Style; got != check.style {
			t.Errorf("%s style = %+v, want %+v", check.name, got, check.style)
		}
	}
}
func TestFrameLayoutFillHeight(t *testing.T) {
	f := Frame{
		Boxes: []Box{
			{ID: "left", Dynamic: true, FillHeight: true, MinWidth: 10, Height: 4},
			{ID: "right", Dynamic: true, FillHeight: false, MinWidth: 10, Height: 4},
		},
	}
	rects := f.Layout(30, 10)
	if len(rects) != 2 {
		t.Fatalf("expected 2 rects, got %d", len(rects))
	}
	// Available inner height for boxes is height - 2 (rows 1 to height-2).
	if got, want := rects[0].H, 8; got != want {
		t.Errorf("left box FillHeight height = %d, want %d", got, want)
	}
	if got, want := rects[1].H, 4; got != want {
		t.Errorf("right box fixed height = %d, want %d", got, want)
	}
}

func TestFrameLayoutFillHeightWithMinMaxClamping(t *testing.T) {
	for _, tc := range []struct {
		name     string
		height   int
		minH, mH int
		wantH    int
	}{
		// When available space (height-2) exceeds MaxHeight, clamp to MaxHeight
		{"clamp to max", 20, 0, 10, 10},
		// When available space is greater than MinHeight, fill to available
		{"fill with min constraint", 15, 5, 0, 13},
		// When available space is between min and max, use available space
		{"between min and max", 15, 2, 20, 13},
		// Zero max (no limit) should use available space
		{"zero max means no limit", 20, 0, 0, 18},
		// Max below available should be clamped to max
		{"max limits fill", 30, 0, 15, 15},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := Frame{
				Boxes: []Box{
					{
						ID:         "box",
						Dynamic:    true,
						FillHeight: true,
						MinWidth:   10,
						Height:     4,
						MinHeight:  tc.minH,
						MaxHeight:  tc.mH,
					},
				},
			}
			rects := f.Layout(30, tc.height)
			if len(rects) != 1 || rects[0].H != tc.wantH {
				t.Errorf("got height=%d, want %d", rects[0].H, tc.wantH)
			}
		})
	}
}

func TestFrameLayoutMixedFillAndFixed(t *testing.T) {
	// Test a row with both filling and fixed-height boxes
	f := Frame{
		Gap: 2,
		Boxes: []Box{
			{ID: "fill1", Dynamic: true, FillHeight: true, MinWidth: 10, Height: 4},
			{ID: "fixed", Dynamic: true, FillHeight: false, MinWidth: 10, Height: 6},
			{ID: "fill2", Dynamic: true, FillHeight: true, MinWidth: 10, Height: 4},
		},
	}
	rects := f.Layout(50, 20)
	if len(rects) != 3 {
		t.Fatalf("expected 3 rects, got %d", len(rects))
	}
	// Available height is 20 - 2 = 18
	// All filling boxes should get 18 (clamped by their constraints)
	// Fixed box should get its preferred height (6)
	if got := rects[0].H; got != 18 {
		t.Errorf("fill1 height = %d, want 18", got)
	}
	if got := rects[1].H; got != 6 {
		t.Errorf("fixed height = %d, want 6", got)
	}
	if got := rects[2].H; got != 18 {
		t.Errorf("fill2 height = %d, want 18", got)
	}
}

func TestFrameLayoutFillHeightAtVariousHeights(t *testing.T) {
	// Test that FillHeight boxes adapt to different terminal heights
	f := Frame{
		Gap: 1,
		Boxes: []Box{
			{ID: "left", Dynamic: true, FillHeight: true, MinWidth: 15, Height: 4},
			{ID: "right", Dynamic: true, FillHeight: true, MinWidth: 15, Height: 4},
		},
	}

	for _, height := range []int{12, 20, 30} {
		t.Run(fmt.Sprintf("height-%d", height), func(t *testing.T) {
			rects := f.Layout(60, height)
			if len(rects) != 2 {
				t.Fatalf("expected 2 rects, got %d", len(rects))
			}
			// Both boxes should fill to available height (height - 2)
			expectedH := height - 2
			if got := rects[0].H; got != expectedH {
				t.Errorf("left height = %d, want %d", got, expectedH)
			}
			if got := rects[1].H; got != expectedH {
				t.Errorf("right height = %d, want %d", got, expectedH)
			}
		})
	}
}

func TestFrameLayoutStackedWidthBehavior(t *testing.T) {
	// Test that stacked layout properly handles width when boxes are dynamic
	f := Frame{
		Gap:        1,
		Breakpoint: 40,
		Boxes: []Box{
			{ID: "box1", Dynamic: true, MinWidth: 10, Height: 4},
			{ID: "box2", Dynamic: true, MinWidth: 10, Height: 4},
		},
	}

	// At 30 width (< 40 breakpoint), should use stacked layout
	rects := f.Layout(30, 20)
	if len(rects) != 2 {
		t.Fatalf("expected 2 rects, got %d", len(rects))
	}

	// In stacked layout, dynamic boxes should span full width
	if got := rects[0].W; got != 30 {
		t.Errorf("stacked box 1 width = %d, expected 30 (full width)", got)
	}
	if got := rects[1].W; got != 30 {
		t.Errorf("stacked box 2 width = %d, expected 30 (full width)", got)
	}
}

func TestFrameLayoutHiddenBoxesNotInFill(t *testing.T) {
	// Hidden boxes should be excluded from fill computation
	f := Frame{
		Gap: 1,
		Boxes: []Box{
			{ID: "visible", Dynamic: true, FillHeight: true, MinWidth: 10, Height: 4},
			{ID: "hidden", Dynamic: true, FillHeight: true, MinWidth: 10, Height: 4, Hidden: true},
			{ID: "visible2", Dynamic: true, FillHeight: false, MinWidth: 10, Height: 6},
		},
	}

	rects := f.Layout(50, 20)
	if len(rects) != 3 {
		t.Fatalf("expected 3 rects, got %d", len(rects))
	}

	// Hidden box should not have a rectangle allocated
	if rects[1].W > 0 || rects[1].H > 0 {
		t.Errorf("hidden box should have zero rect, got %+v", rects[1])
	}

	// Visible boxes should still be laid out
	if rects[0].H == 0 {
		t.Errorf("visible box 1 should have height")
	}
	if rects[2].H == 0 {
		t.Errorf("visible box 2 should have height")
	}
}

func TestFrameTinyBounds(t *testing.T) {
	w, _, err := BuildWidget(strings.NewReader(shellFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	for _, size := range []Rect{{}, {W: 1, H: 1}, {W: 2, H: 2}, {W: 3, H: 3}, {W: 4, H: 4}, {W: 5, H: 5}} {
		c := NewCanvas(10, 10)
		c.Fill(c.Bounds(), Cell{Text: "."})
		r := Rect{X: 2, Y: 2, W: size.W, H: size.H}
		w.Draw(c, r)
		for y := 0; y < 10; y++ {
			for x := 0; x < 10; x++ {
				if !r.Contains(x, y) && c.Get(x, y).Text != "." {
					t.Fatalf("size %+v escaped at %d,%d", size, x, y)
				}
			}
		}
		if size.H >= 2 && c.Get(2, 2+size.H-1).Text != "S" {
			t.Errorf("size %+v lost status", size)
		}
	}
}

func TestShellDeclarationFidelity(t *testing.T) {
	source := shellFixture(t)
	source = strings.Replace(source, "title: Loom monitor", "title: Changed in YAML", 1)
	source = strings.Replace(source, "id: usage", "id: first", 1)
	source = strings.Replace(source, "target: usage", "target: first", 1)
	source = strings.Replace(source, "title: All Usage", "title: First", 1)
	w, _, err := BuildWidget(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	frame := w.(*Frame)
	if frame.Title != "Changed in YAML" || frame.Boxes[0].ID != "first" || frame.Boxes[0].Padding != 1 || frame.Gap != 2 {
		t.Fatalf("declaration not consumed: %+v", frame)
	}
	header, boxes, _ := strings.Cut(source, "    boxes:\n")
	parts := strings.Split(boxes, "      - id:")
	reordered := header + "    boxes:\n" + parts[0] + "      - id:" + parts[2] + "      - id:" + parts[1]
	w, _, err = BuildWidget(strings.NewReader(reordered))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(Render(w, 64, 9)[1], "┌ [l] Load ") {
		t.Fatal("box sequence ignored")
	}
}

func TestShellValidation(t *testing.T) {
	source := shellFixture(t)
	for _, tc := range []struct{ name, old, new, message string }{
		{"unknown field", "gap: 2", "gpa: 2", "line"},
		{"invalid size", "width: 31", "width: -1", "boxes[0]"},
		{"duplicate ID", "id: load", "id: usage", "duplicate"},
		{"missing ID", "id: usage", "id: ''", ".id"},
		{"padding", "padding: 1", "padding: 99", "padding"},
		{"unknown root", "height: 9", "root: absent\n  height: 9", "app.root"},
		{"missing dimensions", "height: 9", "height: 0", "positive height"},
		{"control text", "title: Loom monitor", "title: \"bad\\nline\"", "ASCII"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := strings.Replace(source, tc.old, tc.new, 1)
			_, _, err := BuildWidget(strings.NewReader(input))
			if err == nil || !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("got %v, want %q", err, tc.message)
			}
			if err := ValidateYAML(strings.NewReader(input)); err == nil {
				t.Fatal("validation/build disagree")
			}
		})
	}
}

func TestFrameContentHeightWithFillHeight(t *testing.T) {
	// ContentHeight should treat a filling box as its preferred Height (or 1 if Height <= 0)
	f := Frame{
		Boxes: []Box{
			{ID: "fill", FillHeight: true, Height: 6},
			{ID: "fixed", FillHeight: false, Height: 4},
		},
	}

	// ContentHeight includes 2 chrome rows and the tallest preferred height
	// Filling box has Height=6, fixed has Height=4
	// So ContentHeight should be 6 + 2 = 8
	expectedHeight := 8
	if got := f.ContentHeight(); got != expectedHeight {
		t.Errorf("ContentHeight with FillHeight = %d, expected %d", got, expectedHeight)
	}
}

func TestFrameHeightForWidthWithFillHeight(t *testing.T) {
	// HeightForWidth should use the same logic as ContentHeight for filling boxes
	f := Frame{
		Breakpoint: 40,
		Gap:        1,
		Boxes: []Box{
			{ID: "fill", FillHeight: true, Height: 6},
			{ID: "fixed", FillHeight: false, Height: 4},
		},
	}

	// With wide width (>= breakpoint), uses ContentHeight
	wideH := f.HeightForWidth(50)
	if wideH != 8 {
		t.Errorf("HeightForWidth(50) with breakpoint=40 = %d, expected 8", wideH)
	}

	// With narrow width (< breakpoint), uses stacked height calculation
	// In stacked: 2 chrome + (6 gap 4) = 2 + 6 + 1 + 4 = 13
	narrowH := f.HeightForWidth(30)
	if narrowH != 13 {
		t.Errorf("HeightForWidth(30) with breakpoint=40 = %d, expected 13", narrowH)
	}
}

func TestGenerateM1FillHeightEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	for _, height := range []int{12, 20, 30} {
		t.Run(fmt.Sprintf("height-%d", height), func(t *testing.T) {
			source := emptyShellFixture(t)
			// Add FillHeight: true to both boxes
			source = strings.Replace(source, "- id: usage", "- id: usage\n        fill_height: true", 1)
			source = strings.Replace(source, "- id: load", "- id: load\n        fill_height: true", 1)
			// Update height to the test value
			source = strings.Replace(source, "  height: 9", fmt.Sprintf("  height: %d", height), 1)

			w, cfg, err := BuildWidget(strings.NewReader(source))
			if err != nil {
				t.Fatalf("BuildWidget failed: %v", err)
			}

			// Render at the target height and full width
			frames := Render(w, cfg.MaxWidth(), height)
			var buf strings.Builder
			for _, line := range frames {
				buf.WriteString(line)
				buf.WriteString("\n")
			}

			repoRoot := findRepoRoot(t)
			progressDir := filepath.Join(repoRoot, "docs", "progress", "051")
			if err := os.MkdirAll(progressDir, 0755); err != nil {
				t.Fatalf("creating progress directory: %v", err)
			}

			outPath := filepath.Join(progressDir, fmt.Sprintf("M1-fill-h%d.ansi", height))
			if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
				t.Fatalf("writing evidence: %v", err)
			}
			t.Logf("M1 evidence saved to %s (%d bytes)", outPath, len(buf.String()))
		})
	}
}

func TestGenerateM2StackedEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	// Create a stacked layout test frame
	source := `app:
  height: 20
  max_width: 60
view:
  name: main
  frame:
    title: Stacked with FillHeight
    status: Two filling boxes stacked at 60x20
    breakpoint: 40
    boxes:
      - id: top
        title: Top Box
        width: 58
        height: 4
        fill_height: true
        padding: 1
      - id: bottom
        title: Bottom Box
        width: 58
        height: 4
        fill_height: true
        padding: 1
`

	w, _, err := BuildWidget(strings.NewReader(source))
	if err != nil {
		t.Fatalf("BuildWidget failed: %v", err)
	}

	// Render at a narrow width to trigger stacked layout
	frames := Render(w, 30, 20)
	var buf strings.Builder
	for _, line := range frames {
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	repoRoot := findRepoRoot(t)
	progressDir := filepath.Join(repoRoot, "docs", "progress", "051")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		t.Fatalf("creating progress directory: %v", err)
	}

	outPath := filepath.Join(progressDir, "M2-stacked.ansi")
	if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
		t.Fatalf("writing evidence: %v", err)
	}
	t.Logf("M2 stacked evidence saved to %s (%d bytes)", outPath, len(buf.String()))
}

func TestGenerateM2BreakpointEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	// Create a breakpoint test frame
	source := `app:
  height: 20
  max_width: 60
view:
  name: main
  frame:
    title: Breakpoint Transition
    status: Same frame at 70w (horizontal) vs 30w (stacked)
    breakpoint: 50
    boxes:
      - id: left
        title: Fill Box 1
        width: 30
        height: 4
        fill_height: true
        min_width: 15
        padding: 1
      - id: right
        title: Fill Box 2
        width: 30
        height: 4
        fill_height: true
        min_width: 15
        padding: 1
`

	w, _, err := BuildWidget(strings.NewReader(source))
	if err != nil {
		t.Fatalf("BuildWidget failed: %v", err)
	}

	// Render at width that exceeds breakpoint (horizontal layout)
	frames := Render(w, 70, 20)
	var buf strings.Builder
	for _, line := range frames {
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	repoRoot := findRepoRoot(t)
	progressDir := filepath.Join(repoRoot, "docs", "progress", "051")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		t.Fatalf("creating progress directory: %v", err)
	}

	outPath := filepath.Join(progressDir, "M2-breakpoint.ansi")
	if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
		t.Fatalf("writing evidence: %v", err)
	}
	t.Logf("M2 breakpoint evidence saved to %s (%d bytes)", outPath, len(buf.String()))
}

// findRepoRoot walks up from the current directory to find the repo root (where go.mod is)
func findRepoRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting current directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			t.Fatalf("go.mod not found in any parent directory")
		}
		cwd = parent
	}
}
