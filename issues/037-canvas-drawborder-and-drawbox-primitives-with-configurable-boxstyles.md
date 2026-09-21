# 037 — Canvas DrawBorder and DrawBox Primitives with Configurable BoxStyles

**Status**: Closed — delivered via lean sprint (M1-M4); haiku dev, luna:low fixed evidence
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

---

## Resolved Design (host decisions, binding for the sprint)

- `BoxStyle` is an enum-like type with Sharp, Rounded, Double, ASCII (String and parse helpers).
  Glyph sets live in the spec system (docs/Spec.md: spec YAML is the source of truth, Go must not
  duplicate glyph values); extend `spec/box.yaml` or add a sibling spec file, and validate-spec must accept it.
- `Canvas.DrawBorder(r Rect, style BoxStyle, s Style)`: border only, clips to the canvas, no-op when
  width or height is below 2.
- `Canvas.DrawBox(r Rect, style BoxStyle, title string, s Style)`: border plus a left-aligned title
  in the top edge, display-width aware (runes, wide runes, combining marks), truncated with an
  ellipsis when it does not fit, never splitting a cluster (see 038).
  Centered titles are out of scope.
- `Box.Draw` and `Popup.Draw` migrate onto the primitive. Their existing golden tests must pass
  unchanged; if a golden must change, stop and record why in your report instead of editing it.
- The termaid example is in another repo: out of scope.

## Milestones (lean sprint, dev agent: haiku)

Host reviews only diffs and test output; this ticket is the only channel. Root-package tests
needing `/dev/tty` fail before this work; ignore them. Commit each milestone (message ends
'(issue 037 MX)'), stage only your files, never docs/README.md, keep the repo root free of stray
binaries. Evidence: frames produced by code, gated on env `LOOM_EVIDENCE=1`, written to repo-root
`docs/progress/037/` (find root by walking up to `go.mod`). gofmt touched files.

### M1 - BoxStyle, DrawBorder, DrawBox
- Spec entry, `BoxStyle`, both primitives, tests: each style, degenerate sizes, clipping at the canvas
  edge, multi-byte and wide-rune titles, truncation.
- Evidence: `M1-boxstyle-gallery.ansi` (all four styles, plain and titled), `M1-titles.ansi` (CJK,
  umlaut, emoji, long title truncated, tiny widths).

### M2 - Migrate Box.Draw and Popup.Draw
- Migrate both onto the primitives with all existing golden tests unchanged; delete the now-dead
  private border loops.
- Evidence: `M2-box.ansi` and `M2-popup.ansi` rendered through the migrated code.

### M1-M2 Review (host)
Delivered: 25a0ba3, 3c10c9d. Tests, vet, validate-spec green; goldens untouched; glyphs come from
the spec. Defects:
- `M2-popup.ansi` is unreadable: the three popups overlap, so borders cross ('Popup 1 │────').
- After your evidence run, frames of other tickets were modified in the working tree
  (docs/progress/061 and 062). I restored them. Never run evidence generation with
  `LOOM_EVIDENCE=1` across `./...`; run it only for the tests of this ticket
  (`go test -run <YourEvidenceTest> .`) and check `git status` before committing.
- Popup is only ever drawn Sharp; nothing proves Popup output equals the pre-migration output beyond the goldens.

### M3 - Pre-Work / Required Refinements
1. Regenerate `M2-popup.ansi` with three NON-overlapping popups (different positions/sizes: short title,
   long truncated title, CJK title), readable at 80x24.
2. Add a test that `Popup.Draw` output equals a direct `Canvas.DrawBox(..., BoxBorderStyleSharp, ...)`
   for the same rect, title and style (cell by cell, including a truncated CJK title).
3. Commit '(issue 037 M3)', stage only your files, `git status` clean afterwards.

### M3 Review (host)
Delivered 30937f7. The equivalence test is accepted. `M2-popup.ansi` is still not clean: popup 'OK'
overlaps a caption line and another box, and 'Popup 1 (short)' has no visible box. Second failed try
at this frame, so the step moves to a stronger developer (escalation ladder).

### M4 - Evidence Fix Only
Rewrite the M2 popup evidence test so the frame is built from a fixed grid: three popups in three
separate columns of one row (each at least 3 rows apart from any caption text, no overlap of any
cell), captions on their own row above each popup. Assert in the test that no two popup rects
intersect and no caption cell is inside a popup rect. Regenerate `docs/progress/037/M2-popup.ansi`
by running only that test with `LOOM_EVIDENCE=1`. Commit '(issue 037 M4)', stage only your files.
