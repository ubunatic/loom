# 040 — Treemap theming: theme 1 (current) vs. theme 2 (quad/halfblock sub-cell rendering)

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Enhancement
**Category**: Feature
**Related**: `graph/treemap.go` (`RenderTreemap`, `TreemapOptions`, `drawTreemapBox`, `layoutTreemap`), `graph/treemap_aggregate.go` (`AggregateTreemap`, `TreemapSegment`), `examples/treemap/main.go`, issue 014 (configurable graph colors and glyph presentation — related but general, not treemap-specific)

---

## 1. Problem & Motivation

`graph.RenderTreemap` currently renders each box one way only: a
box-drawing border (`┌─┐│└─┘`) when the box is at least 3 columns by 2 rows
and `!opts.NoBorder`, or a solid glyph/ANSI-background fill when it isn't.
Requested (2026-09-14, mid-session on the treemap work): add a second,
selectable visual theme.

- **Theme 1** = current rendering, unchanged, and the default (so existing
  callers/tests keep working with no behavior change).
- **Theme 2** = a denser, sub-cell rendering style using Unicode quadrant
  block characters (`▘ ▝ ▖ ▗ ▚ ▞ ▙ ▟ ▛ ▜ █`, U+2596–U+259F plus `█`) and
  half-blocks (`▀ ▄ ▌ ▐`) colored via ANSI to approximate each node's
  background color at half- or quarter-cell resolution, instead of
  whole-cell box-drawing borders. The intent (per the request) is smoother,
  more "filled-in" looking box edges — closer to how terminal image
  renderers (e.g. `chafa`) use half-blocks for 2x vertical resolution.
- In theme 2, any cell outside every box (gutters/padding, if the layout
  ever has any — today's `layoutTreemap` is gap-free, so this may only
  matter if a future change introduces spacing) must stay the standard
  terminal background; only cells that belong to a box get colored via
  quad/half-block glyphs.

## 2. Scope

**In scope:**
- A way to select the theme, e.g. a `Theme` (or similarly named) field on
  `TreemapOptions`, with theme 1 (current behavior) as the zero value so
  every existing `TreemapOptions{}` literal and test is unaffected.
- Theme 2's rendering logic: deciding what "quad/halfblock edges that match
  the node BG" concretely means at `layoutTreemap`'s existing whole-cell
  rect resolution — e.g. rendering interior cells as solid colored blocks
  and only using quadrant/half-block glyphs at a box's boundary against
  neighboring boxes or the outer background, to fake sub-cell edge
  precision without changing the underlying integer-cell layout algorithm.
- Whether/how labels and the numbered-marker/legend fallback (see
  `RenderTreemap`'s doc comment) apply under theme 2 — presumably
  unchanged, but confirm text still reads legibly against a full-bleed
  colored background with no border cell to reserve for it.
- Tests mirroring the existing treemap coverage (dimensions, ANSI
  reset-on-change, never-panics, legend/marker interaction) for theme 2.

**Out of scope (unless discovered to be required):**
- Changing `layoutTreemap`'s rect-allocation algorithm itself (theme 2 is a
  rendering concern layered on the existing integer-cell rects, not a
  layout-precision upgrade).
- A general theming system beyond these two treemap-specific themes.

## 3. Acceptance Criteria

- `TreemapOptions` gains a theme selector; theme 1 is the default and its
  rendered output is byte-for-byte unchanged from before this ticket for
  every existing test case.
- Theme 2 renders using quadrant/half-block glyphs colored to approximate
  each node's `BackgroundANSI`, with cells outside all boxes left at the
  plain terminal background (no ANSI wrap).
- `go test ./...`, `go vet ./...`, and `gofmt -l` stay clean.
- `examples/treemap` demonstrates both themes (e.g. a `--theme` flag) for a
  quick before/after comparison against a live process tree.

## 4. Verification Guidance

- Unit tests per theme covering: exact grid dimensions preserved, ANSI
  code correctness/reset behavior, ANSI cells outside boxes never receive a
  background code, and the marker/legend fallback still functions.
- Manual/visual check via `go run ./examples/treemap --theme 2` (or
  equivalent) against real snapshot data, comparing against theme 1's
  output for the same segments.

## 5. Notes / Open Questions

- Exact quadrant-glyph selection strategy for theme 2 is not yet designed;
  this ticket does not mandate a specific edge-rendering algorithm, only
  the two-theme contract and behavioral guarantees above. Resolve the
  concrete approach during implementation and record it here or in the doc
  comment.
- Confirm with the requester whether theme 2 should also drop labels/
  numbered markers in very small boxes differently than theme 1, given the
  denser visual style, or keep the existing fallback behavior verbatim.
