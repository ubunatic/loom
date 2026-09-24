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

## Milestones (lean sprint, developer codex:terra:med)

Decisions:
- Results are keyed by terminal profile (`TERM` + `TERM_PROGRAM`), one JSON per profile under
  `docs/data/loomoji-widths/`. A glyph counts as "not measured yet" per profile.
- The `CSI 6n` auto-probe is out of scope; it's a follow-up once the manual workflow works.

- **M1 (persistence)**: store, merge, unmeasured filter, JSON + text report, format README; round-trip tests.
- **M2 (measure widget)**: keyboard-only loom widget with numbered glyphs between markers.
  Keys `1`/`2`/`?`, comment entry, progress indicator, incremental save. Deterministic widget tests.
- **M3 (CLI wiring)**: `loomoji debug --measure`, terminal profile detection, loads M1 data; parsing tests.
  The user does a manual smoke test in a real terminal.

### M1 delivered: persistence (c5d4116)
`MeasurementStore`, one per terminal profile. Includes Merge (keeps existing answers), Unmeasured,
JSON and text save with atomic writes, and a format README. `0` = unsure. `make test-q1` green.

### M2 Pre-Work / Required Refinements
- `Merge` never overwrites, so the TUI can't correct an answer or add a comment afterwards. Add an explicit
  `Set`/`Update` (last write wins) for in-session edits and going back to the previous glyph. Keep `Merge` for load-time merging.
- `LoadMeasurementStore` fails when the file is missing. On a first run, treat that as an empty store
  (needed for M3; add a test).

### M2 delivered: measure widget (fa28ea7, 58ca4d0)
Keyboard-only widget: answers, unsure, comments, going back, progress, incremental save. The pre-work
(Update, empty store when the file is missing) is done. The fix was a test assertion only. The suite hasn't
passed on 58ca4d0 yet; M3's test run must confirm M2.

### M3 delivered: CLI wiring (7dc924b)
`loomoji debug --measure` with a per-terminal profile, loading/saving only unmeasured glyphs. `make test-q1` green (confirms M2 too),
`make install` done.
Remaining: the user's manual smoke test in a real terminal (`go run ./examples/loomoji debug --measure`).
Caveat: the data dir `docs/data/loomoji-widths` is relative to the current directory, so run it from the repo root.

### M4: page table UI (user request after the M3 demo)
- Show 10 unmeasured glyphs at once as a table: row number 0-9, glyph between markers, codepoints,
  loom's computed width, current answer, comment.
- Each answer starts at loom's computed width. Keys `0`-`9` toggle that row between 1 and 2.
- Up/down arrows select a row. `c` edits the selected row's comment (Enter saves, Esc cancels).
  `?` marks the selected row unsure.
- Enter (or PgDn) saves the whole page as answered and moves to the next 10. PgUp goes back one page.
  `q` quits after saving the pages already confirmed.
- Update the widget tests. Rows on a page that isn't confirmed stay unmeasured on disk.
