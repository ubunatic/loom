# 048 — Discrepancy between Unicode / Loom width calculation and terminal rendering for emoji in box titles

**Status**: Closed — resolved 2026-09-24; flag width finding no longer reproduces (measure and check-box agree on 2 columns)
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

## 5. Measured data (issue 114, VTE 0.84 / xterm-256color, 2026-09-24)

Data: `docs/data/loomoji-widths/xterm-256color--unknown.{json,txt}` (1132 glyphs). The measured value is
"loom width + padding at which the next marker lines up". Findings:

1. **VS16 on a base that is text by default** (`Emoji_Presentation=No` + U+FE0F), about 110 glyphs. VTE moves the cursor
   by the base's width (1) and ignores VS16, but draws the glyph 2 wide. Loom is wrong both ways:
   U+2xxx bases (`⚠️ ☀️ ❤️`) count 1 and need +1 padding; U+1F3xx-1F6xx bases (`🏖️ 🗂️ 🛠️`) count 2 but the terminal
   advances 1, so they also need +1. Fix: width 2, emitted as `glyph + " "` (the terminal advances 1 for the glyph, plus 1 for the space).
2. **`✊` U+270A** (`Emoji_Presentation=Yes`, EAW=W): true width 2, loom says 1. Plain table bug.
3. **ZWJ sequences**: VTE doesn't join them. It advances by the sum of the parts and draws them overlapping
   (`❤️‍🔥` = 3, `🐈‍⬛` looks right at 3). Width = sum of the parts' terminal advances, not 2.
4. **Long arrows `⟵ ⟶ ⟷ ⟹ ⟺`** (EAW=N, width 1): the font draws them wider. Nothing lines up cleanly ("never centers");
   avoid them in layout, or pad them.

## 6. Resolution & Authoritative Spec Architecture (2026-09-24)

- Specced `vs16_vte_mode: pad-1` in `spec/emoji.yaml` and schema `spec/schemas/emoji.schema.json`.
- Implemented `measure.VTEMode`, `measure.ApplyRenderMode`, and `measure.ApplyVTEMode` in `measure/spec.go`.
- Unified Loomoji with Loom's `measure` package, eliminating shadow tables.
- Added PTY category width test in `examples/loomoji/loomoji_category_width_pty_test.go`.
- Documented in [`docs/EmojiWidth.md`](../docs/EmojiWidth.md).

## Finding 2026-09-27 (132, ruler mockup)
Loom disagrees with itself about the DE flag (🇩🇪): the ANSI box check (TestAllAnsiAssetsHaveValidBoxes)
counts it as 4 columns, while ansiviewer's cluster replay (132, via ParseANSI) places the next text
2 columns later. The terminals tested in 096 draw it as 2 columns or as a 2-column glyph.
