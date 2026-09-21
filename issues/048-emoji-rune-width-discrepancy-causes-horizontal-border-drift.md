# 048 — Discrepancy between Unicode / Loom width calculation and terminal rendering for emoji in box titles

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug / Rendering
**Related**: `measure/measure.go`, `canvas.go`, `frame.go`, `rawscreen.go`

---

## 1. Problem & Motivation

When emojis or pictorial symbols such as `🖼` (U+1F5BC `FRAMED PICTURE`) are used in a `loom.Box` title, terminal emulators (e.g. `tilix`, `foot`, VTE-based terminals) render the glyph across **1 column** when in text/monochrome presentation (East Asian Width Neutral, single column in `wcwidth`), whereas Loom's `measure.RuneWidth(r)` categorizes the range `0x1f000`..`0x1faff` unconditionally as width **2**.

Because Loom assumes the glyph takes 2 cells, it advances by 2 columns when building the row and skips one trailing border dash. When the terminal emulator renders the glyph in 1 column, all subsequent characters in that row shift left by 1 column, causing:
1. The right border of the first box to drift left.
2. The gap between boxes and the second box's top-left corner to be visibly misaligned.

---

## 2. Technical Evidence

1. In `measure/measure.go`:
   ```go
   if r >= 0x1f000 && r <= 0x1faff {
       return 2
   }
   ```
2. Character `🖼` (U+1F5BC):
   - East Asian Width: `N` (Neutral).
   - Standard POSIX `wcwidth(0x1F5BC)`: `1`.
   - Without Variation Selector-16 (`\uFE0F`), text terminals render it as a single-column monochrome box icon.
3. Because `Canvas.Row()` outputs the rune and advances Loom's grid by 2, the terminal advances its cursor by only 1, offsetting the entire rest of the line.

---

## 3. Proposed Resolution

1. Audit and refine `measure.RuneWidth` against standard Unicode East Asian Width / `wcwidth` tables and Emoji Presentation specifications (distinguishing default emoji presentation vs text presentation / VS-16).
2. For applications requiring deterministic 1-column icons in borders and titles, standard 1-column geometric/symbol glyphs (like `▣`, `▧`, `▦`, `◈`, `•`, `*`) or explicit bracket tags should be used to guarantee single-cell alignment across all terminals.

---

## 4. Related finding from lean sprint 096 (2026-09-21)

`loom.StringWidth` reports a ZWJ family sequence as 6 columns and a regional-indicator flag as 4,
while terminals show both as 2. The `examples/textrender` example records these as
`knownDivergences`. Include ZWJ sequences and flag pairs in the audit in section 3.
