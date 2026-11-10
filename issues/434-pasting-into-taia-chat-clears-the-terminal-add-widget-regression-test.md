# 434 — Pasting into taia chat clears the terminal: add widget regression test

**Status**: Closed — not reproducible: owner paste with ESC[2J in foot and gnome-terminal did not clear; PTY regression test kept
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

### M2 — App-level quit confirmation (owner clarification)

#### Pre-Work / Required Refinements

- M1 host diff review: the demo clears unselected text correctly, but empty Ctrl+C stays unhandled and never exits. The owner explicitly requires the outer gallery app to show a hint on the first empty-input Ctrl+C and quit on the next. M1's empty-Ctrl+C-stays-open assertions must be extended, not simply removed.
- Required sequence from nonempty unselected input: clear entry (stay running, no quit arm), Ctrl+C on empty input shows a visible "Press Ctrl+C again to quit" hint, next Ctrl+C exits. From initially empty input it takes two presses. Keep existing selection-copy precedence; this clarification does not authorize replacing selection-copy with destructive clearing.
- Confirmation belongs to the in-repo outer app, not the TextEntry library. Ordinary typing, paste, focus/navigation or another action cancels pending confirmation; clearing text itself must not arm it. Prefer a simple consecutive-press state machine without sleeps/timers.
- Preserve F10 and other existing gallery quit behavior. Test the actual command's process exit and the visible hint, not only EventResult or retained draft. Preserve paste/display regressions and assert continued input after cancellation.
- Developer inspected the outer path; host approves M2 write set: `cmd/loom/widgets.go`, `cmd/loom/main_test.go`, `cmd/loom/widgets_paste_pty_test.go`, this issue, and the existing sprint study. Opt in only for single `--show TextEntry`, leaving other demos unchanged. Use an app-level state machine and a visible confirmation row above existing controls without removing F10. Add focused routing/cancellation tests and actual CLI PTY exit proof via Session.Wait. No library API/semantic change or external edits authorized.
- Reset on typing, paste, navigation, mouse/focus changes and other actions, including gallery F8/F9 actions. Verify the consumed clear and selected-copy path never arms confirmation; retain F10/Ctrl+Q immediate exit. Check short-height geometry when adding a hint row, preserve the input/cursor contract and surrounding display, and avoid stale hint cells after cancellation.
- Clipboard limitation from M1 remains explicit: preserving selected text does not itself prove copy delivery. Do not label that PTY assertion a clipboard-delivery test.

#### M2 delivered

- The app-level confirmation is enabled only for a single `--show TextEntry` demo. TextEntry receives Ctrl+C first, preserving selected-text behavior; unselected nonempty input clears and consumes without arming quit. Ctrl+C on empty input displays `Press Ctrl+C again to quit`; the next consecutive Ctrl+C quits. No TextEntry library/API semantics changed.
- Typing, navigation, paste, mouse activity, focus changes, F8 and F9 cancel confirmation. The app repaints after ignored events that cancel it, preventing a stale hint. F10 and Ctrl+Q remain immediate. The confirmation occupies a row above existing controls; one-row height uses compact inline text and zero-height drawing is guarded.
- `cmd/loom/main_test.go` checks selection precedence, clear-vs-arm behavior, hint/control rows, cancellation paths, immediate quit bindings, short heights, and opt-in boundaries. `cmd/loom/widgets_paste_pty_test.go` drives the actual CLI process through selection, clear, confirmation, cancellation, multiline bracketed paste, and final confirmation; `Session.Wait` confirms process exit.
- Verification: `go test ./cmd/loom` passed; `go vet ./...` passed. The single `make test-q1` run exited 2 at `internal/session/TestSessionDetachPrefixDAndReattach` (`session_pty_test.go:167`, retained screen/tabs repaint assertion). Full output: `/tmp/loom-issue434-m2-test-q1.log`; no rerun was made. `make install` passed. Host verification is still required by the local test-tier rule.
- Manual checklist: run `loom widgets --show TextEntry`. Press Ctrl+A, Ctrl+C and check the selected draft remains (this does not establish desktop clipboard delivery). Press Right, Ctrl+C and check the entry clears while F2/F9/F10 remain. Press Ctrl+C and check the hint appears; type `x` and check the hint clears and typing works. Ctrl+C clears `x` without quitting or arming; paste `after\nclear` and check both lines and surrounding controls remain; Ctrl+C clears the paste, Ctrl+C shows the hint, and another Ctrl+C exits. In a fresh run with an initially empty entry, two Ctrl+C presses exit. Check F10 and Ctrl+Q exit immediately.
- M2 validates this in-repository gallery behavior only. The original Taia terminal-clearing report remains unreproduced and unattributed; this issue remains open. Desktop clipboard receipt still requires a separate manual check.

### M3 — Host review hardening

#### Pre-Work / Required Refinements

- Host reviewed every changed assertion in M1/M2. M2 removed the embedded `ESC[2J` payload from the existing Ctrl+C/paste PTY test. Restore that control-containing payload or retain it in an equally strong explicit real-command test; do not reduce the original display-preservation coverage to plain text alone.
- In `themedGallery.Draw`, subtracting a hint row after reserving controls and confirmation can make `content.H` negative at height 2. Clamp the final content height to zero before drawing the child; add a child-rectangle assertion for heights 0, 1, 2, 3, including a hint-bearing child. Do not claim no-panic alone proves valid geometry.
- Approved write set remains M2's three cmd/loom files, this issue, existing study. No library change or unrelated session fix. Preserve confirmation state/reset/exit coverage.
- M2 focused tests/vet/install passed, but full `make test-q1` failed in untouched `internal/session/TestSessionDetachPrefixDAndReattach`; record this accurately, do not weaken its assertions or claim a full pass. Source changes in M3 permit the next quota verification. Host will independently verify after review.

#### M3 delivered

- Restored the control-containing `\x1b[2J` sequence inside the bracketed paste sent through the real `loom widgets --show TextEntry` PTY path; the test still verifies pasted lines and surrounding controls, then confirms Ctrl+C handling and process exit.
- `themedGallery.Draw` now clamps the child rectangle height after all control, confirmation and hint-row reservations. `TestThemedGalleryChildHeightNeverNegative` records child rectangles at heights 0–3 with and without a hint-capable child, and checks the height is nonnegative.
- Regression-first evidence: before the clamp, the new test failed at height 2 with `H:-1` for both child variants. The PTY control-payload regression passed. After the clamp, `go test ./cmd/loom -count=1`, `go vet ./...`, and the single `make test-q1` all passed. Full quota output: `/tmp/loom-issue434-m3-test-q1.log`.
- No `internal/session` assertion or unrelated source was changed. M2 behavior and tests remain intact. Host review and independent quota verification remain pending; issue 434 stays open because the original Taia report remains unreproduced and unattributed.

### Host verification

Host reviewed M1/M2/M3 source and assertion diffs, required restoration of the terminal-clear paste payload and a nonnegative-height regression, and independently ran `make test-q1` after M3. The host run passed (exit 0); full log: `/tmp/loom-issue434-host-test-q1.log`. Developer install passed on final source. Gallery behavior is delivered; original Taia report and desktop clipboard receipt remain unverified, so this issue stays open. No push performed.

## Goal

/goal Re-verify the report against current taia and loom code and recent history; reproduce and identify the responsible layer. Add a loom chat-input widget regression test exercising paste and asserting the text is inserted without clearing the terminal or losing the surrounding display, and fix any confirmed loom defect. Verify the fix and report any taia-side follow-up needed. Stop and report if blocked by a user decision or denied permission; on a loop, hard architectural or OS blocker, or no progress, clear the goal and write a failure report explaining why.

## Owner manual check (2026-11-09)

Owner pasted `/tmp/loom-434-paste.txt` (three lines, `ESC[2J` inside line two) into `loom widgets --show TextEntry`. No screen clear in either terminal. foot showed `line [2J two`; gnome-terminal (VTE; Tilix also uses VTE) showed `line ␛[2J two`. The escape byte never reached the screen as a control sequence: foot or TextEntry drops it, and VTE appears to replace it with U+241B before delivery (not verified at byte level). Original Taia report not reproduced; the clearing is attributed to neither loom nor a current terminal. Gallery coverage and the real-command PTY test with the `\x1b[2J` payload remain as regression guards. Closing as not reproducible.
