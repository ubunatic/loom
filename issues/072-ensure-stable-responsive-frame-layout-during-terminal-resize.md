# 072 — Ensure stable responsive frame layout during terminal resize

**Status**: Open
**Priority**: P1 (High)
**Severity**: Moderate
**Category**: Bug
**Related**: [008](008-responsive-declared-box-layout.md), [071](071-reflow-stacked-dynamic-frames-within-narrow-terminal-heights.md), `frame.go`, `frame_layout_test.go`, filebrowser demo

## 1. Problem & Motivation

While horizontally resizing the filebrowser terminal, especially while shrinking
it, the responsive frame briefly renders unstable intermediate geometry: pane
boxes stop filling their available width and expose full-width background gaps
between rows. The UI becomes clean again once resizing stops, which points to a
resize-stream/render synchronization or intermediate-layout issue rather than a
persistently invalid final layout. This is a regression risk in the responsive
layout contract from #008, not an application-specific filebrowser concern.

The layout must remain bounded and visually coherent for every resize result,
including breakpoint crossings and narrow widths. Intermediate terminal sizes
must not leave stale rectangles, alternating uncovered rows, or panes positioned
using the previous width.

## 2. Desired Behavior

- Horizontal resize in both directions produces stable, deterministic frame
  rectangles at every rendered resize step, not only after the terminal settles.
- Dynamic panes reflow cleanly across the breakpoint and fill the width of the
  active stacked layout without uncovered row bands.
- Borders, titles, status rows, gaps, focus state, and background composition
  remain aligned after repeated shrink/grow cycles.
- Very narrow terminals remain bounded and safe; panes may be omitted only when
  their documented minimum geometry cannot fit.
- Existing fixed-height, wide two-column, and stacked-height behavior remains
  unchanged.

## 3. Implementation & Verification Plan

### M1 — Reproduce and isolate

1. Add a headless resize sequence covering wide → breakpoint → narrow and
   narrow → wide transitions using the filebrowser-style frame.
2. Capture frame rectangles and identify whether the defect is stale bounds,
   width allocation, breakpoint reflow, or paint clipping.

### M2 — Stabilize framework layout

1. Fix the generic frame/layout path so each resize computes bounded rectangles
   from the current dimensions only.
2. Preserve dynamic fill, minimum constraints, gaps, hidden boxes, and chrome
   allocation across breakpoint transitions.
3. Add regression coverage for repeated shrink/grow cycles and narrow widths.

### M3 — PTY and visual verification

1. Run the filebrowser PTY path through wide, shrinking, narrow, and growing
   terminal sizes.
2. Verify no uncovered row bands, clipping, overlap, stale pane placement, or
   terminal restoration regressions.
3. Run the repository's required test, vet, and geometry checks.

## 4. Acceptance Criteria

- Responsive frame rectangles stay within the current frame bounds after every
  resize step.
- The filebrowser remains visually stable while shrinking and expanding across
  its breakpoint.
- No stale-width gaps or alternating uncovered rows appear between pane rows
  during the resize stream or after it settles.
- Headless regression tests and the manual PTY resize check pass.
