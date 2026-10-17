# 333 — Align lint and package test targets with CI

**Status**: Closed — implemented in 1271e76; lint, short, full race and regular suites pass on Go 1.25.0; container runtime verification tracked in 335
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Enhancement
**Related**:

---

## 1. Problem & Motivation
Local and CI lint/test commands need a consistent vocabulary and a fast package tier.

## 2. Technical Specification / Findings
Add lint, test-pkg and test-short targets; skip PTY/media/integration in short mode and run lint in GitHub CI.

## 3. Implementation & Verification Plan
Run lint, focused race and short tests, full tests and make container-test before the final commit.
