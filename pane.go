// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package loom is an inline terminal UI library.
//
// A Pane reserves a fixed-height region below the shell cursor, renders
// a Widget tree into a Canvas on every frame, and dispatches keyboard and
// mouse events. The terminal is always restored on exit, panic, or signal.
//
// Typical usage:
//
//	items := []loom.Item{{Name: "git"}, {Name: "make"}}
//	choice := loom.NewChoice(items)
//	pane, err := loom.New(12)
//	if err != nil { log.Fatal(err) }
//	defer pane.Close()
//	pane.Run(choice)
//	if item, ok := choice.Selected(); ok {
//	    fmt.Println(item.Name)
//	}
//
// See docs/TuiInput.md for the constraints that govern /dev/tty usage,
// EINTR handling, and ZSH widget compatibility.
package loom

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

var paneOwnership struct {
	sync.Mutex
	active bool
}

// DefaultMaxCols is the default maximum canvas width loaded from SpeccedDefaults.
var DefaultMaxCols = SpeccedDefaults.Pane.MaxCols

// Pane manages an inline terminal region and drives the widget event loop.
type Pane struct {
	tty      *os.File
	fd       int
	oldState *term.State
	rows     int // reserved height (clamped to the current terminal)
	wantRows int // desired height; rows is restored toward this when space allows
	startRow int // 1-based terminal row of the pane top
	cols     int // terminal width at Open time
	MaxCols  int // canvas width cap; 0 = use terminal width
	restored bool
	ownsTTY  bool

	// mouse tracking is enabled with EnableMouse.
	mouse      bool
	mouseMode  int
	Resizeable bool

	// DisableDefaultQuit suppresses the fallback exit behavior for unhandled
	// Esc, Ctrl-C, Ctrl-Q, and q keys when the active widget returns false from HandleKey.
	DisableDefaultQuit bool

	// Background is composited after the root widget on every frame.
	Background Background
	// ReduceMotion disables animated background ticks while retaining its static
	// rendering on normal foreground redraws.
	ReduceMotion bool
	// Metrics, when non-nil, receives rolling redraw measurements.
	Metrics *RenderMetrics
	// ResizeConfig controls runtime resize rendering behaviors.
	ResizeConfig ResizeConfig
	// widthGuardActive tracks whether the pane is inside a SIGWINCH burst window.
	widthGuardActive bool
	// winchMeter measures SIGWINCH arrival rate for the adaptive guard.
	winchMeter  WinchMeter
	altActive   bool
	fullActive  bool // primary-screen layout is full screen (from row 1)
	inlineStart int  // pane top before the full-screen layout took over
	savedStart  int
	savedRows   int
	// adaptiveN is the adaptive guard width latched at the last SIGWINCH; 0
	// means not measurable (fewer than two events), so the manual n applies.
	adaptiveN int
	// WidthGuardDuration configures the burst timeout before restoring full width.
	// When 0, defaults to 1 second.
	WidthGuardDuration time.Duration

	// winch carries SIGWINCH notifications so the Run loop reflows on a
	// terminal window resize. Buffered (cap 1) to coalesce resize bursts.
	winch chan os.Signal

	// readerDone is closed by Run's input-reader goroutine when it exits. close()
	// waits on it so a lingering blocked Read cannot steal input meant for
	// whatever reads the terminal after the pane (e.g. the calling shell).
	readerDone chan struct{}
	interrupts chan os.Signal
	help       *Popup
}

// RenderMetrics reports recent completed redraw performance.
type RenderMetrics struct {
	LoomFPS        float64
	AstraFPS       float64
	AstraTargetFPS float64
	RedrawTime     time.Duration
	windowStart    time.Time
	loomFrames     int
	astraFrames    int
}

func (m *RenderMetrics) record(now time.Time, astra bool, elapsed time.Duration) {
	if m.windowStart.IsZero() {
		m.windowStart = now
	}
	m.loomFrames++
	if astra {
		m.astraFrames++
	}
	if elapsed > 0 {
		m.RedrawTime = elapsed
	}
	if d := now.Sub(m.windowStart); d >= time.Second {
		m.LoomFPS = float64(m.loomFrames) / d.Seconds()
		m.AstraFPS = float64(m.astraFrames) / d.Seconds()
		m.loomFrames, m.astraFrames, m.windowStart = 0, 0, now
	}
}

// AnimatedBackground is a background that needs periodic redraws.
type AnimatedBackground interface {
	Background
	DrawBackgroundAt(*Canvas, Rect, time.Time)
}

