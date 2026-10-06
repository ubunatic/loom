# 276 — Container-owned coordinate translation and mouse event clipping

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Architecture
**Related**: [273](273-audit-richtextedit-workarounds-and-plan-loom-library-fixes.md)

---

## 1. Problem & Motivation
In `gallery/richtextedit.go:52-70`, `richTextEditDemo` calculates `w.viewH`, manually subtracts `mouse.Y -= w.viewH` (`mouselocal.Y`), and manually routes out-of-bounds mouse events (`mouse.Y < w.viewH || mouse.Y >= w.viewH+w.editRect.H`) to `w.edit`. Having parent widgets manually translate coordinates and dispatch out-of-bounds mouse events is error-prone.

## 2. Technical Specification / Findings
- Standardize child-relative coordinate translation and clipping in Loom layout containers (`VStack`, `Split`, `Viewport`, `Grid`).
- Containers should automatically translate `MouseEvent.X/Y` to child bounds and handle clipping or event bubbling without requiring custom math in host widgets or demos.

## 3. Implementation & Verification Plan
/goal Standardize mouse event coordinate translation and clipping across composite containers so host widgets do not need manual coordinate math.

> Update (2026-10-06): the gallery RTE demo no longer has the read-only preview row, so its `viewH` offset and out-of-bounds routing are gone. The ticket still applies to containers in general; find a current example before starting.
