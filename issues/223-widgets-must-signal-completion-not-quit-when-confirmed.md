# 223 — Widgets must signal completion, not Quit, when confirmed

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Refactor
**Related**: 209, 212

---

## 1. Problem & Motivation
Choice (choice.go) returns `QuitResult()` when an item is confirmed. Embedded in a larger app this ends the whole app. 212 fixed the gallery by clearing `result.Quit` from children in `cmd/loom/widgets.go`. That is a caller-side workaround; the library rule is to fix the widget contract.

## 2. Technical Specification / Findings
Separate "this widget's interaction is done" (e.g. a `Done`/`Submitted` flag on EventResult) from "quit the app". A standalone prompt runner may map Done to exit; containers do not. Audit other widgets that return QuitResult on confirm.

## 3. Implementation & Verification Plan
Remove the Quit masking from `themedGallery`; the 212 PTY test stays green. Add a test that a standalone Choice prompt still exits on confirm.

/goal Confirming a widget never quits its host app without caller masking, proven by tests; or stop and report when blocked on a user decision.
