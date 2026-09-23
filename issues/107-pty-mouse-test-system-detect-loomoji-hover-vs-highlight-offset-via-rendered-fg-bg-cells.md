# 107 — PTY mouse test system: detect loomoji hover-vs-highlight offset via rendered FG/BG cells

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Testing
**Related**: `internal/ptytest/vt.go`, `internal/ptytest/session.go`,
`examples/loomoji/loomoji/loomoji.go`,
`examples/filebrowser/filebrowser/browser_093_pty_test.go`,
[089](089-add-mouse-cursor-position-hints-and-configurable-visual-effects.md)

## Goal

Build a black-box mouse test system that **detects and reports** the current
loomoji bug where the hovered cell ≠ the highlighted cell — purely by using the
running app like a user and observing rendered terminal FG/BG cells.

Done when a PTY test: opens loomoji → sends SGR hover reports over grid cells →
reads the rendered screen with colors → locates the cell(s) carrying the hover
BG → compares with the hovered terminal cell (term chars, or pixels if needed)
→ reports the (dx, dy) offset per probe. The test must currently expose the
bug (fail or log a non-zero offset).

## Constraints

- **Do not change any code in loomoji. Do not fix the bug.** Detection and
  reporting only; the fix is a separate ticket.
- Black-box only: no internal state access, no debug flags or status-line probes
  added to the app.

## Notes

- `ptytest.VT` stores only runes (`[][]rune`) and ignores SGR, so hover BG is
  currently invisible to tests. Likely needs a cell grid with FG/BG/attr and
  `CSI … m` parsing (16/256/truecolor, reverse).
- SGR mouse coords are 1-based; hover = `\x1b[<35;X;YM`.
- Probe edges: first/last cell of an item, both halves of wide (emoji) cells,
  cells near the new left/top padding (commits `b71f8c9`, `95926f9`).
- `findPTYText` uses rune index, not display width — wrong after wide runes.
- Prefer waiting on synchronized frames (`Frames()`) over `time.Sleep`.
- Re-verify against live code/commits before starting.
