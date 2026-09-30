# 203 — Gallery key routing: typing and Esc quit the app

**Status**: Closed
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: 195, 197-202 gallery

---

## 1. Problem & Motivation
User feedback on `loom widgets`: TextArea and TextInput quit on any key, so typing is impossible; in Popup, Esc quits the app instead of closing the popup; Dialog closes on Enter and does not reopen when its tab is selected again.

## 2. Technical Specification / Findings
Root cause (host read, 201): `themedGallery` in `cmd/loom/widgets.go` forwards only the legacy `HandleKey(e) bool` (true = quit) and `HandleMouse`, hiding `ConsumeKey`/`ConsumeMouse` `EventResult` of the wrapped Tabs; the dispatcher then applies fallback quit keys (Esc, q, text). Fix the wrapper to forward `EventResult`. Modal demos (Popup, Dialog) must consume Esc/Enter while open; a closed Dialog/Popup demo needs a way to reopen (e.g. Enter or a button, and on tab reselect). If loom lacks key capture for modals, add a minimal non-breaking one.

## 3. Implementation & Verification Plan
Proof: each item gets a regression test that fails before the fix, driven through the real `loom widgets --show <Name>` binary in a PTY (`gallery/gallery_test.go`, `internal/ptytest`), plus a `docs/progress/<Widget>.ansi` capture where the look changes.

/goal Typing works in TextArea/TextInput, Esc closes Popup without quitting, Dialog can be reopened; each proven by PTY test; or stop and report when blocked on a user decision.
