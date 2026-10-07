# 286 — Prevent search panel quit keys from losing unsaved editor changes

**Status**: Closed — Integrated unsaved quit guards; full suite and PTY search quit/dialog capture pass.
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

Updated review (2026-10-07, PR head a9f1a7d): ^Q/^C/^D now route through handleQuit while search is focused. Independent PTY probes with unsaved edits show the Save changes prompt, keep the app alive and preserve the saved file. The full make test-q1 suite passes. Keep open pending integration and complete Save/Discard/Cancel/failed-save coverage; the original loss path is fixed on the PR branch.

## 3. Implementation & Verification Plan
/goal Search-panel close/quit keys cannot discard unsaved document changes without an explicit user decision; verify the behavior, or stop and report when blocked on a user decision or denied permission.

Acceptance: PTY and integration coverage for ^Q/^C/^D/Esc with modified and clean documents, including Save/Discard/Cancel and failed save. Verify make test-q1 and make install.

## Integration verification (2026-10-07)

Integrated PR #16 unsaved quit protection, plus a host guard for QuitResult from editor/browser children. Final full suite passes and PTY unsaved ^Q over search opens the dialog without exiting; ^P captures that dialog. Earlier ^C/^D probes also passed.
