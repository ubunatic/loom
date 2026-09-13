# 037 — Canvas DrawBorder and DrawBox Primitives with Configurable BoxStyles

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [frame.go](file:///home/uwe/projects/loom/frame.go), [popup.go](file:///home/uwe/projects/loom/popup.go), [examples/viewer/viewer.go](file:///home/uwe/projects/termaid/examples/viewer/viewer.go)

---

## 1. Problem & Motivation
Loom draws box/panel borders in at least three independent places, each with its
own copy of the corner/edge-fill logic:

- `Box.Draw` (`frame.go:94-125`) fills corners and edges cell-by-cell from a
  YAML-configured `BoxBorder` (top/bottom/left/right glyphs + title
  prefix/suffix), loaded once from the embedded `spec/box.yaml`.
- `Popup.Draw` (`popup.go:27-69`) builds top/bottom border strings by hand with
  hardcoded `┌ ─ ┐ │ └ ┘` runes and a manual row loop for the side walls —
  unrelated to `Box`'s `BoxBorder` struct or spec file.
- The termaid example app's help modal (`examples/viewer/viewer.go:177-192`)
  hand-rolls a third border, this time with `╭ ─ ╮ │ ╰ ╯` (rounded) runes and
  its own manual side-wall loop, because neither `Box` nor `Popup` exposes a
  reusable, style-selectable primitive it can call into.

Termaid itself already has a `BorderStyle` enum (`BorderStyleRounded`,
`BorderStyleSharp`, `BorderStyleDouble`, `BorderStyleAscii`) for diagram
rendering, but Loom has no equivalent on the `Canvas`/`Box`/`Popup` side, so a
consumer drawing both a diagram panel and a UI box/popup cannot share one
border mechanism or style enum across them.

## 2. Proposed Solution
Add a `Canvas`-level primitive so `Box`, `Popup`, and application code all draw
borders through the same path:

1. A `BoxStyle` type (or reuse/alias termaid's `BorderStyle` shape) enumerating
   at least Sharp (`┌─┐│└┘`), Rounded (`╭─╮│╰╯`), Double (`╔═╗║╚╝`), and ASCII
   (`+-+|++`) glyph sets.
2. `func (c *Canvas) DrawBorder(r Rect, style BoxStyle, s Style)` — draws just
   the border of `r` (no fill, no title).
3. `func (c *Canvas) DrawBox(r Rect, style BoxStyle, title string, s Style)` —
   border plus a rune/cluster-aware centered or left-aligned title, replacing
   the ad hoc title-fitting logic duplicated in `Popup.Draw`.
4. Migrate `Box.Draw` and `Popup.Draw` to call the new primitive instead of
   their private per-cell/per-row loops; `Box`'s existing YAML-driven
   `BoxBorder` can map onto a custom `BoxStyle` for backward compatibility with
   `spec/box.yaml`.

## 3. Verification & Acceptance
- Unit tests for `DrawBorder`/`DrawBox` covering each `BoxStyle`, degenerate
  sizes (width/height < 2), and title truncation/centering with multi-byte
  titles (see [038](038-fix-multi-byte-utf-8-string-truncation-in-popup-title-and-borders.md)).
- `Box.Draw` and `Popup.Draw` golden-render tests still pass unchanged after
  migrating onto the shared primitive.
- The termaid example's help modal can be rewritten to call
  `canvas.DrawBox(...)` instead of its own hand-rolled border block.
