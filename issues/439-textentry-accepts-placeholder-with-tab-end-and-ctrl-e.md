# 439 — TextEntry accepts placeholder with Tab End and Ctrl+E

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 422, 438; https://codeberg.org/harnez/taia/issues/110

---

## 1. Problem & Motivation

Owner requests the TextEntry widget itself accept its visible placeholder using Tab, End or Ctrl+E only when actual input is empty and the cursor is at the beginning. Taia #110 planned similar behavior but listed Home; the current owner request specifies End, so do not add Home. Existing issue 422 delivered SetPlaceholder/Placeholder/AcceptPlaceholder.

## 2. Technical Specification / Findings

Reuse AcceptPlaceholder: accepting inserts the full placeholder without submitting, caret at end, one undo step and one OnChange. Never treat placeholder as actual Value before acceptance. Only consume acceptance keys for empty actual input, nonempty placeholder and caret at start. Otherwise preserve existing key behavior, especially Tab/Shift+Tab focus routing, Home/End navigation, Ctrl+E behavior, Ctrl+C and Ctrl+U. Document Tab consumption as an intentional widget action consistent with AGENTS.md. No caller-specific workaround or Taia edit.

## 3. Implementation & Verification Plan

Read-only discovery then exact developer write-set/plan approval. Add reproduction-first library tests, negative key cases, Unicode/multiline placeholders, undo/redo/notifications and host/gallery routing. Verify real `loom widgets --show TextEntry` via PTY and installed binary; preserve existing tests. Update canonical spec/docs and upgrade notes if semantic compatibility changes. Run quota checks and install. Developer terra:med, host reviewer, no external implementation.

## Goal

/goal Make loom.TextEntry accept its visible placeholder with Tab, End and Ctrl+E under the empty-input/start-caret guard, preserving all other routing/editing semantics and proving one-step acceptance without submit in unit and live terminal tests; stop and report when blocked on a user decision or denied permission. On a loop, hard architectural or OS blocker, or no progress, clear the goal and record a failure report explaining why.
