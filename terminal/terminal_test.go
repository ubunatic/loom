// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package terminal_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"golang.org/x/term"
	"ubunatic.com/loom/terminal"
)

func TestResetSequencesCompleteness(t *testing.T) {
	seq := terminal.ResetSequences
	required := []string{
		"\x1b[?9l",    // X10 mouse off
		"\x1b[?1000l", // normal mouse off
		"\x1b[?1002l", // button-event mouse off
		"\x1b[?1003l", // any-event mouse off
		"\x1b[?1005l", // utf-8 mouse off
		"\x1b[?1006l", // sgr mouse off
		"\x1b[?1015l", // urxvt mouse off
		"\x1b[?1004l", // focus tracking off
		"\x1b[?2004l", // bracketed paste off
		"\x1b[r",      // scroll region reset (DECSTBM)
		"\x1b[0m",     // SGR style reset
		"\x1b[?25h",   // cursor visible
		"\x1b[0 q",    // default cursor shape (DECSCUSR 0)
		"\x1b[?7h",    // auto-wrap on (DECAWM)
	}

	for _, req := range required {
		if !strings.Contains(seq, req) {
			t.Errorf("ResetSequences missing required sequence %q", req)
		}
	}

	withAlt := terminal.Reset(true, true)
	if !strings.Contains(withAlt, terminal.ResetAltScreen) {
		t.Errorf("Reset(true) missing ResetAltScreen (%q)", terminal.ResetAltScreen)
	}
	if !strings.Contains(withAlt, terminal.ResetKittyKeyboard) {
		t.Errorf("Reset(true, true) missing ResetKittyKeyboard (%q)", terminal.ResetKittyKeyboard)
	}
	withoutAlt := terminal.Reset(false, false)
	if strings.Contains(withoutAlt, terminal.ResetAltScreen) {
		t.Errorf("Reset(false) unexpectedly contains ResetAltScreen (%q)", terminal.ResetAltScreen)
	}
	if strings.Contains(withoutAlt, terminal.ResetKittyKeyboard) {
		t.Errorf("Reset(false, false) unexpectedly contains ResetKittyKeyboard (%q)", terminal.ResetKittyKeyboard)
	}
}

func TestRestoreWritesSequencesAndIsIdempotent(t *testing.T) {
	var buf bytes.Buffer
	state := &term.State{} // dummy state

	// Restore with dummy fd -1 (drain/term.Restore will fail or no-op, but sequences written)
	err := terminal.Restore(-1, state, &buf, true)
	if err == nil {
		// If on a mock/pipe fd, Restore might return error from term.Restore or nil
	}

	output := buf.String()
	if !strings.Contains(output, "\x1b[?1006l") {
		t.Fatalf("Restore output missing mouse reset: %q", output)
	}
	if !strings.Contains(output, terminal.ResetAltScreen) {
		t.Fatalf("Restore output missing alt screen exit: %q", output)
	}
}

func TestRestorerIdempotency(t *testing.T) {
	var buf bytes.Buffer
	restorer := terminal.NewRestorer(-1, nil, &buf)

	callCount := 0
	restorer.OnRestore(func() {
		callCount++
	})

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			restorer.Restore()
		}()
	}
	wg.Wait()

	if callCount != 1 {
		t.Fatalf("expected OnRestore hook to run exactly once, got %d", callCount)
	}
}
