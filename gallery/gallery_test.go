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

	"gopkg.in/yaml.v3"
	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/ptytest"
	"ubunatic.com/loom/spec"
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

func TestGalleryChoiceQuitContractPTY(t *testing.T) {
	waitForAbsent := func(t *testing.T, s *ptytest.Session, text string) {
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
	bin := filepath.Join(t.TempDir(), "loom")
	if output, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	for _, names := range [][]string{{"Choice"}, {"Choice", "Popup"}} {
		t.Run(strings.Join(names, "+"), func(t *testing.T) {
			s := ptytest.Start(t, 100, 30, bin, append([]string{"widgets", "--show", "--theme", "plain"}, names...)...)
			s.WaitFor("fuzzy-browser-widget", 5*time.Second)
			for y, line := range s.Screen() {
				if i := strings.Index(line, "fuzzy-browser-widget"); i >= 0 {
					x := utf8.RuneCountInString(line[:i])
					s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", x+1, y+1, x+1, y+1)))
					break
				}
			}
			s.WaitFor("▶ fuzzy-browser-widget", 5*time.Second)
			// Enter confirms a Choice too; neither confirmation may quit a gallery.
			s.Send("\r\x1b[20~") // Enter, then F9: a new theme proves the process remains live.
			s.WaitFor("Theme: "+nextGalleryTheme(), 5*time.Second)
			s.Send("\x1b[21~") // F10 quits even when Choice consumes text keys.
			if err := s.Wait(5 * time.Second); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, key := range []string{"q", "\x1b[21~", "\x1b"} {
		t.Run(fmt.Sprintf("exit-%q", key), func(t *testing.T) {
			s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "--theme", "plain", "ProgressBar")
			s.WaitFor("Theme: plain", 5*time.Second)
			s.Send(key)
			if err := s.Wait(5 * time.Second); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("popup consumes first escape", func(t *testing.T) {
		s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "--theme", "plain", "Popup")
		s.WaitFor("Gallery popup", 5*time.Second)
		s.Send("\x1b")
		waitForAbsent(t, s, "Gallery popup")
		s.Send("\x1b[20~")
		s.WaitFor("Theme: "+nextGalleryTheme(), 5*time.Second)
		s.Send("\x1b")
		if err := s.Wait(5 * time.Second); err != nil {
			t.Fatal(err)
		}
	})
}

func nextGalleryTheme() string {
	names := loom.ThemeNames()
	for i, name := range names {
		if name == "plain" {
			return names[(i+1)%len(names)]
		}
	}
	return ""
}

func TestGalleryTextInputFocusPTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	if output, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	for _, target := range []string{"Placeholder:", "Masked:", "Enter"} {
		t.Run(target, func(t *testing.T) {
			s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "TextInput", "Choice")
			s.WaitFor("Masked:", 5*time.Second)
			if target == "Enter" {
				s.Send("\rZ")
			} else {
				for y, line := range s.Screen() {
					if i := strings.Index(line, target); i >= 0 {
						x := utf8.RuneCountInString(line[:i]) + utf8.RuneCountInString(target) + 2
						s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dmZ", x+1, y+1, x+1, y+1)))
						break
					}
				}
			}
			if target == "Masked:" {
				s.WaitFor(strings.Repeat("•", 14), 5*time.Second)
			} else {
				s.WaitFor("Placeholder:  Z", 5*time.Second)
			}
			if strings.Contains(strings.Join(s.Screen(), "\n"), "Ada LovelaceZ") {
				t.Fatal("typing reached the first field")
			}
		})
	}
}

func TestGalleryTimerControlsPTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	if output, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "Timer", "Choice")
	s.WaitFor("04:", 5*time.Second)
	s.WaitFor("04:10", 5*time.Second)
	click := func(text string) {
		t.Helper()
		s.WaitFor(text, 5*time.Second)
		for y, line := range s.Screen() {
			if i := strings.Index(line, text); i >= 0 {
				x := utf8.RuneCountInString(line[:i])
				s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", x+1, y+1, x+1, y+1)))
				return
			}
		}
	}
	click("[Stop]")
	time.Sleep(200 * time.Millisecond)
	stopped := strings.Join(s.Screen(), "\n")
	time.Sleep(1200 * time.Millisecond)
	if strings.Join(s.Screen(), "\n") != stopped {
		t.Fatal("Stop button did not pause timer")
	}
	// Reset followed by a wait proves reset restores the duration and stops ticking.
	click("[Reset]")
	s.WaitFor("04:12", 5*time.Second)
	time.Sleep(1200 * time.Millisecond)
	if !strings.Contains(strings.Join(s.Screen(), "\n"), "04:12") {
		t.Fatal("reset timer kept running")
	}
	click("[Start]")
	s.WaitFor("04:10", 5*time.Second)
	s.Send("r")
	s.WaitFor("04:12", 5*time.Second)
	s.Send(" ")
	s.WaitFor("04:10", 5*time.Second)
	s.Send(" ")
	time.Sleep(200 * time.Millisecond)
	paused := strings.Join(s.Screen(), "\n")
	time.Sleep(1200 * time.Millisecond)
	if strings.Join(s.Screen(), "\n") != paused {
		t.Fatal("Space did not stop timer")
	}
}

func TestGalleryThemeFooterSurfacePTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	if output, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	for _, name := range []string{"Choice", "FilePicker", "Media", "Table", "Tree"} {
		t.Run(name, func(t *testing.T) {
			s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "--theme", "plain", "-W", "100", "-H", "24", name)
			s.WaitFor("Theme: plain", 5*time.Second)
			s.Send("\x1b[20~")
			themeName := nextGalleryTheme()
			s.WaitFor("Theme: "+themeName, 5*time.Second)
			for y, line := range s.Screen() {
				if strings.Contains(line, "Theme: "+themeName) {
					for x := 70; x < 100; x++ {
						if !ptyColorMatches(s.Cell(x, y).Style.BG, loom.Theme(themeName).NormalBG.Color()) {
							t.Fatalf("footer column %d inherited child background: %+v", x, s.Cell(x, y).Style)
						}
					}
					return
				}
			}
			t.Fatal("footer missing")
		})
	}
}

func TestGalleryDialogReselectionPTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	if output, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "Dialog", "ProgressBar")
	clickTab := func(text string) {
		t.Helper()
		for y, line := range s.Screen() {
			if i := strings.Index(line, text); i >= 0 {
				x := utf8.RuneCountInString(line[:i])
				s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", x+1, y+1, x+1, y+1)))
				return
			}
		}
		t.Fatalf("tab %q missing", text)
	}
	for _, away := range []bool{true, false} {
		s.WaitFor("Save changes", 5*time.Second)
		s.Send("\r")
		deadline := time.Now().Add(5 * time.Second)
		for strings.Contains(strings.Join(s.Screen(), "\n"), "Save changes") && time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond)
		}
		if strings.Contains(strings.Join(s.Screen(), "\n"), "Save changes") {
			t.Fatal("dialog did not close")
		}
		if away {
			clickTab("ProgressBar")
			s.WaitFor("/100 files", 5*time.Second)
		}
		clickTab("Dialog")
		s.WaitFor("Save changes", 5*time.Second)
	}
}

func TestGalleryTableClickSelectionPTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	if output, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "Table", "Choice")
	s.WaitFor("Unit tests", 5*time.Second)
	locate := func(text string) (int, int) {
		t.Helper()
		for y, line := range s.Screen() {
			if i := strings.Index(line, text); i >= 0 {
				return utf8.RuneCountInString(line[:i]), y
			}
		}
		t.Fatalf("%q missing", text)
		return -1, -1
	}
	sx, sy := locate("Compile")
	selected := s.Cell(sx, sy).Style
	x, y := locate("Unit tests")
	s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", x+1, y+1, x+1, y+1)))
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if s.Cell(x, y).Style == selected && s.Cell(sx, sy).Style != selected {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("click did not transfer selection to second data row")
}

func TestGalleryMenuAcceleratorTogglePTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	if output, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "MenuBar", "Choice")
	s.WaitFor("Autosave", 5*time.Second)
	s.Send("\x1b") // close the initially open demo menu
	waitAbsent := func(text string) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if !strings.Contains(strings.Join(s.Screen(), "\n"), text) {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatalf("%q remained visible", text)
	}
	waitAbsent("Autosave")
	s.Send("\x1bf")
	s.WaitFor("Autosave", 5*time.Second)
	s.Send("\x1bf")
	waitAbsent("Autosave")
	s.Send("\x1bf")
	s.WaitFor("Autosave", 5*time.Second)
	s.Send("\x1be")
	s.WaitFor("Undo", 5*time.Second)
	waitAbsent("Autosave")
	s.Send("\x1be")
	waitAbsent("Undo")
	s.Send("\x1bh")
	s.WaitFor("Keyboard shortcuts", 5*time.Second)
	s.Send("\x1bh")
	waitAbsent("Keyboard shortcuts")
}

func TestGalleryPaintCanvasTunePTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	if output, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "PaintCanvas", "Choice")
	s.WaitFor("Smoothing: 0", 5*time.Second)
	for y, line := range s.Screen() {
		if i := strings.Index(line, "[Tune]"); i >= 0 {
			x := utf8.RuneCountInString(line[:i])
			s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", x+1, y+1, x+1, y+1)))
			break
		}
	}
	s.WaitFor("Smoothing: 1", 5*time.Second)
	s.Send("t")
	s.WaitFor("Smoothing: 2", 5*time.Second)
	s.Send("t")
	s.WaitFor("Smoothing: 0", 5*time.Second)
}

func TestGalleryNumberInputRangeAlignmentPTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	if output, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "NumberInput")
	s.WaitFor("◂    7.50 ▸", 5*time.Second)
	s.Send("\r\x1b[H" + strings.Repeat("\x1b[3~", 4) + "-100.00\r")
	s.WaitFor("◂ -100.00 ▸", 5*time.Second)
	s.Send("\x1b[D")
	s.Send("\x1b[20~")
	s.WaitFor("Theme: "+nextGalleryTheme(), 5*time.Second)
	s.WaitFor("◂ -100.00 ▸", 5*time.Second)
	s.Send("\r\x1b[H" + strings.Repeat("\x1b[3~", 7) + "100.00\r")
	s.WaitFor("◂  100.00 ▸", 5*time.Second)
	s.Send("\x1b[C")
	s.Send("\r\x1b[H" + strings.Repeat("\x1b[3~", 6) + "5.00\r")
	s.WaitFor("◂    5.00 ▸", 5*time.Second)
}

func TestGalleryNumberInputMousePTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	if output, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "NumberInput")
	s.WaitFor("◂    7.50 ▸", 5*time.Second)

	locate := func(text string) (int, int) {
		t.Helper()
		for y, line := range s.Screen() {
			if i := strings.Index(line, text); i >= 0 {
				return utf8.RuneCountInString(line[:i]), y
			}
		}
		t.Fatalf("%q missing", text)
		return -1, -1
	}

	// Click right stepper (▸) to step up.
	rx, ry := locate("▸")
	s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", rx+1, ry+1, rx+1, ry+1)))
	s.WaitFor("◂    8.00 ▸", 5*time.Second)

	// Click left stepper (◂) to step down.
	lx, ly := locate("◂")
	s.SendRaw([]byte(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", lx+1, ly+1, lx+1, ly+1)))
	s.WaitFor("◂    7.50 ▸", 5*time.Second)

	// Mouse wheel up to step up.
	s.SendRaw([]byte(fmt.Sprintf("\x1b[<64;%d;%dM", rx+1, ry+1)))
	s.WaitFor("◂    8.00 ▸", 5*time.Second)

	// Mouse wheel down to step down.
	s.SendRaw([]byte(fmt.Sprintf("\x1b[<65;%d;%dM", rx+1, ry+1)))
	s.WaitFor("◂    7.50 ▸", 5*time.Second)
}

func TestGalleryProgressBarFillFadeRepeat(t *testing.T) {
	w, err := New("ProgressBar")
	if err != nil {
		t.Fatal(err)
	}
	ticker, ok := w.(loom.Ticker)
	if !ok || ticker.TickInterval() <= 0 {
		t.Fatal("demo has no animation cadence")
	}
	value := w.(interface{ Value() float64 })
	for cycle := 0; cycle < 3; cycle++ {
		for step := 0; step <= 100; step += 10 {
			if got := value.Value(); got != float64(step) {
				t.Fatalf("cycle %d: value=%g, want %d", cycle, got, step)
			}
			if !strings.Contains(strings.Join(loom.Render(w, 80, 4), ""), fmt.Sprintf("%d%%", step)) {
				t.Fatalf("step %d not rendered", step)
			}
			if step < 100 {
				ticker.Tick(time.Time{})
			}
		}
		ticker.Tick(time.Time{})
		cells := renderDemoCells(w, 4)
		if !cells[0][0].Style.Dim || value.Value() != 100 {
			t.Fatal("completed bar did not fade at 100")
		}
		ticker.Tick(time.Time{})
		for _, row := range renderDemoCells(w, 4) {
			for _, cell := range row {
				if strings.TrimSpace(cell.Text) != "" {
					t.Fatal("fade did not disappear before restart")
				}
			}
		}
		ticker.Tick(time.Time{})
	}
}

func TestGalleryMediaRetainsViewportAcrossTabSwitches(t *testing.T) {
	w := demos["Media"]()
	tabs := loom.NewTabs(loom.Tab{Title: "Media", Widget: w}, loom.Tab{Title: "Choice", Widget: demos["Choice"]()})
	tabs.SetKeys(loom.TabsKeys{Previous: "shift-tab", Next: "tab"})
	c := loom.NewCanvas(50, 18)
	r := loom.Rect{X: 3, Y: 2, W: 40, H: 14}
	draw := func() []loom.Cell {
		tabs.Draw(c, r)
		var cells []loom.Cell
		for y := r.Y + 2; y < r.Y+r.H; y++ {
			for x := r.X; x < r.X+r.W; x++ {
				cells = append(cells, c.Get(x, y))
			}
		}
		return cells
	}
	draw()
	tabs.ConsumeKey(loom.KeyEvent{Text: "+"})
	for _, key := range []string{"right", "down", "left", "up", "right", "down"} {
		before := draw()
		if res := tabs.ConsumeKey(loom.KeyEvent{Key: key}); !res.Consumed || tabs.Focus() != 0 {
			t.Fatalf("media pan key %q escaped to Tabs: %#v focus=%d", key, res, tabs.Focus())
		}
		if reflect.DeepEqual(before, draw()) {
			t.Fatalf("media pan key %q did not move viewport", key)
		}
	}
	beforeDrag := draw()
	for _, e := range []loom.MouseEvent{
		{Action: loom.MousePress, Button: loom.MouseLeft, X: 20, Y: 5},
		{Action: loom.MouseDrag, Button: loom.MouseLeft, X: 15, Y: 7},
		{Action: loom.MouseDrag, Button: loom.MouseLeft, X: 9, Y: 13},
		{Action: loom.MouseDrag, Button: loom.MouseLeft, X: -3, Y: 0},
		{Action: loom.MouseRelease, Button: loom.MouseLeft, X: -3, Y: 0},
	} {
		if res := tabs.ConsumeMouse(e); !res.Consumed || res.Quit || tabs.Focus() != 0 {
			t.Fatalf("media drag escaped to Tabs: %#v result=%#v", e, res)
		}
	}
	panned := draw()
	if reflect.DeepEqual(beforeDrag, panned) {
		t.Fatal("drag did not change viewport")
	}
	tabs.ConsumeKey(loom.KeyEvent{Key: "tab"})
	draw()
	tabs.ConsumeKey(loom.KeyEvent{Key: "shift-tab"})
	if tabs.Focus() != 0 || tabs.Tabs[0].Widget != w || !reflect.DeepEqual(panned, draw()) {
		t.Fatal("tab switching changed media instance, zoom or viewport")
	}
}

func TestGalleryMediaZoomPanPTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	if output, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	s := ptytest.Start(t, 80, 24, bin, "widgets", "--show", "-W", "40", "-H", "16", "Media", "Choice")
	latestFrame := func() [][]ptytest.Cell {
		frames := s.CellFrames()
		if len(frames) == 0 {
			return nil
		}
		return frames[len(frames)-1]
	}
	waitFrame := func(text string) [][]ptytest.Cell {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			frame := latestFrame()
			for _, row := range frame {
				var line strings.Builder
				for _, cell := range row {
					line.WriteRune(cell.Rune)
				}
				if strings.Contains(line.String(), text) {
					return frame
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatalf("%q missing from completed PTY frame:\n%s", text, strings.Join(s.Screen(), "\n"))
		return nil
	}
	// Sample image pixels and colors from one completed synchronized frame.
	imageState := func(frame [][]ptytest.Cell) []ptytest.Cell {
		var cells []ptytest.Cell
		if len(frame) < 10 {
			return nil
		}
		for y := 3; y < 10; y++ {
			cells = append(cells, frame[y][5:35]...)
		}
		return cells
	}
	waitChanged := func(before []ptytest.Cell) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if frame := latestFrame(); frame != nil && !reflect.DeepEqual(imageState(frame), before) {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatal("media viewport did not move on PTY")
	}
	waitFrame("1.00x")
	s.Send("++")
	before := imageState(waitFrame("1.56x"))
	s.Send("\x1b[C\x1b[B")
	waitChanged(before)
	before = imageState(latestFrame())
	// Drag across controls and beyond the panel, then release on the tab bar.
	s.SendRaw([]byte("\x1b[<0;21;6M\x1b[<32;16;8M\x1b[<32;10;14M\x1b[<32;45;2M\x1b[<0;45;2m"))
	waitChanged(before)
	// A changed zoom label is a barrier after all queued drags and release.
	// It also verifies that zooming while panned preserves the view state.
	s.Send("+")
	panned := imageState(waitFrame("1.95x"))
	s.Send("\t")
	waitFrame("filebrowser-widget")
	s.Send("\x1b[Z")
	if !reflect.DeepEqual(imageState(waitFrame("1.95x")), panned) {
		t.Fatal("tab switching changed media viewport on PTY")
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
		"SearchBar": {Text: "x"},
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
		"FilePicker":  {locate: func([]string) (int, int) { return 3, 1 }, state: func(w loom.Widget) any { entry, _ := w.(*loom.FilePicker).Selected(); return entry.Name }},
		"Form":        {locate: func([]string) (int, int) { return 2, 2 }, state: func(w loom.Widget) any { return w.(*loom.Form).FocusIndex() }},
		"MenuBar":     {locate: func([]string) (int, int) { return 1, 0 }, state: func(w loom.Widget) any { return w.(*loom.MenuBar).Open }},
		"NumberInput": {locate: func(rows []string) (int, int) { return runeColumn(rows[0], "▸"), 0 }, state: func(w loom.Widget) any { return *w.(*loom.NumberInput).Value }},
		"Paginator":   {locate: func([]string) (int, int) { return 6, 0 }, state: func(w loom.Widget) any { return w.(*loom.Paginator).Page }},
		"Tabs":        {locate: func(rows []string) (int, int) { return runeColumn(rows[1], "Details"), 1 }, state: func(w loom.Widget) any { return w.(*loom.Tabs).Focus() }},
		"Tree":        {locate: func([]string) (int, int) { return 0, 0 }, state: func(w loom.Widget) any { return len(w.(*loom.Tree).VisibleNodes()) }},
		"Toggle":      {locate: func([]string) (int, int) { return 1, 0 }, state: func(w loom.Widget) any { return w.(*loom.Toggle).String() }},
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
	build := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom")
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
	// All also contains app.go. Wait for content unique to the standalone
	// Tree so disclosure coordinates come from the newly selected panel.
	s.WaitFor("tree.go", 5*time.Second)
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

func TestWidgetsPTYSizeAndF9ThemePropagation(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	build := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom")
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
	// demo so one F9 redraw can be checked across the tab bar, modal frame, and
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

	s.SendRaw([]byte("\x1b[20~"))
	s.WaitFor("Theme: "+next, 5*time.Second)
	if got := s.Cell(tabX, tabY).Style; got == oldTabStyle || !ptyColorMatches(got.FG, loom.Theme(next).HeaderFG.Color()) {
		t.Fatalf("tab bar cell did not recolor after F9: %v", got)
	}
	if got := s.Cell(frameX, frameY).Style; got == oldFrameStyle || !ptyColorMatches(got.BG, loom.Theme(next).NormalBG.Color()) {
		t.Fatalf("modal frame cell did not recolor after F9: %v", got)
	}
	if got := s.Cell(width-1, footerY-2).Style; got == oldBackgroundStyle || !ptyColorMatches(got.BG, loom.Theme(next).NormalBG.Color()) {
		t.Fatalf("gallery background did not recolor after F9: got %v", got)
	}
	s.Send("\t")
	s.WaitFor("filebrowser-widget", 5*time.Second)
	demoX, demoY = findText("filebrowser-widget")
	if got := s.Cell(demoX, demoY).Style; got == oldDemoStyle || !ptyColorMatches(got.FG, loom.Theme(next).NormalFG.Color()) {
		t.Fatalf("active demo cell did not use the new theme after F9: got %v", got)
	}
}

func TestRicherDemosPTY(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	build := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build loom binary: %v\n%s", err, output)
	}
	checks := []struct {
		name, visible string
		animated      bool
	}{
		{"ProgressBar", "/100 files", true}, {"Spinner", "Syncing workspace", true},
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
		s.WaitFor("◂    7.50 ▸", 5*time.Second)
		before := strings.Join(s.Screen(), "\n")
		start := strings.Index(before, "◂    7.50 ▸")
		if start < 0 {
			t.Fatal("initial number input missing")
		}
		s.Send("\x1b[C")
		s.WaitFor("◂    8.00 ▸", 5*time.Second)
		after := strings.Join(s.Screen(), "\n")
		left, right := strings.Index(after, "◂    8.00"), strings.Index(after, "▸")
		if left < 0 || right < 0 || utf8.RuneCountInString(after[left:right]) != utf8.RuneCountInString("◂    7.50 ") {
			t.Fatalf("fixed-width NumberInput changed its footprint after stepping:\n%s", after)
		}
	})
	t.Run("StopwatchControls", func(t *testing.T) {
		s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "Stopwatch")
		s.WaitFor("00:", 5*time.Second)
		s.Send(" ")
		time.Sleep(200 * time.Millisecond)
		paused := strings.Join(s.Screen(), "\n")
		time.Sleep(1100 * time.Millisecond)
		if strings.Join(s.Screen(), "\n") != paused {
			t.Fatal("Space did not pause the stopwatch")
		}
		s.Send("r")
		s.WaitFor("00:00", 5*time.Second)
		s.Send(" ")
		s.WaitFor("00:01", 5*time.Second)
	})
}

func ptyColorMatches(got ptytest.Color, want loom.Color) bool {
	gr, gg, gb, gok := got.RGB()
	wr, wg, wb, wok := want.RGB()
	return gok && wok && gr == wr && gg == wg && gb == wb
}

func TestPaintCanvasPTYMouseDragDrawsBrailleLine(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "loom")
	build := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom")
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
	build := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build loom binary: %v\n%s", err, output)
	}
	start := func(t *testing.T, name string) *ptytest.Session {
		t.Helper()
		s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "--theme", "plain", name)
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
		s.WaitFor("Ada Lovelace", 5*time.Second)
		s.Send("Z")
		s.WaitFor("Ada LovelaceZ", 5*time.Second)
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
	if output, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/cmd/loom").CombinedOutput(); err != nil {
		t.Fatalf("build loom: %v\n%s", err, output)
	}
	start := func(t *testing.T, name string) *ptytest.Session {
		t.Helper()
		// Nest demos in outer tabs so their draw origins are nonzero.
		s := ptytest.Start(t, 100, 30, bin, "widgets", "--show", "--theme", "plain", name, "Choice")
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

func TestAllTabInitialActiveAndLayout(t *testing.T) {
	tabs := NewAll()
	if got := tabs.Focus(); got != 0 {
		t.Fatalf("initial active tab = %d, want 0", got)
	}
	if len(tabs.Tabs) != len(Names())+1 {
		t.Fatalf("len(tabs.Tabs) = %d, want %d", len(tabs.Tabs), len(Names())+1)
	}
	if tabs.Tabs[0].Title != "All" {
		t.Fatalf("first tab title = %q, want All", tabs.Tabs[0].Title)
	}
	for i, name := range Names() {
		if tabs.Tabs[i+1].Title != name {
			t.Fatalf("tab %d title = %q, want %q", i+1, tabs.Tabs[i+1].Title, name)
		}
	}

	// Verify cycling through tabs
	tabs.ConsumeKey(loom.KeyEvent{Key: "tab"})
	if tabs.Focus() != 1 {
		t.Fatalf("after Tab, active tab = %d, want 1", tabs.Focus())
	}
	tabs.ConsumeKey(loom.KeyEvent{Key: "shift-tab"})
	if tabs.Focus() != 0 {
		t.Fatalf("after Shift-Tab, active tab = %d, want 0", tabs.Focus())
	}

	// Verify 3-column composite layout on the All tab
	grid, ok := tabs.Tabs[0].Widget.(*loom.Grid)
	if !ok {
		t.Fatalf("first tab widget is %T, want *loom.Grid", tabs.Tabs[0].Widget)
	}
	if grid.Cols != 3 {
		t.Fatalf("grid columns = %d, want 3", grid.Cols)
	}
	if len(grid.Children) != 24 {
		t.Fatalf("grid child count = %d, want 24", len(grid.Children))
	}

	c := loom.NewCanvas(100, 30)
	tabs.Draw(c, c.Bounds())

	// Verify exact 3-column cell boundaries (r.X + k*cellW)
	r0 := grid.ChildRect(0)
	panelX := r0.X
	cellW := r0.W
	cellH := r0.H
	if cellW <= 0 || cellH <= 0 {
		t.Fatalf("cell size invalid: %dx%d", cellW, cellH)
	}
	for i := range grid.Children {
		k := i % grid.Cols
		row := i / grid.Cols
		wantX := panelX + k*cellW
		wantY := r0.Y + row*cellH
		rect := grid.ChildRect(i)
		if rect.X != wantX || rect.Y != wantY || rect.W != cellW || rect.H != cellH {
			t.Fatalf("child %d rect = %+v, want X=%d Y=%d W=%d H=%d", i, rect, wantX, wantY, cellW, cellH)
		}
	}

	plainRows := make([]string, 30)
	for y := 0; y < 30; y++ {
		for x := 0; x < 100; x++ {
			plainRows[y] += c.Get(x, y).Text
		}
	}

	findCol := func(line, text string) int {
		idx := strings.Index(line, text)
		if idx < 0 {
			return -1
		}
		return utf8.RuneCountInString(line[:idx])
	}

	// Verify row 0 widgets: Button, Toggle, Checkbox in 3 distinct columns
	var row0Idx = -1
	for y, line := range plainRows {
		if strings.Contains(line, "Click Me") && strings.Contains(line, "Toggle") && strings.Contains(line, "Checkbox") {
			row0Idx = y
			break
		}
	}
	if row0Idx < 0 {
		t.Fatalf("row with Button, Toggle, Checkbox not found in render:\n%s", strings.Join(plainRows, "\n"))
	}
	line0 := plainRows[row0Idx]
	colBtn := findCol(line0, "Click Me")
	colTgl := findCol(line0, "Toggle")
	colChk := findCol(line0, "Checkbox")
	if !(colBtn < colTgl && colTgl < colChk) {
		t.Fatalf("expected 3 columns (Button < Toggle < Checkbox), got cols %d, %d, %d", colBtn, colTgl, colChk)
	}

	// Verify row 1 widgets: NumberInput, Badge, PillCluster
	var row1Idx = -1
	for y, line := range plainRows {
		if strings.Contains(line, "42.00") && strings.Contains(line, "Badge") && strings.Contains(line, "API") {
			row1Idx = y
			break
		}
	}
	if row1Idx < 0 {
		t.Fatalf("row with NumberInput, Badge, PillCluster not found in render:\n%s", strings.Join(plainRows, "\n"))
	}
	line1 := plainRows[row1Idx]
	colNum := findCol(line1, "42.00")
	colBdg := findCol(line1, "Badge")
	colPill := findCol(line1, "API")
	if !(colNum < colBdg && colBdg < colPill) {
		t.Fatalf("expected 3 columns (NumberInput < Badge < PillCluster), got cols %d, %d, %d", colNum, colBdg, colPill)
	}
}

func TestAllTabKeyboardInteractions(t *testing.T) {
	tabs := NewAll()
	grid, ok := tabs.Tabs[0].Widget.(*loom.Grid)
	if !ok {
		t.Fatalf("tab widget is %T, want *loom.Grid", tabs.Tabs[0].Widget)
	}
	c := loom.NewCanvas(100, 30)
	tabs.Draw(c, c.Bounds())

	btn, ok := grid.Children[0].(*buttonDemo)
	if !ok {
		t.Fatalf("child 0 is %T, want *buttonDemo", grid.Children[0])
	}
	toggle, ok := grid.Children[1].(*loom.Toggle)
	if !ok {
		t.Fatalf("child 1 is %T, want *loom.Toggle", grid.Children[1])
	}
	cb, ok := grid.Children[2].(*checkboxDemo)
	if !ok {
		t.Fatalf("child 2 is %T, want *checkboxDemo", grid.Children[2])
	}
	numInput, ok := grid.Children[3].(*loom.NumberInput)
	if !ok {
		t.Fatalf("child 3 is %T, want *loom.NumberInput", grid.Children[3])
	}
	badge, ok := grid.Children[4].(*badgeDemo)
	if !ok {
		t.Fatalf("child 4 is %T, want *badgeDemo", grid.Children[4])
	}

	// 1. Initial focus is cell 0 (Button). Pressing Enter clicks it.
	if grid.Focus() != 0 {
		t.Fatalf("initial grid focus = %d, want 0", grid.Focus())
	}
	tabs.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if btn.clicked != 1 {
		t.Fatalf("after Enter, button clicks = %d, want 1", btn.clicked)
	}

	// 2. Press "right" -> moves to cell 1 (Toggle). Press Enter -> toggles.
	tabs.ConsumeKey(loom.KeyEvent{Key: "right"})
	if grid.Focus() != 1 {
		t.Fatalf("after right, grid focus = %d, want 1", grid.Focus())
	}
	if !*toggle.Value {
		t.Fatal("toggle value initially should be true")
	}
	tabs.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if *toggle.Value {
		t.Fatal("after Enter, toggle value should be false")
	}

	// 3. Press "right" -> moves to cell 2 (Checkbox). Press Space -> toggles.
	tabs.ConsumeKey(loom.KeyEvent{Key: "right"})
	if grid.Focus() != 2 {
		t.Fatalf("after right, grid focus = %d, want 2", grid.Focus())
	}
	if !cb.checked {
		t.Fatal("checkbox initially should be true")
	}
	tabs.ConsumeKey(loom.KeyEvent{Key: "space"})
	if cb.checked {
		t.Fatal("after Space, checkbox should be false")
	}

	// 4. Navigate down and left:
	tabs.ConsumeKey(loom.KeyEvent{Key: "down"})
	if grid.Focus() != 5 {
		t.Fatalf("after down, grid focus = %d, want 5", grid.Focus())
	}
	tabs.ConsumeKey(loom.KeyEvent{Key: "left"})
	if grid.Focus() != 4 {
		t.Fatalf("after left, grid focus = %d, want 4", grid.Focus())
	}
	if badge.status != "Active" {
		t.Fatalf("badge status = %q, want Active", badge.status)
	}
	tabs.ConsumeKey(loom.KeyEvent{Key: "space"})
	if badge.status != "Idle" {
		t.Fatalf("after space, badge status = %q, want Idle", badge.status)
	}

	// 5. Navigate left to cell 3 (NumberInput) and step it:
	tabs.ConsumeKey(loom.KeyEvent{Key: "left"})
	if grid.Focus() != 3 {
		t.Fatalf("after left, grid focus = %d, want 3", grid.Focus())
	}
	beforeVal := *numInput.Value
	tabs.ConsumeKey(loom.KeyEvent{Key: "right"})
	if *numInput.Value != beforeVal+1.0 {
		t.Fatalf("after right, NumberInput value = %f, want %f", *numInput.Value, beforeVal+1.0)
	}
}

func TestAllTabMouseInteractions(t *testing.T) {
	tabs := NewAll()
	grid, ok := tabs.Tabs[0].Widget.(*loom.Grid)
	if !ok {
		t.Fatalf("tab widget is %T, want *loom.Grid", tabs.Tabs[0].Widget)
	}
	c := loom.NewCanvas(100, 30)
	tabs.Draw(c, c.Bounds())

	btn, ok := grid.Children[0].(*buttonDemo)
	if !ok {
		t.Fatalf("child 0 is %T, want *buttonDemo", grid.Children[0])
	}
	toggle, ok := grid.Children[1].(*loom.Toggle)
	if !ok {
		t.Fatalf("child 1 is %T, want *loom.Toggle", grid.Children[1])
	}
	cb, ok := grid.Children[2].(*checkboxDemo)
	if !ok {
		t.Fatalf("child 2 is %T, want *checkboxDemo", grid.Children[2])
	}
	numInput, ok := grid.Children[3].(*loom.NumberInput)
	if !ok {
		t.Fatalf("child 3 is %T, want *loom.NumberInput", grid.Children[3])
	}

	findLoc := func(target string) (int, int) {
		for y := 0; y < 30; y++ {
			var row strings.Builder
			for x := 0; x < 100; x++ {
				row.WriteString(c.Get(x, y).Text)
			}
			s := row.String()
			idx := strings.Index(s, target)
			if idx >= 0 {
				return utf8.RuneCountInString(s[:idx]), y
			}
		}
		return -1, -1
	}

	// 1. Click Button
	btnX, btnY := findLoc("[ Click Me ]")
	if btnX < 0 {
		t.Fatal("Button not found on canvas")
	}
	initialClicks := btn.clicked
	tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: btnX + 1, Y: btnY})
	if btn.clicked != initialClicks+1 {
		t.Fatalf("after click, button clicks = %d, want %d", btn.clicked, initialClicks+1)
	}
	if grid.Focus() != 0 {
		t.Fatalf("after button click, grid focus = %d, want 0", grid.Focus())
	}

	// 2. Click Toggle inside [✓] mark
	tglX, tglY := findLoc("[✓] Toggle")
	if tglX < 0 {
		t.Fatal("Toggle not found on canvas")
	}
	tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: tglX + 1, Y: tglY})
	if *toggle.Value {
		t.Fatal("after toggle click, value should be false")
	}
	if grid.Focus() != 1 {
		t.Fatalf("after toggle click, grid focus = %d, want 1", grid.Focus())
	}

	// 3. Click Checkbox
	chkX, chkY := findLoc("[x] Checkbox")
	if chkX < 0 {
		t.Fatal("Checkbox not found on canvas")
	}
	tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: chkX + 1, Y: chkY})
	if cb.checked {
		t.Fatal("after checkbox click, checked should be false")
	}
	if grid.Focus() != 2 {
		t.Fatalf("after checkbox click, grid focus = %d, want 2", grid.Focus())
	}

	// 4. Click NumberInput stepper ▸
	tabs.Draw(c, c.Bounds())
	stepperX, stepperY := findLoc("▸")
	if stepperX < 0 {
		t.Fatal("stepper ▸ not found on canvas")
	}
	beforeNum := *numInput.Value
	tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: stepperX, Y: stepperY})
	if *numInput.Value != beforeNum+1.0 {
		t.Fatalf("after stepper click, NumberInput = %f, want %f", *numInput.Value, beforeNum+1.0)
	}
	if grid.Focus() != 3 {
		t.Fatalf("after numberinput click, grid focus = %d, want 3", grid.Focus())
	}

	// 5. Mouse scroll on NumberInput
	tabs.Draw(c, c.Bounds())
	numX, numY := findLoc("43.00")
	if numX < 0 {
		t.Fatal("43.00 not found on canvas")
	}
	tabs.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollUp, Button: loom.MouseNone, X: numX, Y: numY})
	if *numInput.Value != beforeNum+2.0 {
		t.Fatalf("after scroll up, NumberInput = %f, want %f", *numInput.Value, beforeNum+2.0)
	}
}

