# 244 — Gallery All tab: focused Grid cell hides the astra background

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: 233 (astra cadence), 243 (cell backgrounds)

---

/goal With astra active (F8), the focused All-tab cell shows the astra decoration over its focus BG,
fixed in the library without hacks; or stop and report when blocked on a user decision.

## 1. Problem & Motivation
User report: in `loom widgets` → All with the astra background, the selected/active cell shows no astra
decoration.

## 2. Technical Specification / Findings
- `grid.go:136` fills the focused cell with `c.Fill` (foreground cells), which claims them, so
  `PaintDecoration` skips them. Per docs/AnimatedBackgrounds.md, a surface (`PaintSurface`) keeps its
  color and still lets decoration through.
- Likely shares the fix with 243; do 243 first or together. No gallery-side workaround.

## 3. Implementation & Verification Plan
- Developer: not luna (user choice).
- Test: Grid with astra decoration; free cells of the focused cell get decoration and keep the focus BG.
- `make test-q1`, `make install`, PTY check with F8 astra.
