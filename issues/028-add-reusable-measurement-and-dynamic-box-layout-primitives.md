<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 028 — Add reusable measurement and dynamic box layout primitives

**Status**: Closed — resolved in 5e8cac5
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Architecture
**Related**: [008](008-responsive-declared-box-layout.md), [010](010-geometry-and-visual-evidence-milestone-before-rich-content.md), [011](011-aligned-dashboard-rows-and-ansi-safe-truncation.md), [024](024-port-harnez-rograph-primitives-with-provenance.md), [Loom graph](../graph), [Loom frame](../frame.go), [Loom canvas](../canvas.go), [Loom truncation](../truncate.go)
**Depends on**: [010](010-geometry-and-visual-evidence-milestone-before-rich-content.md), [011](011-aligned-dashboard-rows-and-ansi-safe-truncation.md); build on [008](008-responsive-declared-box-layout.md).

---

## 1. Problem & Motivation

Loom currently has fixed `Box.Width`/`Box.Height` declarations. `Frame.Layout` only chooses bounded side-by-side placement or breakpoint stacking. `canvas.go`, `rows.go`, `truncate.go`, and `table.go` already perform related width calculations, but measurement and content-sizing policy is not a reusable library primitive. As more widgets and graph-backed rows are added, each consumer can drift on ANSI handling, Unicode cell width, padding, minimums, and shrink behavior.

Extract a small public Loom library feature, analogous to `loom/graph`, for visible-width/text measurement and content sizing first, then use it as the basis for dynamic box sizing/layout. Widgets should report preferred/minimum/max dimensions from visible content; parents should allocate rectangles deterministically within terminal bounds.

### Provenance and read-only findings

The request was informed by read-only sibling inspection; neither sibling is to be modified.

- Harnez revision `1dd5e83216865554e8a502cce0da061b98513442`: [internal/usage/watch.go](../../harnez/internal/usage/watch.go) contains `stripANSI`, `visLen`, `truncateVisible`, a wide measurement pass, and natural-width panel planning; [internal/uix/uix.go](../../harnez/internal/uix/uix.go) contains min/preferred/max widths, wrapping, gaps, and stretch allocation. Its helper counts runes after ANSI stripping, so it is not a complete Unicode cell-width oracle.
- Voxi revision `7a64d8ef6642ee52b5472c03453dbedaa013ffcf`: [internal/monitor/render.go](../../voxi/internal/monitor/render.go) contains `RuneDisplayWidth`, `StringDisplayWidth`, ANSI-preserving `TruncateLineANSI`, and bordered-box sizing; [internal/monitor/monitor_test.go](../../voxi/internal/monitor/monitor_test.go) covers ANSI, emoji, truncation, and exact line widths. These are behavioral references, not packages to import.
- Loom already has `RuneWidth`/`StringWidth`, ANSI-aware cluster truncation, graph width padding, and frame breakpoint tests. The implementation must consolidate or delegate these behaviors rather than create a second incompatible width policy.

## 2. Technical Specification / Findings

### Phase A: measurement and content sizing

- Add a stable exported package/API under a Loom library boundary (for example `measure` or `layout`; choose one and document it) exposing visible cell width, ANSI-aware fit/truncation/padding, and content-to-size measurement for one or more lines.
- Specify the terminal-text policy: ANSI controls consume no cells; combining marks are zero-width; supported wide glyphs consume two cells; truncation never splits a display cluster or leaves unterminated styling. Do not claim exhaustive terminal/Unicode conformance without an independent oracle.
- Define empty content, zero/negative constraints, borders/padding/chrome, minimum/maximum dimensions, and height-from-content semantics. Results are deterministic and non-negative.
- Keep the API independent of renderer, YAML spec, producer, and terminal driver so `loom/graph`, rows, tables, and text widgets can consume it.

### Phase B: dynamic box sizing/layout

- Replace the assumption that every box supplies fixed dimensions with layout inputs containing minimum, preferred/content, maximum, stretch/shrink policy, visibility, and order. Preserve an explicit fixed-size compatibility option.
- Allocate width/height, gaps, borders, padding, and frame chrome deterministically. Prefer content-sized boxes, shrink or wrap at declared floors, and make tiny-terminal behavior explicit; never produce negative, overlapping, or border-drifting rectangles.
- Generalize `Frame.Layout`/`HeightForWidth` only after Phase A can calculate the same result independently. Preserve declaration order, breakpoint stacking, visibility actions, graph exact-width guarantees, and existing `Stack`/`Grid` consumers unless a migration is documented.
- Keep coordinate arithmetic in library layout code, not the monitor example or data collectors.

