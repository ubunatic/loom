// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package media provides a loom widget backed by cati's terminal-cell renderers.
package media

import (
	"context"
	"fmt"
	"image"
	"image/draw"
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
	mu         sync.RWMutex
	cacheMu    sync.Mutex
	loadMu     sync.Mutex
	load       *renderLoad
	image      image.Image
	mode       Mode
	theme      loom.ThemeColors
	frames     <-chan image.Image
	interval   time.Duration
	path       string
	fps        float64
	closed     bool
	openStream func(context.Context, string, float64) (<-chan image.Image, func(), error)
	poster     image.Image
	stop       func()
	playing    bool
	zoom       float64
	panX       float64
	panY       float64
	lastRect   loom.Rect
	dragX      int
	dragY      int
	dragging   bool
	controls   loom.Rect
	playButton loom.Rect
	zoomOut    loom.Rect
	zoomIn     loom.Rect
	keys       *loom.KeyMap
	progress   *loom.ProgressBar
	help       *loom.KeyHelp
	renderer   func(image.Image, Mode, int, int) (*core.Grid, error)
	cache      renderCache
}

type renderLoad struct {
	image image.Image
	cols  int
	rows  int
	done  chan renderResult
}

type renderResult struct {
	grid *core.Grid
	err  error
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
	keys := loom.NewKeyMapWithLabels(map[string][]string{
		"play": {"space", " "}, "zoom_in": {"+", "="}, "zoom_out": {"-", "_"}, "pan": {"←/→/↑/↓", "left", "right", "up", "down"}, "reset": {"0"},
	}, map[string]string{"play": "Play/Pause", "zoom_in": "Zoom in", "zoom_out": "Zoom out", "pan": "Pan", "reset": "Reset"})
	bar := loom.NewProgressBar()
	bar.Options.Width = 8
	bar.Indeterminate = true
	return &Widget{image: src, mode: mode, theme: loom.Theme("plain"), zoom: 1, keys: keys, progress: bar, help: loom.NewKeyHelp(keys)}, nil
}

// NewImageWithTheme creates a still-image widget using the supplied theme.
func NewImageWithTheme(src image.Image, mode Mode, theme loom.ThemeColors) (*Widget, error) {
	w, err := NewImage(src, mode)
	if err != nil {
		return nil, err
	}
	w.theme = theme
	return w, nil
}

// ApplyTheme updates the theme used for media status indicators.
func (w *Widget) ApplyTheme(theme loom.ThemeColors) {
	if w == nil {
		return
	}
	w.mu.Lock()
	w.theme = theme
	w.mu.Unlock()
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
	return newVideo(path, mode, fps, image.NewRGBA(image.Rect(0, 0, 1, 1)))
}

// NewVideoWithPoster opens a video stream and displays poster until its first
// decoded frame arrives.
func NewVideoWithPoster(path string, mode Mode, fps float64, poster image.Image) (*Widget, error) {
	if poster == nil || (reflect.ValueOf(poster).Kind() == reflect.Pointer && reflect.ValueOf(poster).IsNil()) {
		return nil, fmt.Errorf("media: poster image is nil")
	}
	return newVideo(path, mode, fps, poster)
}

func newVideoWithOpener(path string, mode Mode, fps float64, poster image.Image, open func(context.Context, string, float64) (<-chan image.Image, func(), error)) (*Widget, error) {
	widget, err := NewImage(poster, mode)
	if err != nil {
		return nil, err
	}
	if math.IsNaN(fps) || math.IsInf(fps, 0) {
		return nil, fmt.Errorf("media: invalid frame rate %v", fps)
	}
	if fps <= 0 {
		fps = 24
	}
	widget.path = path
	widget.fps = fps
	widget.poster = poster
	widget.image = poster
	widget.interval = time.Duration(float64(time.Second) / fps)
	if widget.interval <= 0 {
		widget.interval = time.Nanosecond
	}
	widget.openStream = open
	if widget.openStream == nil {
		widget.openStream = func(ctx context.Context, path string, fps float64) (<-chan image.Image, func(), error) {
			return halfblock.OpenVideoStream(ctx, path, fps, 0, 0)
		}
	}
	if err := widget.startStream(); err != nil {
		return nil, err
	}
	return widget, nil
}

func newVideo(path string, mode Mode, fps float64, poster image.Image) (*Widget, error) {
	return newVideoWithOpener(path, mode, fps, poster, nil)
}

