# 038 — Fix Multi-byte UTF-8 String Truncation in Popup Title and Borders

**Status**: Closed — completed in lean sprint 2026-09-20
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [popup.go](file:///home/uwe/projects/loom/popup.go), [frame.go](file:///home/uwe/projects/loom/frame.go)

---

## 1. Problem & Motivation
`Popup.Draw` builds and truncates its title using byte length and byte
slicing, not rune/display width:

```go
// popup.go:45-48
title := " " + p.Title + " "
if len(title) > pw-2 {
    title = title[:pw-2]
}
```

`len(title)` counts UTF-8 bytes, not runes or terminal columns, so:
- A title within its rune/display-width budget can still be misjudged as too
  long (or too short) whenever it contains any multi-byte character — e.g. an
  emoji like the `📖` glyph the termaid viewer example uses in its own header
  text (`examples/viewer/viewer.go:84`), or any accented/CJK text.
- When truncation *does* trigger, `title[:pw-2]` is a raw byte slice. Cutting
  in the middle of a multi-byte UTF-8 sequence produces an invalid/corrupted
  string, which renders as replacement characters or broken glyphs in the
  terminal.

`frame.go` already solves this correctly for `Box` titles via `StringWidth`,
`textClusters`, and `writeBounded` (`frame.go:71-76`, `510-520`), which measure
and truncate by display width and grapheme cluster, never by raw byte index.
`Popup` is the one remaining widget still using the byte-based approach.

## 2. Proposed Solution
- Replace `len(title)` with `StringWidth(title)` (or equivalent) for the
  length check in `Popup.Draw`.
- Replace `title[:pw-2]` with cluster-aware truncation, reusing
  `textClusters`/`writeBounded` from `frame.go` (or the `DrawBox` primitive
  from [037](037-canvas-drawborder-and-drawbox-primitives-with-configurable-boxstyles.md),
  if sequenced after it) instead of a raw byte slice.

## 3. Verification & Acceptance
- Unit test: a `Popup` with a title containing multi-byte runes (emoji,
  accented characters, CJK) at widths that force truncation renders valid
  UTF-8 with correct display width, no split runes, and no panic.
- Unit test: a multi-byte title that fits within `pw-2` display columns is
  *not* truncated, even though its byte length may exceed `pw-2`.
- Regression test confirms existing ASCII-title `Popup` golden renders are
  unchanged.
