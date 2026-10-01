// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package media

import (
	"context"
	"errors"
	"image"
	"image/color"
	"math"
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

func TestMediaControlsZoomPanClampAndReset(t *testing.T) {
	w, err := NewImage(solidImage(40, 20, color.RGBA{A: 255}), ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	w.Draw(loom.NewCanvas(40, 12), loom.Rect{W: 40, H: 12})
	w.setZoom(1.25)
	if w.zoom != 1.25 {
		t.Fatalf("zoom = %v, want 1.25", w.zoom)
	}
	w.pan(-4, 4)
	if w.panX != -1 || w.panY != 1 {
		t.Fatalf("pan = (%v,%v), want clamped (-1,1)", w.panX, w.panY)
	}
	w.setZoom(1.0 / 1.25)
	if w.zoom != 1 || w.panX != 0 || w.panY != 0 {
		t.Fatalf("zoom out did not reset fit/pan: zoom=%v pan=(%v,%v)", w.zoom, w.panX, w.panY)
	}
	for i := 0; i < 20; i++ {
		w.setZoom(1.25)
	}
	if w.zoom != 8 {
		t.Fatalf("zoom limit = %v, want 8", w.zoom)
	}
	if result := w.ConsumeKey(loom.KeyEvent{Text: "0"}); result.Quit || !result.Consumed {
		t.Fatal("reset key was not consumed")
	}
	if w.zoom != 1 {
		t.Fatalf("reset zoom = %v, want 1", w.zoom)
	}
}

func TestZoomPreservesViewportCenter(t *testing.T) {
	w, err := NewImage(solidImage(100, 60, color.RGBA{A: 255}), ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	w.Draw(loom.NewCanvas(20, 12), loom.Rect{W: 20, H: 12})
	w.pan(.5, -.4)
	// At 1x the source center is (70,22). At 2x that becomes (140,44).
	w.setZoom(2)
	if math.Abs(w.panX-4.0/9) > 1e-9 || math.Abs(w.panY+.32) > 1e-9 {
		t.Fatalf("2x pan = (%v,%v), want (4/9,-.32)", w.panX, w.panY)
	}
	w.setZoom(.5)
	if math.Abs(w.panX-.5) > 1e-9 || math.Abs(w.panY+.4) > 1e-9 {
		t.Fatalf("round-trip pan = (%v,%v), want (.5,-.4)", w.panX, w.panY)
	}
	w.pan(2, -2)
	w.setZoom(.5)
	if w.panX != 1 || w.panY != -1 {
		t.Fatalf("zoom-out edge clamp = (%v,%v), want (1,-1)", w.panX, w.panY)
	}
	w.setZoom(.5)
	if w.panX != 1 || w.panY != 0 {
		t.Fatalf("independent axis clamp = (%v,%v), want (1,0)", w.panX, w.panY)
	}
}

func TestDragTracksPixelsAcrossControlsAndBounds(t *testing.T) {
	for _, mode := range []Mode{ModeHalfblock, ModeQuadblock, ModeSextant} {
		t.Run(string(mode), func(t *testing.T) {
			w, err := NewImage(solidImage(100, 60, color.RGBA{A: 255}), mode)
			if err != nil {
				t.Fatal(err)
			}
			w.Draw(loom.NewCanvas(30, 12), loom.Rect{X: 3, Y: 2, W: 24, H: 10})
			w.setZoom(2)
			for _, e := range []loom.MouseEvent{
				{Action: loom.MousePress, Button: loom.MouseLeft, X: 18, Y: 3},
				{Action: loom.MouseDrag, Button: loom.MouseLeft, X: 15, Y: 5},
				{Action: loom.MouseDrag, Button: loom.MouseLeft, X: 9, Y: 9},
				// Repeated press reports during a drag must not click the minus button.
				{Action: loom.MousePress, Button: loom.MouseLeft, X: 9, Y: 9},
				{Action: loom.MouseDrag, Button: loom.MouseLeft, X: -2, Y: 13},
			} {
				if res := w.ConsumeMouse(e); !res.Consumed || res.Quit {
					t.Fatalf("event %#v was not handled: %#v", e, res)
				}
			}
			cw, ch := pxPerCell(mode)
			wantX := 40 * float64(cw) / float64(200-24*cw)
			wantY := -20 * float64(ch) / float64(120-8*ch)
			if w.zoom != 2 || math.Abs(w.panX-wantX) > 1e-9 || math.Abs(w.panY-wantY) > 1e-9 {
				t.Fatalf("drag state: zoom=%v pan=(%v,%v), want 2 (%v,%v)", w.zoom, w.panX, w.panY, wantX, wantY)
			}
			if res := w.ConsumeMouse(loom.MouseEvent{Action: loom.MouseRelease, X: -2, Y: 13}); !res.Consumed || w.dragging {
				t.Fatal("release did not end drag")
			}
			if res := w.ConsumeMouse(loom.MouseEvent{Action: loom.MouseDrag, X: 0, Y: 0}); res.Consumed || math.Abs(w.panX-wantX) > 1e-9 {
				t.Fatal("drag after release changed pan")
			}
		})
	}
}

func TestPanUsesRoundedPixelOverflow(t *testing.T) {
	w, err := NewImage(solidImage(13, 13, color.RGBA{A: 255}), ModeQuadblock)
	if err != nil {
		t.Fatal(err)
	}
	w.Draw(loom.NewCanvas(7, 9), loom.Rect{W: 7, H: 9})
	w.setZoom(1.05) // 13*1.05 rounds to 14 pixels: exactly seven cells.
	if w.canPanLocked() {
		t.Fatal("rounded image fits but canPan reports overflow")
	}
	w.pan(.5, -.5)
	if w.panX != 0 || w.panY != 0 {
		t.Fatalf("pan changed without pixel overflow: (%v,%v)", w.panX, w.panY)
	}
}

func TestMediaControlsPlayPauseKeysAndMouseWheel(t *testing.T) {
	w := newStreamingWidget(make(chan image.Image), 24, nil)
	if result := w.ConsumeKey(loom.KeyEvent{Text: " "}); result.Quit || !result.Consumed || w.IsPlaying() {
		t.Fatal("space did not pause playback")
	}
	if result := w.ConsumeKey(loom.KeyEvent{Text: " "}); result.Quit || !result.Consumed || !w.IsPlaying() {
		t.Fatal("space did not resume playback")
	}
	if result := w.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollUp, X: 0, Y: 0}); result.Quit || !result.Consumed || w.zoom <= 1 {
		t.Fatal("wheel up did not zoom in")
	}
	if result := w.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollDown, X: 0, Y: 0}); result.Quit || !result.Consumed || w.zoom != 1 {
		t.Fatal("wheel down did not return to fit")
	}
}

