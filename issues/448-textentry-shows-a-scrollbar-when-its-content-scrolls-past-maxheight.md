# 448 — TextEntry shows a scrollbar when its content scrolls past MaxHeight

**Status**: Closed — bar in the last column, text wraps one column narrower while it shows; thumb drag with pane capture, track click, Esc restore; middle scrub left out
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 447

---

## 1. Problem & Motivation
A long `TextEntry` grows to `MaxHeight` rows and then scrolls, but nothing shows that more rows exist above or below (found in the `loom widgets --show` gallery). `View` and `Viewport` already draw a scrollbar; `TextEntry` does not (`text_entry.go` has no scrollbar code).

## 2. Technical Specification / Findings
- Open design point: the last cell of every row is reserved for the caret (see issue 447), so the bar cannot take a column inside the text width without changing wrap width. Options: widen the reserved column to also hold the bar, or draw the bar over the reserved caret column on rows where the caret is not.
- The bar shows only when the rows exceed `MaxHeight`, so short entries look as before.
- Reuse the scrollbar thumb helpers in `scrollbar.go`; mouse drag on the bar is optional and can be a follow-up.
- A hosting app (taia) draws its own chat-log scrollbar and is unaffected unless it adopts this one.

## 3. Implementation & Verification Plan
- Decide the column layout, then draw the bar in `TextEntry.Draw` and keep `Measure`, `rows` and `Cursor` consistent with the wrap width.
- Tests: no bar when the rows fit; thumb size and position when they overflow; wrap width equals the content width in both cases.
- Document the behavior in `docs/Widgets.md` (TextEntry section) and check the gallery with `loom widgets --show`.