// BackgroundCadence optionally lets an animated background choose its redraw
// interval. Implementations that do not provide it use the spec default.
type BackgroundCadence interface {
	BackgroundInterval() time.Duration
}

// New opens /dev/tty, enters raw mode, and reserves height rows below the
// current cursor position. Call Close (or defer it) to restore the terminal.
//
// Always uses /dev/tty — never os.Stdin — so the pane works inside shells
// and command substitution (e.g. `foo=$(bar)`), where stdin may be a pipe.
// See docs/TuiInput.md §1.
func New(height int) (*Pane, error) {
	if height < 1 {
		height = 1
	}
	if !claimPaneOwnership() {
		return nil, fmt.Errorf("loom: another pane already owns the terminal")
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		releasePaneOwnership()
		return nil, fmt.Errorf("loom: open /dev/tty: %w", err)
	}
	fd := int(tty.Fd())

	old, err := term.MakeRaw(fd)
	if err != nil {
		_ = tty.Close()
		releasePaneOwnership()
		return nil, fmt.Errorf("loom: raw mode: %w", err)
	}

	cols, termRows := termSize(fd)
	if cols < 1 {
		cols = 80
	}
	wantRows := height
	if height > termRows-1 {
		height = termRows - 1
	}

	if Debug {
		fmt.Fprintf(os.Stderr, "loom: [debug] terminal: %dx%d\r\n", cols, termRows)
	}

	// Query cursor position using DSR
	cy, cx, err := queryCursor(tty)
	if err != nil || cy < 1 || cy > termRows {
		cy = termRows
		cx = 1
	}

	actualCy := cy
	if Debug {
		// We print 2 more debug lines below, which will advance the cursor by 2 lines.
		// Adjust actualCy accordingly (capped to termRows) for the TUI reservation.
		actualCy = cy + 2
		if actualCy > termRows {
			actualCy = termRows
		}
	}

	startRow, toScroll := reserveRegion(actualCy, termRows, height)
	if Debug {
		fmt.Fprintf(os.Stderr, "loom: [debug] cursor: (%d,%d)\r\n", cx, cy)
		fmt.Fprintf(os.Stderr, "loom: [debug] placement: start_row: %d, scroll: %d\r\n", startRow, toScroll)
	}

	if toScroll > 0 {
		tty.WriteString(fmt.Sprintf("\x1b[%dS", toScroll)) // scroll up toScroll lines
	}

	// Move cursor to startRow so drawing begins at the correct relative row
	tty.WriteString(fmt.Sprintf("\x1b[%d;1H", startRow))

	p := &Pane{
		tty:          tty,
		fd:           fd,
		oldState:     old,
		rows:         height,
		wantRows:     wantRows,
		startRow:     startRow,
		cols:         cols,
		MaxCols:      DefaultMaxCols,
		ResizeConfig: DefaultResizeConfig(),
		ownsTTY:      true,
	}
	p.installSignalHandler()
	return p, nil
}

func claimPaneOwnership() bool {
	paneOwnership.Lock()
	defer paneOwnership.Unlock()
	if paneOwnership.active {
		return false
	}
	paneOwnership.active = true
	return true
}

func releasePaneOwnership() {
	paneOwnership.Lock()
	paneOwnership.active = false
	paneOwnership.Unlock()
}

func queryCursor(tty *os.File) (row, col int, err error) {
	if _, err = tty.WriteString("\x1b[6n"); err != nil {
		return 0, 0, err
	}
	// Poll synchronously: a timed-out reader goroutine would steal later keys.
	deadline := time.Now().Add(100 * time.Millisecond)
	poll := []unix.PollFd{{Fd: int32(tty.Fd()), Events: unix.POLLIN}}
	var buf [32]byte
	for n := 0; n < len(buf); {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return 0, 0, fmt.Errorf("cursor query timed out")
		}
		ready, err := unix.Poll(poll, max(1, int(remaining.Milliseconds())))
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return 0, 0, err
		}
		if ready == 0 {
			continue
		}
		m, err := tty.Read(buf[n : n+1])
		if err != nil {
			return 0, 0, err
		}
		if m == 0 {
			return 0, 0, fmt.Errorf("cursor query: input closed")
		}
		n += m
		if buf[n-1] == 'R' {
			_, err := fmt.Sscanf(string(buf[:n]), "\x1b[%d;%dR", &row, &col)
			return row, col, err
		}
	}
	return 0, 0, fmt.Errorf("cursor query: response too long")
}

