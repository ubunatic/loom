# 155 — Built-in Modal and Dialog Overlay Primitive

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [issues/156-canvas-subcanvas-and-blit-layer-compositing.md](156-canvas-subcanvas-and-blit-layer-compositing.md), [loom-games](../loom-games)

---

## 1. Problem & Motivation

Applications frequently require centered or positioned modal dialog overlays (e.g. Game Over dialogs, confirmation prompts, alert boxes) rendered on top of existing widget content. Currently, developers must manually calculate centered `Rect` coordinates, clear background cells with `c.Fill(rect, blank)`, draw borders, and position text line-by-line.

While `loom.Popup` exists, it is specialized for command `:help` or single widget wrappers rather than lightweight modal dialog rendering on `Canvas`.

## 2. Technical Specification / Findings

- Add a high-level `Dialog` or `Overlay` drawing helper on `Canvas` or as a reusable component:
  ```go
  type Dialog struct {
      Title       string
      Rect        loom.Rect // or auto-center within parent
      Style       loom.Style
      BorderStyle loom.BoxBorderStyle
      Fill        bool
  }

  func (c *Canvas) DrawDialog(d Dialog, renderContent func(inner loom.Rect))
  ```
- Handles canvas bounds checking, background cell filling/clearing, border rendering with title, and centered alignment calculations.

## 3. Implementation & Verification Plan

### Goal
Provide a declarative dialog and modal overlay rendering API for `loom.Canvas`.

### Acceptance Criteria
- [ ] Support centered and explicitly placed dialog overlays.
- [ ] Auto-clearing background cells beneath the dialog to prevent background text bleed.
- [ ] Border styling and title truncation using standard `loom.DrawBox` rules.
- [ ] Unit tests and canvas visual parity tests in `canvas_test.go`.

## Sprint goal (roadmap 180)
/goal Ship a `Dialog` on top of `Popup` (title, body, button row, centered or placed, background clear) with canvas tests and an example for human review; stop and report when button or placement style needs a user decision.
