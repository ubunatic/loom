# 147 — Detect unclosed boxes and prioritize box validation errors

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature / CLI Tooling
**Related**: [129](129-build-loom-cli-tool-for-tui-asset-validation-measurement-and-interactive-viewing.md), [137](137-feedback-cli-asset-tools-workflow-and-multi-box-ansi-validation.md)

---

## 1. Problem & Motivation

`loom check-box` reports alignment errors but does not identify unclosed boxes. A missing corner can cause cascading arrow/alignment diagnostics that obscure the root cause and create noisy output.

## 2. Technical Specification / Findings

Add a flag or verb to detect unclosed boxes. When a box is unclosed, report that structural error prominently and avoid or de-prioritize cascading diagnostics attributable to it. Keep unrelated validation errors visible.

## 3. Implementation & Verification Plan

- **/goal**: Let users detect unclosed boxes and distinguish root-cause structural errors from cascading alignment/arrow noise, or stop and report when blocked on a user decision or denied permission.
