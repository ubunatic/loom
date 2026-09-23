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

## Milestones

Preflight (HEAD 63ea328): premise holds — `VT.cells` is `[][]rune`, no SGR
state; loomoji handles `loom.MouseHover`.

### M1 — VT colour cells

- `ptytest.VT` tracks per-cell FG/BG/attrs via SGR (`CSI … m`: reset, 16,
  256, truecolor, reverse); existing rune-only API unchanged.
- Session accessor for the colour grid (current and per-frame if cheap).
- Unit tests feeding raw SGR byte streams; wide runes keep their colour on
  both halves.

M1 delivered (396efdb): VT colour cells — `Cell{Rune,Style}` grid, SGR
16/256/truecolor/attrs/reverse, EL/ED pen-BG fill, `Cells`/`Cell`/`CellFrames`
on VT and Session; ptytest tests green, loomoji untouched.

### M2 — Loomoji hover probe

Pre-Work / Required Refinements (from M1 review):

- Add `Style.Effective() (fg, bg Color)` (swaps under Reverse) with a unit
  test; the probe must compare effective BG, not raw BG.
- `lineFeed` scroll inserts rows with `Style{}`; use the pen BG like ED/EL.


- PTY test under `examples/loomoji/` using only `SendRaw` SGR hover reports
  and the M1 colour grid; loomoji source untouched (`git diff` on
  `examples/loomoji/loomoji/` must stay empty).
- Probes per ticket Notes edges; logs (dx, dy) per probe; currently exposes
  the offset bug.

M2 first pass (5df53d4) rejected in review: the offset table is an artifact.
Pre-Work style fixes (`Style.Effective`, lineFeed pen BG) accepted.

### M2 rework — differential hover detection

Pre-Work / Required Refinements (from M2 review):

- `findHighlighted` returns the first cell with a hard-coded accent RGB; the
  startup focus at (1,3) always wins, so nearly every probe reports (1,3).
- Black-box: no colours or geometry copied from loomoji source. Take a
  baseline grid before each hover (and move the pointer off-grid between
  probes); the hover highlight = cells whose effective BG differs from the
  baseline. Report the full changed-cell bounding box.
- Derive item positions from the rendered screen (emoji cell columns via
  display width), not hard-coded x values.
- Wait until the screen is stable after the hover (no new frame for a short
  settle window), not the first new frame; no fixed 250 ms fallback that
  silently returns a stale grid.
- Distinguish "hover changes nothing" (status NO_CHANGE) from an offset.

M2 rework pass (d009530) reviewed: differential detection and screen-derived
geometry accepted; results still an artifact (timing).

Pre-Work / Required Refinements (from M2 rework review):

- `waitForSettle` returns after 30 ms without any new frame, so a render
  arriving later is read as NO_CHANGE and bleeds into the next probe. The only
  "offset" (probe at (1,5) → change (1,4)..(4,4)) is exactly the previous
  probe's item.
- After each `SendRaw`, require at least one new frame (generous timeout,
  e.g. 2 s) before the quiet window; a timeout with no frame = NO_CHANGE,
  logged as such. Same for the off-grid reset step.
- Add a sanity assertion that at least one on-grid hover produces a change;
  otherwise fail with "hover not observable" instead of reporting offsets.
