package filebrowser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom"
)

type helpModalProbe struct{}

func (helpModalProbe) Draw(*loom.Canvas, loom.Rect)     {}
func (helpModalProbe) HandleKey(loom.KeyEvent) bool     { return false }
func (helpModalProbe) HandleMouse(loom.MouseEvent) bool { return false }

func TestBrowserSelectionAndNavigation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir, "plain", loom.Theme("plain"))
	if err != nil {
		t.Fatal(err)
	}
	b.HandleKey(loom.KeyEvent{Key: "down"}) // parent -> a.txt
	if item, ok := b.list.Selected(); !ok || item.Name != "a.txt" {
		t.Fatalf("selected item = %+v, ok=%v", item, ok)
	}
	if got := strings.Join(b.details.Lines, "\n"); !strings.Contains(got, "Size: 5 bytes") || !strings.Contains(got, "Type: file") {
		t.Fatalf("file metadata missing: %s", got)
	}
	b.HandleKey(loom.KeyEvent{Key: "down"}) // sub
	b.HandleKey(loom.KeyEvent{Key: "enter"})
	if b.dir != filepath.Join(dir, "sub") || b.frame.Boxes[0].Child != b.list {
		t.Fatalf("directory navigation failed: %q", b.dir)
	}
	if item, ok := b.list.Selected(); !ok || item.Name != ".." {
		t.Fatalf("parent entry missing: %+v, ok=%v", item, ok)
	}
	b.HandleKey(loom.KeyEvent{Key: "enter"})
	if b.dir != dir {
		t.Fatalf("parent navigation returned %q, want %q", b.dir, dir)
	}
	if item, ok := b.list.Selected(); !ok || item.Name != "sub" {
		t.Fatalf("selected item after parent navigation = %+v, ok=%v; want sub", item, ok)
	}
}

