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

func TestMediaDemoPTYPaintsOnlyItsImageRect(t *testing.T) {
	s := ptytest.Start(t, 50, 16, buildMediaDemo(t), writeSolidPNG(t))
	s.WaitFor("Media Demo", 5*time.Second)

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

	redCells := 0
	for y, row := range s.Cells() {
		for x, cell := range row {
			fg, _ := cell.Style.Effective()
			r, g, b, ok := fg.RGB()
			if !ok || r != 240 || g != 24 || b != 16 {
				continue
			}
			redCells++
			if x < 6 || x >= 14 || y < titleY+2 || y >= titleY+6 {
				t.Fatalf("image color escaped its rect at (%d,%d), title row %d", x, y, titleY)
			}
		}
	}
	if redCells == 0 {
		t.Fatal("PTY screen contains no image-colored cells")
	}

	s.Send("q")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("media example exit: %v", err)
	}
}
