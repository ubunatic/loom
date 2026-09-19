# 070 — Prevent text overflow in framework help modals

**Status**: Closed — Framework help modal text is now display-width truncated to its content bounds; added Unicode/narrow-width regression coverage.

---

Reserved placeholder ticket.
# 070 — Prevent text overflow in framework help modals

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: [069](069-render-help-as-a-root-level-modal-overlay.md), `popup.go`, `cmd.go`, filebrowser demo

---

## 1. Problem & Motivation

The filebrowser demo's root-level `:help` modal can render help text wider than
the popup or available terminal canvas. Long command descriptions visibly run
through the modal border or into adjacent cells instead of wrapping or being
clipped. This is a framework behavior problem: applications should not need to
pre-measure or manually truncate help content to use the standard modal.

Investigate the shared popup/text layout path first so the fix applies to other
framework-generated modals and narrow terminals without changing compositor
semantics.

## 2. Desired Behavior

- Help text remains inside the modal's content and border at all supported
  terminal widths.
- Long lines wrap or truncate according to the framework's established text
  layout policy, preserving display-width and rune correctness.
- Modal sizing remains centered and usable on narrow canvases; no panics or
  writes outside canvas bounds occur.
- Existing help input capture, dismissal, root-level ownership, and nested/split
  pane behavior from issue 069 remain unchanged.

## 3. Implementation & Verification Plan

### M1 — Reproduce and locate the shared layout defect

1. Add a headless regression reproducing the overflowing help line at a narrow
   canvas width.
2. Trace whether overflow originates in popup sizing, text wrapping, border
   placement, or canvas writes; keep the fix in the framework layer.

### M2 — Correct modal text layout

1. Implement width-safe help/modal layout using rune and display-width aware
   measurement.
2. Ensure borders, padding, and wrapped lines use the same effective content
   width and remain within canvas bounds.
3. Preserve existing behavior for short help text and injected help runners.

### M3 — Verify compatibility and demo behavior

1. Add regression coverage for narrow and wide canvases, long descriptions,
   Unicode text, dismissal, and split roots.
2. Run `make test-q1`, `go vet ./...`, and the filebrowser PTY smoke test with
   `:help`, resize, dismissal, and terminal restoration.

## 4. Acceptance Criteria

- The filebrowser demo no longer shows help text crossing its modal border at
  narrow or normal terminal sizes.
- Framework modal/help rendering is display-width safe and bounded by the
  target canvas.
- Automated regression tests cover the original overflow and pass with the
  existing suite.
- No changes are made to canvas/compositor layering semantics unless required
  by the confirmed root cause.
