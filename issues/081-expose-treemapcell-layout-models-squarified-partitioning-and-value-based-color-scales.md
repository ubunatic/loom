# 081 — Expose TreemapCell layout models squarified partitioning and value-based color scales

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `docs/studies/2026-09-20-feature-gap-analysis-treemap.md`, `graph/treemap.go`, `examples/treemap/treemap/treemap.go`

---

## 1. Problem & Motivation

Loom's treemap renderer directly writes rune grids and only supports an internal slice-and-dice layout and index-cycling palette. Applications cannot obtain pre-rendered cell geometry (for hit-testing, tooltips, or custom decorations), choose squarified aspect-ratio layouts, or apply continuous value-based color scales (heat maps).

## 2. Desired Behavior & Goal

`/goal`: Expose pre-render `TreemapCell` geometry structures, support squarified layout algorithms alongside slice-and-dice, and provide value-based `ColorScale` options in `graph`.

- Expose `LayoutTreemap` returning `[]TreemapCell` with exact `Rect` coordinates and segment metadata.
- Support `TreemapLayoutSquarified` to minimize aspect ratio distortion in cells.
- Add `ColorScale func(value, min, max float64) CellStyle` for continuous color gradients.

## 3. Implementation Plan

1. Define `TreemapCell` and layout strategy interfaces in `graph/treemap.go`.
2. Implement squarified layout partitioning algorithm.
3. Add color scale mapping helpers and unit tests for geometry and bounds preservation.

---

## Resolved Design (host decisions, binding for the sprint)

- Reuse the existing node/segment types of `graph/treemap.go`; do not rename public API.
- `TreemapLayout` enum: `TreemapLayoutSliceDice` (default, current behavior, output byte-identical) and
  `TreemapLayoutSquarified`.
- `LayoutTreemap(<existing nodes>, w, h int, layout TreemapLayout) []TreemapCell`; `TreemapCell` carries
  an integer `Rect`, the segment metadata (label/id, value, depth or parent) needed for hit-testing.
  Deterministic order, stable for equal values.
- Invariants asserted by property tests (random inputs, fixed seed): cells with positive value tile the
  w x h area exactly (no overlap, no gap, total area == w*h), all rects integral and inside bounds, zero and
  negative values yield no cell, empty input and w or h <= 0 yield no cell and no panic.
- Squarified: on a fixed set of sample distributions, mean worst aspect ratio is not worse than slice-dice
  (assert in a test), and the algorithm follows Bruls-Huizing-van Wijk row logic.
- `ColorScale func(value, min, max float64) CellStyle` plus helpers (`LinearColorScale(from, to)`, a heat
  scale). Clamp outside [min,max]; `min==max` returns the midpoint; NaN and Inf never panic.
- The existing renderer keeps using its current path unless a golden test shows byte-identical output when
  routed through `LayoutTreemap`; keep goldens unchanged.

## Milestones (lean sprint, developer: luna:low)

Host reviews only diffs, test output and evidence frames; this ticket is the only channel. Root-package tests
needing `/dev/tty` fail before this work; ignore them. Commit each milestone (message ends '(issue 081 MX)'),
stage only your files, never docs/README.md, no stray binaries in the repo root; if `.git/index.lock` blocks
the commit say so and the host commits. Evidence: frames produced by code, gated on env `LOOM_EVIDENCE=1`,
written to repo-root `docs/progress/081/` (find root by walking up to `go.mod`); run only this ticket's
evidence test and check `git status` before finishing. View frames ANSI-stripped before you finish; labels on
their own rows, nothing overlapping. gofmt.

### M1 - TreemapCell and LayoutTreemap (slice-dice) with invariants
- Types, function, property tests, golden proof that slice-dice output is unchanged.
- Evidence: `M1-slicedice-s1.ansi`, `M1-slicedice-s2.ansi`, `M1-slicedice-s3.ansi` (three sizes, cell borders and labels).

### M2 - Squarified layout
- Algorithm, invariants, aspect-ratio comparison test.
- Evidence: `M2-squarified-s1.ansi` .. `s3.ansi` (same data and sizes as M1, for comparison).

### M3 - ColorScale
- Scale type, helpers, edge-case tests; a demo render coloring cells by value.
- Evidence: `M3-colorscale.ansi` (heat gradient legend plus a colored treemap at one size).
