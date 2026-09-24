# 113 — Fix filebrowser arrow up/down jumping multiple items

**Status**: Closed — verified single dispatch and filebrowser arrow navigation
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [105](105-unify-filebrowser-navigation-pane-across-examples.md) (shared `NavigationPane`),
[063](063-convert-filebrowser-example-to-a-hostable-widget.md) (hostable widget, c0612e0)

## Goal

One arrow up/down press in `examples/filebrowser` moves the selection by exactly one item,
standalone and hosted.

## Observed

User smoke test after 105/063 (2026-09-24): up/down jumps several items per press. The rest of the
105 checklist (ESC back, parent reselect, `/` filter, click select) looks fine.

**Reproduction evidence (2026-09-24)**

Added a PTY test that starts the standalone filebrowser, waits for the initial `..` row, sends one
down-arrow sequence (`ESC [ B`), and checks the selected row. The assertion fails: the list skips
`alpha.txt` and highlights `beta.txt`. Captured `make test-q1` output reports:

```text
--- FAIL: TestFilebrowserPTYArrowMovesOneItem
    browser_093_pty_test.go:74: one down arrow did not select alpha.txt
        ... Name: beta.txt ...
        ... ▶ beta.txt ...
```

The standalone and framed in-process dispatch probes, plus ansiviewer's framed probe, each moved
one row. The PTY discrepancy is reproduced, but its root cause and a safe production fix remain
unresolved. The quota test command exited 2; no commit was made.

## Notes

- Suspect: one key reaching the list more than once, e.g. via frame focus dispatch and
  `browser.HandleKey`/`ConsumeKey`/`NavigationPane.HandleKey` together. Verify, don't assume.
- Also check `examples/ansiviewer` (same pane since 105 M3).
- Reproduction test first: send one "down" through the real key path (framed/hosted and
  standalone, ideally a PTY test) and assert the selection moves by one. It must fail before the fix.

User smoke of 105 (2026-09-24): ESC back, parent reselect, `/` filter, click select all fine; only arrow multi-jump (this ticket).

Escalated to flash37 (luna: repro only). Hint: the framed in-process path moves one row, the PTY path skips one, so look at terminal input decoding and frame/browser/pane key routing for arrows (e.g. ESC-prefixed sequences, or a key both consumed and forwarded).
