# 453 — TextEntry middle-button scrub and a runtime swap of the send key

**Status**: Closed — middle scrub (shown only while the scrollbar shows) and `SetSubmitKeys` delivered; gallery key Ctrl+T swaps Enter and Alt+Enter
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 448, 452

---

## 1. Problem & Motivation
Owner: middle-button drag did not scroll a `TextEntry` (left out of 448). Chat hosts also want to choose whether Enter or Alt+Enter sends.

## 2. Technical Specification / Findings
- Middle scrub reuses `contentPan` and the pane capture (452), only when the bar shows; a middle press on the bar column or without a bar stays ignored. `TextEntry` has no middle-click paste, so nothing conflicts. Esc restores.
- `SetSubmitKeys(submit, newline)` replaces the bindings at run time; empty lists restore the defaults. The gallery demo toggles with Ctrl+T (F-keys are taken) and the bar shows `Enter` or `Alt+Enter` as the Submit key.
- Unverified: terminals that do not report Alt+Enter distinctly (the default newline list already relies on it).

## 3. Implementation & Verification Plan
Tests in `text_entry_scrollbar_test.go`; docs in `docs/Widgets.md`; check with `loom widgets --show TextEntry`.
