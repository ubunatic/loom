# 076 — Add first-class Split widget with ratio control dividers and nested focus traversal

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `docs/studies/2026-09-20-feature-gap-analysis-split.md`, `frame.go`, `examples/split/split/split.go`

---

## 1. Problem & Motivation

Split layout in Loom currently relies on flat `loom.Box` definitions in `Frame`. Applications have no native way to declare proportional splits (e.g. 30/70 ratio), interactive divider hit-testing / keyboard adjustment, or recursive nested split trees with unified Tab focus and mouse event routing.

## 2. Desired Behavior & Goal

`/goal`: Provide a dedicated, composable `loom.Split` widget supporting horizontal/vertical orientation, configurable ratios, minimum constraints, divider styling, and nested focus/mouse traversal.

- Support `NewSplit(first, second Widget)` with `Orientation`, `Ratio` (0.0–1.0), and `MinFirst` / `MinSecond`.
- Implement `Focusable` and `FocusContainer` contracts so nested split trees participate seamlessly in Tab/Shift-Tab focus cycles.
- Forward mouse events accurately to nested split children.
- Update `examples/split` to consume the new `Split` widget.

## 3. Implementation Plan

1. Implement `Split` struct and constructor in `split.go` (or `widget_split.go`).
2. Implement `Draw`, `HandleKey`, `HandleMouse`, and focus forwarding.
3. Add unit tests for ratio calculation, boundary constraints, and nested focus traversal.
