# 248 — Grid: navigate with Shift-Left and Shift-Right

**Status**: In Progress — extend shifted navigation to all four directions
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 242 (All tab child arrow-key handling)

---

## 1. Problem & Motivation
`Grid` sends key events to its focused child before handling grid navigation. A child that consumes the regular arrows can prevent users from moving focus between grid cells. Add Shift-Left/Right and Shift-Up/Down as grid navigation keys that work regardless of the focused child's handling of regular arrows.

## 2. Technical Specification / Findings
`Grid.ConsumeKey` in `grid.go` currently dispatches to the focused child first. Handle all four shifted arrows as grid-level navigation while preserving regular arrow dispatch to children.

## 3. Implementation & Verification Plan
/goal Shift-Left/Right and Shift-Up/Down move focus across grid cells even when the focused child consumes regular arrows, with tests; or stop and report if blocked on a user decision.

Verify movement and wraparound behavior in both axes with a child that consumes regular arrows, and confirm regular arrows continue to reach that child.
