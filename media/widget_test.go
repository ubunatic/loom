// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package media

import (
	"image"
	"image/color"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestStillImageDrawModesFitAndStayInsideRect(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 12))
	for y := 0; y < 12; y++ {
		for x := 0; x < 32; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x * 7), G: uint8(y * 17), B: 190, A: 255})
		}
	}

	for _, mode := range []Mode{ModeHalfblock, ModeQuadblock, ModeSextant} {
		t.Run(string(mode), func(t *testing.T) {
			widget, err := NewImage(img, mode)
			if err != nil {
				t.Fatal(err)
			}
			canvas := loom.NewCanvas(14, 10)
			r := loom.Rect{X: 3, Y: 2, W: 7, H: 5}
			widget.Draw(canvas, r)

			painted := 0
			for y := 0; y < canvas.Rows(); y++ {
				for x := 0; x < canvas.Cols(); x++ {
					cell := canvas.Get(x, y)
					if !r.Contains(x, y) && cell.Text != " " {
						t.Fatalf("write outside rect at (%d,%d): %#v", x, y, cell)
					}
					if r.Contains(x, y) && cell.Text != " " {
						painted++
					}
				}
			}
			if painted == 0 {
				t.Fatal("image area was not painted")
			}
		})
	}
}

func TestNewImageRejectsUnknownMode(t *testing.T) {
	_, err := NewImage(image.NewRGBA(image.Rect(0, 0, 1, 1)), Mode("unknown"))
	if err == nil {
		t.Fatal("NewImage accepted an unknown render mode")
	}
}
