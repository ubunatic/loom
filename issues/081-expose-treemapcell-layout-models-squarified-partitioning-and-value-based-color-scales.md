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
