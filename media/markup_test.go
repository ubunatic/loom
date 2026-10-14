// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package media

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ubunatic.com/loom"
)

func writeTestPNG(t *testing.T, dir, name string, img image.Image) {
	t.Helper()
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func gradientImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{uint8(x * 20), uint8(y * 20), 128, 255})
		}
	}
	return img
}

func lineWidth(line loom.RichLine) int {
	n := 0
	for _, span := range line.Spans {
		n += loom.StringWidth(span.Text)
	}
	return n
}

func TestMarkupImageSize(t *testing.T) {
	dir := t.TempDir()
	writeTestPNG(t, dir, "g.png", gradientImage(8, 12))
	tests := []struct {
		name       string
		img        loom.MarkupImage
		cols, rows int
	}{
		{"half native", loom.MarkupImage{Src: "g.png", Mode: "halfblock"}, 8, 6},
		{"quad native", loom.MarkupImage{Src: "g.png", Mode: "quadblock"}, 4, 3},
		{"six native", loom.MarkupImage{Src: "g.png", Mode: "sextant"}, 4, 3},
		{"half width", loom.MarkupImage{Src: "g.png", Mode: "halfblock", Width: 4}, 4, 3},
		{"half height", loom.MarkupImage{Src: "g.png", Mode: "halfblock", Height: 3}, 4, 3},
		{"half both", loom.MarkupImage{Src: "g.png", Mode: "halfblock", Width: 2, Height: 3}, 2, 2},
	}
	for _, tt := range tests {
		lines, err := MarkupImage(tt.img, dir)
		if err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		if len(lines) != tt.rows {
			t.Errorf("%s rows = %d, want %d", tt.name, len(lines), tt.rows)
		}
		for i, line := range lines {
			if w := lineWidth(line); w != tt.cols {
				t.Errorf("%s row %d width = %d, want %d", tt.name, i, w, tt.cols)
			}
		}
	}
}

func TestMarkupImageGlyphsAndColors(t *testing.T) {
	dir := t.TempDir()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(1, 0, color.RGBA{0, 255, 0, 255})
	img.Set(0, 1, color.RGBA{0, 0, 255, 255})
	img.Set(1, 1, color.RGBA{0, 0, 255, 255})
	writeTestPNG(t, dir, "rgb.png", img)
	lines, err := MarkupImage(loom.MarkupImage{Src: "rgb.png", Mode: "halfblock"}, dir)
	if err != nil {
		t.Fatal(err)
	}
	blue := loom.ColorRGB(0, 0, 255)
	want := []loom.RichSpan{
		{Text: "▀", Style: loom.Style{FG: loom.ColorRGB(255, 0, 0), BG: blue}},
		{Text: "▀", Style: loom.Style{FG: loom.ColorRGB(0, 255, 0), BG: blue}},
	}
	if len(lines) != 1 || len(lines[0].Spans) != 2 || lines[0].Spans[0] != want[0] || lines[0].Spans[1] != want[1] {
		t.Errorf("lines = %+v, want %+v", lines, want)
	}
}

func TestMarkupImageTransparentAndMerged(t *testing.T) {
	dir := t.TempDir()
	img := image.NewRGBA(image.Rect(0, 0, 4, 2)) // fully transparent
	writeTestPNG(t, dir, "clear.png", img)
	lines, err := MarkupImage(loom.MarkupImage{Src: "clear.png", Mode: "halfblock"}, dir)
	if err != nil {
		t.Fatal(err)
	}
	want := loom.RichSpan{Text: "    "}
	if len(lines) != 1 || len(lines[0].Spans) != 1 || lines[0].Spans[0] != want {
		t.Errorf("lines = %+v, want one default-style run %+v", lines, want)
	}
}

func TestMarkupImageSingleWidthGlyphs(t *testing.T) {
	dir := t.TempDir()
	img := image.NewRGBA(image.Rect(0, 0, 36, 36))
	for y := range 36 {
		for x := range 36 {
			img.Set(x, y, color.RGBA{uint8(x * 7), uint8(y * 7), uint8((x ^ y) * 7), uint8(255 * ((x*y + x) % 3 / 2))})
		}
	}
	writeTestPNG(t, dir, "noise.png", img)
	for _, mode := range []Mode{ModeHalfblock, ModeQuadblock, ModeSextant} {
		lines, err := MarkupImage(loom.MarkupImage{Src: "noise.png", Mode: string(mode)}, dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range lines {
			for _, span := range line.Spans {
				for _, r := range span.Text {
					if loom.RuneWidth(r) != 1 {
						t.Errorf("%s glyph %U width %d, want 1", mode, r, loom.RuneWidth(r))
					}
				}
			}
		}
	}
}

func TestMarkupImageParse(t *testing.T) {
	dir := t.TempDir()
	writeTestPNG(t, dir, "g.png", gradientImage(4, 4))
	path := filepath.Join(dir, "doc.tui")
	doc, err := loom.ParseMarkupWith("[box]\n[image src=g.png]\n[/box]", MarkupOptions(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Lines) != 4 {
		t.Fatalf("lines = %d, want 4 (top, 2 image rows, bottom)", len(doc.Lines))
	}
	for i, line := range doc.Lines {
		if w := lineWidth(line); w != 8 {
			t.Errorf("row %d width = %d, want 8", i, w)
		}
	}
	_, err = loom.ParseMarkupWith("x\n[image src=missing.png]", MarkupOptions(path))
	var me *loom.MarkupError
	if !errors.As(err, &me) || me.Line != 2 || !strings.Contains(me.Msg, "missing.png") {
		t.Errorf("missing file err = %v, want MarkupError at line 2", err)
	}
}
