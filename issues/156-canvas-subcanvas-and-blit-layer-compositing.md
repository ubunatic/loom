# 156 — Canvas SubCanvas and Blit Layer Compositing

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [issues/155-built-in-modal-and-dialog-overlay-primitive.md](155-built-in-modal-and-dialog-overlay-primitive.md), [loom-games](../loom-games)

---

## 1. Problem & Motivation

Rendering modular, decoupled TUI components often requires sub-views to draw relative to their own local coordinate space `(0, 0)` rather than canvas-absolute offsets. In complex applications and games with separate layers (e.g. background arena, HUD, status overlays), drawing everything into a single flat `Canvas` requires manually carrying offset coordinates through all render functions or risking coordinate bleed.

## 2. Technical Specification / Findings

- Add `SubCanvas` and `Blit` support to `Canvas`:
  ```go
  // SubCanvas creates an isolated canvas buffer representing region r.
  func (c *Canvas) SubCanvas(r Rect) *Canvas

  // Blit copies cells from src into dst canvas starting at (dstX, dstY).
  // Optionally supports transparency for unwritten/blank cells.
  func (c *Canvas) Blit(src *Canvas, dstX, dstY int)
  ```
- Simplifies nested view composition and overlay rendering.

## 3. Implementation & Verification Plan

### Goal
Implement lightweight sub-canvas creation and blitting onto parent canvases.

### Acceptance Criteria
- [ ] `SubCanvas` creates a canvas with local coordinate origin `(0, 0)`.
- [ ] `Blit` copies cells with full style and wide-character continuation cell integrity.
- [ ] Comprehensive unit tests in `canvas_test.go`.
