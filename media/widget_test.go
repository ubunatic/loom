// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package media

import (
	"image"
	"image/color"
	"sync/atomic"
	"testing"
	"time"

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

func TestStillImagesDoNotTick(t *testing.T) {
	widget, err := NewImage(image.NewRGBA(image.Rect(0, 0, 2, 2)), ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	if got := widget.TickInterval(); got != 0 {
		t.Fatalf("still image tick interval = %s, want zero", got)
	}
}

func TestVideoTicksAdvanceFramesAndStopAtEnd(t *testing.T) {
	first := image.NewRGBA(image.Rect(0, 0, 2, 2))
	second := image.NewRGBA(image.Rect(0, 0, 3, 2))
	frames := make(chan image.Image, 2)
	frames <- first
	frames <- second
	close(frames)
	widget := newStreamingWidget(frames, 24, func() {})
	if got := widget.TickInterval(); got != time.Second/24 {
		t.Fatalf("video tick interval = %s, want %s", got, time.Second/24)
	}

	widget.Tick(time.Now())
	if got := widget.image; got != second {
		t.Fatalf("current frame = %v, want latest queued frame", got)
	}
	if got := widget.TickInterval(); got != 0 {
		t.Fatalf("finished video tick interval = %s, want zero", got)
	}
}

func TestCloseStopsVideoStreamOnce(t *testing.T) {
	frames := make(chan image.Image)
	var stops atomic.Int32
	widget := newStreamingWidget(frames, 30, func() { stops.Add(1) })
	widget.Close()
	widget.Close()
	if got := stops.Load(); got != 1 {
		t.Fatalf("stream stop calls = %d, want 1", got)
	}
	if got := widget.TickInterval(); got != 0 {
		t.Fatalf("closed video tick interval = %s, want zero", got)
	}
}

func TestNewImageRejectsUnknownMode(t *testing.T) {
	_, err := NewImage(image.NewRGBA(image.Rect(0, 0, 1, 1)), Mode("unknown"))
	if err == nil {
		t.Fatal("NewImage accepted an unknown render mode")
	}
}