func reserveRegion(cy, rows, want int) (startRow, toScroll int) {
	startRow = cy
	if rows-cy < want {
		if d := want - (rows - cy) - 1; d > 0 {
			toScroll = d
			startRow -= d
		}
	}
	if startRow < 1 {
		startRow = 1
	}
	return startRow, toScroll
}

// EnableMouse turns on SGR mouse tracking (button press, release, hover, scroll).
// Must be called before Run. Mouse events are delivered to the root Widget's
// HandleMouse method.
func (p *Pane) EnableMouse() {
	p.setMouseMode(1003)
}

// EnableMouseClicks tracks clicks and wheel events without any-motion reports.
// Call it before Run; Close restores the terminal's normal mouse behavior.
func (p *Pane) EnableMouseClicks() {
	p.setMouseMode(1000)
}

func (p *Pane) setMouseMode(mode int) {
	if p.mouseMode != 0 {
		p.tty.WriteString(fmt.Sprintf("\x1b[?%dl", p.mouseMode)) //nolint:errcheck
	}
	p.mouse, p.mouseMode = true, mode
	p.tty.WriteString(fmt.Sprintf("\x1b[?%dh\x1b[?1006h", mode)) //nolint:errcheck
}

func (p *Pane) disableMouse() {
	if p.mouseMode == 0 {
		return
	}
	p.tty.WriteString(fmt.Sprintf("\x1b[?%dl\x1b[?1006l", p.mouseMode)) //nolint:errcheck
	p.mouse, p.mouseMode = false, 0
}

// Resize adjusts the pane height dynamically.
// If the height increases and overflows the terminal screen, it scrolls the terminal up
// to reserve the required space. If it decreases, it clears the abandoned lines.
func (p *Pane) Resize(newHeight int) {
	if newHeight < 1 {
		newHeight = 1
	}
	cols, termRows := termSize(p.fd)
	if cols < 1 {
		cols = 80
	}
	p.wantRows = newHeight
	if p.fullActive || p.altActive {
		return // the layout owns the rows; wantRows applies when inline again
	}
	if newHeight > termRows-1 {
		newHeight = termRows - 1
	}

	if newHeight == p.rows {
		return
	}

	if newHeight > p.rows {
		// Growing: check for overflow at the bottom of the screen
		overflow := (p.startRow + newHeight - 1) - termRows
		if overflow > 0 {
			// Scroll up to reserve space
			p.tty.WriteString(fmt.Sprintf("\x1b[%dS", overflow)) //nolint:errcheck
			p.startRow -= overflow
			if p.startRow < 1 {
				p.startRow = 1
			}
		}
	} else {
		// The next frame clears abandoned rows together with its UI output.
	}

	p.rows = newHeight
}

// winchBounds recomputes pane placement after a terminal window resize. Given
// the current top row and height and the new terminal row count, it clamps the
// height to fit (leaving the shell prompt line) and lifts the top row if the
// pane would overflow the new bottom. It is pure so the math can be unit-tested.
func winchBounds(startRow, rows, termRows int) (newStartRow, newRows int) {
	newRows = rows
	if newRows > termRows-1 {
		newRows = termRows - 1
	}
	if newRows < 1 {
		newRows = 1
	}
	newStartRow = startRow
	if overflow := (newStartRow + newRows - 1) - termRows; overflow > 0 {
		newStartRow -= overflow
	}
	if newStartRow < 1 {
		newStartRow = 1
	}
	return newStartRow, newRows
}

// applyWinch re-queries the terminal size after a SIGWINCH and updates the
// pane bounds and canvas width so the next frame reflows. Rendering and stale
// row clearing remain a single operation in the caller.
func (p *Pane) applyWinch(cols *int) {
	newCols, termRows := termSize(p.fd)
	if newCols < 1 {
		newCols = 80
	}

	// Clamp from the desired height, not the currently-clamped one, so the pane
	// grows back toward wantRows when the window is enlarged again (a tiny window
	// must not permanently pin the pane at one row).
	newStartRow, newRows := winchBounds(p.startRow, p.wantRows, termRows)
	if p.fullActive {
		newStartRow = 1
		newRows = max(1, termRows-1)
	}
	if p.altActive {
		newStartRow, newRows = 1, max(1, termRows)
	}

	if p.ResizeConfig.OutOfBandClear && p.tty != nil {
		top := p.startRow
		if newStartRow < top {
			top = newStartRow
		}
		var b strings.Builder
		for row := top; row <= termRows; row++ {
			fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", row)
		}
		p.tty.WriteString(b.String()) //nolint:errcheck
	}

	p.startRow, p.rows = newStartRow, newRows
	p.cols = newCols

	*cols = p.guardedCols(newCols)
}

