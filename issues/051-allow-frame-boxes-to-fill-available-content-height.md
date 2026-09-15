# 051 — Allow Frame boxes to fill available content height

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `frame.go`, `frame_test.go`,
`examples/filebrowser/browser.go`, `288230f`

---

## 1. Problem & Motivation

`Frame.Layout` sizes horizontally arranged boxes from each box's preferred
`Height`. It clamps that height to the available content rows but does not offer
a way to request all available rows. Applications must know that the frame
reserves one title row and one status row, then duplicate that arithmetic.

The filebrowser initially used `Height: 16` inside a 20-row frame, leaving two
unused rows between its boxes and status bar. Removing the visual gap required
changing both boxes to the application-specific constant `Height: 18`. That
works for the example's current dimensions but does not express the actual
intent: fill the frame content area and remain flush with the status bar.

## 2. Design Questions

Add an explicit, spec-compatible way for a box or horizontal row of boxes to
consume the available frame content height. The contract must answer:

- Whether fill is a `Box` property, a height constraint mode, or a frame-level
  cross-axis alignment.
- How preferred, minimum, and maximum heights constrain a filling box.
- Whether all boxes in a horizontal row fill equally or may retain independent
  heights and alignment.
- How fill behaves when the frame switches to its stacked layout, where height
  is the primary allocation axis rather than the cross axis.
- How `ContentHeight` and `HeightForWidth` report a preferred height for a frame
  containing fill semantics without depending on the most recent draw size.

Do not assign special meaning to zero without checking compatibility with
existing validation and programmatic callers. Prefer a named intent that is
clear in both Go and YAML.

## 3. Verification & Acceptance

- Filebrowser boxes reach the status row without hard-coding the frame's
  content height.
- Resizing the available frame height grows and shrinks filling boxes while
  respecting minimum and maximum constraints.
- Horizontal and stacked layouts have documented, tested behavior.
- Hidden boxes and breakpoint transitions do not leave stale gaps.
- Existing fixed-height frames preserve their layout.
- `go test ./...` and `go vet ./...` pass.
