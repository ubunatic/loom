# 071 — Reflow stacked dynamic frames within narrow terminal heights

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: `frame.go`, `frame_layout_test.go`, filebrowser demo

---

## 1. Problem & Motivation

When a responsive `Frame` crosses its breakpoint, dynamic boxes are stacked
vertically. In the filebrowser demo, both boxes retain an 18-row preferred
height, so a narrow terminal places the metadata box below the visible canvas.
The wide two-column layout is usable, but the narrow flow clips or hides the
second pane instead of reallocating the available height.

This is a framework layout issue: applications should be able to declare
preferred dynamic box sizes without manually implementing terminal-height
breakpoint logic.

## 2. Desired Behavior

- Stacked dynamic boxes remain within the available frame height whenever the
  terminal is too short for all preferred heights.
- Each visible box receives at least its valid minimum height when possible,
  with remaining rows distributed according to the existing layout policy.
- Box order, gap, borders, focus routing, and wide two-column behavior remain
  unchanged.
- Very small terminals continue to omit rectangles that cannot fit a complete
  box, without panics or out-of-bounds writes.

## 3. Implementation & Verification Plan

### M1 — Reproduce and locate the height-allocation defect

1. Add a headless layout regression using the filebrowser-style dynamic boxes,
   breakpoint, preferred heights, and a narrow canvas.
2. Confirm whether the defect is in stacked allocation, `HeightForWidth`, or
   fallback rectangle placement.

### M2 — Reflow stacked dynamic boxes

1. Make stacked dynamic allocation consume the available frame height rather
   than placing preferred-height boxes beyond the viewport.
2. Preserve minimum-height constraints, gaps, hidden-box handling, and existing
   wide-layout allocations.
3. Keep the fix in framework layout code; do not special-case the filebrowser.

### M3 — Compatibility verification

1. Cover wide, breakpoint, narrow, short, and tiny frame sizes, including
   rectangle bounds and border rendering.
2. Run `make test-q1`, `go vet ./...`, and the filebrowser PTY smoke path across
   a wide and narrow terminal.

## 4. Acceptance Criteria

- The filebrowser metadata pane remains visible or is validly omitted according
  to available height, rather than being silently placed below the canvas.
- Stacked dynamic frame rectangles stay within the requested frame bounds.
- Existing responsive frame tests and wide layouts pass unchanged.
- No application-specific workaround is required.
