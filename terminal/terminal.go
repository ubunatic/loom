// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package terminal provides safe, unified, idempotent terminal mode restoration
// for loom commands, panes, and external tools that enter raw mode.
package terminal

import (
	"io"
	"os"
	"os/signal"
	"sync"

	"golang.org/x/term"
)

// ResetSequences contains the ANSI escape sequence string that resets every
// terminal mode a child or loom process may set: mouse tracking (9, 1000, 1002,
// 1003, 1005, 1006, 1015 off), focus events (1004 off), bracketed paste (2004 off),
// DECSTBM scroll region reset (r), SGR reset (0m), cursor visibility (25h), default
// cursor shape (0 q), and auto-wrap on (7h).
const ResetSequences = "\x1b[?9l" +
	"\x1b[?1000l" +
	"\x1b[?1002l" +
	"\x1b[?1003l" +
	"\x1b[?1005l" +
	"\x1b[?1006l" +
	"\x1b[?1015l" +
	"\x1b[?1004l" +
	"\x1b[?2004l" +
	"\x1b[r" +
	"\x1b[0m" +
	"\x1b[?25h" +
	"\x1b[0 q" +
	"\x1b[?7h"

// ResetAltScreen is the ANSI sequence to exit the alternate screen buffer.
const ResetAltScreen = "\x1b[?1049l"

// ResetKittyKeyboard is the ANSI sequence to pop the Kitty keyboard protocol state.
const ResetKittyKeyboard = "\x1b[<u"

// Reset returns ResetSequences, plus ResetKittyKeyboard if kittyPop is true,
// and ResetAltScreen if altScreen is true.
func Reset(altScreen bool, kittyPop ...bool) string {
	var s string
	if len(kittyPop) > 0 && kittyPop[0] {
		s = ResetSequences + ResetKittyKeyboard
	} else {
		s = ResetSequences
	}
	if altScreen {
		return s + ResetAltScreen
	}
	return s
}

// Restore writes the full reset sequence set to out, drains unread input on fd,
// and restores state using term.Restore. It is safe to call with an invalid or
// non-terminal fd (such as in tests).
func Restore(fd int, state *term.State, out io.Writer, altScreen bool, kittyPop ...bool) error {
	pop := len(kittyPop) > 0 && kittyPop[0]
	if out != nil {
		_, _ = out.Write([]byte(Reset(altScreen, pop)))
	}
	if fd >= 0 {
		DrainInput(fd)
		if state != nil {
			return term.Restore(fd, state)
		}
	}
	return nil
}

// Restorer manages terminal cleanup hooks, alternate screen state, and signals.
// It ensures that terminal restoration is thread-safe and strictly idempotent.
type Restorer struct {
	fd            int
	state         *term.State
	out           io.Writer
	altScreen     bool
	kittyKeyboard bool

	mu        sync.Mutex
	once      sync.Once
	hooks     []func()
	sigCh     chan os.Signal
	sigDone   chan struct{}
}

// NewRestorer constructs a Restorer for the given terminal file descriptor and output writer.
func NewRestorer(fd int, state *term.State, out io.Writer) *Restorer {
	return &Restorer{
		fd:    fd,
		state: state,
		out:   out,
	}
}

// SetAltScreen configures whether Restore will exit the alternate screen buffer.
func (r *Restorer) SetAltScreen(on bool) {
	r.mu.Lock()
	r.altScreen = on
	r.mu.Unlock()
}

// SetKittyKeyboard configures whether Restore will pop the Kitty keyboard protocol.
func (r *Restorer) SetKittyKeyboard(on bool) {
	r.mu.Lock()
	r.kittyKeyboard = on
	r.mu.Unlock()
}

// AltScreen reports whether the alternate screen buffer exit is enabled.
func (r *Restorer) AltScreen() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.altScreen
}

// OnRestore registers a hook to be run during restoration before mode reset.
func (r *Restorer) OnRestore(fn func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if fn != nil {
		r.hooks = append(r.hooks, fn)
	}
}

// Restore runs all registered cleanup hooks, writes reset sequences, drains input,
// and restores terminal cooked state. Restore is idempotent.
func (r *Restorer) Restore() {
	r.once.Do(func() {
		r.StopSignals()

		r.mu.Lock()
		hooks := r.hooks
		alt := r.altScreen
		kitty := r.kittyKeyboard
		fd := r.fd
		state := r.state
		out := r.out
		r.mu.Unlock()

		for _, fn := range hooks {
			fn()
		}

		_ = Restore(fd, state, out, alt, kitty)
	})
}

// CatchSignals configures the Restorer to listen for the provided signals.
// If any received signal is caught, onSignal is invoked (if non-nil),
// Restore() is executed, and r exits or hands control back.
func (r *Restorer) CatchSignals(onSignal func(os.Signal), sigs ...os.Signal) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sigCh != nil || len(sigs) == 0 {
		return
	}
	r.sigCh = make(chan os.Signal, len(sigs))
	r.sigDone = make(chan struct{})
	signal.Notify(r.sigCh, sigs...)

	go func(ch <-chan os.Signal, done <-chan struct{}) {
		select {
		case sig, ok := <-ch:
			if !ok {
				return
			}
			if onSignal != nil {
				onSignal(sig)
			}
			r.Restore()
		case <-done:
			return
		}
	}(r.sigCh, r.sigDone)
}

// StopSignals unregisters the signal handler.
func (r *Restorer) StopSignals() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sigCh != nil {
		signal.Stop(r.sigCh)
		close(r.sigDone)
		close(r.sigCh)
		r.sigCh = nil
		r.sigDone = nil
	}
}
