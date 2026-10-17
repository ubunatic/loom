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
	"golang.org/x/term"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var paneOwnership struct {
	sync.Mutex
	active bool
}

// DefaultMaxCols is the default maximum canvas width loaded from SpeccedDefaults.
var DefaultMaxCols = SpeccedDefaults.Pane.MaxCols

// paneBeforeSignalHandler is a test hook for a resize after New's size read.
// It is nil in production and only set in isolated test subprocesses.
var paneBeforeSignalHandler func(*Pane)

// Pane manages an inline terminal region and drives the widget event loop.
type Pane struct {
	tty             *os.File
	fd              int
	oldState        *term.State
	rows            int // reserved height (clamped to the current terminal)
	wantRows        int // desired height; rows is restored toward this when space allows
	startRow        int // 1-based terminal row of the pane top
	cursorRowOffset int // original cursor row relative to startRow, after reservation scroll
	cursorCol       int // original 1-based terminal column
	cols            int // terminal width at Open time
	MaxCols         int // canvas width cap; 0 = use terminal width
	maxColsExplicit bool
	cursorShape     cursorShapeState
	restored        bool
	ownsTTY         bool

	// DebugMode overlays a 1-cell ruler around the pane and insets widget bounds by 1 cell.
	DebugMode bool

	// mouse tracking is enabled with EnableMouse.
	mouse            bool
	mouseMode        int
	baseMouseMode    int
	modalMouseGrab   bool
	modalMouse       ModalMouseCapture
	cursorProximity  bool
	cursorStarTrail  bool
	cursorPressPulse bool
	mouseX, mouseY   int
	mouseKnown       bool
	cursorTrail      []CursorTrailPoint
	cursorPulse      *CursorPulse
	Resizeable       bool

	// InlineOnly opts out of the automatic switch to the alternate screen: the
	// pane stays inline even when it is nearly as big as the terminal. Loom's
	// normal mode is to switch when the dimensions get resize-error prone.
	InlineOnly bool

	// DisableDefaultQuit suppresses the fallback exit behavior for unhandled
	// quit keys (spec quit_keys), Ctrl-C and, when enabled, Esc.
	DisableDefaultQuit bool
	// DisableGlobalF10Quit lets a pane keep F10 within its widget tree.
	// The default is to quit before dispatching F10 to any widget.
	DisableGlobalF10Quit bool
	// EscapeQuits makes an unhandled Esc close the app. Esc never quits by
	// default (spec escape_quits); overlays and widgets consume it first.
	EscapeQuits bool
	// KittyKeyboard opts in to the kitty keyboard protocol (progressive
	// enhancement flag 1, "disambiguate escape codes") while Run is active, so
	// Shift+Enter, Ctrl+Enter, Ctrl+I and friends arrive as distinct CSI-u keys.
	// Loom pushes the flag on start and pops it on close, panic exit and around
	// Suspend. Terminals without support ignore it. Spec: pane.kitty_keyboard.
	KittyKeyboard bool
	// OnCloseRequest, when non-nil, is asked before the pane closes because of
	// a quit key or an interrupt/termination signal. Returning CloseVeto keeps
	// the loop running; the app may call Quit later (e.g. after a dialog).
	OnCloseRequest func(reason CloseReason) CloseDecision
	// quitting is set by Quit and makes Run return at the next loop turn.
	quitting atomic.Bool

	// Background is composited after the root widget on every frame.
	Background Background
	// BackgroundOnRedraw disables independent animation ticks for animated
	// backgrounds. The background still renders whenever the pane redraws.
	// The zero value keeps animated backgrounds updating continuously.
	BackgroundOnRedraw bool
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
	pasteMode   bool
	kittyKeys   bool // kitty keyboard flag 1 is pushed on the current screen
	fullActive  bool // primary-screen layout is full screen (from row 1)
	staleRows   int  // rows below the pane left over from a shrink, erased by the next frame
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
	// reader is the running input reader; reads is the Run loop's input
	// channel, kept so Suspend can restart the reader on it.
	reader *inputReader
	reads  chan paneRead
	// suspended is set while Suspend runs its callback; resumeRepaint makes
	// the next frame re-read the terminal size and repaint every row.
	suspended      bool
	resumeRepaint  bool
	interrupts     chan os.Signal
	invalidate     chan struct{}
	invalidateOnce sync.Once
	// postMu guards posted and postClosed; Post appends, the run loop drains.
	postMu      sync.Mutex
	posted      []func()
	postClosed  bool
	tickerReset chan struct{}
	help        *Popup
}

// CloseReason tells OnCloseRequest why the pane wants to close.
type CloseReason int

const (
	// CloseReasonQuitKey is a quit key from the spec (Ctrl-Q, F10) or Esc when
	// EscapeQuits is enabled.
	CloseReasonQuitKey CloseReason = iota
	// CloseReasonInterrupt is Ctrl-C or an interrupt, hangup or termination signal.
	CloseReasonInterrupt
)

// CloseDecision is the answer of OnCloseRequest.
type CloseDecision int

const (
	// CloseAllow lets the pane close. It is the zero value.
	CloseAllow CloseDecision = iota
	// CloseVeto keeps the pane running.
	CloseVeto
)

// Quit asks Run to return at the next loop turn. It is safe to call from
// widget callbacks and from other goroutines, and bypasses OnCloseRequest.
func (p *Pane) Quit() {
	p.quitting.Store(true)
	p.Invalidate()
}

// requestClose consults OnCloseRequest and reports whether the pane should close.
func (p *Pane) requestClose(reason CloseReason) bool {
	if p.OnCloseRequest == nil {
		return true
	}
	return p.OnCloseRequest(reason) != CloseVeto
}

const maxPendingPasteSize = 1 << 20

// ResetTicker postpones the next ticker update by its current interval. It is
// safe to call from widget callbacks, including callbacks on another goroutine.
func (p *Pane) ResetTicker() {
	select {
	case p.tickerReset <- struct{}{}:
	default:
	}
}