func TestScaleForViewPanChangesCrop(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 60))
	for y := 0; y < 60; y++ {
		for x := 0; x < 100; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x), G: uint8(y), A: 255})
		}
	}
	left := scaleForView(img, ModeHalfblock, 20, 10, 2, -1, -1)
	right := scaleForView(img, ModeHalfblock, 20, 10, 2, 1, 1)
	a, b := left.Bounds(), right.Bounds()
	if a.Dx() != 20 || a.Dy() != 20 || b.Dx() != 20 || b.Dy() != 20 {
		t.Fatalf("zoomed images = %v and %v, want panel pixel size (20x20)", a, b)
	}
	if left.At(a.Min.X, a.Min.Y) == right.At(b.Min.X, b.Min.Y) {
		t.Fatal("pan did not change the visible source crop")
	}
	if left.At(0, 0) != left.At(1, 0) {
		t.Fatal("zoomed source pixels were not enlarged across adjacent output pixels")
	}
}

func TestNativeCellSizeAndScaleForView(t *testing.T) {
	for _, tc := range []struct {
		mode  Mode
		pxW   int
		pxH   int
		wantW int
		wantH int
	}{
		{ModeHalfblock, 1, 2, 13, 7},
		{ModeQuadblock, 2, 2, 7, 7},
		{ModeSextant, 2, 3, 7, 5},
	} {
		t.Run(string(tc.mode), func(t *testing.T) {
			img := solidImage(13, 13, color.RGBA{R: 255, A: 255})
			w, h := nativeCellSize(tc.mode, img)
			if w != tc.wantW || h != tc.wantH {
				t.Fatalf("nativeCellSize = %dx%d, want %dx%d", w, h, tc.wantW, tc.wantH)
			}
			oneX := scaleForView(img, tc.mode, 30, 30, 1, 0, 0)
			small := scaleForView(img, tc.mode, 30, 30, .5, 0, 0)
			if oneX.Bounds().Dx() != 13 || oneX.Bounds().Dy() != 13 {
				t.Fatalf("1x image dimensions = %v, want 13x13", oneX.Bounds())
			}
			if small.Bounds().Dx() != 7 || small.Bounds().Dy() != 7 {
				t.Fatalf("0.5x image dimensions = %v, want 7x7", small.Bounds())
			}
			cropped := scaleForView(img, tc.mode, 3, 2, 1, 0, 0)
			if cropped.Bounds().Dx() != 3*tc.pxW || cropped.Bounds().Dy() != 2*tc.pxH {
				t.Fatalf("cropped dimensions = %v, want %dx%d", cropped.Bounds(), 3*tc.pxW, 2*tc.pxH)
			}
		})
	}
}

