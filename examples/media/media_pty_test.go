// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"codeberg.org/ubunatic/loom/internal/ptytest"
	"codeberg.org/ubunatic/loom/media"
)

func buildMediaDemo(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "loom-media")
	out, err := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/examples/media").CombinedOutput()
	if err != nil {
		t.Fatalf("build media example: %v\n%s", err, out)
	}
	return bin
}

func TestMediaCommandHelp(t *testing.T) {
	cmd := newCommand()
	cmd.SetArgs([]string{"--help"})
	var output strings.Builder
	cmd.SetOut(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("show help: %v", err)
	}
	for _, want := range []string{"loom-media <image-or-video> [mode]", "--mode", "--help"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("help output does not contain %q: %s", want, output.String())
		}
	}
}

func TestMediaCommandModeArgumentAndFlag(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{name: "default", args: []string{"image.png"}, want: "halfblock"},
		{name: "positional", args: []string{"image.png", "quadblock"}, want: "quadblock"},
		{name: "flag", args: []string{"--mode", "sextant", "image.png"}, want: "sextant"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mode := media.Mode("")
			path := ""
			cmd := newCommandWithRun(func(gotPath string, gotMode media.Mode) error {
				path, mode = gotPath, gotMode
				return nil
			})
			cmd.SetArgs(tc.args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("execute: %v", err)
			}
			if path != "image.png" || string(mode) != tc.want {
				t.Fatalf("mode = %q, want %q", mode, tc.want)
			}
		})
	}
}

func writeSolidPNG(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "red.png")
	img := image.NewRGBA(image.Rect(0, 0, 16, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 16; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 240, G: 24, B: 16, A: 255})
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, img); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeVideo(t *testing.T) string {
	t.Helper()
	for _, tool := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is unavailable: %v", tool, err)
		}
	}
	path := filepath.Join(t.TempDir(), "colors.mp4")
	cmd := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=c=red:s=32x16:d=2:r=8", "-f", "lavfi", "-i", "color=c=blue:s=32x16:d=2:r=8", "-filter_complex", "[0:v][1:v]concat=n=2:v=1:a=0,format=yuv420p", "-an", "-c:v", "mpeg4", "-y", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("ffmpeg cannot create the test video: %v: %s", err, output)
	}
	return path
}

func TestMediaDemoPTYFillsAvailableWidth(t *testing.T) {
	const cols, rows = 120, 40
	s := ptytest.Start(t, cols, rows, buildMediaDemo(t), writeSolidPNG(t))
	// The media request promotes this full-width pane to the alternate screen,
	// which provides all 40 terminal rows. 39 was the old inline reservation.
	s.WaitFor("Cols: 120  Rows: 40  Mode: halfblock", 5*time.Second)
	if screen := s.Screen(); len(screen) != rows {
		t.Fatalf("rendered %d rows, want terminal height %d", len(screen), rows)
	}

	var titleY = -1
	for y, row := range s.Screen() {
		if strings.Contains(row, "Media Demo") {
			titleY = y
			break
		}
	}
	if titleY < 0 {
		t.Fatal("media title was not rendered")
	}
	status := s.Screen()[rows-1]
	if !strings.Contains(status, "Cols: 120  Rows: 40  Mode: halfblock") {
		t.Fatalf("bottom row does not contain status dimensions and mode: %q", status)
	}
	if !strings.Contains(status, "Media: ") || !strings.Contains(status, "red.png") {
		t.Fatalf("bottom row does not contain media path: %q", status)
	}

	redCells := 0
	maxX := -1
	for y, row := range s.Cells() {
		for x, cell := range row {
			fg, _ := cell.Style.Effective()
			r, g, b, ok := fg.RGB()
			if !ok || r != 240 || g != 24 || b != 16 {
				continue
			}
			redCells++
			if y < titleY+1 {
				t.Fatalf("image color reached title row at (%d,%d), title row %d", x, y, titleY)
			}
			if x > maxX {
				maxX = x
			}
		}
	}
	if redCells == 0 {
		t.Fatal("PTY screen contains no image-colored cells")
	}
	if maxX < 50 || maxX >= cols {
		t.Fatalf("aspect-fitted image reached column %d, want image content beyond column 50 within terminal width %d", maxX, cols)
	}
	for y, row := range s.Cells() {
		if y <= titleY {
			continue
		}
		if len(row) != cols {
			t.Fatalf("rendered row %d has %d cells, want full terminal width %d", y, len(row), cols)
		}
	}

	s.Send("q")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("media example exit: %v", err)
	}
}

