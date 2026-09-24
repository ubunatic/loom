<!-- SPDX-FileCopyrightText: 2026 Uwe Jugel -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

# Emoji & Unicode Width Measurement

This document records the architecture, invariants, and terminal mitigation strategies for emoji and sequence width measurement in Loom.

---

## 1. Core Principles & Single Source of Truth

Loom treats cell widths and rendering transformations as spec-governed data:
- **Spec Author**: [`spec/emoji.yaml`](../spec/emoji.yaml), validated against JSON schema [`spec/schemas/emoji.schema.json`](../spec/schemas/emoji.schema.json).
- **Core Runtime**: [`measure`](../measure/), specifically [`measure/spec.go`](../measure/spec.go) and [`measure/measure.go`](../measure/measure.go).
- **Invariants**:
  1. **No Shadow Tables**: User applications, examples (e.g. `loomoji`), and internal widgets must never maintain hardcoded emoji width maps or shadow tables in Go source. All width logic and terminal transformation modes must originate from `spec/emoji.yaml` and be accessed via `measure.*`.
  2. **Deterministic Line Widths**: Text layouts, category grids, box borders, and UI bars must have identical, deterministic line lengths across all frames.

---

## 2. Terminal Rendering Discrepancies (VTE & XTerm)

Terminal emulators (particularly VTE-based terminals such as GNOME Terminal, Tilix, Ptyxis, and XTerm derivatives) often diverge from Unicode / POSIX `wcwidth` in how they advance the cursor and render glyphs:

### A. Variation Selector-16 (`\uFE0F`) on Single-Cell Bases
- **Behavior**: For base symbols (e.g. `☠`, `⌨`, `☀️`, `☁`, `⚠️`, `☢️`, `☣️`) followed by `\uFE0F`, VTE advances the cursor by the base glyph's `wcwidth` (1 cell), but the terminal font draws the emoji across 2 cells.
- **Resulting Bug**: Without padding, subsequent characters in the row drift left by 1 column, causing right box borders to shift left and row ends to become ragged.
- **Mitigation (`pad-1`)**: The spec configures `vs16_vte_mode: pad-1`. In VTE mode, `measure.ApplyVTEMode(glyph)` outputs `glyph + " "` (the glyph advances the cursor 1 cell, the trailing space advances 1 cell, totaling the 2 cells Loom allocated).

### B. Arrow Ligatures & Font Extensions
- **Behavior**: Long arrows (`⟵`, `⟶`, `⟷`, `⟹`, `⟺`) have East Asian Width Neutral (1 cell in POSIX `wcwidth`), but monospace fonts render wide ligature extensions that clip or drift when rendered in 1 column.
- **Mitigation**: Specced as `width: 2` with `vte_mode: pad-1`.

### C. Multi-part ZWJ Sequences & Overlaps
- **Behavior**: Multi-character sequences with Zero Width Joiner (`\u200D`) like `🐻‍❄️` (Bear + ZWJ + Snowflake + VS16) or `👁️‍🗨️` (Eye + ZWJ + Speech Bubble + VS16) may not be joined into a single ligature by all terminal renderers, rendering as separate consecutive glyphs.
- **Mitigation**: Specced with explicit widths (`4` and `3`) and `vte_mode: pad-1` to maintain visual row alignment.

### D. Naturally Wide VS16 Glyphs
- **Behavior**: Certain glyphs such as `⭐️` (U+2B50 + U+FE0F) advance by 2 cells naturally in VTE fonts.
- **Mitigation**: Specced with explicit override `vte_mode: default` so no unnecessary trailing pad is appended.

---

## 3. Measurement Pipeline Architecture

```mermaid
flowchart TD
    SpecFile["spec/emoji.yaml"] -->|LoadEmojiSpecYAML| Runtime["measure Runtime (atomic.Pointer)"]
    Schema["spec/schemas/emoji.schema.json"] -.->|make validate-spec| SpecFile

    Runtime --> MeasureFuncs["measure.StringWidth / RuneWidth"]
    Runtime --> ModeFuncs["measure.VTEMode / ApplyVTEMode"]

    App["Loom Application / Loomoji"] -->|Measure Bounds| MeasureFuncs
    App -->|Transform for Terminal| ModeFuncs
    ModeFuncs --> RenderOutput["Rendered ANSI Stream"]
```

### Key API Functions (`measure` package)

- `measure.StringWidth(s string) int`: Returns the total visual cell width for string `s`, parsing ANSI escapes, wide runes, combining sequences, and spec overrides.
- `measure.RuneWidth(r rune) int`: Fast-path visual width for a single Unicode rune.
- `measure.VTEMode(glyph string) string`: Looks up the specced VTE render mode (`default`, `pad-1`, `no-vs16`, `no-vs16-pad`, `force-vs16`, `split-zwj`, `base-only`).
- `measure.ApplyRenderMode(glyph string, mode string) string`: Applies a specific render transformation.
- `measure.ApplyVTEMode(glyph string) string`: Applies the specced VTE render mode to `glyph`.

---

## 4. Verification & Testing

1. **Spec Validation**: `make validate-spec` validates `spec/emoji.yaml` against its JSON Schema, including negative control assertions.
2. **Category Width PTY Test**: [`examples/loomoji/loomoji_category_width_pty_test.go`](../examples/loomoji/loomoji_category_width_pty_test.go) executes Loomoji across all categories in a real virtual terminal PTY, asserting that every line has uniform column width.
3. **Debug Tools**:
   - `loomoji debug --grid -W <w>`: Renders all emojis partitioned by their VTE cell width (1, 2, 3, 4) in a grid to visually inspect column alignment.
   - `loomoji debug --measure`: Interactive TUI tool to measure glyph alignment against terminal cell columns.
