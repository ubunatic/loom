# 286 — Prevent search panel quit keys from losing unsaved editor changes

**Status**: Open
**Priority**: P0 (Critical)
**Severity**: Critical
**Category**: Bug
**Related**: [PR #16](https://github.com/ubunatic/loom/pull/16), [282](282-add-f2-side-panel-with-file-browser-to-loom-edit.md), [283](283-add-f3-and-ctrl-f-search-panel-to-loom-edit.md)

---

## 1. Problem & Motivation
PR #16 loses unsaved edits through search quit keys. PTY reproduction: open `hello`, type `X`, press ^F, then ^Q. The app exits without a Save/Discard/Cancel prompt, and the file still contains `hello`.

## 2. Technical Specification / Findings
cmd/loom/edit.go:577–583 returns SearchBar.ConsumeKey results directly. SearchBar returns QuitResult for ^Q, ^C and ^D, so a child search widget can terminate the host. Contain child termination and route any application quit through unsaved-change protection. Fix library gaps in the library. Event-routing work starts on codex:sol:med.

Observed at PR head b8a59ca; this branch has not been merged locally. Before implementation, check live code and recent commits and reverify the finding.

## 3. Implementation & Verification Plan
/goal Search-panel close/quit keys cannot discard unsaved document changes without an explicit user decision; verify the behavior, or stop and report when blocked on a user decision or denied permission.

Acceptance: PTY and integration coverage for ^Q/^C/^D/Esc with modified and clean documents, including Save/Discard/Cancel and failed save. Verify make test-q1 and make install.
