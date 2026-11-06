# 434 — Pasting into taia chat clears the terminal: add widget regression test

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: issues/181, issues/417, issues/418

---

## Problem

The user reports that copying and pasting into the chat input in taia clears the terminal. The cause is unknown and may lie in loom; do not assume it does. Bracketed paste support was previously delivered in issue 181.

## Decision

Validate the library contract using the in-repository host `loom widgets --show TextEntry`; external application code remains read-only. The owner proposed clear-on-first-Ctrl+C and parent handling on the next press. Demonstrate this as input-state-based gallery policy, not a new TextEntry library contract: offer the key to TextEntry first (selected text copies), otherwise clear nonempty input and consume it, and leave an empty input's Ctrl+C unhandled for the existing parent behavior. This reversible example preserves the current selection-copy contract and needs no library API change.

## Milestones

### M1 — Gallery host regression and Ctrl+C example

- Searcher and terra:low developer inspected current code/history read-only. Existing commit `49e5bad` already covers modeled chat canvas preservation and Pane.run PTY paste. Neither worker reproduced the original Taia report or established the responsible layer.
- Approved write set: `gallery/coverage.go`, `gallery/coverage_test.go`, `cmd/loom/widgets_paste_pty_test.go`, this ticket, and one bounded sprint study in `docs/studies/`. No library semantic/API change, no external repository edits, no changes to issue 436.
- Add regression-first gallery routing coverage: selected Ctrl+C copies before parent handling; unselected nonempty Ctrl+C clears only the entry; empty Ctrl+C remains ignored by the wrapper. Verify following keys and paste remain usable, including multiline/control payloads where supported.
- Exercise the real `loom widgets --show TextEntry` command via PTY and assert surrounding display is retained after paste and clearing. Preserve all existing assertions. Separate clipboard-protocol evidence from real desktop clipboard delivery; do not claim the latter from unit or PTY tests alone.
- Verify actual top-level parent routing rather than assuming an empty-input Ctrl+C exits. Record concrete expected manual steps and any unsupported part.
- Developer runs one `make test-q1` after checking for unrelated work; host reviews and independently verifies before acceptance. Install before reporting final results.
- Record verified evidence honestly. Keep this ticket open if the original report cannot be reproduced or attributed; gallery coverage alone does not establish a Taia fix.

#### M1 delivered

- The TextEntry gallery parent offers Ctrl+C to `TextEntry` first. A selected range remains selected/copied; otherwise nonempty input is cleared and consumed; empty input remains ignored by the wrapper.
- Gallery unit coverage checks selection precedence, clearing, empty routing, and continued typing/paste. The real `loom widgets --show TextEntry` PTY test checks selected and unselected Ctrl+C, confirms the top-level gallery stays running after empty Ctrl+C, and verifies paste still works with surrounding gallery controls visible.
- Regression-first evidence: `go test ./gallery -run '^TestTextEntryDemoCtrlCPolicy$' -count=1` failed before the gallery parent implementation at the unselected nonempty Ctrl+C case. The targeted tests and full changed-package tests passed after implementation.
- Existing library tests cover OSC 52 encoding and TextEntry clipboard-hook behavior. The gallery PTY does not establish desktop clipboard delivery; the gallery wrapper does not expose the private clipboard binding path.
- Manual checklist: run `loom widgets --show TextEntry`; press Ctrl+A then Ctrl+C and confirm the draft remains; press Right then Ctrl+C and confirm the input clears while the gallery controls remain; press Ctrl+C again and confirm the gallery stays open; paste `after\nclear` and confirm it appears without losing the F2/F9/F10 controls. Desktop clipboard receipt remains a separate manual check.
- The reported Taia terminal clearing was not reproduced or attributed in M1. No Taia fix is claimed; keep this issue open pending a reproduction and layer attribution.

## Goal

/goal Re-verify the report against current taia and loom code and recent history; reproduce and identify the responsible layer. Add a loom chat-input widget regression test exercising paste and asserting the text is inserted without clearing the terminal or losing the surrounding display, and fix any confirmed loom defect. Verify the fix and report any taia-side follow-up needed. Stop and report if blocked by a user decision or denied permission; on a loop, hard architectural or OS blocker, or no progress, clear the goal and write a failure report explaining why.
