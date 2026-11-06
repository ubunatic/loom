# 438 — TextEntry Ctrl+U visual-line clearing and previous-line navigation

**Status**: Closed — Implemented TextEntry Ctrl+U visual-prefix clearing and hard-line navigation with documented soft-boundary behavior; final developer and independent host test-q1 passed, installed-binary PTY smoke verified suffix preservation and Ctrl+C confirmation/exit.
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: https://codeberg.org/harnez/taia/issues/151, https://codeberg.org/harnez/taia/issues/106

---

## 1. Problem & Motivation

Owner request: "Now see ^U ticket in ../taia and apply ^U to the TextEntry widget". Taia #151 requests clearing from cursor to current line beginning; at a line beginning Ctrl+U moves to the previous line end without deleting the newline, and at the first line beginning stays in place. Repeated presses alternate clearing and moving backward. Support visible soft-wrapped segments where feasible and define boundary behavior explicitly.

## 2. Technical Specification / Findings

Recheck live TextEntry behavior, wrapping/cursor mapping and recent history before implementation. Fix the library rather than adding gallery/app workarounds. Taia is read-only context; no external implementation authorized. Preserve selection, undo, paste, Unicode display-width and existing Ctrl+C contracts unless a documented intentional change is approved. Ctrl+L is not requested.

## 3. Implementation & Verification Plan

Use a bounded read-only Searcher, then a developer plan with exact files and visual-boundary semantics before approving writes. Add regression-first tests for line interior/start, first line, empty text, multiline, newline preservation, repeated presses, soft-wrap boundaries and display-width edge cases; validate the real `loom widgets --show TextEntry` path. Run quota verification and install before reporting. Record any caller interception that prevents the widget key from reaching TextEntry without editing that caller.

## Decision

Use existing cursor affinity rather than redesigning movement/render/history: inside a visual segment, delete from segment start to caret; at a hard-line start, move before the preceding newline without deleting it; at offset zero, ignore. Soft boundaries share one caret offset and are displayed on the following row. At such a boundary, Ctrl+U clears the preceding visual segment immediately (documented exception to alternating navigation/clearing), preserving suffix text and all unselected hard newlines. Example at width four: `abcdefgh|ij` -> `abcd|ij` -> `|ij`. This keeps repeated presses useful without a broad cursor-affinity redesign.

Delete an explicit selection only, matching Backspace/Delete (selected newlines may be deleted). Each deletion is one undo step/OnChange; navigation creates neither and must preserve redo/clipboard. Use latest layout width; before layout fall back to newline-delimited lines. At the first visual segment's beginning, no-op. Ctrl+L is out of scope.

## M1 — Implementation and validation

Searcher and read-only developer plan confirmed the missing handler and shared soft-wrap offsets. Approved write set: `text_entry.go`, `text_entry_test.go`, `gallery/coverage_test.go`, `cmd/loom/widgets_paste_pty_test.go`, `spec/widgets.yaml`, `docs/Widgets.md`, `docs/Upgrading.md`, this ticket, and one bounded sprint study. No app workaround or external edits. Add an Unreleased compatibility note for formerly ignored Ctrl+U becoming consumed.

Tests: regression first, hard-line interiors/starts/empty/first line/trailing and consecutive newlines, repeated visual clearing with retained suffix, prompt width, CJK/combining/emoji graphemes, selection by keys/mouse, undo/redo/notifications/EventResult, unknown layout and resize. Exercise actual CLI via byte 0x15 and state-based cursor/screen/exit assertions, retaining paste and Ctrl+C tests. Developer runs one logged `make test-q1`, installs, commits evidence; host reviews and independently verifies. No known baseline failure: previous host full quota run passed. Tree was clean; preserve any concurrent unrelated work. Model changed by owner to terra:med in AGENTS.md, commit e60554f.

## M2 — Host review coverage hardening

### Pre-Work / Required Refinements

Host reviewed all M1 source and assertion diffs. Implementation matches the chosen bounded semantics; existing assertions were preserved. Complete the following acceptance gaps before signoff:

- Add explicit clipboard preservation tests for deletion and hard-start navigation: seed internal clipboard through supported key behavior, bind a recording host hook, and prove Ctrl+U neither changes clipboard nor emits a host write. Prove paste still inserts the original copied text.
- Add explicit single-step undo/redo and exact OnChange assertions for visual-prefix deletion (not only selected-range deletion), including retained suffix/newlines and repeated clearing.
- Strengthen real gallery PTY coverage: drive multiline input, assert cursor coordinates before and after Ctrl+U at a hard-line start and inside a line, and retained surrounding display. Also verify repeated clearing of a known-width soft-wrapped suffix with actual cursor/screen state where feasible; do not claim text-only screen checks prove caret navigation. State-based waits only.
- Preserve Ctrl+C exit confirmation coverage. Approved writes: existing M1 core/test/PT Y files only as required, this ticket and existing study; no caller workarounds or unrelated session changes.
- M1 focused packages/vet/install passed; its full quota run failed in untouched `internal/session/TestSessionDetachPrefixDAndReattach`, also intermittently seen in previous ticket. Record actual results, do not weaken or fix that unrelated test. After test/source edits run one quota suite, install, update evidence and commit. Leave ticket open pending host review/verification.

## Host acceptance

Host reviewed M1/M2 implementation and every changed assertion, then independently ran `make test-q1`: exit 0 (log `/tmp/loom-issue438-host-test-q1.log`). Host also exercised the installed `loom widgets --show TextEntry` under a real PTY: Ctrl+U inside `prefixZZ` retained `ZZ`; Ctrl+C then cleared input, showed confirmation, and exited successfully. The first smoke probe's raw-output marker was split by ANSI styling; stripping styling fixed the probe, not product code. The developer's multiline cursor-coordinate tests, repeated soft-wrap clearing, clipboard preservation and exact undo notifications passed in both final full runs. No Taia edits or Ctrl+L changes. Ready to close with the documented soft-boundary behavior.

## Goal

/goal Deliver Ctrl+U clearing and previous-line navigation in loom.TextEntry matching Taia #151, including visual-line clearing where feasible, with explicit tested wrap-boundary behavior, gallery/live terminal evidence and passing repository verification; stop and report when blocked on a user decision or denied permission. On a loop, hard architectural or OS blocker, or no progress, clear the goal and record a failure report explaining why.

## Developer evidence

Regression first: the soft-boundary test failed because Ctrl+U was ignored. Implementation and verification are recorded in the accompanying study; the ticket remains open for host acceptance.

## M2 developer evidence

- Clipboard regression seeds the internal clipboard with Ctrl+C, proves hard-start navigation and visual deletion do not call the host clipboard hook, then pastes the original copied value.
- Visual-prefix tests assert retained suffix/newlines, each OnChange value, and independent undo/redo steps.
- CLI PTY coverage sends multiline paste and Ctrl+U byte 0x15; it asserts terminal cursor coordinates at a hard-line start and inside the line, retained surrounding text, and repeated deletion at a known soft-wrap boundary. Existing Ctrl+C confirmation checks remain intact.
- `go vet ./...` passed. `go test . ./gallery ./cmd/loom` passed. One `make test-q1` passed in full with no `--- FAIL`; output: `/tmp/loom-issue-438-m2-test-q1.log`. `make install` passed.
- M1's quota failure was in untouched `internal/session/TestSessionDetachPrefixDAndReattach`; it did not recur in M2. Ticket remains open for host acceptance.