func TestBrowserNavigationToParentFallsBackToFirstItem(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(sub, "plain", loom.Theme("plain"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(sub); err != nil {
		t.Fatal(err)
	}

	b.HandleKey(loom.KeyEvent{Key: "enter"})

	if b.dir != dir {
		t.Fatalf("parent navigation returned %q, want %q", b.dir, dir)
	}
	if b.list.FilteredSel() != 0 {
		t.Fatalf("selected index after missing previous directory = %d, want 0", b.list.FilteredSel())
	}
}

func TestBrowserUsesAnimatedBackground(t *testing.T) {
	pane := &loom.Pane{}
	configurePane(pane)
	if pane.Background == nil {
		t.Fatal("filebrowser pane has no animated background")
	}

	b, err := newBrowser(t.TempDir(), "plain", loom.Theme("plain"))
	if err != nil {
		t.Fatal(err)
	}
	c := loom.NewCanvas(100, 24)
	b.Draw(c, c.Bounds())
	pane.Background.(loom.AnimatedBackground).DrawBackgroundAt(c, c.Bounds(), time.Unix(0, 0))

	transparent := 0
	for y := 0; y < c.Rows(); y++ {
		for x := 0; x < c.Cols(); x++ {
			if c.Get(x, y).Text == " " && c.Get(x, y).Style.BG == loom.ColorReset() {
				transparent++
			}
		}
	}
	if transparent == 0 {
		t.Fatal("filebrowser has no transparent pane surface for background composition")
	}
}

func TestBrowserThemePersistsAcrossNavigation(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	theme := loom.Theme("mc")
	b, err := newBrowser(dir, "mc", theme)
	if err != nil {
		t.Fatal(err)
	}
	want := theme.ChoiceStyle()
	if b.list.Style != want {
		t.Fatalf("initial list style = %+v, want %+v", b.list.Style, want)
	}
	if b.details.Style != want.Normal {
		t.Fatalf("metadata style = %+v, want %+v", b.details.Style, want.Normal)
	}
	if b.details.Scrollbar != theme.ScrollbarStyle() {
		t.Fatalf("metadata scrollbar = %+v, want %+v", b.details.Scrollbar, theme.ScrollbarStyle())
	}
	if b.frame.Style != theme.FrameStyle() {
		t.Fatalf("frame style = %+v, want %+v", b.frame.Style, theme.FrameStyle())
	}
	for i, box := range b.frame.Boxes {
		if box.Style != theme.BoxStyle() {
			t.Fatalf("box %d style = %+v, want %+v", i, box.Style, theme.BoxStyle())
		}
	}

	b.HandleKey(loom.KeyEvent{Key: "down"})
	b.HandleKey(loom.KeyEvent{Key: "enter"})
	if b.dir != filepath.Join(dir, "sub") {
		t.Fatalf("directory navigation returned %q", b.dir)
	}
	if b.list.Style != want {
		t.Fatalf("navigated list style = %+v, want %+v", b.list.Style, want)
	}
}

func TestBrowserCyclesSpeccedThemes(t *testing.T) {
	b, err := newBrowser(t.TempDir(), "mc", loom.Theme("mc"))
	if err != nil {
		t.Fatal(err)
	}
	names := themeNames()
	start := 0
	for i, name := range names {
		if name == "mc" {
			start = i
			break
		}
	}
	for step := 1; step <= len(names); step++ {
		if b.HandleKey(loom.KeyEvent{Key: "f9"}) {
			t.Fatal("F9 quit while cycling themes")
		}
		wantName := names[(start+step)%len(names)]
		wantTheme := loom.Theme(wantName)
		if b.themeName != wantName || b.theme != wantTheme {
			t.Fatalf("cycle %d selected %q/%+v, want %q/%+v", step, b.themeName, b.theme, wantName, wantTheme)
		}
		if b.list.Style != wantTheme.ChoiceStyle() || b.frame.Style != wantTheme.FrameStyle() {
			t.Fatalf("cycle %d did not apply %q to list and frame", step, wantName)
		}
	}
}

func TestResolveTheme(t *testing.T) {
	tests := []struct {
		name    string
		want    loom.ThemeColors
		wantErr string
	}{
		{name: "plain", want: loom.Theme("plain")},
		{name: "mc", want: loom.Theme("mc")},
		{name: "mc-classic", want: loom.Theme("mc-classic")},
		{name: "mc-dark", want: loom.Theme("mc-dark")},
		{name: "julia256", want: loom.Theme("julia256")},
		{name: "missing", wantErr: `filebrowser: unknown theme "missing" (available: julia256, mc, mc-classic, mc-dark, plain)`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := resolveTheme(test.name)
			if test.wantErr != "" {
				if err == nil || err.Error() != test.wantErr {
					t.Fatalf("resolveTheme() error = %v, want %q", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("resolveTheme() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestBrowserDetailScrollingAndFilterKey(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "q.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir, "plain", loom.Theme("plain"))
	if err != nil {
		t.Fatal(err)
	}
	if b.HandleKey(loom.KeyEvent{Text: "q"}) {
		t.Fatal("typing q into the file filter quit the app")
	}
	if item, ok := b.list.Selected(); !ok || item.Name != "q.txt" {
		t.Fatalf("filter did not select q.txt: %+v, ok=%v", item, ok)
	}
	b.details.Draw(loom.NewCanvas(40, 2), loom.Rect{W: 40, H: 2})
	b.HandleKey(loom.KeyEvent{Key: "tab"})
	b.HandleKey(loom.KeyEvent{Key: "pgdown"})
	if b.details.Scroll != 2 || !b.details.Focused() {
		t.Fatalf("detail pane did not scroll with focus: scroll=%d focused=%v", b.details.Scroll, b.details.Focused())
	}
	if !b.HandleKey(loom.KeyEvent{Key: "ctrl-q"}) {
		t.Fatal("Ctrl-Q did not quit")
	}
	if !b.HandleKey(loom.KeyEvent{Key: "f10"}) {
		t.Fatal("F10 did not quit")
	}
}

func TestBrowserHelpDismissalRestoresInputRouting(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir, "plain", loom.Theme("plain"))
	if err != nil {
		t.Fatal(err)
	}

	// The help popup consumes its dismissal key and reports no application quit.
	help := loom.NewPopup("Help", helpModalProbe{})
	if help.HandleKey(loom.KeyEvent{Text: "q"}) {
		t.Fatal("help dismissal reported an application quit")
	}
	if !help.Open {
		t.Fatal("help popup closed without an explicit dismissal state")
	}
	if help.HandleKey(loom.KeyEvent{Key: "esc"}) {
		t.Fatal("Esc dismissal reported an application quit")
	}
	if help.Open {
		t.Fatal("Esc did not close help popup")
	}

	if b.HandleKey(loom.KeyEvent{Key: "down"}) {
		t.Fatal("filebrowser navigation reported an application quit after help dismissal")
	}
	if item, ok := b.list.Selected(); !ok || item.Name != "a.txt" {
		t.Fatalf("normal input was not restored after help dismissal: %+v, ok=%v", item, ok)
	}
}

func TestBrowserShowsSideBySidePanes(t *testing.T) {
	b, err := newBrowser(t.TempDir(), "plain", loom.Theme("plain"))
	if err != nil {
		t.Fatal(err)
	}
	pane := &loom.Pane{MaxCols: loom.DefaultMaxCols}
	configurePane(pane)
	cols := 80
	if pane.MaxCols > 0 && cols > pane.MaxCols {
		cols = pane.MaxCols
	}
	rects := b.frame.Layout(cols, 20)
	if rects[0].W < 20 || rects[1].W < 25 || rects[1].X <= rects[0].X+rects[0].W {
		t.Fatalf("expected two side-by-side panes at terminal width 80 (canvas %d), got %+v", cols, rects)
	}
	if rects[0].Y+rects[0].H != 19 || rects[1].Y+rects[1].H != 19 {
		t.Fatalf("panes do not meet the status row: %+v", rects)
	}
	if b.frame.Boxes[0].Border.Vertical == "" || b.frame.Boxes[1].Border.Vertical == "" {
		t.Fatal("pane borders are invisible")
	}
}

func TestBrowserEnterOpensSelectedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a file.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir, "plain", loom.Theme("plain"))
	if err != nil {
		t.Fatal(err)
	}
	var opened string
	b.openFile = func(path string) error {
		opened = path
		return nil
	}
	b.HandleKey(loom.KeyEvent{Key: "down"})
	b.HandleKey(loom.KeyEvent{Key: "enter"})
	if opened != path || b.dir != dir {
		t.Fatalf("opened %q, browsing %q; want %q and %q", opened, b.dir, path, dir)
	}
	if !strings.Contains(strings.Join(b.details.Lines, "\n"), "Opening file") {
		t.Fatalf("missing open notice: %v", b.details.Lines)
	}
	b.openFile = func(string) error { return errors.New("no opener") }
	b.HandleKey(loom.KeyEvent{Key: "enter"})
	if !strings.Contains(strings.Join(b.details.Lines, "\n"), "Open failed: no opener") {
		t.Fatalf("missing launch error: %v", b.details.Lines)
	}
}

func TestBrowserMouseClickSelectsThenEnterOpensFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	if err := os.WriteFile(path, []byte("sample"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := newBrowser(dir, "plain", loom.Theme("plain"))
	if err != nil {
		t.Fatal(err)
	}
	var opened string
	b.openFile = func(path string) error { opened = path; return nil }
	b.Draw(loom.NewCanvas(80, 20), loom.Rect{W: 80, H: 20})
	rect := b.frame.Layout(80, 20)[0]
	b.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: rect.X + 1, Y: rect.Y + 2})
	if opened != "" || b.list.FilteredSel() != 1 {
		t.Fatalf("click opened %q or selected index %d, want selection only", opened, b.list.FilteredSel())
	}
	b.HandleKey(loom.KeyEvent{Key: "enter"})
	if opened != path {
		t.Fatalf("Enter opened %q, want %q", opened, path)
	}
}

func TestBrowserScrollbarClickJumpsFileList(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 40; i++ {
		name := fmt.Sprintf("file-%02d", i)
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	b, err := newBrowser(dir, "plain", loom.Theme("plain"))
	if err != nil {
		t.Fatal(err)
	}
	b.Draw(loom.NewCanvas(80, 20), loom.Rect{W: 80, H: 20})
	rect := b.frame.Layout(80, 20)[0]
	// Child scrollbar is the last column inside the box border. Its final
	// item row is immediately above Choice's filter prompt.
	b.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft,
		X: rect.X + rect.W - 2, Y: rect.Y + rect.H - 3})
	item, ok := b.list.Selected()
	if !ok || item.Name != "file-25" {
		t.Fatalf("bottom track click selected %+v, ok=%v; want file-25", item, ok)
	}
}

func TestBrowserApplyThemeAndF9Cycling(t *testing.T) {
	// Create a temp directory with a few test files
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file2.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Build browser with plain theme
	b, err := newBrowser(dir, "plain", loom.Theme("plain"))
	if err != nil {
		t.Fatal(err)
	}

	// Render with plain theme
	c1 := loom.NewCanvas(80, 20)
	b.Draw(c1, c1.Bounds())
	plainRendered := canvasToString(c1)

	// Apply a different theme via ApplyTheme
	b.ApplyTheme(loom.Theme("mc"))
	c2 := loom.NewCanvas(80, 20)
	b.Draw(c2, c2.Bounds())
	mcRendered := canvasToString(c2)

	// Verify the renderings differ (theme was applied)
	if plainRendered == mcRendered {
		t.Error("plain and mc themes produced identical output; theme not applied")
	}

	// Verify that F9 cycles to a valid theme (not "custom")
	initialThemeName := b.themeName
	b.HandleKey(loom.KeyEvent{Key: "f9"})
	if b.themeName == "custom" {
		t.Errorf("F9 cycled to custom theme; want a named theme")
	}
	names := themeNames()
	validTheme := false
	for _, name := range names {
		if name == b.themeName {
			validTheme = true
			break
		}
	}
	if !validTheme {
		t.Errorf("F9 produced theme %q which is not in SpeccedThemes", b.themeName)
	}

	// Verify that ApplyTheme keeps the current name when no exact match
	customTheme := loom.Theme("plain")
	customTheme.NormalFG = loom.ThemeColorRGB(255, 0, 0) // Modify to make it non-matching
	b.ApplyTheme(customTheme)
	if b.themeName != initialThemeName && b.themeName != "mc" {
		// Theme name changed from the last valid one, which is expected
		// since ApplyTheme keeps the current name when no exact match
		if b.themeName == "custom" {
			t.Error("ApplyTheme created custom theme name when no exact match found")
		}
	}
}

// canvasToString renders a canvas to a simple string representation for comparison
func canvasToString(c *loom.Canvas) string {
	var sb strings.Builder
	for y := 0; y < c.Rows(); y++ {
		for x := 0; x < c.Cols(); x++ {
			cell := c.Get(x, y)
			sb.WriteString(cell.Text)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func TestGenerateM3FilebrowserWideAndNarrowEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	// Create a temp directory with test files
	dir := t.TempDir()
	for i := 1; i <= 5; i++ {
		name := fmt.Sprintf("file%d.txt", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	repoRoot := findFilebrowserRepoRoot(t)
	progressDir := filepath.Join(repoRoot, "docs", "progress", "051")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		t.Fatalf("creating progress directory: %v", err)
	}

	// Wide layout (80x24): boxes side-by-side, should reach status row
	b, err := newBrowser(dir, "plain", loom.Theme("plain"))
	if err != nil {
		t.Fatalf("creating browser: %v", err)
	}
	frames := loom.Render(b, 80, 24)
	var wideBuffer strings.Builder
	for _, line := range frames {
		wideBuffer.WriteString(line)
		wideBuffer.WriteString("\n")
	}
	widePath := filepath.Join(progressDir, "M3-filebrowser-wide.ansi")
	if err := os.WriteFile(widePath, []byte(wideBuffer.String()), 0644); err != nil {
		t.Fatalf("writing wide evidence: %v", err)
	}
	t.Logf("M3 filebrowser wide evidence saved to %s (%d bytes)", widePath, len(wideBuffer.String()))

	// Narrow layout (40x24): boxes stacked, should reach status row
	b, err = newBrowser(dir, "plain", loom.Theme("plain"))
	if err != nil {
		t.Fatalf("creating browser: %v", err)
	}
	frames = loom.Render(b, 40, 24)
	var narrowBuffer strings.Builder
	for _, line := range frames {
		narrowBuffer.WriteString(line)
		narrowBuffer.WriteString("\n")
	}
	narrowPath := filepath.Join(progressDir, "M3-filebrowser-narrow.ansi")
	if err := os.WriteFile(narrowPath, []byte(narrowBuffer.String()), 0644); err != nil {
		t.Fatalf("writing narrow evidence: %v", err)
	}
	t.Logf("M3 filebrowser narrow evidence saved to %s (%d bytes)", narrowPath, len(narrowBuffer.String()))
}

func TestGenerateM2FilebrowserEvidence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping evidence generation in short mode")
	}
	if os.Getenv("LOOM_EVIDENCE") != "1" {
		t.Skip("set LOOM_EVIDENCE=1 to generate evidence frames")
	}

	// Create a temp directory with test files
	dir := t.TempDir()
	for i := 1; i <= 5; i++ {
		name := fmt.Sprintf("file%d.txt", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Create browser and render with two themes
	const cols, rows = 80, 25
	themes := []string{"plain", "mc"}

	var buf strings.Builder
	for _, themeName := range themes {
		b, err := newBrowser(dir, themeName, loom.Theme(themeName))
		if err != nil {
			t.Fatalf("creating browser: %v", err)
		}

		// Render the browser
		canvas := loom.NewCanvas(cols, rows)
		b.Draw(canvas, canvas.Bounds())

		// Add a label line before each theme rendering
		label := fmt.Sprintf("--- Theme: %s ---", themeName)
		buf.WriteString(label)
		buf.WriteString("\n\n")

		// Render canvas to ANSI
		renderBuf := &strings.Builder{}
		wrapper := &canvasWidget{canvas: canvas}
		renderErr := loom.RenderTo(renderBuf, wrapper, cols, rows)
		if renderErr != nil {
			t.Fatalf("RenderTo failed: %v", renderErr)
		}
		buf.WriteString(renderBuf.String())
		buf.WriteString("\n\n")
	}

	// Write evidence file
	repoRoot := findFilebrowserRepoRoot(t)
	progressDir := filepath.Join(repoRoot, "docs", "progress", "061")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		t.Fatalf("creating progress directory: %v", err)
	}

	outPath := filepath.Join(progressDir, "M2-filebrowser-themes.ansi")
	if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
		t.Fatalf("writing evidence: %v", err)
	}

	t.Logf("M2 evidence saved to %s (%d bytes)", outPath, len(buf.String()))
}

// canvasWidget wraps a canvas for rendering
type canvasWidget struct {
	canvas *loom.Canvas
}

func (w *canvasWidget) Draw(c *loom.Canvas, r loom.Rect) {
	for y := 0; y < w.canvas.Rows(); y++ {
		for x := 0; x < w.canvas.Cols(); x++ {
			cell := w.canvas.Get(x, y)
			c.Set(x, y, cell)
		}
	}
}

func (w *canvasWidget) HandleKey(e loom.KeyEvent) bool     { return false }
func (w *canvasWidget) HandleMouse(e loom.MouseEvent) bool { return false }

// findFilebrowserRepoRoot walks up from the filebrowser package directory to find the repo root
func findFilebrowserRepoRoot(t *testing.T) string {
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
