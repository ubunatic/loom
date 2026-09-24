<!-- SPDX-FileCopyrightText: 2026 Uwe Jugel -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

# Emoji Width Measurement Consolidation and VTE Modes

This case study records the consolidation of emoji width measurement between Loom and Loomoji, the elevation of VTE render modes to the authoritative spec system, and PTY verification across categories.

---

## 1. Context & Problem Statement

Prior to this work, Loomoji had a private `measure_store.go` with localized heuristics, duplicated shadow tables, and custom rendering overrides that diverged from `measure` and `spec/emoji.yaml`. This created several friction points:
- Glyphs like `☠️`, `⌨️`, `☀️`, `⚠️` rendered with correct width in `loomoji debug --measure`, but appeared misaligned or unpadded in `loomoji debug --grid` and standard layouts.
- VS16 emojis were handled via ad-hoc strings in Loomoji rather than via authoritative spec configuration.
- There was no automated PTY test ensuring that cycling through all emoji categories maintained equal line widths without horizontal drift.

---

## 2. Technical Interventions

1. **Measurement Consolidation (`129d2fb`)**:
   - Replaced duplicate width evaluation with `measure.StringWidth(glyph)`.
   - Removed hardcoded shadow tables in Loomoji.
   - Refined `measure.StringWidth` cluster parsing to correctly handle zero-width combining runes and variation selectors in cluster heads.
   - Added keybinding `f` in Loomoji for interactive category cycling.
   - Added `loomoji debug --grid -W <w>` command to view emoji sets partitioned by terminal cell width.
   - Added `examples/loomoji/loomoji_category_width_pty_test.go` to test that all lines across all categories have uniform width in a PTY.

2. **Authoritative Spec Elevation (`52df081`)**:
   - Extended `spec/schemas/emoji.schema.json` and `spec/emoji.yaml` with `vs16_vte_mode: pad-1` and per-glyph `vte_mode` overrides.
   - Added `measure.VTEMode`, `measure.ApplyRenderMode`, and `measure.ApplyVTEMode` to `measure/spec.go`.
   - Wired Loomoji (`examples/loomoji/loomoji/measure_store.go`) directly to `measure.VTEMode` and `measure.ApplyRenderMode`.

3. **Flakiness Isolation in Test Suite**:
   - Isolated asynchronous `SIGWINCH` delivery in `TestPaneNonTickerIdleGuard` (`ticker_test.go`) by setting `p.ResizeConfig.ResizeHandling = false` during non-ticker idle tests.

---

## 3. Results & Invariants

- `spec/emoji.yaml` is now the single source of truth for both display widths and VTE rendering transformation modes across the entire repository.
- `make test-q1` passes cleanly including all spec validation checks and negative controls.
- `make install` refreshes `loomoji` and library binaries.
