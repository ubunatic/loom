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

## Decision and M1 write authority

Searcher verified `AcceptPlaceholder()` already inserts the full value atomically with undo/notification support. Add a narrow normalized-key guard in TextEntry.ConsumeKey for exactly Tab, End and Ctrl+E, before normal navigation/host fallback: accept only when actual value is empty, placeholder nonempty, and caret/anchor at zero. Return consumed based on successful acceptance even if OnChange subsequently changes the value. Do not alter AcceptPlaceholder's public API, other keys, or add Home acceptance. Otherwise preserve the existing result, including ignored Tab/Shift+Tab. Do not implement Ctrl+E navigation as unrelated scope.

Approved complete write set: `text_entry.go`, `text_entry_test.go`, `gallery/coverage.go`, `gallery/coverage_test.go`, `cmd/loom/widgets_paste_pty_test.go`, `spec/widgets.yaml`, `docs/Widgets.md`, `docs/GalleryCoverage.md`, `docs/Upgrading.md`, this ticket and one bounded sprint study. Gallery advertises built-in acceptance but must not intercept these keys; preserve F2/F3 controls and Ctrl+C/Ctrl+U behavior. Document Tab's guarded action and caller-visible key consumption.

Regression-first coverage: all three exact normalized keys; full Unicode/multiline insertion; caret at end; one OnChange/undo step, redo and no OnSubmit; OnChange mutation; no placeholder, nonempty input/selection and modified keys (Shift+Tab, Home, Shift+End, Ctrl+End) preserve prior behavior. Test actual byte paths (`0x09`, End sequence, `0x05`) and real gallery placeholder acceptance with retained display/cursor. Explicitly test container/focus Tab fallback when no placeholder or nonempty input, not just library return values. Use state-based waits, no sleeps. No known baseline failures (last host full run passed), preserve concurrent unrelated changes. Developer terra:med, host reviewer; developer runs one logged make test-q1, installs and commits without push or closing; host reviews and independently verifies.

## Host acceptance

Host reviewed all implementation, test and canonical documentation diffs, then independently ran `make test-q1`: exit 0 (log `/tmp/loom-issue439-host-test-q1.log`). Host exercised the installed `loom widgets --show TextEntry` in separate real PTYs for Tab, End and Ctrl+E: F2 established placeholder/offset zero; each key accepted it and moved the offset to 18 (end of `Write an update…`); F10 exited successfully each time. Guarded widget acceptance, ordinary Tab focus traversal, notifications/history and no-submit behavior are covered by the final green suites. No external application changes; ready to close.

## Goal

/goal Make loom.TextEntry accept its visible placeholder with Tab, End and Ctrl+E under the empty-input/start-caret guard, preserving all other routing/editing semantics and proving one-step acceptance without submit in unit and live terminal tests; stop and report when blocked on a user decision or denied permission. On a loop, hard architectural or OS blocker, or no progress, clear the goal and record a failure report explaining why.