func TestNewAllDemoDirect(t *testing.T) {
	widget, err := New("All")
	if err != nil {
		t.Fatalf("New(\"All\") error = %v", err)
	}
	grid, ok := widget.(*loom.Grid)
	if !ok {
		t.Fatalf("New(\"All\") returned %T, want *loom.Grid", widget)
	}
	if grid.Cols != 3 {
		t.Fatalf("grid.Cols = %d, want 3", grid.Cols)
	}
	if len(grid.Children) != 24 {
		t.Fatalf("grid children = %d, want 24", len(grid.Children))
	}
}

func TestAllTabAddedWidgetsStayInCellsAndRouteInput(t *testing.T) {
	tabs := NewAll()
	grid := tabs.Tabs[0].Widget.(*loom.Grid)
	c := loom.NewCanvas(100, 40)
	tabs.Draw(c, c.Bounds())

	added := []struct {
		index int
		name  string
	}{
		{12, "TextInput"}, {13, "Choice"}, {14, "DatePicker"},
		{15, "KeyHelp"}, {16, "MenuBar"}, {17, "Chart"},
		{18, "Table"}, {19, "Tree"}, {20, "Dialog"},
		{21, "Popup"}, {22, "TextArea"}, {23, "Viewport"},
	}
	clickCell := func(index int) {
		rect := grid.ChildRect(index)
		grid.ConsumeMouse(loom.MouseEvent{
			Action: loom.MousePress, Button: loom.MouseLeft,
			X: rect.X - gridRectX(grid) + rect.W/2,
			Y: rect.Y - gridRectY(grid) + rect.H/2,
		})
	}
	for _, item := range added {
		t.Run(item.name, func(t *testing.T) {
			rect := grid.ChildRect(item.index)
			if rect.W <= 0 || rect.H <= 0 {
				t.Fatalf("%s cell has invalid bounds: %+v", item.name, rect)
			}
			// Draw the widget into an isolated cell and ensure no writes escape it.
			cellCanvas := loom.NewCanvas(rect.W+4, rect.H+4)
			cellCanvas.Fill(cellCanvas.Bounds(), loom.Cell{Text: "?"})
			grid.Children[item.index].Draw(cellCanvas, loom.Rect{X: 2, Y: 2, W: rect.W, H: rect.H})
			drew := false
			for y := 0; y < rect.H+4; y++ {
				for x := 0; x < rect.W+4; x++ {
					if x >= 2 && x < rect.W+2 && y >= 2 && y < rect.H+2 {
						if cellCanvas.Get(x, y).Text != "?" {
							drew = true
						}
						continue
					}
					if got := cellCanvas.Get(x, y).Text; got != "?" {
						t.Fatalf("%s drew outside its cell at (%d,%d): %q", item.name, x, y, got)
					}
				}
			}
			if !drew {
				t.Fatalf("%s drew nothing inside its cell", item.name)
			}

			clickCell(item.index)
			if grid.Focus() != item.index {
				t.Fatalf("focus = %d, want cell %d", grid.Focus(), item.index)
			}
			if focusable, ok := grid.Children[item.index].(loom.Focusable); ok && !focusable.Focused() {
				t.Fatalf("%s did not receive focus", item.name)
			}
			grid.ConsumeKey(loom.KeyEvent{Key: "down"})

			clickCell(0)
			clickCell(item.index)
			if grid.Focus() != item.index {
				t.Fatalf("click in %s cell focused %d, want %d", item.name, grid.Focus(), item.index)
			}
		})
	}

	clickCell(20)
	tabs.ConsumeKey(loom.KeyEvent{Key: "tab"})
	if tabs.Focus() != 1 {
		t.Fatalf("Tab from Dialog cell selected tab %d, want 1", tabs.Focus())
	}
	tabs.Select(0)
	clickCell(21)
	tabs.ConsumeKey(loom.KeyEvent{Key: "shift-tab"})
	if tabs.Focus() != len(tabs.Tabs)-1 {
		t.Fatalf("Shift-Tab from Popup cell selected tab %d, want %d", tabs.Focus(), len(tabs.Tabs)-1)
	}

	tabs.Select(0)
	clickCell(12)
	textInput := grid.Children[12].(*textInputWidget).input
	before := textInput.Value()
	grid.ConsumeKey(loom.KeyEvent{Text: "!"})
	if got := textInput.Value(); got != before+"!" {
		t.Fatalf("TextInput value after key = %q, want %q", got, before+"!")
	}

	clickCell(0)
	outside := loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: c.Bounds().W, Y: 0}
	if res := grid.ConsumeMouse(outside); res.Consumed {
		t.Fatal("click outside grid cells was consumed")
	}
	last := grid.ChildRect(len(grid.Children) - 1)
	outerEdgeX := last.X + last.W - gridRectX(grid)
	if res := grid.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: outerEdgeX, Y: last.Y - gridRectY(grid)}); res.Consumed {
		t.Fatal("click on the outer grid edge was consumed")
	}
}

func gridRectX(grid *loom.Grid) int { return grid.ChildRect(0).X }

func gridRectY(grid *loom.Grid) int { return grid.ChildRect(0).Y }
