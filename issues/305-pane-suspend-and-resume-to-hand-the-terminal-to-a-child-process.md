# 305 — Pane Suspend and Resume to hand the terminal to a child process

**Status**: Closed — Pane.Suspend stops the poll-based reader (replaying held bytes), leaves and restores terminal modes around fn, repaints at the new size; PTY tests plus examples/suspend (Ctrl+G $EDITOR)
**Priority**: P1 (High)
**Severity**: Feature
**Category**: Pane / Terminal Control
**Related**: [066](066-nested-help-pane-opens-a-second-pane-racing-the-outer-pane-tty-reader.md), harnez #737, #740, #742

---

## 1. Problem & Motivation
A Loom app cannot run an interactive child (`$EDITOR`, a PTY session) on its
terminal. `Pane` offers only `Run` and `Close`. harnez tui needs this for
Ctrl+G ($EDITOR, harnez #740) and Ctrl+O (foreground PTY, harnez #742).
Closing and reopening the Pane is no fix: the stdin reader goroutine
(`pane.go:1076`) keeps reading and steals the child's input (same race as 066).

## 2. Technical Specification / Findings
- Add `Pane.Suspend(fn func() error) error`, called from the event loop (key handler):
  1. stop the input reader without losing buffered bytes;
  2. leave alt screen, disable mouse and bracketed paste, show cursor, restore cooked mode;
  3. run `fn` (the caller starts the child on the real tty);
  4. re-enter raw mode, restore screen/mouse/paste modes, re-read size, restart reader, full repaint.
- Ignore SIGWINCH repaints while suspended; apply the latest size on resume.
- Restore the terminal even if `fn` panics or returns an error.

## 3. Implementation & Verification Plan
- PTY test: a key triggers `Suspend` running a child that reads one line; the
  child gets the line, the Pane gets no stray key, and the screen repaints after.
- PTY test: resize during suspend; resumed frame matches the new size.
- Example in `examples/` that opens `$EDITOR` on Ctrl+G.
