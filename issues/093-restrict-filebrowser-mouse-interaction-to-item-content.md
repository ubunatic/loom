# 093 — Restrict filebrowser mouse interaction to item content

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Bug
**Related**: None

---

## Problem & Motivation

In the filebrowser, clicking empty space in an item's row currently can select
or activate that item. This makes hit behavior surprising, especially in wide
rows where the pointer is far from the visible item.

## Scope

Restrict mouse interaction to the item's actual interactive region: item text
and any visible frame or decoration belonging to that item. Clicking whitespace
elsewhere in the row must not select, activate, or change the current item.

Preserve normal keyboard navigation, row layout, selection state, and intended
interactions for item decorations. Define behavior for truncated text, icons,
multi-column metadata, and rows with no visible interactive content.

## Goal

/goal: Make filebrowser mouse hit-testing precise so only an item's text, frame,
or decoration responds to clicks, while empty row space is inert and cannot
change selection or activation state.

## Verification

Add focused coverage for clicks on text, whitespace before/after text, frames,
icons/decorations, truncated labels, and neighboring rows. Verify selection and
activation remain correct for mouse and keyboard interaction.
Extend the PTY tests to click item content and row whitespace and prove that
only the intended hit regions change selection or activation.
