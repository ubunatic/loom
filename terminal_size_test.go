package loom

import "testing"

func TestTerminalSize(t *testing.T) {
	cols, rows, err := TerminalSize()
	if err != nil {
		// Non-interactive test runners need not have a controlling terminal.
		if cols != 0 || rows != 0 {
			t.Fatalf("TerminalSize error returned dimensions %dx%d: %v", cols, rows, err)
		}
		return
	}
	if cols < 1 || rows < 1 {
		t.Fatalf("TerminalSize = %dx%d with nil error", cols, rows)
	}
}