// ScreenMode is where the pane draws: a few rows below the prompt, the whole
// terminal, or the whole alternate screen.
type ScreenMode int

const (
	ScreenInline ScreenMode = iota // reserved rows below the prompt
	ScreenFull                     // full terminal from row 1 on the primary screen
	ScreenAlt                      // full terminal on the alternate screen
)

// SetScreenMode selects the layout by setting the resize modes it consists of.
// The next frame performs the switch. Auto full screen, when enabled, can still
// promote an inline pane; disable it in ResizeConfig to force ScreenInline.
func (p *Pane) SetScreenMode(m ScreenMode) {
	p.ResizeConfig.FullScreenBuffer = m == ScreenFull
	p.ResizeConfig.AltScreen = m == ScreenAlt
}

// Screen reports the layout currently in force, including an automatic
// promotion to full screen.
func (p *Pane) Screen() ScreenMode {
	switch {
	case p.altActive:
		return ScreenAlt
	case p.fullActive:
		return ScreenFull
	}
	return ScreenInline
}

// wantScreen derives the layout the config asks for right now. full applies to
// the primary screen only; it is ignored while alt is wanted. Auto full screen
// compares the wanted height (not the clamped one) with the terminal height.
func (p *Pane) wantScreen() (full, alt bool) {
	cfg := p.ResizeConfig
	auto := false
	if cfg.AutoFullscreen {
		_, termRows := termSize(p.fd)
		auto = cfg.QuasiFullscreen(p.wantRows, termRows)
	}
	return cfg.FullScreenBuffer || auto, cfg.AltScreen || (auto && cfg.FullAlt)
}

// changeScreen moves between layouts. Leaving alt first restores the primary
// bounds, the primary full/inline change happens on the primary screen, and
// entering alt comes last so it saves the final primary bounds.
func (p *Pane) changeScreen(full, alt bool) {
	if p.altActive && !alt {
		p.switchAltScreen(false)
	}
	if !alt && full != p.fullActive {
		p.switchFull(full)
	}
	if alt && !p.altActive {
		p.switchAltScreen(true)
	}
}

// switchFull changes the primary-screen layout between inline and full screen.
// The old frame is erased row by row; the next frame paints the new one.
func (p *Pane) switchFull(on bool) {
	if p.tty != nil {
		var b strings.Builder
		for i := 0; i < p.rows; i++ {
			fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", p.startRow+i)
		}
		p.tty.WriteString(b.String()) //nolint:errcheck
	}
	if on {
		p.inlineStart = p.startRow
	} else {
		p.startRow = p.inlineStart
	}
	p.fullActive = on
}

// switchAltScreen enters or leaves the alternate screen buffer. The primary
// screen's pane bounds are kept and restored on leave.
func (p *Pane) switchAltScreen(on bool) {
	if p.tty == nil || on == p.altActive {
		return
	}
	if on {
		p.savedStart, p.savedRows = p.startRow, p.rows
		// Erase the frame already drawn on the primary screen, or it stays in
		// the scrollback after leaving; park the cursor where the pane started
		// so leaving restores it there.
		var b strings.Builder
		for i := 0; i < p.rows; i++ {
			fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", p.startRow+i)
		}
		fmt.Fprintf(&b, "\x1b[%d;1H\x1b[?1049h\x1b[2J", p.startRow)
		p.tty.WriteString(b.String()) //nolint:errcheck
	} else {
		p.tty.WriteString("\x1b[?1049l") //nolint:errcheck
		p.startRow, p.rows = p.savedStart, p.savedRows
	}
	p.altActive = on
}

// Run renders root on every frame and dispatches events until the root signals
// quit or an unrecoverable error occurs.
func (p *Pane) Run(root Widget) error {
	return p.run(context.Background(), root, nil, nil, nil)
}

