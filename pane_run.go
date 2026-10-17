// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"context"
	"fmt"
	"os/signal"
	"strings"
	"time"
	"ubunatic.com/loom/measure"
)

// Run renders root on every frame and dispatches events until the root signals
// quit or an unrecoverable error occurs.
func (p *Pane) Run(root Widget) error {
	return p.run(context.Background(), root, nil, nil, nil)
}

// Invalidate requests a repaint. It is safe to call from any goroutine and never blocks.
func (p *Pane) Invalidate() {
	p.invalidateOnce.Do(func() { p.invalidate = make(chan struct{}, 1) })
	select {
	case p.invalidate <- struct{}{}:
	default:
	}
}

// Post queues fn to run on the event-loop goroutine before the next frame,
// then requests a repaint. It is safe to call from any goroutine, never blocks
// on the loop and never drops callbacks; queued callbacks run in FIFO order and
// a burst is drawn once. Callbacks posted before Run start on its first loop
// turn. Post after Close is a no-op.
func (p *Pane) Post(fn func()) {
	if fn == nil {
		return
	}
	p.postMu.Lock()
	if p.postClosed {
		p.postMu.Unlock()
		return
	}
	p.posted = append(p.posted, fn)
	p.postMu.Unlock()
	p.Invalidate()
}