func TestMediaDemoPTYZoomControlsRespondToKeyAndClick(t *testing.T) {
	const cols, rows = 80, 24
	s := ptytest.Start(t, cols, rows, buildMediaDemo(t), writeSolidPNG(t))
	s.WaitFor("Cols: 80", 5*time.Second)
	s.Send("+")
	s.WaitFor("1.25x", 3*time.Second)
	// Locate the visible control in terminal coordinates; the pane may start
	// below the shell cursor, so its screen row is not a fixed constant.
	controlRow, controlCol := -1, -1
	for row, line := range s.Screen() {
		if i := strings.Index(line, "[+]"); i >= 0 {
			col := utf8.RuneCountInString(line[:i])
			controlRow, controlCol = row, col
			break
		}
	}
	if controlRow < 0 {
		t.Fatal("zoom-in control is not visible")
	}
	press := fmt.Sprintf("\x1b[<0;%d;%dM", controlCol+2, controlRow+1)
	release := fmt.Sprintf("\x1b[<0;%d;%dm", controlCol+2, controlRow+1)
	s.SendRaw([]byte(press))
	s.SendRaw([]byte(release))
	s.WaitFor("1.56x", 3*time.Second)
	s.Send("q")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("media example exit: %v", err)
	}
}

func TestMediaDemoPTYPlaysVideo(t *testing.T) {
	const cols, rows = 120, 40
	s := ptytest.Start(t, cols, rows, buildMediaDemo(t), writeVideo(t))
	// The media request promotes this full-width pane to the alternate screen,
	// which provides all 40 terminal rows. 39 was the old inline reservation.
	s.WaitFor("Cols: 120  Rows: 40  Mode: halfblock", 5*time.Second)
	if screen := s.Screen(); len(screen) != rows {
		t.Fatalf("rendered %d rows, want full terminal height %d", len(screen), rows)
	}
	if status := s.Screen()[rows-1]; !strings.Contains(status, "Cols: 120  Rows: 40  Mode: halfblock") || !strings.Contains(status, "colors.mp4") {
		t.Fatalf("bottom row does not contain terminal status and video path: %q", status)
	}
	maxX := -1
	deadline := time.Now().Add(5 * time.Second)
	seenRed, seenBlue := false, false
	for time.Now().Before(deadline) && !(seenRed && seenBlue && maxX >= 50) {
		for y, row := range s.Cells() {
			if len(row) != cols {
				t.Fatalf("rendered row %d has %d cells, want full terminal width %d", y, len(row), cols)
			}
			for x, cell := range row {
				fg, _ := cell.Style.Effective()
				r, g, b, ok := fg.RGB()
				if !ok {
					continue
				}
				isRed := r > 180 && g < 80 && b < 80
				isBlue := b > 120 && r < 80 && g < 80
				if (isRed || isBlue) && x > maxX {
					maxX = x
				}
				seenRed = seenRed || isRed
				seenBlue = seenBlue || isBlue
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	if maxX < 50 {
		t.Fatalf("video content reached column %d, want content beyond column 50", maxX)
	}
	if !seenRed || !seenBlue {
		t.Fatalf("video colors observed: red=%t blue=%t", seenRed, seenBlue)
	}
	s.Send("q")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("media example exit: %v", err)
	}
}