## 3. Acceptance Criteria

- [x] A documented Loom library package provides reusable visible-cell measurement, ANSI-aware fit/truncation/padding, and content sizing; existing duplicate helpers are removed, delegated, or explicitly justified.
- [x] Tests cover ASCII, ANSI SGR/control sequences, combining marks, representative wide glyphs, graph glyphs, empty/short/long content, exact-width padding, truncation boundaries, and malformed/unterminated ANSI without panics.
- [x] Independent geometry checks verify measured and rendered lines agree at known terminal columns, including a negative control catching rune-count/byte-count drift.
- [x] Dynamic box layout allocates content/preferred widths and heights with min/max/fixed constraints, gaps, chrome, and padding; it wraps or stacks predictably and never emits invalid or overlapping rectangles.
- [x] Existing fixed boxes, breakpoint stacking, visibility, rows/table alignment, and `loom/graph` exact-width behavior remain compatible, with migration tests for current fixtures.
- [x] Public API and Unicode/ANSI policy are documented; provenance and adaptation notes identify the sibling evidence without copying internal sibling packages or inventing attribution.

## 4. Implementation & Verification Plan

1. Inventory Loom width/truncation/size call sites and record the API decision. Establish a known-column canary before changing behavior.
2. Extract/consolidate measurement and text-fit primitives against the existing Canvas/cell contract; add focused table tests and independent expected widths.
3. Add content-sizing contracts and adapters for representative Frame, Rows, Table, and graph consumers. Verify empty, tiny, wide, combining, wide-glyph, and styled content.
4. Implement dynamic box allocation on those contracts. Test preferred packing, min-floor shrink, max/stretch, gaps/chrome, breakpoint fallback, hidden boxes, and repeated resize in headless and PTY-compatible fixtures.
5. Run `gofmt` on changed Go files, `go vet ./...`, `go test ./...`, `go test -race ./...`, `make test`, and existing geometry/replay checks. Record dimensions, commands, and unattended visual limitations in delivery evidence.

## Audit — 2026-09-10

Issue 028 is complete in the bounded scope defined above. The `measure` package
owns cell-aware width, fit, padding, truncation, wrapping, insets, and validated
content sizing; root compatibility helpers delegate to it. The `layout` package
provides deterministic min/preferred/max allocation with gaps, hidden items,
shrink floors, capped stretch, unused capped space, and overflow checks.

`Stack.Measured` and `Box.Dynamic` consume these policies without changing
legacy fixed/equal-share behavior. Dynamic boxes can derive preferred dimensions
from rows, title/footer, border and padding; the monitor schema accepts either
fixed dimensions or an explicit dynamic declaration. Geometry, visibility,
resize, malformed ANSI, graph one-cell glyph, schema, race and replay evidence
passes.

The Unicode policy remains intentionally bounded: emoji ZWJ, variation-selector
presentation, flags, and terminal-specific ambiguous-width modes are not
guaranteed. Human visual confirmation is still not recorded; automated ANSI
replay is the accepted unattended evidence.

Fresh `GOWORK=off make test`, `GOWORK=off go test -race ./...` and
`GOWORK=off make watch-pty` pass, including existing fixed-layout regressions.
Continue from the shared policy instead of extracting a second width API.

### Sprint delivery — 2026-09-10

- `bd395b0`: renderer-independent exact cell fitting/padding; root truncation
  delegates to `measure`, and Table clipping is cluster/cell safe.
- `29e32d7`: constrained text sizing with cell-aware wrapping, insets, and
  validated min/preferred/max dimensions.
- `d145846`: deterministic renderer-independent allocation planner.
- `47fd3f3`: opt-in measured Stack allocation while preserving legacy behavior.
- `3690af1`, `8945149`: dynamic Box sizing/integration, schema declarations,
  overflow checks, and consistent measured height reporting.


## 5. Scope Limits

- This filing changes ticket/index files only; no Loom source code is authorized in this request.
- No CSS/flexbox/general UI engine, constraint solver, animation, terminal emulator, raster screenshot stack, or exhaustive Unicode conformance suite.
- No new collectors, producer adapters, graph renderer redesign, color/palette framework, or sibling-repository changes. Graph integration is limited to consuming the measurement contract and preserving graph width semantics.
- No silent breaking public API change: compatibility shims or a separately reviewed migration are required where exported behavior is affected.
