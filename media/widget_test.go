// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package media

import (
	"errors"
	"image"
	"image/color"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom"
	"ubunatic.com/cati/v1/core"
	"ubunatic.com/cati/v1/halfblock"
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
					if !r.Contains(x, y) && (cell.Text != " " || cell.Style != loom.Reset) {
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

func TestWideAndTallImagesKeepAspectAndCenterInRect(t *testing.T) {
	for _, tc := range []struct {
		name       string
		width      int
		height     int
		wantAspect float64
	}{
		{name: "wide", width: 40, height: 10, wantAspect: 4},
		{name: "tall", width: 10, height: 40, wantAspect: 0.25},
	} {
		t.Run(tc.name, func(t *testing.T) {
			img := solidImage(tc.width, tc.height, color.RGBA{R: 220, G: 30, B: 40, A: 255})
			widget, err := NewImage(img, ModeHalfblock)
			if err != nil {
				t.Fatal(err)
			}
			canvas := loom.NewCanvas(22, 14)
			r := loom.Rect{X: 3, Y: 2, W: 12, H: 8}
			widget.Draw(canvas, r)
			bounds, found := coloredBounds(canvas, r, color.RGBA{R: 220, G: 30, B: 40, A: 255})
			if !found {
				t.Fatal("image produced no colored cells")
			}
			actualAspect := float64(bounds.W) / float64(bounds.H*2)
			if delta := absFloat(actualAspect - tc.wantAspect); delta > 1 {
				t.Fatalf("rendered aspect %.2f, source aspect %.2f (bounds %#v)", actualAspect, tc.wantAspect, bounds)
			}
			if absInt((bounds.X+bounds.W/2)-(r.X+r.W/2)) > 1 || absInt((bounds.Y+bounds.H/2)-(r.Y+r.H/2)) > 1 {
				t.Fatalf("image bounds %#v are not centered in %#v", bounds, r)
			}
		})
	}
}

func TestDrawClearsLetterboxCellsAndCachesByImageAndSize(t *testing.T) {
	red := solidImage(40, 10, color.RGBA{R: 220, A: 255})
	green := solidImage(10, 40, color.RGBA{G: 220, A: 255})
	widget, err := NewImage(red, ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	widget.renderer = func(img image.Image, _ Mode, cols, rows int) (*core.Grid, error) {
		calls.Add(1)
		return halfblock.RenderToGrid(img, cols, halfblock.Options{Rows: rows})
	}
	canvas := loom.NewCanvas(20, 12)
	r := loom.Rect{X: 2, Y: 1, W: 12, H: 8}
	widget.Draw(canvas, r)
	widget.Draw(canvas, r)
	if got := calls.Load(); got != 1 {
		t.Fatalf("same image and size rendered %d times, want 1", got)
	}

	widget.mu.Lock()
	widget.image = green
	widget.mu.Unlock()
	widget.Draw(canvas, r)
	if got := calls.Load(); got != 2 {
		t.Fatalf("new image rendered %d times total, want 2", got)
	}
	if _, found := coloredBounds(canvas, r, color.RGBA{R: 220, A: 255}); found {
		t.Fatal("letterbox retained cells from the previous frame")
	}
	if _, found := coloredBounds(canvas, r, color.RGBA{G: 220, A: 255}); !found {
		t.Fatal("new frame was not rendered")
	}

	resized := r
	resized.W--
	widget.Draw(canvas, resized)
	if got := calls.Load(); got != 3 {
		t.Fatalf("resized rect rendered %d times total, want 3", got)
	}
}

func TestDrawShowsShortErrorMessage(t *testing.T) {
	widget, err := NewImage(solidImage(4, 4, color.RGBA{A: 255}), ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	widget.renderer = func(image.Image, Mode, int, int) (*core.Grid, error) {
		return nil, errors.New("synthetic render failure")
	}
	canvas := loom.NewCanvas(14, 8)
	r := loom.Rect{X: 2, Y: 1, W: 10, H: 4}
	widget.Draw(canvas, r)
	var message strings.Builder
	for x := r.X; x < r.X+r.W; x++ {
		message.WriteString(canvas.Get(x, r.Y).Text)
	}
	if !strings.Contains(message.String(), "render") {
		t.Fatalf("render failure message = %q", message.String())
	}
}

func solidImage(width, height int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

func coloredBounds(c *loom.Canvas, r loom.Rect, want color.RGBA) (loom.Rect, bool) {
	bounds := loom.Rect{X: r.X + r.W, Y: r.Y + r.H}
	found := false
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			style := c.Get(x, y).Style
			fgR, fgG, fgB, fgOK := style.FG.RGB()
			bgR, bgG, bgB, bgOK := style.BG.RGB()
			if (fgOK && fgR == want.R && fgG == want.G && fgB == want.B) ||
				(bgOK && bgR == want.R && bgG == want.G && bgB == want.B) {
				found = true
				if x < bounds.X {
					bounds.X = x
				}
				if x > bounds.W {
					bounds.W = x
				}
				if y < bounds.Y {
					bounds.Y = y
				}
				if y > bounds.H {
					bounds.H = y
				}
			}
		}
	}
	if !found {
		return loom.Rect{}, false
	}
	return loom.Rect{X: bounds.X, Y: bounds.Y, W: bounds.W - bounds.X + 1, H: bounds.H - bounds.Y + 1}, true
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func absFloat(n float64) float64 {
	if n < 0 {
		return -n
	}
	return n
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
