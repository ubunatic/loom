// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package media provides a loom widget backed by cati's terminal-cell renderers.
package media

import (
	"context"
	"fmt"
	"image"
	"math"
	"reflect"
	"sync"
	"time"

	"codeberg.org/ubunatic/loom"
	"ubunatic.com/cati/v1/core"
	"ubunatic.com/cati/v1/halfblock"
	"ubunatic.com/cati/v1/quadblock"
	"ubunatic.com/cati/v1/sextant"
)

// Mode selects cati's cell geometry for media rendering.
type Mode string

const (
	ModeHalfblock Mode = "halfblock"
	ModeQuadblock Mode = "quadblock"
	ModeSextant   Mode = "sextant"
)

// Widget renders a still image or video frame inside the rectangle supplied by loom.
type Widget struct {
	mu        sync.RWMutex
	cacheMu   sync.Mutex
	image     image.Image
	mode      Mode
	frames    <-chan image.Image
	interval  time.Duration
	stop      func()
	playing   bool
	closeOnce sync.Once
	renderer  func(image.Image, Mode, int, int) (*core.Grid, error)
	cache     renderCache
}

type renderCache struct {
	image image.Image
	cols  int
	rows  int
	grid  *core.Grid
	err   error
	valid bool
}

// NewImage creates a still-image widget using mode. A zero mode selects
// half-block rendering.
func NewImage(src image.Image, mode Mode) (*Widget, error) {
	if src == nil {
		return nil, fmt.Errorf("media: image is nil")
	}
	if mode == "" {
		mode = ModeHalfblock
	}
	switch mode {
	case ModeHalfblock, ModeQuadblock, ModeSextant:
	default:
		return nil, fmt.Errorf("media: unsupported render mode %q", mode)
	}
	return &Widget{image: src, mode: mode}, nil
}

// LoadImage decodes a PNG, JPEG, or SVG image using cati's image loader.
func LoadImage(path string, mode Mode) (*Widget, error) {
	src, err := halfblock.LoadImage(path)
	if err != nil {
		return nil, err
	}
	return NewImage(src, mode)
}

// NewVideo opens a cati video frame stream. A non-positive fps uses 24 frames
// per second; ffprobe and ffmpeg must be available on PATH.
func NewVideo(path string, mode Mode, fps float64) (*Widget, error) {
	still, err := NewImage(image.NewRGBA(image.Rect(0, 0, 1, 1)), mode)
	if err != nil {
		return nil, err
	}
	if math.IsNaN(fps) || math.IsInf(fps, 0) {
		return nil, fmt.Errorf("media: invalid frame rate %v", fps)
	}
	if fps <= 0 {
		fps = 24
	}
	ctx, cancel := context.WithCancel(context.Background())
	frames, stop, err := halfblock.OpenVideoStream(ctx, path, fps, 0, 0)
	if err != nil {
		cancel()
		return nil, err
	}
	still.frames = frames
	still.interval = time.Duration(float64(time.Second) / fps)
	if still.interval <= 0 {
		still.interval = time.Nanosecond
	}
	still.stop = func() {
		cancel()
		// OpenVideoStream's cleanup and reader both call cmd.Wait. Drain until
		// its reader has finished waiting and closed the channel before cleanup
		// checks its completion signal, avoiding concurrent waits.
		for range frames {
		}
		stop()
	}
	still.playing = true
	return still, nil
}

// Mode returns the widget's selected renderer mode.
func (w *Widget) Mode() Mode {
	if w == nil {
		return ""
	}
	return w.mode
}

// Draw paints a cati-rendered image, clipped to r.
func (w *Widget) Draw(c *loom.Canvas, r loom.Rect) {
	if w == nil || c == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	w.mu.RLock()
	img := w.image
	w.mu.RUnlock()
	if img == nil {
		return
	}
	for y := 0; y < r.H; y++ {
		for x := 0; x < r.W; x++ {
			c.Set(r.X+x, r.Y+y, loom.Cell{Text: " ", Style: loom.Reset})
		}
	}
	grid, err := w.renderCached(img, r.W, r.H)
	if err != nil {
		message := "render error"
		if len(message) > r.W {
			message = message[:r.W]
		}
		c.Write(r.X, r.Y, message, loom.Style{FG: loom.ColorRGB(255, 96, 96)})
		return
	}
	if grid == nil {
		return
	}
	offsetX := (r.W - grid.Width) / 2
	offsetY := (r.H - grid.Height) / 2
	for y, row := range grid.Cells {
		for x, cell := range row {
			dstX, dstY := r.X+offsetX+x, r.Y+offsetY+y
			if dstX < r.X || dstX >= r.X+r.W || dstY < r.Y || dstY >= r.Y+r.H {
				continue
			}
			c.Set(dstX, dstY, canvasCell(cell))
		}
	}
}

