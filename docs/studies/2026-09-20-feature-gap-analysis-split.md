---
title: Split Example Feature Gap Analysis
---
# Split Example Feature Gap Analysis

## Scope

Reviewed `examples/split/main.go` and `examples/split/split/split.go` against
the current `loom.Frame`, `loom.Box`, and focus APIs.

## Findings

- Basic horizontal split layout is already a framework feature. The example
  creates two `loom.Box` values in `split.Run`; `Frame.Layout` in `frame.go`
  allocates them, applies `Gap`, and supports `Dynamic`, `MinWidth`, and the
  `Breakpoint` stack behavior. This is not hard-coded application layout.
- The example has no split-ratio model or resize interaction. It hard-codes
  `{Width: 34, Dynamic: true, MinWidth: 12}` for both panes and relies on
  `Frame`'s preferred/minimum allocation. A user cannot move a divider or
  persist a 30/70 (or similar) ratio.
- Focus traversal for the flat example is already framework behavior:
  `Frame.HandleKey` handles Tab/Shift-Tab and `Frame.applyFocus` calls the
  child's `Focusable` methods. The custom `scrollPane` in
  `examples/split/split/split.go` only adds the `▶` rendering cue; that cue
  could be a reusable framework decoration or default focus style.
- Nested splits are structurally possible because `Box.Child` is a
  `loom.Widget`, but there is no dedicated split tree or documented nested
  focus contract. In particular, `Box.HandleMouse` returns false, so mouse
  routing stops at a box instead of reliably reaching a nested `Frame`.
  Outer `Frame` focus also only recognizes children implementing `Focusable`;
  `Frame` itself does not currently expose that interface.
- Divider rendering, hit-testing, keyboard resizing, ratio persistence, and
  minimum-size redistribution are absent from the framework. Implementing
  these in the example would duplicate layout policy.

## Proposed Loom Features

Prefer a composable split widget over making `Frame` carry all split-specific
state. Rough API sketch:

```go
type Split struct {
	Orientation Orientation // Horizontal or Vertical
	Gap         int
	Ratio       float64     // 0..1; optional preferred ratio
	MinFirst    int
	MinSecond   int
	Divider     DividerStyle
	First       Widget
	Second      Widget
}

func NewSplit(first, second Widget) *Split
func (s *Split) SetRatio(r float64)
func (s *Split) Ratio() float64
func (s *Split) HandleKey(KeyEvent) bool
func (s *Split) HandleMouse(MouseEvent) bool
```

Add a common focus contract so split trees can participate in one traversal:

```go
type Focusable interface {
	Focused() bool
	SetFocus(bool)
}

type FocusContainer interface {
	Focusable
	FocusNext() bool
	FocusPrevious() bool
}
```

`Frame` and `Split` should implement `FocusContainer`; `Box.HandleMouse` should
forward translated events to `Child` when the child supports mouse handling.
The framework should provide a default focus marker/style so examples do not
need the `scrollPane` wrapper solely to draw `▶`.

## Opportunity

Refactor `examples/split/split/split.go` to compose two `loom.View`s with
`loom.NewSplit`, expose divider controls in the status line, and use nested
`Split` values as a regression/demo for recursive layout and focus traversal.
Keep `Frame` as the chrome/action wrapper and retain its existing flat-box API
for compatibility.
