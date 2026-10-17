# 332 — Split rich text editor and pane implementation by concern

**Status**: Closed — implemented in 323ee29, 204 declaration ASTs unchanged, full suite and API baseline passing
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Enhancement
**Related**:

---

## 1. Problem & Motivation
Large editor and pane files obscure concerns and make reviews difficult.

## 2. Technical Specification / Findings
Move declarations without changing bodies into concern-specific files for rendering, input, selection, editing, history, screen management and lifecycle.

## 3. Implementation & Verification Plan
Compare declarations before and after, verify exported API snapshot and run go test ./....
