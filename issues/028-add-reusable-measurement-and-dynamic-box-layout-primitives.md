# 028 — Add reusable measurement and dynamic box layout primitives

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Feature
**Related**: [Roadmap](../docs/Roadmap.md), [Spec](../docs/Spec.md), [025](025-integrate-graph-renderers-into-declarative-monitor.md), [027](027-introduce-first-spec-driven-collector-prototype.md), [Harnez uix](../../harnez/internal/uix/uix.go), [Voxi monitor rendering](../../voxi/internal/monitor/render.go)
**Roadmap stage**: 6–7 UI foundation; prerequisite for richer dynamic monitor panels

---

## 1. Problem & Motivation

Loom's frame width can follow the terminal, but its boxes still use fixed
declared width and height. The current layout clips preferred rectangles and
switches between a horizontal row and a vertical stack; it does not measure
content or distribute available space according to minimum, preferred, and
maximum sizes. The monitor therefore needs application-specific dimensions and
cannot naturally grow or shrink its panels as content changes.

Loom should gain reusable measurement and layout primitives, analogous to
`loom/graph`, before more monitor content is added. The first extraction should
be evidence-led and small rather than a wholesale port of either sibling UI.

## 2. Technical Specification / Findings

Read-only inspection found two useful precedents:

- Harnez `internal/uix/uix.go` provides a compact `Box` model with
  `MinWidth`, `PrefWidth`, `MaxWidth`, `Stretch`, `Priority`, and `Order`;
  `Layout` filters disabled boxes, wraps them into rows, and distributes extra
  width across stretchable boxes. Its `renderBox` is ASCII and rune-count
  based, so its planner is reusable but its renderer is not sufficient for
  Loom's ANSI/cell contract.
- Voxi `internal/monitor/render.go` provides terminal-cell measurement through
  `RuneDisplayWidth` and `StringDisplayWidth`, ignores ANSI sequences, and
  preserves ANSI styling while truncating visible content with
  `TruncateLineANSI`. `RenderBoxLines` derives content width from the outer box,
  pads measured lines, and keeps borders intact. Its monitor-specific box
  assembly and color/domain code should remain outside Loom.

The Loom extraction must preserve combining/wide-rune and ANSI-aware geometry,
minimum border/padding sizes, and deterministic rendering. It must not make
collectors, widgets, or application labels responsible for coordinates.

## 3. Implementation & Verification Plan

1. Add a Loom measurement package or bounded public helpers with explicit cell
   width semantics, ANSI handling, and tests derived from the existing canvas
   contract. Record provenance and intentional differences from Harnez/Voxi.
2. Add a pure layout planner with min/preferred/max size, fixed and flexible
   sizing, visibility, gaps, wrapping/stacking, and deterministic remainder
   distribution. Keep it independent of terminal I/O and data collection.
3. Adapt `Frame`/`Box` to consume planned rectangles, including content-aware
   height for rows and plain footers, while retaining compatibility with the
   current fixed declarations.
4. Extend the validated YAML vocabulary only for consumed size policies;
   preserve fixed-width documents and reject invalid or undersized policies.
5. Migrate the monitor example to exercise dynamic sizing at 80-, 64-, 40-,
   and tiny-column widths, with changing and long content.

### Acceptance criteria

- [ ] Reusable Loom measurement helpers report terminal cell width correctly
  for plain, ANSI-styled, combining, and wide-rune text.
- [ ] A pure planner supports fixed plus flexible boxes with minimum and
  maximum bounds, hidden boxes, gaps, stacking/wrapping, and stable remainder
  allocation.
- [ ] Box borders, padding, rows, footers, truncation, and ANSI styling remain
  geometrically valid at wide, narrow, tiny, and content-changing sizes.
- [ ] Existing fixed-layout YAML documents remain compatible; new sizing
  fields are schema-validated and consumed without duplicated defaults.
- [ ] The monitor demonstrates dynamic box sizing without collector/rendering
  coupling, and independent geometry tests cover the resulting layouts.
- [ ] Copied/adapted code has source commit/path provenance and Loom-specific
  differences are documented; no sibling repository is modified.

## Verification

Run measurement unit tests, planner table tests, schema negative controls,
wide/slim/tiny geometry replay, `GOWORK=off go vet ./...`,
`GOWORK=off go test ./...`, race tests, and the monitor PTY smoke test. Compare
visible cell positions independently of ANSI replay. Include long labels,
combining marks, wide glyphs, hidden/restored boxes, and changing footer/row
content.

## Scope limits

No general reactive layout engine, animation, terminal theme discovery,
collector implementation, external eventing, sibling mutation, or wholesale
copy of Harnez/Voxi monitor code. This ticket establishes measurement and
layout primitives; source parsing and declarative collection remain separate
issues.
