# 142 — Document theme selection and listing in the Go library

**Status**: Closed — Implemented in 4ae804b (M1): added ThemeNames and ThemeExists helpers and updated docs/Themes.md with Go API examples
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Documentation
**Related**:

---

## 1. Problem & Motivation

Go app authors need a clear, discoverable way to select a built-in Loom theme and get the supported theme names. Without this guidance, callers may guess names; `Theme(name)` silently returns `plain` for unknown names.

## 2. Technical Specification / Findings

Document the Go library API for applying themes and listing built-in names, using the current spec-backed API as the source of truth. If the existing API is awkward or insufficient for app authors, add a small theme-listing API. Explain how applications should validate user-provided names instead of silently accepting an unknown name.

## 3. Implementation & Verification Plan

- **/goal**: Make it straightforward for Go app authors to discover, list, validate, and apply built-in Loom themes, or stop and report when blocked on a user decision or denied permission.
