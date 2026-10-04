# 262 — Flaky TestPaneFirstDrawUsesScreenBounds: PTY width 99 instead of 100

**Status**: Closed — duplicate of 247; observation moved there
**Priority**: P2 (Medium)
**Severity**: Normal
**Category**: Bug
**Related**: `pane_pty_test.go:70`, commit c0326c1 (alternate-screen bounds before first draw), issue 261 (where it surfaced)

---

/goal Make TestPaneFirstDrawUsesScreenBounds deterministic (fix the race or the test setup, not a loosened assertion), verified by several consecutive green runs, or stop and report when blocked.

## Observed

During issue 261 M1 (which touched only RichTextEdit), `make test-q1` failed three times in a row, each time in a different subtest:
`pane_pty_test.go:70: first draw at row 7 with {X:0 Y:0 W:99 H:24}, want row 7 and 100x24` (subtests `inline` and `wrapped-request-auto-alt`). The same suite passed earlier in the session (after 258, 260).

## Notes

- Looks like a race between setting the PTY size and the first size probe, or a width reduced by one for a reserved last column. Check which.
