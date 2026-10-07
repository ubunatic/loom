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

	"ubunatic.com/cati/v1/core"
	"ubunatic.com/cati/v1/halfblock"
	"ubunatic.com/cati/v1/quadblock"
	"ubunatic.com/cati/v1/sextant"
	"ubunatic.com/loom"
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
	viewSource image.Image
	viewMode   Mode
	viewCols   int
	viewRows   int
	viewZoom   float64
	viewPanX   float64
	viewPanY   float64
	viewImage  image.Image
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
	return &Widget{image: src, mode: mode, theme: loom.DefaultTheme(), zoom: 1, keys: keys, progress: bar, help: loom.NewKeyHelp(keys)}, nil
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
	slackX, slackY := viewSlack(w.mode, img, imageRect.W, imageRect.H, zoom)
	img = w.cachedScaleForView(img, imageRect.W, imageRect.H, zoom, panX, panY)
	c.PaintSurface(imageRect, loom.Style{FG: theme.NormalFG.Color(), BG: theme.NormalBG.Color()})
	grid, err, loading := w.renderWithThreshold(img, 0, 0)
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
	// Media smaller than the viewport is placed by pan instead of centered.
	if slackX > 0 && imageRect.W > grid.Width {
		offsetX = panOffset(imageRect.W-grid.Width, panX)
	}
	if slackY > 0 && imageRect.H > grid.Height {
		offsetY = panOffset(imageRect.H-grid.Height, panY)
	}
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

func pxPerCell(mode Mode) (int, int) {
	switch mode {
	case ModeQuadblock:
		return 2, 2
	case ModeSextant:
		return 2, 3
	default:
		return 1, 2
	}
}

func nativeCellSize(mode Mode, img image.Image) (int, int) {
	if img == nil {
		return 0, 0
	}
	w, h := pxPerCell(mode)
	b := img.Bounds()
	return (b.Dx() + w - 1) / w, (b.Dy() + h - 1) / h
}

func (w *Widget) cachedScaleForView(src image.Image, cols, rows int, zoom, panX, panY float64) image.Image {
	w.mu.Lock()
	defer w.mu.Unlock()
	if sameImage(w.viewSource, src) && w.viewMode == w.mode && w.viewCols == cols && w.viewRows == rows && w.viewZoom == zoom && w.viewPanX == panX && w.viewPanY == panY && w.viewImage != nil {
		return w.viewImage
	}
	out := scaleForView(src, w.mode, cols, rows, zoom, panX, panY)
	w.viewSource, w.viewMode = src, w.mode
	w.viewCols, w.viewRows = cols, rows
	w.viewZoom, w.viewPanX, w.viewPanY = zoom, panX, panY
	w.viewImage = out
	return out
}

// viewGeometry describes the zoomed image, viewport and overflow in pixels.
func viewGeometry(mode Mode, img image.Image, cols, rows int, zoom float64) (zoomedW, zoomedH, visibleW, visibleH, maxX, maxY int) {
	if img == nil || cols <= 0 || rows <= 0 || zoom <= 0 || img.Bounds().Empty() {
		return
	}
	cellW, cellH := pxPerCell(mode)
	zoomedW = max(1, int(math.Round(float64(img.Bounds().Dx())*zoom)))
	zoomedH = max(1, int(math.Round(float64(img.Bounds().Dy())*zoom)))
	visibleW, visibleH = min(zoomedW, cols*cellW), min(zoomedH, rows*cellH)
	maxX, maxY = zoomedW-visibleW, zoomedH-visibleH
	return
}

// viewSlack returns the spare viewport pixels per axis when the zoomed image is smaller than the viewport.
func viewSlack(mode Mode, img image.Image, cols, rows int, zoom float64) (slackX, slackY int) {
	dx, dy := panDelta(mode, img, cols, rows, zoom)
	return max(0, dx), max(0, dy)
}

// panDelta returns viewport minus zoomed size in pixels per axis. Positive means
// spare room, negative means overflow; the image origin is delta/2*(1+pan).
func panDelta(mode Mode, img image.Image, cols, rows int, zoom float64) (dx, dy int) {
	zoomedW, zoomedH, _, _, _, _ := viewGeometry(mode, img, cols, rows, zoom)
	if zoomedW == 0 || zoomedH == 0 {
		return 0, 0
	}
	cellW, cellH := pxPerCell(mode)
	return cols*cellW - zoomedW, rows*cellH - zoomedH
}