// runPosted runs the callbacks queued by Post and reports whether any ran.
// Callbacks posted while it runs wait for the next loop turn.
func (p *Pane) runPosted() bool {
	p.postMu.Lock()
	fns := p.posted
	p.posted = nil
	p.postMu.Unlock()
	for _, fn := range fns {
		p.postMu.Lock()
		closed := p.postClosed
		p.postMu.Unlock()
		if closed {
			break // a callback closed the pane; drop the rest of the batch
		}
		fn()
	}
	return len(fns) > 0
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
	p.invalidateOnce.Do(func() { p.invalidate = make(chan struct{}, 1) })
	bindInvalidationTree(root, p.Invalidate)
	bindClipboardTree(root, func(text string) {
		if p.tty != nil {
			_ = WriteOSC52(p.tty, text)
		}
	})
	p.tickerReset = make(chan struct{}, 1)
	bindTickerControlTree(root, p.ResetTicker)
	if requester, ok := UnwrapWidget(root).(PaneRequester); ok {
		request := requester.PaneRequest()
		if !p.mouse && request.Mouse > 0 {
			p.baseMouseMode = request.Mouse
			p.setMouseMode(request.Mouse)
		}
		if !p.Resizeable {
			p.Resizeable = request.Resizeable
		}
		p.applyRequestedMaxCols(request.MaxCols)
		if request.OwnsQuit {
			p.DisableDefaultQuit = true
			p.DisableGlobalF10Quit = true
		}
	}
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
		// the same path.
		p.inlineStart, p.fullActive = p.startRow, true
		p.applyWinch(&cols)
	}
	full, alt := p.wantScreen()
	p.changeScreen(full, alt)
	if alt {
		// Entering the alternate screen saves the inline bounds. Allocate its
		// full terminal area before the first canvas, just as later switches do.
		p.applyWinch(&cols)
	}
	p.setBracketedPaste(true)
	defer p.setBracketedPaste(false)
	if p.wantKittyKeyboard() {
		p.setKittyKeyboard(true)
		defer p.setKittyKeyboard(false)
	}
	// Probe in the selected screen before starting the input reader so the
	// DSR reply cannot be consumed as a key event. Refresh after both signal
	// registration and the probe, before allocating the first canvas.
	measure.DetectZWJMode(p.tty, p.tty, p.startRow)
	p.refreshStartupSize()
	cols = p.cols
	if p.MaxCols > 0 && cols > p.MaxCols {
		cols = p.MaxCols
	}
	canvas := NewCanvas(cols, p.rows)
	var backgroundClock backgroundTicker
	backgroundClock.reconcile(p.Background, p.ReduceMotion, p.BackgroundOnRedraw)
	defer backgroundClock.stop()
	if p.Metrics != nil && backgroundClock.interval > 0 {
		p.Metrics.AstraTargetFPS = 1 / backgroundClock.interval.Seconds()
	}
	var cursorFrames <-chan time.Time
	var cursorTicker *time.Ticker
	var cursorFrameInterval time.Duration
	defer func() {
		if cursorTicker != nil {
			cursorTicker.Stop()
		}
	}()

	// Read input in a goroutine and forward it on a channel so the main loop can
	// select between input and SIGWINCH. os.File.Read retries EINTR via Go's poll
	// layer (docs/TuiInput.md §2), so a signal alone cannot wake a blocking read —
	// hence the separate reader (see inputReader). Suspend stops and restarts it.
	reads := make(chan paneRead)
	p.reads = reads
	p.startReader(nil)
	defer p.stopReader()

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
		if p.resumeRepaint {
			// Back from Suspend: the terminal may have been resized and the
			// child changed the screen and the auto-wrap flag; take the
			// current size and paint every row from scratch.
			p.resumeRepaint = false
			autoWrapDisabled = false
			p.applyWinch(&cols)
			canvas = NewCanvas(cols, p.rows)
			clearRows = 0
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
			wasAlt := p.altActive
			p.changeScreen(full, alt)
			if p.altActive != wasAlt && p.ResizeConfig.FullLeakGuard && p.ResizeConfig.AutoWrap && p.tty != nil {
				// Terminals restore the auto-wrap flag with the cursor on
				// ?1049; without this a wide row wraps and leaks into the
				// scrollback.
				p.tty.WriteString("\x1b[?7l") //nolint:errcheck
				autoWrapDisabled = true
			}
			p.applyWinch(&cols)
			clearRows = 0
		}
		if canvas.Rows() != p.rows || canvas.Cols() != cols {
			canvas = NewCanvas(cols, p.rows)
		}
		canvas.Clear()
		if p.cursorProximity && p.mouseKnown {
			canvas.SetCursorPosition(p.mouseX, p.mouseY)
		} else {
			canvas.ClearCursorPosition()
		}
		drawBounds := canvas.Bounds()
		if p.DebugMode && drawBounds.W > 2 && drawBounds.H > 2 {
			DrawRuler(canvas, drawBounds)
			drawBounds = Rect{
				X: drawBounds.X + 1,
				Y: drawBounds.Y + 1,
				W: drawBounds.W - 2,
				H: drawBounds.H - 2,
			}
		}
		root.Draw(canvas, drawBounds)
		if p.help != nil {
			p.help.Draw(canvas, canvas.Bounds())
		}
		modal, _ := ActiveModalMouse(root)
		p.modalMouseGrab = modal != nil || (p.help != nil && p.help.Open)
		p.reconcileMouseMode()
		drawn := time.Now()
		canvas.ComposeBackground(p.Background, canvas.Bounds(), drawn)
		if p.cursorProximity {
			canvas.ApplyCursorBrighten()
		}
		if p.cursorStarTrail {
			canvas.ApplyCursorStarTrail(p.cursorTrail, time.Now())
		}
		if p.cursorPressPulse && p.cursorPulse != nil {
			canvas.ApplyCursorPulse(*p.cursorPulse, time.Now())
		}
		composed := time.Now()
		out := &countingWriter{w: p.tty}
		p.applyCanvasCursorShape(out, canvas)
		canvas.FlushWithConfig(out, p.startRow, clearRows+p.staleRows, p.ResizeConfig)
		clearRows, p.staleRows = 0, 0
		if p.Metrics != nil {
			done := time.Now()
			p.Metrics.record(done, astra, framePhases{
				total: done.Sub(started), draw: drawn.Sub(started),
				background: composed.Sub(drawn), flush: done.Sub(composed), bytes: out.n,
			})
		}
	}

	// pending carries over bytes left from a previous reads result that could
	// not yet be decoded as a complete key event — either a lone ESC that
	// might be the start of a longer escape sequence, or a CSI/SS3 prefix cut
	// short by a read boundary. pendingC fires escKeyTimeout after such bytes
	// arrive with nothing more following, so a standalone ESC keypress still
	// resolves promptly instead of waiting forever for bytes that will never
	// come.
	escKeyTimeout := SpeccedDefaults.Pane.EscKeyTimeout
	var pending []byte
	var pendingC <-chan time.Time

	var guardTimer *time.Timer
	var guardTimerC <-chan time.Time
	guardDuration := p.WidthGuardDuration
	if guardDuration <= 0 {
		guardDuration = SpeccedDefaults.Pane.GuardDuration
	}
	defer func() {
		if guardTimer != nil {
			guardTimer.Stop()
		}
		p.widthGuardActive = false
	}()

	dirty := true
	animationDirty := false
	tickLast := make(map[Widget]time.Time)
	var tickTimer *time.Timer
	var tickC <-chan time.Time
	var tickInterval time.Duration
	defer func() {
		if tickTimer != nil {
			tickTimer.Stop()
		}
	}()
	for {
		// A ready input or resize event may win select even after cancellation.
		// Check before doing another frame so a canceled Draw stays final.
		if err := ctx.Err(); err != nil {
			return err
		}
		if p.runPosted() {
			dirty = true
		}
		if backgroundClock.reconcile(p.Background, p.ReduceMotion, p.BackgroundOnRedraw) && p.Metrics != nil {
			p.Metrics.AstraTargetFPS = 0
			if backgroundClock.interval > 0 {
				p.Metrics.AstraTargetFPS = 1 / backgroundClock.interval.Seconds()
			}
		}
		now := time.Now()
		p.expireCursorEffects(now)
		wantedCursorFrame := cursorAnimationFrameInterval(p.cursorTrail, p.cursorPulse, now)
		if wantedCursorFrame != cursorFrameInterval {
			if cursorTicker != nil {
				cursorTicker.Stop()
				cursorTicker = nil
				cursorFrames = nil
			}
			cursorFrameInterval = wantedCursorFrame
			if wantedCursorFrame > 0 {
				cursorTicker = time.NewTicker(wantedCursorFrame)
				cursorFrames = cursorTicker.C
			}
		}
		interval := shortestTickInterval(root)
		if interval <= 0 {
			tickC = nil
			tickInterval = 0
			if tickTimer != nil {
				tickTimer.Stop()
			}
		} else {
			if tickTimer == nil {
				tickTimer = time.NewTimer(interval)
				tickInterval = interval
			} else if tickInterval != interval {
				if !tickTimer.Stop() {
					select {
					case <-tickTimer.C:
					default:
					}
				}
				tickTimer.Reset(interval)
				tickInterval = interval
			}
			tickC = tickTimer.C
		}
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

		if p.quitting.Load() {
			return nil
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
			if p.requestClose(CloseReasonInterrupt) {
				return nil
			}
			dirty = true
		case now := <-samples:
			if err := collect(now); err != nil {
				return err
			}
		case <-frames:
			dirty = true
		case now := <-tickC:
			tickTree(root, now, tickLast)
			tickInterval = 0
			dirty = true
		case <-p.invalidate:
			dirty = true
		case <-p.tickerReset:
			// Force the timer to adopt the current interval with a fresh countdown.
			tickInterval = -1
		case <-backgroundClock.frames:
			dirty = true
			animationDirty = true
		case now := <-cursorFrames:
			p.expireCursorEffects(now)
			dirty = true
		case <-guardTimerC:
			guardTimerC = nil
			p.widthGuardActive = false
			p.adaptiveN = 0
			// Settle against the current kernel size, not a stale WINCH
			// sample: an emulator may finish resizing before delivery.
			oldCols, oldRows, oldStart := cols, p.rows, p.startRow
			p.applyWinch(&cols)
			dirty = true
			clearRows = max(clearRows, oldStart+oldRows-p.startRow-p.rows)
			if cols != oldCols || p.rows != oldRows || p.startRow != oldStart {
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
			p.triggerCursorPulse(time.Now())
			if quit, handled := p.handleHelpKey(ke); handled {
				if quit {
					return nil
				}
				continue
			}
			if p.dispatchKey(root, ke) {
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
				pending = append(pending, raw...)
				raw = pending
			}
			pending = nil
			pendingC = nil

			// Try mouse first (SGR: \x1b[<…M/m). A single read may carry several
			// reports (1003 any-motion tracking floods them), so drain the whole
			// buffer report-by-report and dispatch each in order.
			if p.mouse && len(raw) >= 3 && raw[0] == 27 && raw[1] == '[' && raw[2] == '<' {
				var quit bool
				raw, quit = p.dispatchMouseRead(root, canvas, raw, redraw)
				if quit {
					return nil
				}
				if len(raw) == 0 {
					continue
				}
				// Process keys following the mouse reports in this same read.
			}

			// Drain every complete key event out of raw (mirroring the mouse
			// loop above) instead of decoding only the first one, so several
			// key sequences coalesced into one read are all dispatched, in
			// order. A trailing prefix that could still be the start of a
			// longer escape sequence is held as pending rather than decoded
			// wrong.
			quit := false
			for len(raw) > 0 {
				const pasteStart = "\x1b[200~"
				if strings.HasPrefix(string(raw), pasteStart) {
					paste, used, complete := DecodePaste(raw)
					if !complete {
						if len(raw) > maxPendingPasteSize {
							paste = PasteEvent{Text: string(raw[len(pasteStart):])}
							p.triggerCursorPulse(time.Now())
							if DispatchPasteEvent(root, paste).Quit {
								quit = true
							}
							pending = nil
							pendingC = nil
							break
						}
						pending = raw
						pendingC = nil
						dirty = false
						break
					}
					raw = raw[used:]
					p.triggerCursorPulse(time.Now())
					if DispatchPasteEvent(root, paste).Quit {
						quit = true
						break
					}
					continue
				}
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
				p.triggerCursorPulse(time.Now())
				if quit, handled := p.handleHelpKey(ke); handled {
					if quit {
						return nil
					}
					continue
				}
				if p.dispatchKey(root, ke) {
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
