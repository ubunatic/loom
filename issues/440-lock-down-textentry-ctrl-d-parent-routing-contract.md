# 440 — Lock down TextEntry Ctrl+D parent routing contract

**Status**: Closed — Existing Ctrl+D ignored contract documented and locked by decoded/named state-preservation and child-first Frame parent tests; developer and independent host test-q1 passed.
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Test
**Related**:

---

## 1. Problem & Motivation

Owner: "^D should also exit the entry, i.e. be sent to the parent widget". Ctrl+D is a host key, not deletion, submission or widget-owned application quit. Searcher verified TextEntry currently ignores it, so lock down/document the existing contract rather than invent a behavior change.

## 2. Technical Specification / Findings

TextEntry.ConsumeKey falls back to Ignored for ctrl-d; normalizer preserves it; DecodeKey maps raw 0x04 to ctrl-d. Frame delegates to focused child then bindings. Gallery wrappers do not handle Ctrl+D. Preserve value, cursor, selection, placeholder, clipboard/history and notification/submit state. Parent may handle focus/exit itself. Do not add a global/app Ctrl+D quit binding or change default quit keys.

Searcher noted apparent confusion between session quit bindings and pane defaults; verify the exact live defaults/accessor and gallery DisableDefaultQuit read-only before documenting app behavior. Do not treat session bindings as pane bindings. Report true discrepancies but do not expand this ticket into session fixes.

## 3. Implementation & Verification Plan

Approved write set: `text_entry_test.go`, `gallery/coverage_test.go` if needed, `docs/Widgets.md`, this ticket and one bounded study. Production change only if a real violation is reproduced; stop for write-set approval first. Test empty/nonempty/selected/placeholder state with decoded raw 0x04 and named key; assert Ignored/non-Quit/no mutation, parent receives the exact event via child-first routing (use real Frame binding or parent probe), and parent action runs once. Verify key output and unchanged clipboard/history/notifications without confusing ignored entry with app exit. Preserve Ctrl+C/U/placeholder and all existing assertions. Unit and parent routing tests, one make test-q1, install. Developer terra:med; host reviewer. No external edits/push or caller workaround.

## Host acceptance

Host reviewed all test assertions and key-policy documentation. Independent `make test-q1` passed (exit 0, log `/tmp/loom-issue440-host-test-q1.log`). Regression covers named and decoded Ctrl+D across empty, text, selected and placeholder states, exact state/history/clipboard preservation, and child-first Frame parent action. No production change was needed. Pane defaults are Ctrl+Q/F10 with separate Ctrl+C fallback; the gallery disables defaults. Ctrl+D is not a new app quit binding. Installed by developer; ready to close.

## Goal

/goal Prove and document TextEntry Ctrl+D is ignored and offered to its parent without mutation or widget-owned quit, with decoded-byte parent-routing regression and preserved existing behavior; stop and report when blocked on a user decision or denied permission. On a loop, hard architectural or OS blocker, or no progress, clear the goal and record a failure report explaining why.
