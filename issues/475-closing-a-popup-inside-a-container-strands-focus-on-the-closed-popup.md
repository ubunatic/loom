# 475 — Closing a Popup inside a container strands focus on the closed popup

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: 458, 316, 469

---

## 1. Problem & Motivation
A Popup placed as a container child (the documented Popup example puts it in a Frame,
`spec/widgets.yaml:243`) keeps the container's focus after it closes. Stack keeps its focused
child index (`stack.go:138-151`), and a closed Popup returns `Ignored` for every key
(`popup.go:77-84`), so the next keys are lost until the user presses Tab. 458 batch 8
(`docs/studies/tuios/report-desktop.md`): canary A7 and host canary `TestPopupStack` give
`open=false value="" consumed=false` after Tab, open, Esc, `x`. Pane's own help overlay does not
have this problem: it takes keys without changing the root's focus index (`pane_dispatch.go:51`).
A desktop app returns focus to the widget that had it before the popup opened.

## 2. Technical Specification / Findings
- Fix the library, not the caller: when a focused Popup closes, focus returns to the child that
  was focused before it opened (or the nearest focusable child if that is gone). Decide where the
  memory lives (container or Popup) and how the container learns about the close (EventResult,
  focus change notification).
- Check Frame and other containers with the same focus index logic, and mouse: a click while the
  popup is closed must not go to it.
- Event routing: start on terra:med (project rule).

## 3. Implementation & Verification Plan
/goal After a Popup inside a Stack or Frame closes (Esc, outside click, programmatic), the next key
reaches the widget that was focused before the popup opened, with tests; stop and report when
blocked on an API decision or denied permission; on a loop or no progress, clear the goal and write
a failure report.

- Re-check live code first; port `TestPopupStack` (scratchpad `b8-canary`) into a loom test per
  container and close path.

## Jules test-draft (GH #123, 2026-11-10)
Session jules-routing-tests-b added `popup_focus_return_test.go` (`TestPopupFocusReturn`). Host run on
the PR head: fails as the bug predicts: after the popup closes, a typed `x` never reaches the input
(value stays empty), because `Popup.ConsumeKey` returns `Ignored()` when closed (`popup.go:79-80`)
and `Stack.ConsumeKey` keeps the focus on it (`stack.go:143`). References hold. Pick the test from
`git fetch github pull/123/head`.