func TestScaleForViewCentersSmallImageAndPanMovesCrop(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 20), G: uint8(y * 20), A: 255})
		}
	}
	w, err := NewImage(img, ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	w.renderer = func(image.Image, Mode, int, int) (*core.Grid, error) {
		cells := make([][]core.Cell, 5)
		for y := range cells {
			cells[y] = make([]core.Cell, 10)
			for x := range cells[y] {
				cells[y][x] = core.Cell{Ch: 'X'}
			}
		}
		return &core.Grid{Width: 10, Height: 5, Cells: cells}, nil
	}
	canvas := loom.NewCanvas(20, 10)
	w.Draw(canvas, loom.Rect{X: 0, Y: 0, W: 20, H: 10})
	// 10x5 native cells occupy x=5..14 and y=1..5 once controls take two rows.
	if canvas.Get(4, 2).Text != " " || canvas.Get(5, 2).Text == " " {
		t.Fatal("small media was not centered in the available image viewport")
	}
	center := scaleForView(img, ModeHalfblock, 3, 2, 2, 0, 0)
	left := scaleForView(img, ModeHalfblock, 3, 2, 2, -1, 0)
	right := scaleForView(img, ModeHalfblock, 3, 2, 2, 1, 0)
	if center.At(0, 0) == left.At(0, 0) || center.At(0, 0) == right.At(0, 0) || left.At(0, 0) == right.At(0, 0) {
		t.Fatal("centered and panned crops did not show distinct source areas")
	}
}

func TestScaleForViewHandlesNonZeroBoundsAndZoomFloor(t *testing.T) {
	img := image.NewRGBA(image.Rect(4, 7, 24, 17))
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 7), G: uint8(y * 11), A: 255})
		}
	}
	out := scaleForView(img, ModeHalfblock, 5, 3, 2, 0, 0)
	if out.Bounds().Dx() != 5 || out.Bounds().Dy() != 6 {
		t.Fatalf("non-zero-bound crop = %v, want 5x6", out.Bounds())
	}
	w, err := NewImage(solidImage(80, 40, color.RGBA{A: 255}), ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 40; i++ {
		w.setZoom(1 / 1.25)
	}
	if w.zoom != 1.0/16 {
		t.Fatalf("zoom floor = %v, want 1/16", w.zoom)
	}
}

func TestZoomKeepsRenderedAreaAndMagnifiesPixels(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 80, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 80; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 3), G: uint8(y * 5), B: 120, A: 255})
		}
	}
	w, err := NewImage(img, ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	canvas := loom.NewCanvas(40, 8)
	r := loom.Rect{X: 0, Y: 0, W: 40, H: 8}
	countImageCells := func() int {
		count := 0
		for y := 0; y < 6; y++ {
			for x := 0; x < 40; x++ {
				if canvas.Get(x, y).Text != " " {
					count++
				}
			}
		}
		return count
	}
	w.Draw(canvas, r)
	fitCells := countImageCells()
	w.setZoom(2)
	w.Draw(canvas, r)
	zoomCells := countImageCells()
	if fitCells == 0 || zoomCells != fitCells {
		t.Fatalf("rendered cell count at fit/2x = %d/%d, want same non-zero area", fitCells, zoomCells)
	}

	zoomed := scaleForView(img, ModeHalfblock, 40, 6, 2, 0, 0)
	if zoomed.At(0, 20) != zoomed.At(1, 20) {
		t.Fatal("source detail was not enlarged across adjacent output pixels")
	}
}