// RunWatch runs collection and redraw on independent timers. Collection and
// widget callbacks share the event-loop goroutine; Draw must not collect data.
// The caller owns Close, as with Run. Collection callbacks must not block.
func (p *Pane) RunWatch(ctx context.Context, root Widget, cadence Cadence, collect func(time.Time) error) error {
	if err := cadence.Validate(); err != nil {
		return err
	}
	if collect == nil {
		return fmt.Errorf("loom: watch collection callback required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	samples := time.NewTicker(cadence.Collect)
	defer samples.Stop()
	frames := time.NewTicker(cadence.Redraw)
	defer frames.Stop()
	if err := collect(time.Now()); err != nil {
		return err
	}
	return p.run(ctx, root, samples.C, frames.C, collect)
}

func (p *Pane) run(ctx context.Context, root Widget, samples, frames <-chan time.Time, collect func(time.Time) error) error {
	previousHelpRequest := paneHelpRequest
	paneHelpRequest = func(cmds []Cmd) {
		p.help = NewPopup("Help", newHelpWidget(cmds))
	}
	defer func() { paneHelpRequest = previousHelpRequest }()
	defer func() {
		if r := recover(); r != nil {
			p.close()
			panic(r)
		}
	}()

	if p.winch == nil {
		p.installSignalHandler()
		defer func() {
			if p.winch != nil {
				signal.Stop(p.winch)
			}
			if p.interrupts != nil {
				signal.Stop(p.interrupts)
			}
		}()
	}

	cols := p.cols
	if full, alt := p.wantScreen(); full && !alt {
		// The full-screen layout owns the usable terminal buffer from row 1.
		// Recompute it before allocating the first canvas; later Winch events use
		// the same path. (An alternate-screen start is entered by the first redraw.)
		p.inlineStart, p.fullActive = p.startRow, true
		p.applyWinch(&cols)
	}
	if p.MaxCols > 0 && cols > p.MaxCols {
		cols = p.MaxCols
	}
	canvas := NewCanvas(cols, p.rows)
	var backgroundFrames <-chan time.Time
	var backgroundTicker *time.Ticker
	if _, ok := p.Background.(AnimatedBackground); ok && !p.ReduceMotion {
		interval := SpeccedBackground.RedrawInterval
		if cadence, ok := p.Background.(BackgroundCadence); ok {
			interval = cadence.BackgroundInterval()
		}
		if interval > 0 {
			backgroundTicker = time.NewTicker(interval)
			backgroundFrames = backgroundTicker.C
			defer backgroundTicker.Stop()
		}
		if p.Metrics != nil && interval > 0 {
			p.Metrics.AstraTargetFPS = 1 / interval.Seconds()
		}
	}

	// Read input in a goroutine and forward it on a channel so the main loop can
	// select between input and SIGWINCH. os.File.Read retries EINTR via Go's poll
	// layer (docs/TuiInput.md §2), so a signal alone cannot wake a blocking read —
	// hence the separate reader. The goroutine exits when the tty is closed (on
	// Close) or when done is signalled.
	type readResult struct {
		data []byte
		err  error
	}
	reads := make(chan readResult)
	done := make(chan struct{})
	defer close(done)
	p.readerDone = make(chan struct{})
	go func() {
		defer close(p.readerDone) // let close() join us so no read is left racing
		// Poll with a short timeout instead of blocking in Read: os.File.Fd (called
		// in New) detaches the tty from the runtime poller and puts it in blocking
		// mode, so Close cannot interrupt a bare Read — the goroutine would leak and
		// race the next terminal reader for input. Polling lets us observe done and
		// exit on teardown. EINTR (e.g. from a delivered SIGWINCH) just re-polls.
		pfd := []unix.PollFd{{Fd: int32(p.fd), Events: unix.POLLIN}}
		for {
			select {
			case <-done:
				return
			default:
			}
			n, perr := unix.Poll(pfd, 50)
			if perr != nil {
				if perr == unix.EINTR {
					continue
				}
				select {
				case reads <- readResult{err: perr}:
				case <-done:
				}
				return
			}
			if n == 0 {
				continue // timeout — loop back to re-check done
			}
			buf := make([]byte, 64)
			m, rerr := p.tty.Read(buf)
			if m > 0 {
				select {
				case reads <- readResult{data: buf[:m]}:
				case <-done:
					return
				}
			}
			if rerr != nil || m == 0 {
				select {
				case reads <- readResult{err: rerr}:
				case <-done:
				}
				return
			}
		}
	}()

	// redraw paints one frame. It is a closure so the mouse loop can repaint
	// after each event in a burst, letting the highlight follow the pointer
	// rather than jumping once after the whole burst is drained.
	autoWrapDisabled := false
	clearRows := 0
	lastMaxCols := p.MaxCols
	redraw := func(astra bool) {
		started := time.Now()
		if p.MaxCols != lastMaxCols {
			// The app changed the width cap: take it up on this frame.
			lastMaxCols = p.MaxCols
			cols = p.cols
			if p.MaxCols > 0 && cols > p.MaxCols {
				cols = p.MaxCols
			}
		}
		if p.ResizeConfig.AutoWrap && !autoWrapDisabled && p.tty != nil {
			p.tty.WriteString("\x1b[?7l") //nolint:errcheck
			autoWrapDisabled = true
		} else if !p.ResizeConfig.AutoWrap && autoWrapDisabled && p.tty != nil {
			p.tty.WriteString("\x1b[?7h") //nolint:errcheck
			autoWrapDisabled = false
		}
		if canvas.Rows() != p.rows {
			canvas = NewCanvas(cols, p.rows)
		}
		if full, alt := p.wantScreen(); alt != p.altActive || (!alt && full != p.fullActive) {
			p.changeScreen(full, alt)
			p.applyWinch(&cols)
			clearRows = 0
		}
		if canvas.Rows() != p.rows || canvas.Cols() != cols {
			canvas = NewCanvas(cols, p.rows)
		}
		canvas.Clear()
		root.Draw(canvas, canvas.Bounds())
		if p.help != nil {
			p.help.Draw(canvas, canvas.Bounds())
		}
		canvas.ComposeBackground(p.Background, canvas.Bounds(), time.Now())
		canvas.FlushWithConfig(p.tty, p.startRow, clearRows, p.ResizeConfig)
		clearRows = 0
		if p.Metrics != nil {
			p.Metrics.record(time.Now(), astra, time.Since(started))
		}
	}

	// pending carries over bytes left from a previous reads result that could
	// not yet be decoded as a complete key event — either a lone ESC that
	// might be the start of a longer escape sequence, or a CSI/SS3 prefix cut
	// short by a read boundary. pendingC fires escKeyTimeout after such bytes
	// arrive with nothing more following, so a standalone ESC keypress still
	// resolves promptly instead of waiting forever for bytes that will never
	// come.
	const escKeyTimeout = 50 * time.Millisecond
	var pending []byte
	var pendingC <-chan time.Time

	var guardTimer *time.Timer
	var guardTimerC <-chan time.Time
	guardDuration := p.WidthGuardDuration
	if guardDuration <= 0 {
		guardDuration = time.Second
	}
	defer func() {
		if guardTimer != nil {
			guardTimer.Stop()
		}
		p.widthGuardActive = false
	}()

	dirty := true
	animationDirty := false
	for {
		if p.widthGuardActive {
			if !p.ResizeConfig.WidthGuard {
				p.widthGuardActive = false
				if guardTimer != nil {
					guardTimer.Stop()
					guardTimerC = nil
				}
			}
			if want := p.guardedCols(p.cols); want != cols {
				cols = want
				canvas = NewCanvas(cols, p.rows)
				dirty = true
			}
		}

		if p.Resizeable {
			targetH := p.wantRows
			if ch, ok := root.(WidthHeighter); ok {
				targetH = ch.HeightForWidth(cols)
			} else if ch, ok := root.(ContentHeighter); ok {
				targetH = ch.ContentHeight()
			}
			if targetH != p.wantRows {
				p.Resize(targetH)
				dirty = true
			}
		}

		if dirty {
			redraw(animationDirty)
		}
		dirty = false
		animationDirty = false

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.interrupts:
			return nil
		case now := <-samples:
			if err := collect(now); err != nil {
				return err
			}
		case <-frames:
			dirty = true
		case <-backgroundFrames:
			dirty = true
			animationDirty = true
		case <-guardTimerC:
			guardTimerC = nil
			p.widthGuardActive = false
			p.adaptiveN = 0
			if full := p.guardedCols(p.cols); full != cols {
				cols = full
				canvas = NewCanvas(cols, p.rows)
				dirty = true
			}
		case <-p.winch:
			p.winchMeter.Record(time.Now())
			if !p.ResizeConfig.ResizeHandling {
				// Diagnostic mode: consume SIGWINCH but leave the current
				// canvas and pane geometry untouched.
				continue
			}
			// Terminal window resized: reflow width/bounds and repaint.
			if p.ResizeConfig.Coalesce {
				for {
					select {
					case <-p.winch:
					default:
						goto resizeCoalesced
					}
				}
			}
		resizeCoalesced:
			// Latch the adaptive width once per resize event so the guard does
			// not jitter as the measured rate decays between events.
			p.adaptiveN = AdaptiveGuardN(p.winchMeter.Rate(time.Now()), p.winchMeter.Events(time.Now()), 0)
			if p.ResizeConfig.WidthGuard {
				p.widthGuardActive = true
				if guardTimer == nil {
					guardTimer = time.NewTimer(guardDuration)
				} else {
					if !guardTimer.Stop() {
						select {
						case <-guardTimer.C:
						default:
						}
					}
					guardTimer.Reset(guardDuration)
				}
				guardTimerC = guardTimer.C
			} else {
				p.widthGuardActive = false
				if guardTimer != nil {
					guardTimer.Stop()
					guardTimerC = nil
				}
			}

			oldStartRow, oldRows := p.startRow, p.rows
			p.applyWinch(&cols)
			oldBottom := oldStartRow + oldRows
			newBottom := p.startRow + p.rows
			if oldBottom > newBottom {
				// A vertical resize can move the pane as well as change its
				// height. Clear the abandoned tail from the new frame's bottom.
				clearRows = oldBottom - newBottom
			}
			canvas = NewCanvas(cols, p.rows)
			dirty = true
			continue
		case <-pendingC:
			// No further bytes arrived in time to complete the pending prefix;
			// resolve it now (a lone ESC decodes as a plain "esc" keypress).
			ke := DecodeKey(pending)
			pending = nil
			pendingC = nil
			dirty = true
			if p.handleHelpKey(ke) || root.HandleKey(ke) || p.handleKeyFallback(ke) {
				return nil
			}
		case rr := <-reads:
			if rr.err != nil {
				return fmt.Errorf("loom: input: %w", rr.err)
			}
			if len(rr.data) == 0 {
				return nil
			}
			dirty = true
			raw := rr.data
			if len(pending) > 0 {
				raw = append(pending, raw...)
			}
			pending = nil
			pendingC = nil

			// Try mouse first (SGR: \x1b[<…M/m). A single read may carry several
			// reports (1003 any-motion tracking floods them), so drain the whole
			// buffer report-by-report and dispatch each in order.
			if p.mouse && len(raw) >= 3 && raw[0] == 27 && raw[1] == '[' && raw[2] == '<' {
				quit := false
				for len(raw) > 0 {
					me, used, ok := scanMouse(raw)
					if !ok {
						break
					}
					// Translate absolute terminal coords to pane-relative 1-based:
					// the pane top (canvas row 0) is terminal row startRow, but widgets
					// hit-test as if their first row were Y=1. Without this, mouse only
					// lined up when the pane happened to sit at row 1.
					me.Y -= p.startRow - 1
					if p.handleHelpMouse(me) || root.HandleMouse(me) {
						quit = true
						break
					}
					raw = raw[used:]
					redraw(false) // repaint after each event so the pointer is followed live
				}
				if quit {
					return nil
				}
				continue
			}

			// Drain every complete key event out of raw (mirroring the mouse
			// loop above) instead of decoding only the first one, so several
			// key sequences coalesced into one read are all dispatched, in
			// order. A trailing prefix that could still be the start of a
			// longer escape sequence is held as pending rather than decoded
			// wrong.
			quit := false
			for len(raw) > 0 {
				ke, used, ok := scanKey(raw)
				if !ok {
					pending = append([]byte(nil), raw...)
					pendingC = time.After(escKeyTimeout)
					raw = nil
					break
				}
				raw = raw[used:]
				if used == 0 {
					// scanKey must always make progress; guard against a stall.
					break
				}
				if p.handleHelpKey(ke) || root.HandleKey(ke) || p.handleKeyFallback(ke) {
					quit = true
					break
				}
			}
			if quit {
				return nil
			}
		}
	}
}

func (p *Pane) handleHelpKey(e KeyEvent) bool {
	if p.help == nil {
		return false
	}
	if e.Key == "esc" {
		p.help = nil
		return true
	}
	close := p.help.HandleKey(e)
	if close || !p.help.Open {
		p.help = nil
	}
	return true
}

func (p *Pane) handleHelpMouse(e MouseEvent) bool {
	if p.help == nil {
		return false
	}
	p.help.HandleMouse(e)
	return true
}

var defaultQuitKeyMap = func() map[string]bool {
	m := make(map[string]bool, len(SpeccedDefaults.FallbackQuitKeys))
	for _, k := range SpeccedDefaults.FallbackQuitKeys {
		m[k] = true
	}
	return m
}()

// handleKeyFallback reports whether an unhandled key should trigger a default
// safeguard exit based on embedded spec/defaults.yaml.
func (p *Pane) handleKeyFallback(ke KeyEvent) bool {
	if p.DisableDefaultQuit {
		return false
	}
	key := ke.Key
	if key == "" {
		key = ke.Text
	}
	return defaultQuitKeyMap[key]
}

// Close tears down the pane: clears the reserved region, restores terminal
// state, and closes /dev/tty. Safe to call multiple times.
func (p *Pane) Close() { p.close() }

func (p *Pane) close() {
	if p.restored {
		return
	}
	p.restored = true
	if p.ownsTTY {
		releasePaneOwnership()
		p.ownsTTY = false
	}

	if p.winch != nil {
		signal.Stop(p.winch)
	}
	if p.interrupts != nil {
		signal.Stop(p.interrupts)
	}

	p.disableMouse()

	var b strings.Builder
	if p.altActive {
		// Leaving the alternate screen restores the shell's screen and cursor.
		b.WriteString("\x1b[?1049l")
		p.altActive = false
	} else {
		// Clear reserved region.
		for i := 0; i < p.rows; i++ {
			b.WriteString(fmt.Sprintf("\x1b[%d;1H\x1b[2K", p.startRow+i))
		}
		b.WriteString(fmt.Sprintf("\x1b[%d;1H", p.startRow))
	}
	b.WriteString("\x1b[?25h")    // restore cursor visibility (a prompt-less frame may have hidden it)
	b.WriteString("\x1b[?7h")     // restore auto-wrap
	p.tty.WriteString(b.String()) //nolint:errcheck

	// Discard any unread input before leaving raw mode: keys pressed during the
	// session (the closing Enter, leftover mouse/DSR reports) otherwise sit in the
	// tty queue and leak into the shell or the next cooked reader as a stray line.
	p.drainInput()

	_ = term.Restore(p.fd, p.oldState)
	_ = p.tty.Close()

	// Wait for the reader goroutine to notice done (it polls on a 50ms cycle) and
	// exit, so it cannot consume a keystroke meant for the next terminal reader.
	// Bounded so teardown can never hang.
	if p.readerDone != nil {
		select {
		case <-p.readerDone:
		case <-time.After(300 * time.Millisecond):
		}
	}
}

// drainInput flushes the terminal input queue (TCIFLUSH) so raw-mode leftovers
// don't bleed into whatever reads the terminal next.
func (p *Pane) drainInput() {
	const tcflsh = 0x540B // Linux TCFLSH; argument 0 selects TCIFLUSH (input queue)
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, uintptr(p.fd), tcflsh, 0)
}

