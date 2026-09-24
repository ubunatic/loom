# 115 — Apply measured emoji widths and feedback from docs/data/loomoji-widths/vte*

**Status**: Closed — M1-M3 delivered in 6d040b3 with spec/emoji.yaml, measure/spec.go, VTE column, and suite green
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Rendering / Core
**Related**: [048](048-emoji-rune-width-discrepancy-causes-horizontal-border-drift.md),
[109](109-detect-loomoji-emoji-width-drift-stray-right-edge-fragments-and-ragged-rows.md),
[110](110-detect-loomoji-row-background-bands-overflowing-the-grid-panel.md),
[114](114-loomoji-debug-measure-interactive-tui-to-record-real-terminal-emoji-widths.md),
`docs/data/loomoji-widths/vte-8401.json`, `docs/data/loomoji-widths/vte-8401.txt`,
`measure/measure.go`, `spec/emoji.yaml`

## /goal

Incorporate the real terminal emoji width measurements and review feedback recorded in `docs/data/loomoji-widths/vte*` (`vte-8401.json`, `vte-8401.txt`) into Loom's core width calculation (`measure/`), `spec/emoji.yaml`, and rendering/padding logic.

## Why

Issue 114 provided the interactive measurement and review tool, producing empirical data for real-world terminal rendering (specifically VTE / terminal profiles). The recorded data identifies glyphs where Loom's computed width deviates from actual terminal drawing and cursor advancement (e.g. VS16 sequences, text-presentation vs emoji-presentation differences, ZWJ sequences, and specific multi-cell glyphs). Applying these findings is the core resolution step for issue 048 and addresses the root causes of emoji-width drift (109) and background overflow (110).

## Requirements

- Parse and analyze discrepancy points and human annotations in `docs/data/loomoji-widths/vte-8401.json` / `vte-8401.txt`.
- Reconcile glyph widths where `measured_width` differs from computed width or where specific comments indicate display padding adjustments:
  - Text-default base characters with VS16 variation selectors (e.g., `☝️`, `☠️`, `✌️`, `✍️`, `👁️`, `🖐️`).
  - Multi-part ZWJ and modifier sequences (e.g., `🐻‍❄️`, `❤️‍🩹`).
  - Single/double-width edge cases (e.g., `✊` U+270A, arrows `➔`, `⟹`, stars `⭐️`, ticket `🎫`).
- Integrate overrides or refined width lookup rules into `measure/` and/or `spec/emoji.yaml`.
- Ensure padding/rendering helpers correctly align glyphs between box borders and grid cells.
- Maintain consistency with automated test environments and terminal emulator VT models.
- **TUI Addendum (`debug --measure`)**: Add a dedicated "VTE" column to the `loomoji debug --measure` table alongside Loom's generic computed width, showing what Loom predicts/evaluates the character width to be in VTE-based terminals.

## Done when

- Measured widths and specific feedback annotations from `docs/data/loomoji-widths/vte*` are integrated into Loom's measurement logic or specs.
- `debug --measure` includes a "VTE" column displaying Loom's predicted VTE width.
- Unit and PTY tests pass and verify the updated width metrics and column rendering.
- `make test-q1` and `make install` succeed.

## Milestones (lean sprint, developer agy:flash37:med)

- **M1 (VTE column in `debug --measure`)**: Add the "VTE" column to the `loomoji debug --measure` paged table and review UI, displaying what Loom predicts/evaluates for VTE terminals alongside computed width. Update table layout & widget tests.
- **M2 (Width reconciliation & spec/measure overrides)**: Integrate measured terminal widths and annotations from `docs/data/loomoji-widths/vte-8401.json` into `spec/emoji.yaml` and `measure/` logic (handling VS16, ZWJ sequences, single/double width edge cases).
- **M3 (End-to-end alignment & regression tests)**: Verify border alignment and rendering in widgets/PTY tests without drift. Ensure `make test-q1` and `make install` pass.

### M1 delivered: VTE column in debug --measure (ada9da8)
Added `VTEWidth` to `Measurement` model, implemented `EvaluateVTEWidth` handling VS16/ZWJ/flags/explicit wide emojis, rendered `%3d` `VTE` column in `MeasureWidget`, added test cases in `measure_widget_test.go` and `measure_store_test.go`. `make test-q1` passed, `make install` complete.

### M2 Pre-Work / Required Refinements
- Inspect `docs/data/loomoji-widths/vte-8401.json` for discrepancies where `measured_width != computed_width` or `answered: true` with specific comments.
- Update `spec/emoji.yaml` width annotations and `measure/measure.go` (or `measure/emoji.go`) width calculation functions so that sequence-aware width calculation correctly computes terminal display width for emojis, VS16 variations, flags, and ZWJ combinations.
- Ensure backwards compatibility with standard rune width calculation for non-emoji text.

