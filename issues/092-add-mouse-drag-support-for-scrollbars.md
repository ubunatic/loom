# 092 — Add mouse-drag support for scrollbars

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Enhancement
**Related**: None

---

## Problem & Motivation

Scrollbars need direct manipulation so users can quickly navigate long content
with a mouse instead of relying only on wheel events, keys, or paging actions.

## Scope

- Make a scrollbar thumb draggable with the primary mouse button.
- Convert pointer movement into proportional scroll-offset updates for both
  vertical and horizontal scrollbars where supported.
- Keep the thumb attached to the pointer throughout the drag, including when
  the pointer leaves the thumb's original cell, and end the drag on release or
  cancellation.
- Define sensible behavior for track clicks, minimum thumb sizes, content that
  does not overflow, viewport edges, and disabled/read-only scrollbars.
- Preserve existing wheel, keyboard, paging, and scrollbar activation behavior.

## Goal

/goal: Let users smoothly drag scrollbar thumbs with the mouse to control
scroll position, with consistent vertical/horizontal behavior and no regression
to existing scrolling interactions.

## Verification

Add focused coverage for drag start/move/release/cancel, pointer-to-offset
mapping, bounds and non-overflow cases, both orientations, and interaction with
existing wheel and keyboard scrolling.