func (p *Pane) installSignalHandler() {
	p.interrupts = make(chan os.Signal, 1)
	signal.Notify(p.interrupts, syscall.SIGINT, syscall.SIGTERM)

	// SIGWINCH goes to its own channel; the Run loop selects on it to reflow.
	// Cap 1 coalesces a burst of resizes (e.g. an interactive drag) into one.
	p.winch = make(chan os.Signal, 1)
	signal.Notify(p.winch, syscall.SIGWINCH)
}

// TerminalSize returns the dimensions of the controlling terminal.
// It returns an error when /dev/tty is unavailable or its size cannot be read.
func TerminalSize() (cols, rows int, err error) {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return 0, 0, fmt.Errorf("loom: open /dev/tty: %w", err)
	}
	defer tty.Close()
	cols, rows, err = term.GetSize(int(tty.Fd()))
	if err != nil {
		return 0, 0, fmt.Errorf("loom: terminal size: %w", err)
	}
	if cols < 1 || rows < 1 {
		return 0, 0, fmt.Errorf("loom: invalid terminal size %dx%d", cols, rows)
	}
	return cols, rows, nil
}

// termSize returns the terminal dimensions via x/term (TIOCGWINSZ under the
// hood on Unix), falling back to a conventional 80x24 on error -- this used
// to hand-roll its own syscall.Syscall(SYS_IOCTL, TIOCGWINSZ, ...) call
// even though term.GetSize (already imported in this file for MakeRaw/
// Restore) does the same thing more portably.
func termSize(fd int) (cols, rows int) {
	cols, rows, err := term.GetSize(fd)
	if err != nil || cols < 1 || rows < 1 {
		return 80, 24
	}
	return cols, rows
}