// startStream starts decoding from the beginning.
func (w *Widget) startStream() error {
	ctx, cancel := context.WithCancel(context.Background())
	frames, stop, err := w.openStream(ctx, w.path, w.fps)
	if err != nil {
		cancel()
		return err
	}
	streamStop := func() {
		cancel()
		// OpenVideoStream's cleanup and reader both call cmd.Wait. Drain until
		// its reader has finished waiting and closed the channel before cleanup
		// checks its completion signal, avoiding concurrent waits.
		for range frames {
		}
		stop()
	}
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		streamStop()
		return fmt.Errorf("media: video widget is closed")
	}
	w.frames = frames
	w.stop = streamStop
	w.playing = true
	if w.image == nil {
		w.image = w.poster
	}
	w.mu.Unlock()
	return nil
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
	theme := w.theme
	zoom, panX, panY := w.zoom, w.panX, w.panY
	playing := w.playing
	w.mu.RUnlock()
	w.mu.Lock()
	w.lastRect = r
	w.playButton, w.zoomOut, w.zoomIn = loom.Rect{}, loom.Rect{}, loom.Rect{}
	controlH := 0
	if r.H >= 3 {
		controlH = 2
	} else if r.H == 2 {
		controlH = 1
	}
	imageRect := r
	imageRect.H = max(0, r.H-controlH)
	w.controls = loom.Rect{X: 0, Y: r.H - controlH, W: r.W, H: controlH}
	if r.W >= 9 && controlH > 0 {
		w.playButton = loom.Rect{X: 0, Y: r.H - 1, W: 6, H: 1}
	}
	if r.W >= 17 && controlH > 0 {
		w.zoomOut = loom.Rect{X: 7, Y: r.H - 1, W: 3, H: 1}
		w.zoomIn = loom.Rect{X: 11, Y: r.H - 1, W: 3, H: 1}
	}
	w.mu.Unlock()
	if img == nil || imageRect.W <= 0 || imageRect.H <= 0 {
		w.drawControls(c, playing, zoom)
		return
	}
	img = cropForView(img, imageRect.W, imageRect.H, zoom, panX, panY)
	for y := 0; y < imageRect.H; y++ {
		for x := 0; x < imageRect.W; x++ {
			c.Set(r.X+x, r.Y+y, loom.Cell{Text: " ", Style: loom.Reset})
		}
	}
	grid, err, loading := w.renderWithThreshold(img, imageRect.W, imageRect.H)
	if loading {
		label := loom.SpeccedDefaults.Media.LoadingLabel
		if loom.StringWidth(label) > imageRect.W {
			label = loom.TruncateText(label, imageRect.W, "")
		}
		c.Write(r.X, r.Y, label, loom.Style{FG: theme.MediaLoadingFG.Color(), Dim: theme.MediaLoadingDim})
		w.drawControls(c, playing, zoom)
		return
	}
	if err != nil {
		message := loom.SpeccedDefaults.Media.RenderErrorLabel
		if loom.StringWidth(message) > imageRect.W {
			message = loom.TruncateText(message, imageRect.W, "")
		}
		c.Write(r.X, r.Y, message, loom.Style{FG: theme.MediaErrorFG.Color()})
		w.drawControls(c, playing, zoom)
		return
	}
	if grid == nil {
		return
	}
	offsetX := (imageRect.W - grid.Width) / 2
	offsetY := (imageRect.H - grid.Height) / 2
	for y, row := range grid.Cells {
		for x, cell := range row {
			dstX, dstY := r.X+offsetX+x, r.Y+offsetY+y
			if dstX < r.X || dstX >= r.X+imageRect.W || dstY < r.Y || dstY >= r.Y+imageRect.H {
				continue
			}
			c.Set(dstX, dstY, canvasCell(cell))
		}
	}
	w.drawControls(c, playing, zoom)
}

func cropForView(src image.Image, cols, rows int, zoom, panX, panY float64) image.Image {
	if src == nil || zoom <= 1 || cols <= 0 || rows <= 0 {
		return src
	}
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	// Terminal cells are approximately twice as tall as they are wide.
	wantAspect := float64(cols) / float64(rows*2)
	cw, ch := float64(sw)/zoom, float64(sh)/zoom
	if cw/ch > wantAspect {
		cw = ch * wantAspect
	} else {
		ch = cw / wantAspect
	}
	maxX, maxY := float64(sw)-cw, float64(sh)-ch
	x0, y0 := float64(b.Min.X)+panX*maxX, float64(b.Min.Y)+panY*maxY
	box := image.Rect(int(x0), int(y0), int(x0+cw), int(y0+ch)).Intersect(b)
	if box.Empty() {
		return src
	}
	out := image.NewRGBA(image.Rect(0, 0, box.Dx(), box.Dy()))
	draw.Draw(out, out.Bounds(), src, box.Min, draw.Src)
	return out
}

