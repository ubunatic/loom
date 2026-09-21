# 081 — Expose TreemapCell layout models squarified partitioning and value-based color scales

**Status**: Closed — delivered via lean sprint (M1-M5); luna:low M1-M4, sol:low M5
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

### M1-M3 Review (host)
Code committed by the host (index.lock blocked the developer). It builds and `go test ./graph` passes,
but the sprint is incomplete:
- No evidence frames exist at all (the developer found no harness). Write one: an evidence test in
  `graph/` (gated on `LOOM_EVIDENCE=1`, repo-root `docs/progress/081/`) that renders cells from
  `LayoutTreemap` into a grid with borders and labels and prints it, plus a small helper for it.
- The tests are thin: two hand-picked cases. The Resolved Design demands property tests, an
  aspect-ratio comparison and edge cases.
- No proof that the existing renderer output is unchanged (golden, byte-identical).

### M4 - Pre-Work / Required Refinements
1. Property test (seeded `rand`, at least 200 random inputs, sizes 1..80 x 1..40, 1..12 segments, some
   zero/negative values): for BOTH layouts cells tile the area exactly (build a w x h occupancy grid: every
   grid cell covered exactly once), all rects inside bounds, no cell for value <= 0, deterministic
   (same input gives identical output), no panic on empty input or w/h <= 0.
2. Aspect test: over at least 5 fixed distributions, mean worst aspect ratio (terminal cell aspect 1:2,
   so use W against H*2) of squarified is <= slice-dice; report the numbers in the test log.
3. Color scale edge tests: min > max, min==max, +/-Inf, NaN, negative range, exact midpoint value, monotonic
   red channel for the heat scale.
4. Golden proof: rendering existing sample segments through `RenderTreemap` is byte-identical to the output
   on the previous commit (`git stash`-free way: commit the current output of the OLD code path as a
   golden string in the test before any routing change; do not route the renderer through LayoutTreemap
   unless it stays byte-identical).
5. Evidence as described above: `M1-slicedice-s1..s3.ansi`, `M2-squarified-s1..s3.ansi` (same data and 3
   sizes each), `M3-colorscale.ansi`. View them ANSI-stripped; labels on own rows; nothing overlapping.
6. Commit '(issue 081 M4)'; if index.lock blocks, stage your files and say so.

### M4 Review (host)
Committed by the host. Tests, golden, color-scale edge cases and evidence harness accepted. Blocking defect:
`layoutTreemapSquarified` just calls the old `layoutTreemap` and leaves the real algorithm in a dead
comment block. Aspect log: slice-dice 2.893 equals squarified 2.893; M2 frames equal M1 frames.
This misses the ticket goal, so the step moves up the escalation ladder (`codex:sol:low`).

### M5 - Real squarified layout (architecture given by the host)
1. Test first: on the five fixed distributions squarified mean worst aspect must be STRICTLY lower
   than slice-dice, and on at least one distribution the cell rects differ between the layouts.
2. Implement squarified in `layoutTreemapSquarified` (delete the dead comment block and the fallback):
   run Bruls-Huizing-van Wijk on FLOAT rectangles (areas proportional to values, sorted descending, rows
   laid along the shorter side, terminal aspect compensated by scaling the height by 2 for the ratio),
   then convert to integers by rounding cumulative edges (round the row start and row end coordinates
   along each axis, and the item start/end inside a row), so neighbours share exact edges and the
   tiling stays gap- and overlap-free. Guarantee at least one cell per positive value where the
   area allows; zero-size results are dropped, never overlapping.
3. All property tests, the golden and the color scale tests stay green; regenerate `M2-squarified-s1..s3.ansi`
   (they must visibly differ from the M1 frames).
4. Commit '(issue 081 M5)'; if index.lock blocks, stage your files and say so.