// panOffset maps a normalized pan in [-1, 1] to a cell offset within slack cells.
func panOffset(slack int, pan float64) int {
	return max(0, min(slack, int(math.Round(float64(slack)/2*(1+pan)))))
}

func scaleForView(src image.Image, mode Mode, cols, rows int, zoom, panX, panY float64) image.Image {
	targetW, targetH, visibleW, visibleH, maxX, maxY := viewGeometry(mode, src, cols, rows, zoom)
	if targetW == 0 || targetH == 0 {
		return src
	}
	scaled := halfblock.ScaleNN(src, targetW, targetH)
	if maxX == 0 && maxY == 0 {
		return scaled
	}
	x0 := int(math.Round(float64(maxX)/2 + panX*float64(maxX)/2))
	y0 := int(math.Round(float64(maxY)/2 + panY*float64(maxY)/2))
	x0 = max(0, min(maxX, x0))
	y0 = max(0, min(maxY, y0))
	sb := scaled.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, visibleW, visibleH))
	for y := 0; y < visibleH; y++ {
		for x := 0; x < visibleW; x++ {
			out.Set(x, y, scaled.At(sb.Min.X+x0+x, sb.Min.Y+y0+y))
		}
	}
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

// ConsumeKey applies media keyboard controls.
func (w *Widget) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	if w == nil {
		return loom.Ignored()
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
			return loom.Ignored()
		}
	case "reset":
		w.mu.Lock()
		w.zoom, w.panX, w.panY = 1, 0, 0
		w.mu.Unlock()
	default:
		return loom.Ignored()
	}
	return loom.Handled()
}

func (w *Widget) setZoom(factor float64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	img := w.image
	cols, rows := max(1, w.lastRect.W), max(1, w.imageHeightLocked())
	oldW, oldH, _, _, oldX, oldY := viewGeometry(w.mode, img, cols, rows, w.zoom)
	w.applyZoomLocked(factor)
	newW, newH, _, _, newX, newY := viewGeometry(w.mode, img, cols, rows, w.zoom)
	// Normalized pan measures displacement from the image center over half
	// the overflow. Scale that displacement to retain the same source center.
	w.panX = zoomPan(w.panX, oldW, newW, oldX, newX)
	w.panY = zoomPan(w.panY, oldH, newH, oldY, newY)
}

// applyZoomLocked multiplies the zoom by factor within the allowed range.
func (w *Widget) applyZoomLocked(factor float64) {
	img := w.image
	nativeW, nativeH := nativeCellSize(w.mode, img)
	maxCells := max(nativeW, nativeH)
	minZoom := 1.0 / 16
	if maxCells > 0 {
		minZoom = math.Max(minZoom, 1/float64(maxCells))
	}
	w.zoom = math.Max(minZoom, math.Min(8, w.zoom*factor))
}

// setZoomAt zooms like setZoom but keeps the image point under the cursor
// (child-local cell coordinates) fixed.
func (w *Widget) setZoomAt(factor float64, cursorX, cursorY int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	img := w.image
	cols, rows := max(1, w.lastRect.W), max(1, w.imageHeightLocked())
	oldW, oldH, _, _, _, _ := viewGeometry(w.mode, img, cols, rows, w.zoom)
	oldDX, oldDY := panDelta(w.mode, img, cols, rows, w.zoom)
	w.applyZoomLocked(factor)
	newW, newH, _, _, _, _ := viewGeometry(w.mode, img, cols, rows, w.zoom)
	newDX, newDY := panDelta(w.mode, img, cols, rows, w.zoom)
	if oldW <= 0 || oldH <= 0 || newW <= 0 || newH <= 0 {
		return
	}
	cellW, cellH := pxPerCell(w.mode)
	cx := (float64(cursorX) + .5) * float64(cellW)
	cy := (float64(cursorY) + .5) * float64(cellH)
	w.panX = anchorPan(w.panX, cx, oldW, oldDX, newW, newDX)
	w.panY = anchorPan(w.panY, cy, oldH, oldDY, newH, newDY)
}

