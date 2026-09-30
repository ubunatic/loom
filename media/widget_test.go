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
	w.setZoom(1.25)
	if w.zoom != 1.25 {
		t.Fatalf("zoom = %v, want 1.25", w.zoom)
	}
	w.pan(-4, 4)
	if w.panX != 0 || w.panY != 1 {
		t.Fatalf("pan = (%v,%v), want clamped (0,1)", w.panX, w.panY)
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
	if quit, used := w.ConsumeKey(loom.KeyEvent{Text: "0"}); quit || !used {
		t.Fatal("reset key was not consumed")
	}
	if w.zoom != 1 {
		t.Fatalf("reset zoom = %v, want 1", w.zoom)
	}
}

func TestMediaControlsPlayPauseKeysAndMouseWheel(t *testing.T) {
	w := newStreamingWidget(make(chan image.Image), 24, nil)
	if quit, used := w.ConsumeKey(loom.KeyEvent{Text: " "}); quit || !used || w.IsPlaying() {
		t.Fatal("space did not pause playback")
	}
	if quit, used := w.ConsumeKey(loom.KeyEvent{Text: " "}); quit || !used || !w.IsPlaying() {
		t.Fatal("space did not resume playback")
	}
	if quit, used := w.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollUp, X: 0, Y: 0}); quit || !used || w.zoom <= 1 {
		t.Fatal("wheel up did not zoom in")
	}
	if quit, used := w.ConsumeMouse(loom.MouseEvent{Action: loom.MouseScrollDown, X: 0, Y: 0}); quit || !used || w.zoom != 1 {
		t.Fatal("wheel down did not return to fit")
	}
}

func TestCropForViewPanChangesCropAndClampsInput(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 60))
	for y := 0; y < 60; y++ {
		for x := 0; x < 100; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x), G: uint8(y), A: 255})
		}
	}
	left := cropForView(img, 20, 10, 2, 0, 0)
	right := cropForView(img, 20, 10, 2, 1, 1)
	a, b := left.Bounds(), right.Bounds()
	if a.Dx() != b.Dx() || a.Dy() != b.Dy() {
		t.Fatalf("crop sizes differ: %v and %v", a, b)
	}
	if left.At(a.Min.X, a.Min.Y) == right.At(b.Min.X, b.Min.Y) {
		t.Fatal("pan did not change the visible source crop")
	}
}

func TestMediaControlBarClickAndDragPan(t *testing.T) {
	w, err := NewImage(solidImage(40, 20, color.RGBA{R: 255, A: 255}), ModeHalfblock)
	if err != nil {
		t.Fatal(err)
	}
	w.Draw(loom.NewCanvas(40, 8), loom.Rect{X: 0, Y: 0, W: 40, H: 8})
	if quit, used := w.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 12, Y: 7}); quit || !used || w.zoom != 1.25 {
		t.Fatalf("zoom button: quit=%v used=%v zoom=%v", quit, used, w.zoom)
	}
	if quit, used := w.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 20, Y: 2}); quit || !used {
		t.Fatal("zoomed image press did not start pan drag")
	}
	if quit, used := w.ConsumeMouse(loom.MouseEvent{Action: loom.MouseDrag, Button: loom.MouseLeft, X: 10, Y: 2}); quit || !used || w.panX <= 0 {
		t.Fatalf("drag pan failed: quit=%v used=%v pan=%v", quit, used, w.panX)
	}
	if quit, used := w.ConsumeMouse(loom.MouseEvent{Action: loom.MouseRelease, Button: loom.MouseLeft}); quit || !used {
		t.Fatal("drag release was not consumed")
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
