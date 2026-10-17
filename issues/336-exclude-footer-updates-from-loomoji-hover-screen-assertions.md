# 336 — Exclude footer updates from Loomoji hover screen assertions

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**:

---

## 1. Problem & Motivation
The full Go 1.25 race suite reported TestLoomoji107HoverProbe failure: a footer-only color change at row 22 was classified as a grid highlight. All measured emoji grid offsets were zero. This test compares the whole terminal instead of the grid/padding area it is meant to assert.

## 2. Technical Specification / Findings
Bound the differential highlight detector by the last discovered emoji grid row. Retain top/padding cells so incorrect highlights outside the target remain detectable, but exclude unrelated footer status changes. Keep the coordinate and color assertions intact.

## 3. Implementation & Verification Plan
Add a unit fixture for footer-only changes, real grid changes mixed with footer changes, and unexpected padding highlights. Re-run full host tests with the race detector and the declared Go toolchain.