func (w *Widget) drawControls(c *loom.Canvas, playing bool, zoom float64) {
	w.mu.RLock()
	controls, play, minus, plus := w.controls, w.playButton, w.zoomOut, w.zoomIn
	base := w.lastRect
	w.mu.RUnlock()
	if controls.H <= 0 {
		return
	}
	style := loom.Style{FG: loom.ColorRGB(235, 235, 235), BG: loom.ColorRGB(38, 48, 60)}
	controls.X += base.X
	controls.Y += base.Y
	c.PaintSurface(controls, style)
	if controls.H > 1 {
		w.help.Draw(c, loom.Rect{X: controls.X, Y: controls.Y, W: controls.W, H: 1})
	}
	play.X += base.X
	play.Y += base.Y
	minus.X += base.X
	minus.Y += base.Y
	plus.X += base.X
	plus.Y += base.Y
	if play.W > 0 {
		label := "▶ Play"
		if playing {
			label = "❚❚ Pause"
		}
		c.Write(play.X, play.Y, label, style)
	}
	if minus.W > 0 {
		c.Write(minus.X, minus.Y, "[-]", style)
		c.Write(plus.X, plus.Y, "[+]", style)
	}
	label := fmt.Sprintf("%.2fx", zoom)
	x := controls.X + 14
	if controls.W >= 24 {
		c.Write(x, controls.Y+controls.H-1, label, style)
		x += len(label) + 1
	}
	if w.progress != nil && playing && controls.H > 1 && x < controls.X+controls.W {
		w.progress.Draw(c, loom.Rect{X: x, Y: controls.Y + controls.H - 1, W: min(8, controls.X+controls.W-x), H: 1})
	}
}

func (w *Widget) renderWithThreshold(img image.Image, cols, rows int) (*core.Grid, error, bool) {
	w.loadMu.Lock()
	if load := w.load; load != nil && sameImage(load.image, img) && load.cols == cols && load.rows == rows {
		select {
		case result := <-load.done:
			w.load = nil
			w.loadMu.Unlock()
			return w.storeRender(img, cols, rows, result.grid, result.err)
		default:
			w.loadMu.Unlock()
			return nil, nil, true
		}
	}
	load := &renderLoad{image: img, cols: cols, rows: rows, done: make(chan renderResult, 1)}
	w.load = load
	w.loadMu.Unlock()
	go func() {
		grid, err := w.renderCached(img, cols, rows)
		load.done <- renderResult{grid: grid, err: err}
	}()
	timer := time.NewTimer(loom.SpeccedDefaults.Media.LoadingThreshold)
	defer timer.Stop()
	select {
	case result := <-load.done:
		w.loadMu.Lock()
		if w.load == load {
			w.load = nil
		}
		w.loadMu.Unlock()
		return result.grid, result.err, false
	case <-timer.C:
		return nil, nil, true
	}
}

