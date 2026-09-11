# 029 — Declarative Centered Layout and Viewport Alignment Primitives

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [docs/HarnezSplashTarget.md](../docs/HarnezSplashTarget.md), [issues/028-add-reusable-measurement-and-dynamic-box-layout-primitives.md](028-add-reusable-measurement-and-dynamic-box-layout-primitives.md)

---

## 1. Problem & Motivation

The splash screen target requires a content block (title, bar, status text, provider indicators, skip hint) to be centered both horizontally and vertically inside arbitrary terminal dimensions without manual coordinate hardcoding. Loom currently has row/box layout primitives, but needs clear viewport centering and alignment primitives in declarative specs and canvas placement.

## 2. Technical Specification

1. **Alignment Primitives**:
   - Support horizontal alignment (`center`, `left`, `right`) and vertical alignment (`center`, `top`, `bottom`) within a layout container.
   - Dynamic offset calculation based on terminal width $W$ and height $H$ versus content dimensions $(w, h)$.
2. **Padding & Clipping**:
   - Compute top/left margin offsets: `(W - w)/2` and `(H - h)/2`.
   - Prevent overflow/wrap when $W < w$ or $H < h$ through safe clipping / min-dimension bounds.
3. **Declarative Spec Support**:
   - Allow containers or views in YAML specs to declare `align: center` / `valign: center` or flex-centered child arrangement.

## 3. Implementation & Verification Plan

- [ ] Add vertical and horizontal centering calculation helpers in layout/canvas package.
- [ ] Support centering in the declarative layout engine / view renderer.
- [ ] Unit tests verifying centered coordinates across odd/even viewport widths and heights.
- [ ] Zero-allocation / safe boundary checks for small viewports.