func TestMediaControlBarClickAndDragPan(t *testing.T) {
	w, err := NewImage(solidImage(40, 20, color.RGBA{R: 255, A: 255}), ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	w.Draw(loom.NewCanvas(40, 8), loom.Rect{X: 0, Y: 0, W: 40, H: 8})
	if result := w.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 12, Y: 7}); result.Quit || !result.Consumed || w.zoom != 1.25 {
		t.Fatalf("zoom button: quit=%v used=%v zoom=%v", result.Quit, result.Consumed, w.zoom)
	}
	if result := w.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 20, Y: 2}); result.Quit || !result.Consumed {
		t.Fatal("zoomed image press did not start pan drag")
	}
	if result := w.ConsumeMouse(loom.MouseEvent{Action: loom.MouseDrag, Button: loom.MouseLeft, X: 10, Y: 2}); result.Quit || !result.Consumed || w.panX <= 0 {
		t.Fatalf("drag pan failed: quit=%v used=%v pan=%v", result.Quit, result.Consumed, w.panX)
	}
	if result := w.ConsumeMouse(loom.MouseEvent{Action: loom.MouseRelease, Button: loom.MouseLeft}); result.Quit || !result.Consumed {
		t.Fatal("drag release was not consumed")
	}
}

func TestWideAndTallImagesCropWithinRectAndCenterGrid(t *testing.T) {
	for _, tc := range []struct {
		name   string
		width  int
		height int
	}{
		{name: "wide", width: 40, height: 10},
		{name: "tall", width: 10, height: 40},
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
			if absInt((bounds.X+bounds.W/2)-(r.X+r.W/2)) > 1 || absInt((bounds.Y+bounds.H/2)-(r.Y+(r.H-2)/2)) > 1 {
				t.Fatalf("image bounds %#v are not centered in image viewport %#v", bounds, r)
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

func TestDrawStatusIndicatorsUseThemeAndSpec(t *testing.T) {
	img := solidImage(2, 2, color.RGBA{A: 255})
	for _, tc := range []struct {
		name  string
		theme loom.ThemeColors
		want  loom.Color
	}{
		{name: "plain", theme: loom.Theme("plain"), want: loom.ColorRGB(255, 96, 96)},
		{name: "custom", theme: func() loom.ThemeColors {
			theme := loom.Theme("plain")
			theme.MediaLoadingFG = loom.ThemeColorRGB(1, 2, 3)
			theme.MediaErrorFG = loom.ThemeColorRGB(4, 5, 6)
			theme.MediaLoadingDim = false
			return theme
		}(), want: loom.ColorRGB(4, 5, 6)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			widget, err := NewImageWithTheme(img, ModeHalfblock, tc.theme)
			if err != nil {
				t.Fatal(err)
			}
			widget.renderer = func(image.Image, Mode, int, int) (*core.Grid, error) {
				return nil, errors.New("synthetic render failure")
			}
			canvas := loom.NewCanvas(20, 2)
			r := loom.Rect{W: 16, H: 1}
			widget.Draw(canvas, r)
			if !canvasContainsText(canvas, r, loom.SpeccedDefaults.Media.RenderErrorLabel[:1]) {
				t.Fatalf("render error label missing: %q", canvas.Row(0))
			}
			if got := canvas.Get(0, 0).Style.FG; got != tc.want {
				t.Errorf("error foreground = %+v, want %+v", got, tc.want)
			}
		})
	}

	widget, err := NewImage(img, ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	widget.renderer = func(image.Image, Mode, int, int) (*core.Grid, error) {
		time.Sleep(100 * time.Millisecond)
		return &core.Grid{Width: 1, Height: 1, Cells: [][]core.Cell{{{Ch: 'X'}}}}, nil
	}
	widget.ApplyTheme(loom.Theme("plain"))
	canvas := loom.NewCanvas(12, 2)
	r := loom.Rect{W: 10, H: 1}
	widget.Draw(canvas, r)
	if !canvasContainsText(canvas, r, loom.SpeccedDefaults.Media.LoadingLabel[:1]) {
		t.Fatalf("loading label missing: %q", canvas.Row(0))
	}
	if got := canvas.Get(0, 0).Style; got.FG != loom.ColorRGB(128, 128, 128) || !got.Dim {
		t.Errorf("plain loading style = %+v, want existing gray foreground and dim", got)
	}
}

func TestDrawLoadsMediaImmediatelyOrInBackground(t *testing.T) {
	for _, tc := range []struct {
		name  string
		delay time.Duration
	}{
		{name: "immediate", delay: 5 * time.Millisecond},
		{name: "background", delay: 80 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			widget, err := NewImage(solidImage(2, 2, color.RGBA{R: 200, A: 255}), ModeHalfblock)
			if err != nil {
				t.Fatal(err)
			}
			widget.renderer = func(image.Image, Mode, int, int) (*core.Grid, error) {
				time.Sleep(tc.delay)
				return &core.Grid{Width: 1, Height: 1, Cells: [][]core.Cell{{{Ch: 'X'}}}}, nil
			}
			canvas := loom.NewCanvas(12, 3)
			r := loom.Rect{X: 1, Y: 1, W: 10, H: 1}
			widget.Draw(canvas, r)
			if tc.delay < 50*time.Millisecond {
				if !canvasContainsText(canvas, r, "X") {
					t.Fatal("immediate draw did not render media")
				}
				return
			}
			loadingFound := false
			for x := r.X; x < r.X+r.W; x++ {
				cell := canvas.Get(x, r.Y)
				if cell.Text == "l" && cell.Style.Dim {
					loadingFound = true
				}
			}
			if !loadingFound {
				t.Fatal("delayed draw did not show a dim loading indicator")
			}
			time.Sleep(tc.delay)
			canvas = loom.NewCanvas(12, 3)
			widget.Draw(canvas, r)
			if !canvasContainsText(canvas, r, "X") {
				t.Fatal("completed draw did not render media")
			}
		})
	}
}

func canvasContainsText(c *loom.Canvas, r loom.Rect, want string) bool {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			if c.Get(x, y).Text == want {
				return true
			}
		}
	}
	return false
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

func TestVideoPosterDrawsBeforeFirstFrame(t *testing.T) {
	poster := solidImage(4, 4, color.RGBA{R: 230, A: 255})
	frames := make(chan image.Image)
	widget, err := newVideoWithOpener("test-video", ModeHalfblock, 24, poster, func(context.Context, string, float64) (<-chan image.Image, func(), error) {
		return frames, func() {}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	canvas := loom.NewCanvas(8, 4)
	widget.Draw(canvas, loom.Rect{W: 8, H: 4})
	if _, ok := coloredBounds(canvas, loom.Rect{W: 8, H: 4}, color.RGBA{R: 230, A: 255}); !ok {
		t.Fatal("poster was not rendered before the first stream frame")
	}
}

func TestNewVideoWithPosterValidatesArguments(t *testing.T) {
	poster := solidImage(2, 2, color.RGBA{A: 255})
	for _, tc := range []struct {
		name string
		mode Mode
		fps  float64
		img  image.Image
	}{
		{name: "nil poster", mode: ModeHalfblock, fps: 24},
		{name: "unsupported mode", mode: Mode("unknown"), fps: 24, img: poster},
		{name: "invalid frame rate", mode: ModeHalfblock, fps: math.NaN(), img: poster},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewVideoWithPoster("does-not-need-to-exist", tc.mode, tc.fps, tc.img); err == nil {
				t.Fatal("NewVideoWithPoster accepted invalid arguments")
			}
		})
	}
}

func TestPausePlayControlsFrameAdvancement(t *testing.T) {
	poster := solidImage(2, 2, color.RGBA{R: 100, A: 255})
	queued := solidImage(3, 2, color.RGBA{G: 100, A: 255})
	frames := make(chan image.Image, 1)
	frames <- queued
	widget, err := newVideoWithOpener("test-video", ModeHalfblock, 24, poster, func(ctx context.Context, _ string, _ float64) (<-chan image.Image, func(), error) {
		go func() {
			<-ctx.Done()
			close(frames)
		}()
		return frames, func() {}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	widget.Pause()
	if widget.IsPlaying() || widget.TickInterval() != 0 {
		t.Fatal("paused widget reports active playback")
	}
	widget.Tick(time.Now())
	if widget.image != poster {
		t.Fatal("paused widget changed the displayed poster frame")
	}
	if got := len(frames); got != 1 {
		t.Fatalf("paused stream has %d queued frames, want 1 unconsumed frame", got)
	}
	resumed := make(chan image.Image, 1)
	widget.Play()
	if !widget.IsPlaying() || widget.TickInterval() != time.Second/24 {
		t.Fatal("Play did not resume playback")
	}
	widget.Tick(time.Now())
	if widget.image != queued {
		t.Fatal("resumed widget did not display the previously queued frame")
	}
	widget.Pause()
	widget.openStream = func(ctx context.Context, _ string, _ float64) (<-chan image.Image, func(), error) {
		go func() {
			<-ctx.Done()
			close(resumed)
		}()
		return resumed, func() {}, nil
	}
	if err := widget.Restart(); err != nil {
		t.Fatalf("Restart paused stream: %v", err)
	}
	resumed <- solidImage(3, 2, color.RGBA{G: 100, A: 255})
	widget.Tick(time.Now())
	if got := widget.image.Bounds().Dx(); got != 3 {
		t.Fatalf("restarted stream frame width = %d, want 3", got)
	}
	widget.Close()
}

func TestRestartRejectsClosedWidget(t *testing.T) {
	frames := make(chan image.Image)
	widget := newStreamingWidget(frames, 24, func() {})
	widget.path = "test-video"
	widget.openStream = func(context.Context, string, float64) (<-chan image.Image, func(), error) {
		return make(chan image.Image), func() {}, nil
	}
	widget.Close()
	if err := widget.Restart(); err == nil {
		t.Fatal("Restart on a closed widget succeeded")
	}
}

func TestRestartAfterStreamCompletion(t *testing.T) {
	frames := make(chan image.Image)
	var opens atomic.Int32
	opener := func(context.Context, string, float64) (<-chan image.Image, func(), error) {
		opens.Add(1)
		frames := make(chan image.Image)
		close(frames)
		return frames, func() {}, nil
	}
	widget, err := newVideoWithOpener("test-video", ModeHalfblock, 30, solidImage(2, 2, color.RGBA{A: 255}), opener)
	if err != nil {
		t.Fatal(err)
	}
	close(frames)
	widget.Tick(time.Now())
	if widget.IsPlaying() {
		t.Fatal("completed stream still reports playback")
	}
	if err := widget.Restart(); err != nil {
		t.Fatalf("Restart after completion: %v", err)
	}
	if opens.Load() != 2 || !widget.IsPlaying() {
		t.Fatalf("restart opens = %d, playing = %v", opens.Load(), widget.IsPlaying())
	}
	widget.Close()
}

func TestNewImageRejectsUnknownMode(t *testing.T) {
	_, err := NewImage(image.NewRGBA(image.Rect(0, 0, 1, 1)), Mode("unknown"))
	if err == nil {
		t.Fatal("NewImage accepted an unknown render mode")
	}
}

// smallWidget draws a 20x10 px image (10 rows of 1x2 px cells at most 10x5 cells) in a 40x12 area.
func smallWidget(t *testing.T) *Widget {
	t.Helper()
	w, err := NewImage(solidImage(20, 10, color.RGBA{G: 255, A: 255}), ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	w.Draw(loom.NewCanvas(40, 12), loom.Rect{W: 40, H: 12})
	return w
}

func TestSmallImagePansWithKeyboard(t *testing.T) {
	w := smallWidget(t)
	if !w.canPanLocked() {
		t.Fatal("small image must be pannable")
	}
	for _, name := range []string{"right", "down"} {
		before := [2]float64{w.panX, w.panY}
		w.ConsumeKey(loom.KeyEvent{Key: name})
		if before == [2]float64{w.panX, w.panY} {
			t.Fatalf("key %s did not change pan", name)
		}
	}
	if w.panX <= 0 || w.panY <= 0 {
		t.Fatalf("pan = (%v,%v), want both positive", w.panX, w.panY)
	}
	w.pan(5, 5)
	if w.panX != 1 || w.panY != 1 {
		t.Fatalf("pan clamp = (%v,%v), want (1,1)", w.panX, w.panY)
	}
}

func TestSmallImageDrawFollowsPan(t *testing.T) {
	w := smallWidget(t)
	firstCol := func() int {
		c := loom.NewCanvas(40, 12)
		w.Draw(c, loom.Rect{W: 40, H: 12})
		for x := 0; x < 40; x++ {
			if cell := c.Get(x, 0); cell.Text != " " && cell.Text != "" {
				return x
			}
		}
		for x := 0; x < 40; x++ {
			if c.Get(x, 3).Text != " " && c.Get(x, 3).Text != "" {
				return x
			}
		}
		return -1
	}
	w.pan(-1, -1)
	left := firstCol()
	w.pan(2, 0)
	right := firstCol()
	if left < 0 || right <= left {
		t.Fatalf("first image column left=%d right=%d, want right > left", left, right)
	}
}

func TestSmallImagePansWithMouseDrag(t *testing.T) {
	w := smallWidget(t)
	for _, e := range []loom.MouseEvent{
		{Action: loom.MousePress, Button: loom.MouseLeft, X: 20, Y: 4},
		{Action: loom.MouseDrag, Button: loom.MouseLeft, X: 24, Y: 6},
	} {
		if res := w.ConsumeMouse(e); !res.Consumed {
			t.Fatalf("event %#v not handled", e)
		}
	}
	// delta = 40 - 20 = 20 px wide, 20 - 10 = 10 px tall (rows 10 * 2px).
	// Dragging content right by 4 px moves pan by 2*4/20 = .4; 2 rows = 4 px, 2*4/10 = .8.
	if math.Abs(w.panX-.4) > 1e-9 || math.Abs(w.panY-.8) > 1e-9 {
		t.Fatalf("pan = (%v,%v), want (.4,.8)", w.panX, w.panY)
	}
}

func TestMouseWheelZoomKeepsCursorPointFixed(t *testing.T) {
	w, err := NewImage(solidImage(100, 60, color.RGBA{A: 255}), ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	w.Draw(loom.NewCanvas(20, 12), loom.Rect{W: 20, H: 12})
	// Viewport 20x20 px; image at 2x is 200x120 px, overflow 180x100.
	w.setZoom(2)
	cursorX, cursorY := 4, 3
	cx, cy := (float64(cursorX)+.5)*1, (float64(cursorY)+.5)*2
	before := func() (float64, float64) {
		dx, dy := panDelta(w.mode, w.image, 20, 10, w.zoom)
		zw, zh, _, _, _, _ := viewGeometry(w.mode, w.image, 20, 10, w.zoom)
		return (cx - float64(dx)/2*(1+w.panX)) / float64(zw), (cy - float64(dy)/2*(1+w.panY)) / float64(zh)
	}
	u0, v0 := before()
	oldPanX := w.panX
	w.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollUp, X: cursorX, Y: cursorY})
	u1, v1 := before()
	if w.zoom != 2*1.25 {
		t.Fatalf("zoom = %v, want 2.5", w.zoom)
	}
	if math.Abs(u0-u1) > 1e-6 || math.Abs(v0-v1) > 1e-6 {
		t.Fatalf("image point under cursor moved: (%v,%v) -> (%v,%v)", u0, v0, u1, v1)
	}
	if w.panX == oldPanX {
		t.Fatal("off-center cursor zoom did not shift pan")
	}
	// Keyboard zoom still preserves the viewport center, not the cursor.
	w.panX, w.panY = 0, 0
	w.ConsumeKey(loom.KeyEvent{Text: "+"})
	if w.panX != 0 || w.panY != 0 {
		t.Fatalf("keyboard zoom changed centered pan to (%v,%v)", w.panX, w.panY)
	}
}

func TestMouseWheelZoomOutAnchorsSmallImage(t *testing.T) {
	w := smallWidget(t)
	w.pan(.5, 0)
	dx0, _ := panDelta(w.mode, w.image, 40, 10, w.zoom)
	cx := 10.5
	u0 := (cx - float64(dx0)/2*(1+w.panX)) / 20
	w.setZoomAt(1/1.25, 10, 2)
	dx1, _ := panDelta(w.mode, w.image, 40, 10, w.zoom)
	zw, _, _, _, _, _ := viewGeometry(w.mode, w.image, 40, 10, w.zoom)
	u1 := (cx - float64(dx1)/2*(1+w.panX)) / float64(zw)
	if math.Abs(u0-u1) > 1e-6 {
		t.Fatalf("cursor fraction %v -> %v", u0, u1)
	}
}

func TestPanDeltaAndViewGeometryUnchangedForOverflow(t *testing.T) {
	img := solidImage(100, 60, color.RGBA{A: 255})
	zw, zh, vw, vh, mx, my := viewGeometry(ModeHalfblock, img, 20, 10, 2)
	if zw != 200 || zh != 120 || vw != 20 || vh != 20 || mx != 180 || my != 100 {
		t.Fatalf("geometry = %d %d %d %d %d %d", zw, zh, vw, vh, mx, my)
	}
	dx, dy := panDelta(ModeHalfblock, img, 20, 10, 2)
	if dx != -180 || dy != -100 {
		t.Fatalf("delta = %d,%d, want -180,-100", dx, dy)
	}
	if sx, sy := viewSlack(ModeHalfblock, img, 20, 10, 2); sx != 0 || sy != 0 {
		t.Fatalf("slack = %d,%d, want 0,0", sx, sy)
	}
}

func TestZoomCrossesSmallToLargeBoundary(t *testing.T) {
	w := smallWidget(t)
	// Viewport is 40x20 px (10 image rows); the 20x10 px image starts smaller than it.
	if dx, dy := panDelta(w.mode, w.image, 40, 10, w.zoom); dx <= 0 || dy <= 0 {
		t.Fatalf("start delta = %d,%d, want both positive", dx, dy)
	}
	cursorX, cursorY := 20, 5
	cx, cy := (float64(cursorX)+.5)*1, (float64(cursorY)+.5)*2
	fraction := func() (float64, float64) {
		dx, dy := panDelta(w.mode, w.image, 40, 10, w.zoom)
		zw, zh, _, _, _, _ := viewGeometry(w.mode, w.image, 40, 10, w.zoom)
		return (cx - float64(dx)/2*(1+w.panX)) / float64(zw), (cy - float64(dy)/2*(1+w.panY)) / float64(zh)
	}
	u0, v0 := fraction()
	w.setZoomAt(2.5, cursorX, cursorY)
	dx, dy := panDelta(w.mode, w.image, 40, 10, w.zoom)
	if dx >= 0 || dy >= 0 {
		t.Fatalf("zoomed delta = %d,%d, want both negative (overflow)", dx, dy)
	}
	u1, v1 := fraction()
	if math.Abs(u0-u1) > 1e-6 || math.Abs(v0-v1) > 1e-6 {
		t.Fatalf("image point under cursor moved across boundary: (%v,%v) -> (%v,%v)", u0, v0, u1, v1)
	}
}

func TestMouseWheelAtControlBar(t *testing.T) {
	w := smallWidget(t)
	imageHeight := w.imageHeightLocked()
	if w.lastRect.H < 3 || imageHeight != w.lastRect.H-2 {
		t.Fatalf("rect H=%d imageHeight=%d, want a 2-row control bar", w.lastRect.H, imageHeight)
	}
	before := w.zoom
	res := w.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollUp, X: 5, Y: imageHeight})
	if !res.Consumed {
		t.Fatal("wheel over control bar not handled")
	}
	if w.zoom <= before {
		t.Fatalf("zoom = %v, want > %v", w.zoom, before)
	}
	for name, pan := range map[string]float64{"panX": w.panX, "panY": w.panY} {
		if math.IsNaN(pan) || math.IsInf(pan, 0) || pan < -1 || pan > 1 {
			t.Fatalf("%s = %v, want finite in [-1,1]", name, pan)
		}
	}
}

func TestMixedAxesPanDelta(t *testing.T) {
	w, err := NewImage(solidImage(200, 10, color.RGBA{B: 255, A: 255}), ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	// 10 cols x 20 image rows (+2 control rows) = 10x40 px viewport.
	w.Draw(loom.NewCanvas(10, 22), loom.Rect{W: 10, H: 22})
	dx, dy := panDelta(w.mode, w.image, 10, 20, w.zoom)
	if dx >= 0 || dy <= 0 {
		t.Fatalf("delta = %d,%d, want x negative (overflow), y positive (slack)", dx, dy)
	}
	for _, e := range []loom.MouseEvent{
		{Action: loom.MousePress, Button: loom.MouseLeft, X: 5, Y: 5},
		{Action: loom.MouseDrag, Button: loom.MouseLeft, X: 6, Y: 6},
	} {
		if res := w.ConsumeMouse(e); !res.Consumed {
			t.Fatalf("event %#v not handled", e)
		}
	}
	// Dragging content right by 1 px shows more of the left: pan decreases on the overflow axis.
	if w.panX >= 0 {
		t.Fatalf("panX = %v, want negative after dragging right on overflow axis", w.panX)
	}
	// Dragging down moves the image down within slack: pan increases.
	if w.panY <= 0 {
		t.Fatalf("panY = %v, want positive after dragging down on slack axis", w.panY)
	}
}