// anchorPan returns the pan that keeps the image point under cursor px fixed.
func anchorPan(pan, cursor float64, oldSize, oldDelta, newSize, newDelta int) float64 {
	if newDelta == 0 {
		return 0
	}
	oldOrigin := float64(oldDelta) / 2 * (1 + pan)
	frac := (cursor - oldOrigin) / float64(oldSize)
	newOrigin := cursor - frac*float64(newSize)
	return clampPan(2*newOrigin/float64(newDelta)-1, abs(newDelta))
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func zoomPan(pan float64, oldSize, newSize, oldOverflow, newOverflow int) float64 {
	if oldSize <= 0 || newOverflow <= 0 {
		return 0
	}
	return clampPan(pan*float64(oldOverflow)*float64(newSize)/float64(oldSize)/float64(newOverflow), newOverflow)
}

func clampPan(pan float64, overflow int) float64 {
	if overflow <= 0 {
		return 0
	}
	return math.Max(-1, math.Min(1, pan))
}
func (w *Widget) pan(dx, dy float64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	cols, rows := max(1, w.lastRect.W), max(1, w.imageHeightLocked())
	dX, dY := panDelta(w.mode, w.image, cols, rows, w.zoom)
	w.panX = clampPan(w.panX+dx, abs(dX))
	w.panY = clampPan(w.panY+dy, abs(dY))
}

// ConsumeMouse uses child-local 0-based coordinates for the control bar and image.
func (w *Widget) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	if w == nil {
		return loom.Ignored()
	}
	w.mu.Lock()
	if e.Action == loom.MouseRelease {
		wasDragging := w.dragging
		w.dragging = false
		w.mu.Unlock()
		if wasDragging {
			return loom.Handled()
		}
		return loom.Ignored()
	}
	if w.dragging && e.Action == loom.MousePress {
		w.mu.Unlock()
		return loom.Handled()
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
			return loom.Handled()
		case w.zoomOut.Contains(e.X, e.Y):
			w.mu.Unlock()
			w.setZoom(1 / 1.25)
			return loom.Handled()
		case w.zoomIn.Contains(e.X, e.Y):
			w.mu.Unlock()
			w.setZoom(1.25)
			return loom.Handled()
		case e.X >= 0 && e.X < w.lastRect.W && e.Y >= 0 && e.Y < w.imageHeightLocked() && w.canPanLocked():
			w.dragging = true
			w.dragX = e.X
			w.dragY = e.Y
			w.mu.Unlock()
			return loom.Handled()
		}
	}
	if e.Action == loom.MouseDrag && w.dragging {
		dx, dy := e.X-w.dragX, e.Y-w.dragY
		w.dragX, w.dragY = e.X, e.Y
		dX, dY := panDelta(w.mode, w.image, max(1, w.lastRect.W), max(1, w.imageHeightLocked()), w.zoom)
		cellW, cellH := pxPerCell(w.mode)
		// Image origin is delta/2*(1+pan), so dragging by n px moves pan by 2n/delta.
		if dX != 0 {
			w.panX = clampPan(w.panX+2*float64(dx)*float64(cellW)/float64(dX), abs(dX))
		}
		if dY != 0 {
			w.panY = clampPan(w.panY+2*float64(dy)*float64(cellH)/float64(dY), abs(dY))
		}
		w.mu.Unlock()
		return loom.Handled()
	}
	w.mu.Unlock()
	if e.Action == loom.MouseScrollUp {
		w.setZoomAt(1.25, e.X, e.Y)
		return loom.Handled()
	}
	if e.Action == loom.MouseScrollDown {
		w.setZoomAt(1/1.25, e.X, e.Y)
		return loom.Handled()
	}
	return loom.Ignored()
}

func (w *Widget) canPanLocked() bool {
	cols, rows := max(1, w.lastRect.W), max(1, w.imageHeightLocked())
	_, _, _, _, maxX, maxY := viewGeometry(w.mode, w.image, cols, rows, w.zoom)
	slackX, slackY := viewSlack(w.mode, w.image, cols, rows, w.zoom)
	return maxX > 0 || maxY > 0 || slackX > 0 || slackY > 0
}

func (w *Widget) imageHeightLocked() int {
	switch {
	case w.lastRect.H >= 3:
		return w.lastRect.H - 2
	case w.lastRect.H == 2:
		return w.lastRect.H - 1
	default:
		return w.lastRect.H
	}
}
