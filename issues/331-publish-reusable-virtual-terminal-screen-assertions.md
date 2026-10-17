# 331 — Publish reusable virtual terminal screen assertions

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Enhancement
**Related**:

---

## 1. Problem & Motivation
Downstream libraries need reusable screen assertions instead of matching ANSI byte sequences.

## 2. Technical Specification / Findings
Promote internal/ptytest VT and styles to public loomtest, retain internal aliases, add alternate-screen state and deterministic tests.

## 3. Implementation & Verification Plan
Cover cells, cursor, SGR, split writes and alternate-screen restoration, and convert three ANSI tests to screen assertions.
