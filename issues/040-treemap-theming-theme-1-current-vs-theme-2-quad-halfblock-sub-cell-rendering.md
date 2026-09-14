# 040 — Treemap theming: theme 1 (current) vs. theme 2 (quad/halfblock sub-cell rendering)

**Status**: Closed — shipped as `TreemapThemeNumbered` (thin seven-eighths edges + a corner number on every box), `--theme 1|2`, tests, verified. Superseded two earlier designs after live visual review; see §7.
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

## 6. Final Shipped Design (2026-09-14)

- `TreemapOptions` gained a `Theme TreemapTheme` field; `TreemapThemeClassic`
  (the original renderer, byte-for-byte unchanged) is the zero value, and
  `TreemapThemeNumbered` is the one surviving alternate style (see §7 for
  what was tried and dropped along the way).
- Every box renders full-bleed under `TreemapThemeNumbered` (no reserved
  border cell, so labels get the box's entire area -- the existing
  full-label/number/legend fallback logic is unchanged, `treemapInner`
  just always returns the full rect since `border` is always false under
  this theme). Each box's own topmost row and rightmost column render with
  a seven-eighths block glyph (`▇` top, `▉` right, U+2587/U+2589) whose
  foreground is this box's own color (its `BackgroundANSI` converted to a
  foreground code via `treemapBackgroundToForeground`) and whose thin
  unfilled sliver carries no background SGR at all -- the ambient terminal
  background shows straight through it, purely locally to each box's own
  rectangle, with no neighbor lookup.
- Every box (not just ones whose full label doesn't fit) gets a
  superscript corner marker, ranked by value across the whole treemap and
  stamped into its own top-right corner cell; a labeled box shows both its
  name and its number, but the legend row still only explains numbers for
  boxes that needed one in place of a label.
- Requires `opts.ANSI` and a non-empty `opts.BackgroundANSI`; without
  either there is no color to render with, so it falls back to exactly the
  same code path `TreemapThemeClassic` + `NoBorder` uses (verified
  byte-identical by test), and to `TreemapThemeClassic`'s
  only-unlabelled-boxes marker/legend behavior.
- `examples/treemap` has `--theme 1|2` (default 1); theme 2 requires
  `--ansi` and errors clearly if it's missing.
- Tests: `graph/treemap_test.go` covers dimensions, no classic border
  glyphs ever drawn, edge/corner glyph placement, no-background-on-glyph-
  cells, ANSI/no-palette and no-ANSI fallback byte-equality against
  classic, full-bleed no-inset labels, every-box corner numbering, legend
  skipping labeled boxes, never-showing-a-truncated-fragment (a real
  regression caught after shipping -- see §7), and never-panics across the
  same fuzz matrix as theme 1, plus `TestTreemapThemeClassicIsZeroValue`.
  `examples/treemap/main_test.go` covers `--theme` flag validation.
- Manually verified against a live process-tree snapshot via
  `go run ./examples/treemap --ansi --theme 2` compared side-by-side with
  `--theme 1`, `go build/vet/test ./...` all clean.

## 7. Design Iteration (what was tried and dropped)

This ticket's final shape is the result of three rounds of live visual
review against real process-tree data, not the first design that passed
its own tests -- worth recording since each round looked correct in
isolation and only broke down once actually viewed in a real terminal:

1. **Half-block boundary blending** (`TreemapThemeBlocks`, shipped
   briefly as `--theme 2`): each box's own right/bottom edge blended with
   its *specific* neighbor's color via a per-cell `owner`-grid lookup.
   Live review found the corners looked messy -- several different
   neighbor-blended cells piling up at a single junction. Reworked to
   drop the neighbor lookup entirely and use a fixed per-box geometric
   rule (right/bottom edge -> foreground-only, no neighbor color) instead
   -- simpler and better looking, but still using half-blocks.
2. **Universal corner numbering** (`TreemapThemeNumbered` / theme 3) and
   **a filled-corner variant** (`TreemapThemeNumberedFilled` / theme 4):
   requested together, both implemented with seven-eighths edges on the
   top row/right column and a superscript corner marker on every box,
   differing only in whether the corner digit's cell got a filled
   background (theme 4) or floated over the ambient background (theme 3).
   Live review: theme 3 looked good, theme 4 looked worse -- removed
   entirely (commit `5fc60d3`) rather than keep unwanted surface area.
3. **A real regression caught only by live rendering, not by the tests
   written alongside it**: `TreemapThemeNumbered`'s pre-pass never
   populated the *other* theme's interior-fallback-marker slice, so
   `drawTreemapBox` always fell through to placing label text regardless
   of fit -- undersized boxes rendered truncated, ellipsis-cut fragments
   ("fir…", "v………") instead of relying on the corner number already
   shown, violating this project's own "full label or a number, never a
   fragment" rule. Fixed in `d8ef5f2` by threading a `hideLabel` flag
   through, and while there, `truncateLabel` (the ellipsis-truncation
   helper) was found to be dead code for *every* theme -- the pre-pass
   already guaranteed the fallback marker was set whenever a label
   wouldn't fit -- and was deleted outright rather than left unreachable.
   A related bug in the same area: label/marker text landing on a box's
   own softened edge stamped a filled background over it, leaking a solid
   colored patch out of the otherwise thin, unstyled seam (`930558a`).

Takeaway for future theme/visual work on this renderer: a design that
passes `go test` and reads correctly in a doc comment can still look
wrong once rendered against real, messy data (long names, many small
boxes, deep nesting) -- budget for at least one live-render review pass
per visual change, not just before the first ship.
