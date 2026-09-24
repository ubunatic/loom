# 115 — Apply measured emoji widths and feedback from docs/data/loomoji-widths/vte*

**Status**: Open
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
