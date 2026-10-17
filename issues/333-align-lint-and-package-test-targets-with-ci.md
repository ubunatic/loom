# 333 — Align lint and package test targets with CI

**Status**: Open
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
