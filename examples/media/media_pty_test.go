// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom/internal/ptytest"
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
	cmd := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=c=red:s=32x16:d=0.5:r=8", "-f", "lavfi", "-i", "color=c=blue:s=32x16:d=0.5:r=8", "-filter_complex", "[0:v][1:v]concat=n=2:v=1:a=0,format=yuv420p", "-an", "-c:v", "mpeg4", "-y", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("ffmpeg cannot create the test video: %v: %s", err, output)
	}
	return path
}

func TestMediaDemoPTYFillsAvailableWidth(t *testing.T) {
	const cols, rows = 120, 40
	s := ptytest.Start(t, cols, rows, buildMediaDemo(t), writeSolidPNG(t))
	s.WaitFor("Cols: 120  Rows: 40  Mode: halfblock", 5*time.Second)
	if screen := s.Screen(); len(screen) != rows {
		t.Fatalf("rendered %d rows, want full terminal height %d", len(screen), rows)
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
	if maxX < 30 || maxX >= cols {
		t.Fatalf("aspect-fitted image reached column %d, want columns 0 through at least 30 within terminal width %d", maxX, cols)
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

func TestMediaDemoPTYPlaysVideo(t *testing.T) {
	const cols, rows = 120, 40
	s := ptytest.Start(t, cols, rows, buildMediaDemo(t), writeVideo(t))
	s.WaitFor("Cols: 120  Rows: 40  Mode: halfblock", 5*time.Second)
	if screen := s.Screen(); len(screen) != rows {
		t.Fatalf("rendered %d rows, want full terminal height %d", len(screen), rows)
	}
	if status := s.Screen()[rows-1]; !strings.Contains(status, "Cols: 120  Rows: 40  Mode: halfblock") || !strings.Contains(status, "colors.mp4") {
		t.Fatalf("bottom row does not contain terminal status and video path: %q", status)
	}
	for y, row := range s.Cells() {
		if len(row) != cols {
			t.Fatalf("rendered row %d has %d cells, want full terminal width %d", y, len(row), cols)
		}
	}
	deadline := time.Now().Add(5 * time.Second)
	seenRed, seenBlue := false, false
	for time.Now().Before(deadline) && !(seenRed && seenBlue) {
		for _, row := range s.Cells() {
			for _, cell := range row {
				fg, _ := cell.Style.Effective()
				r, g, b, ok := fg.RGB()
				if !ok {
					continue
				}
				seenRed = seenRed || (r > 180 && g < 80 && b < 80)
				seenBlue = seenBlue || (b > 120 && r < 80 && g < 80)
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !seenRed || !seenBlue {
		t.Fatalf("video colors observed: red=%t blue=%t", seenRed, seenBlue)
	}
	s.Send("q")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("media example exit: %v", err)
	}
}