func (w *Widget) renderCached(img image.Image, cols, rows int) (*core.Grid, error) {
	if !cacheableImage(img) {
		return w.render(img, cols, rows)
	}
	w.cacheMu.Lock()
	defer w.cacheMu.Unlock()
	if w.cache.valid && sameImage(w.cache.image, img) && w.cache.cols == cols && w.cache.rows == rows {
		return w.cache.grid, w.cache.err
	}
	grid, err := w.render(img, cols, rows)
	w.cache = renderCache{image: img, cols: cols, rows: rows, grid: grid, err: err, valid: true}
	return grid, err
}

func cacheableImage(img image.Image) bool {
	if img == nil {
		return false
	}
	v := reflect.ValueOf(img)
	return v.Kind() == reflect.Pointer || v.Comparable()
}

func sameImage(a, b image.Image) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	if va.Type() != vb.Type() {
		return false
	}
	if va.Kind() == reflect.Pointer {
		return va.Pointer() == vb.Pointer()
	}
	return va.Comparable() && va.Interface() == vb.Interface()
}

func (w *Widget) render(img image.Image, cols, rows int) (*core.Grid, error) {
	if w.renderer != nil {
		return w.renderer(img, w.mode, cols, rows)
	}
	switch w.mode {
	case ModeQuadblock:
		return quadblock.RenderToGrid(img, cols, quadblock.Options{Rows: rows})
	case ModeSextant:
		return sextant.RenderToGrid(img, cols, sextant.Options{Rows: rows})
	default:
		return halfblock.RenderToGrid(img, cols, halfblock.Options{Rows: rows})
	}
}

// Tick advances a video widget to the newest available frame.
func (w *Widget) Tick(time.Time) {
	if w == nil {
		return
	}
	w.mu.Lock()
	if !w.playing {
		w.mu.Unlock()
		return
	}
	var stop func()
	for {
		select {
		case frame, ok := <-w.frames:
			if !ok {
				w.playing = false
				w.frames = nil
				stop, w.stop = w.stop, nil
				w.mu.Unlock()
				if stop != nil {
					stop()
				}
				return
			}
			if frame != nil {
				w.image = frame
			}
		default:
			w.mu.Unlock()
			return
		}
	}
}

// TickInterval is zero for still images and finished or closed videos.
func (w *Widget) TickInterval() time.Duration {
	if w == nil {
		return 0
	}
	w.mu.RLock()
	defer w.mu.RUnlock()
	if !w.playing {
		return 0
	}
	return w.interval
}

// Close stops the video process and releases its frame stream. It is safe to
// call more than once.
func (w *Widget) Close() {
	if w == nil {
		return
	}
	w.closeOnce.Do(func() {
		w.mu.Lock()
		w.playing = false
		w.frames = nil
		stop := w.stop
		w.stop = nil
		w.mu.Unlock()
		if stop != nil {
			stop()
		}
	})
}

func newStreamingWidget(frames <-chan image.Image, fps float64, stop func()) *Widget {
	w, _ := NewImage(image.NewRGBA(image.Rect(0, 0, 1, 1)), ModeHalfblock)
	w.frames = frames
	w.interval = time.Duration(float64(time.Second) / fps)
	w.stop = stop
	w.playing = true
	return w
}

func canvasCell(cell core.Cell) loom.Cell {
	text := " "
	if !cell.Transparent && cell.Ch != 0 {
		text = string(cell.Ch)
	}
	style := loom.Style{}
	if cell.HasFg {
		style.FG = loom.ColorRGB(cell.Fg.R, cell.Fg.G, cell.Fg.B)
	}
	if cell.HasBg {
		style.BG = loom.ColorRGB(cell.Bg.R, cell.Bg.G, cell.Bg.B)
	}
	return loom.Cell{Text: text, Style: style}
}

// HandleKey reports that this widget does not consume keys.
func (*Widget) HandleKey(loom.KeyEvent) bool { return false }

// HandleMouse reports that this widget does not consume mouse events.
func (*Widget) HandleMouse(loom.MouseEvent) bool { return false }
