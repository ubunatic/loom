# 059 — Fix mouse coordinate convention mismatch between Frame and Tabs/Stack/Grid

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: `widget.go`, `pane.go`, `frame.go`, `tabs.go`, `stack.go`,
`grid.go`, `choice.go`, `view.go`, `mouse_test.go`,
[055](055-add-a-tab-panel-widget-for-pane-hosting.md)

---

## 1. Problem & Motivation

**This is a pre-existing bug, not merely a hosting-readiness gap.** Loom has
two mutually incompatible mouse-coordinate conventions in the tree right now,
and the same leaf widget hit-tests correctly under only one of its possible
parents:

- `Pane.run` delivers pane-relative, **1-based** coordinates (`pane.go:539-542`).
- `Frame.HandleMouse` **re-bases** the event into child-local 1-based
  coordinates before delegating: `e.X, e.Y = x-inner.X+1, y-inner.Y+1`
  (`frame.go:475-493`).
- `Tabs.HandleMouse` (`tabs.go:184-201`), `Stack.HandleMouse`
  (`stack.go:106-113`) and `Grid.HandleMouse` forward the event
  **unchanged**.
- Leaf widgets hit-test against their own stored rect from the last `Draw` —
  `choice.go` does `e.Y -= c.lastRect.Y`, `view.go` compares against
  `v.lastRect.X`. That is the *absolute* convention.

So a `Choice` inside a `Frame` receives coordinates already re-based and then
subtracts `lastRect` a second time: clicks land on the wrong row whenever the
box is not at the canvas origin. `mouse_test.go:133-176` only covers a
standalone `Choice`, so nothing catches it. This bug is reachable today by any
app that puts a clickable widget inside a non-origin `Frame` box — it does not
require the hosting work of 055/056 to hit.

Two further routing defects in the same area:

- `Stack.HandleMouse` and `Grid.HandleMouse` route to the **focused** child
  regardless of where the click landed — `stack.go:107-108` admits this in a
  comment ("Without a Rect at dispatch time we delegate to the focused
  child"). Two side-by-side children each receive the other's clicks.
- Neither stores per-child rects during `Draw`, although `Draw` computes them.

## 2. Design — resolved decisions

**Pin one convention: canvas-absolute, 0-based at the widget boundary**, and
document it on `Widget.HandleMouse` in `widget.go:28-29`.

Rationale, matching the owner's "do coordinate translation, or move to
absolute positioning only at the top level":

- It is what `Pane` already produces and what `Choice`, `View` and `Tabs`
  already assume — the majority convention; only `Frame` deviates.
- It matches the `Draw(c *Canvas, r Rect)` contract, where `r` is already
  canvas-absolute (`widget.go:20-22`). One coordinate space for both halves of
  the widget contract is the maintainable answer; per-parent re-basing is the
  classic UI-toolkit bug source precisely because every composite must get the
  translation right and none of them can be tested in isolation.
- Normalizing 1-based terminal coordinates to 0-based canvas coordinates
  happens exactly once, in `Pane`, instead of being re-derived by every leaf
  (`tabs.go:191` and `choice.go` both do this subtraction today).

Work items:

1. `Pane` converts to 0-based canvas coordinates before dispatch; update the
   leaves that currently do their own `-1`.
2. `Frame.HandleMouse` stops re-basing — it keeps its own hit-test (focus on
   click, bounds check against `inner`) but forwards the event unmodified.
3. `Stack`/`Grid`/`Tabs` store per-child rects during `Draw` (as `Tabs` already
   does for `tabCols`, `tabs.go:66-69`) and route by containment, falling back
   to the focused child only when no rect contains the point.
4. Document the convention on `Widget.HandleMouse` and in the composite doc
   comments.

Rejected: keeping per-parent re-basing and fixing the leaves. It makes every
new composite a potential re-introduction of this bug and forbids a widget
from being drawn by one parent and hit-tested by another.

## 3. Verification & Acceptance

- A failing-first regression test reproducing the current bug: a `Choice`
  inside a `Frame` box at a non-origin rect, click on row *n*, assert the
  selected item — red before the fix, green after.
- Nested-host mouse tests covering `Frame{Choice}`, `Tabs{Frame{Choice}}`,
  `Stack{Choice, Choice}` (side-by-side: each click reaches the child whose
  rect contains it, not the focused one) and `Grid`.
- `Widget.HandleMouse` documents the coordinate convention explicitly; every
  composite's `HandleMouse` doc comment states it forwards unmodified.
- No widget performs a second `lastRect` subtraction after the `Pane`
  normalization (grep review of `choice.go`, `view.go`, `tabs.go`).
- Existing `mouse_test.go` cases still pass (adjusted only where they encode
  the old 1-based input convention).
- `go test ./...` and `go vet ./...` pass.