func (w *Widget) storeRender(img image.Image, cols, rows int, grid *core.Grid, err error) (*core.Grid, error, bool) {
	w.cacheMu.Lock()
	w.cache = renderCache{image: img, cols: cols, rows: rows, grid: grid, err: err, valid: true}
	w.cacheMu.Unlock()
	return grid, err, false
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

// Pause stops frame advancement and preserves the current image and stream
// queue. Play resumes consuming frames buffered while paused.
func (w *Widget) Pause() {
	if w == nil {
		return
	}
	w.mu.Lock()
	if !w.playing {
		w.mu.Unlock()
		return
	}
	w.playing = false
	w.mu.Unlock()
}

// Play resumes a paused stream or starts a completed stream from the beginning.
// A closed widget cannot be restarted.
func (w *Widget) Play() {
	if w == nil {
		return
	}
	w.mu.RLock()
	if w.playing || w.closed {
		w.mu.RUnlock()
		return
	}
	paused := w.frames != nil
	if !paused && w.path == "" {
		w.mu.RUnlock()
		return
	}
	w.mu.RUnlock()
	if paused {
		w.mu.Lock()
		if !w.closed && !w.playing {
			w.playing = true
		}
		w.mu.Unlock()
		return
	}
	_ = w.startStream()
}

// IsPlaying reports whether the video stream is advancing.
func (w *Widget) IsPlaying() bool {
	if w == nil {
		return false
	}
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.playing
}

// Restart closes any active stream and opens a fresh stream from the beginning.
func (w *Widget) Restart() error {
	if w == nil {
		return fmt.Errorf("media: widget is nil")
	}
	w.mu.Lock()
	if w.closed || w.path == "" {
		w.mu.Unlock()
		return fmt.Errorf("media: video widget is closed or has no stream")
	}
	w.playing = false
	stop := w.stop
	w.stop = nil
	w.mu.Unlock()
	if stop != nil {
		stop()
	}
	return w.startStream()
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
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return
	}
	w.closed = true
	w.playing = false
	w.frames = nil
	stop := w.stop
	w.stop = nil
	w.mu.Unlock()
	if stop != nil {
		stop()
	}
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

// HandleKey preserves the legacy quit-only widget contract.
func (*Widget) HandleKey(loom.KeyEvent) bool { return false }

// ConsumeKey applies media keyboard controls and reports consumption separately from quit.
func (w *Widget) ConsumeKey(e loom.KeyEvent) (quit, consumed bool) {
	if w == nil {
		return false, false
	}
	switch w.keys.Action(e) {
	case "play":
		if w.IsPlaying() {
			w.Pause()
		} else {
			w.Play()
		}
	case "zoom_in":
		w.setZoom(1.25)
	case "zoom_out":
		w.setZoom(1 / 1.25)
	case "pan":
		switch e.Name() {
		case "left":
			w.pan(-.12, 0)
		case "right":
			w.pan(.12, 0)
		case "up":
			w.pan(0, -.12)
		case "down":
			w.pan(0, .12)
		default:
			return false, false
		}
	case "reset":
		w.mu.Lock()
		w.zoom, w.panX, w.panY = 1, 0, 0
		w.mu.Unlock()
	default:
		return false, false
	}
	return false, true
}

func (w *Widget) setZoom(factor float64) {
	w.mu.Lock()
	w.zoom = math.Max(1, math.Min(8, w.zoom*factor))
	if w.zoom == 1 {
		w.panX, w.panY = 0, 0
	}
	w.mu.Unlock()
}
func (w *Widget) pan(dx, dy float64) {
	w.mu.Lock()
	if w.zoom > 1 {
		w.panX = math.Max(0, math.Min(1, w.panX+dx))
		w.panY = math.Max(0, math.Min(1, w.panY+dy))
	}
	w.mu.Unlock()
}

// HandleMouse preserves the legacy quit-only widget contract.
func (*Widget) HandleMouse(loom.MouseEvent) bool { return false }

// ConsumeMouse uses child-local 0-based coordinates for the control bar and image.
// ConsumeMouseEvent adapts ConsumeMouse to loom's MouseConsumer dispatch so the
// widget receives clicks when mounted directly as a pane root.
func (w *Widget) ConsumeMouseEvent(e loom.MouseEvent) loom.EventResult {
	quit, consumed := w.ConsumeMouse(e)
	return loom.EventResult{Consumed: consumed, Quit: quit}
}

func (w *Widget) ConsumeMouse(e loom.MouseEvent) (quit, consumed bool) {
	if w == nil {
		return false, false
	}
	w.mu.Lock()
	if e.Action == loom.MouseRelease {
		wasDragging := w.dragging
		w.dragging = false
		w.mu.Unlock()
		return false, wasDragging
	}
	if e.Action == loom.MousePress && e.Button == loom.MouseLeft {
		switch {
		case w.playButton.Contains(e.X, e.Y):
			w.mu.Unlock()
			if w.IsPlaying() {
				w.Pause()
			} else {
				w.Play()
			}
			return false, true
		case w.zoomOut.Contains(e.X, e.Y):
			w.mu.Unlock()
			w.setZoom(1 / 1.25)
			return false, true
		case w.zoomIn.Contains(e.X, e.Y):
			w.mu.Unlock()
			w.setZoom(1.25)
			return false, true
		case e.Y < w.lastRect.H-2 && e.X < w.lastRect.W && w.zoom > 1:
			w.dragging = true
			w.dragX = e.X
			w.dragY = e.Y
			w.mu.Unlock()
			return false, true
		}
	}
	if e.Action == loom.MouseDrag && w.dragging {
		dx, dy := e.X-w.dragX, e.Y-w.dragY
		w.dragX, w.dragY = e.X, e.Y
		width, height := max(1, w.lastRect.W), max(1, w.lastRect.H-2)
		w.mu.Unlock()
		w.pan(-float64(dx)/float64(width), -float64(dy)/float64(height))
		return false, true
	}
	w.mu.Unlock()
	if e.Action == loom.MouseScrollUp {
		w.setZoom(1.25)
		return false, true
	}
	if e.Action == loom.MouseScrollDown {
		w.setZoom(1 / 1.25)
		return false, true
	}
	return false, false
}
