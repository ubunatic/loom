# 330 — Snapshot exported package APIs and document stability policy

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Enhancement
**Related**:

---

## 1. Problem & Motivation
Snapshot all importable non-internal packages so additions, removals and signature changes require explicit review.

## 2. Technical Specification / Findings
Discover public packages with go list, capture a deterministic exported API, and document stable versus experimental packages.

## 3. Implementation & Verification Plan
Verify baseline generation and rejection of changed or missing snapshots, then run the full suite.
