# 114 — loomoji debug --measure: interactive TUI to record real terminal emoji widths

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Tooling / Rendering
**Related**: [048](048-emoji-rune-width-discrepancy-causes-horizontal-border-drift.md),
[109](109-detect-loomoji-emoji-width-drift-stray-right-edge-fragments-and-ragged-rows.md),
[110](110-detect-loomoji-row-background-bands-overflowing-the-grid-panel.md),
`examples/loomoji`, `measure/measure.go`, `internal/ptytest/vt.go`

## /goal

`loomoji debug --measure` lets the user record, for every emoji loomoji allows, how many columns
their real terminal draws it (1 or 2). It can also save a free-text comment per glyph.
The results go into a width table that is checked in to the repo. This table is the reference data
for fixing 048 and detecting 109/110.

## Why

The width loom computes (`measure.RuneWidth`) disagrees with real terminals on some emoji,
so loomoji rows end at different columns. Automated tests can't see it: `ptytest.VT` measures
width with the same `measure` function. Only a human looking at a real terminal can tell for sure.

## Requirements

- Build a loom TUI; it must be fast to operate by keyboard alone. It lists the glyphs as a numbered set,
  each drawn between reference markers, so a 1- or 2-column width is easy to see.
- Per glyph: mark `1`, `2` or "other/unsure" with a single keypress, and optionally type a comment.
  Comments are for manual review later.
- Store the results in `docs/data/` as JSON (machine-readable) plus a text report (readable).
  Record for each glyph:
  - codepoints
  - the width loom currently computes
  - the measured width
  - the comment
  - the terminal: `$TERM`, `$TERM_PROGRAM` if available
- Re-runs are incremental: only glyphs that haven't been measured are shown. Answers already saved
  are never discarded.
- The glyph set is whatever loomoji allows, including VS16 and ZWJ sequences and flags,
  so 048 §4 is covered too.

## Open questions

- Should results be kept per terminal (one file per terminal, or keyed by terminal), and does
  "not measured yet" mean for this terminal only?
- Could a `CSI 6n` cursor-position probe auto-fill the answers, with the human only confirming
  disagreements? Record it as a follow-up if it's out of scope here.

## Done when

- The command works end to end and a second run skips glyphs that are already measured.
- The data files are written under `docs/data/` with a short format note (README or header).
- A test covers the load/merge/save round trip and the "unmeasured only" filtering.
