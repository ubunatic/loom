# 249 — Grid: support inner and full cell borders

**Status**: Closed — resolved in 3d07274
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 238 (Gallery All grid borders and content-fitted row heights)

---

## 1. Problem & Motivation
`Grid` has no border option, so callers cannot visually separate cells. Add an opt-in border mode with `inner` borders between cells and `full` borders that also outline the grid perimeter.

## 2. Technical Specification / Findings
Implement this on the reusable `Grid` widget so the gallery and other callers can use it without custom drawing. Check `spec/` for Grid defaults or option values and extend the YAML source of truth and generated/spec-backed code where appropriate; avoid duplicating spec-owned values in Go.

## 3. Implementation & Verification Plan
/goal `Grid` supports disabled, inner, and full border modes with rendering and layout tests, and any spec-owned values are defined in YAML; or stop and report if blocked on a user decision about border behavior.

Verify border placement for each mode, including narrow and incomplete rows, and confirm child drawing and mouse hit testing still use the intended cell bounds.
