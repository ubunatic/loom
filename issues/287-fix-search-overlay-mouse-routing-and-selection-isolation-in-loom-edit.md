# 287 — Fix search overlay mouse routing and selection isolation in loom edit

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [PR #16](https://github.com/ubunatic/loom/pull/16), [282](282-add-f2-side-panel-with-file-browser-to-loom-edit.md), [283](283-add-f3-and-ctrl-f-search-panel-to-loom-edit.md)

---

## 1. Problem & Motivation
PR #16 routes search-panel clicks to the document underneath. PTY reproduction at 100×24 with --mousegrab: open ^F, click the query field at column 55/row 5, type X, then ^S; X is saved into the second document line instead of the query. Search highlights also call SetSelection, visibly opening the formatting popover.

## 2. Technical Specification / Findings
Stale reference: when this was written, cmd/loom/edit.go:623–660 did not route searchRect before editorRect. The live code is internal/editor/edit.go, where search is routed first (ConsumeMouse 821–889). Synchronize widget focus, prevent mouse hover/scroll from stealing focus, and render search highlights separately from editing selections/popovers. Refresh matches after document edits and restore usable editor focus when closing search. Use standard Loom widgets and container-owned routing; file/link library gaps rather than app hacks. Start on codex:sol:med.

Observed at PR head b8a59ca; this branch has not been merged locally. Before implementation, check live code and recent commits and reverify the finding.

Updated review (2026-10-07, PR head a9f1a7d): searchRect is now routed before editorRect. Repeating the query-field click/type/save PTY probe leaves the document unchanged. Remaining gaps: highlightSearchMatch still calls SetSelection; no widget focus synchronization for the editor, stale-match refresh after document edits, or restriction of focus changes to mouse presses. Mode hit regions are still manually calculated. Keep open; this is a partial fix.

## 3. Implementation & Verification Plan
Updated review (2026-10-07, PR head 9f41045): selection highlighting, editor focus synchronization, hover/scroll focus changes, stale matches and manual mode hit regions remain unchanged. The full make test-q1 suite passes but does not resolve these findings. Keep open.

/goal Search UI receives its own input and highlights matches without changing editing selection or accidentally editing the document; verify the behavior, or stop and report when blocked on a user decision or denied permission.

Acceptance: Cover query/mode/control clicks, overlay occlusion, focus changes, stale matches after edits, and saved document integrity. Verify real PTY interaction, make test-q1 and make install.

## Integration verification (2026-10-07)

Host integration adds display-only RichTextEdit.Highlights in the library, replaces search SetSelection/ClearSelection, synchronizes editor/search focus on draw, retains press-only mouse focus and edit-driven match refresh, and uses standard HintBar mode hit testing. Tab/Shift-Tab moves container focus and ^R switches mode. Selection isolation, editing cursor and bounds regression tests pass. Keep open for complete mouse/control/focus coverage.

## Sprint Plan (lean sprint, 2026-10-30)

Searcher audit (read-only, code-read, not yet run): the live editor is `internal/editor/edit.go`; `cmd/loom/edit.go` is a 51-line wrapper, so the `cmd/loom/edit.go:623–660` reference above is stale. Search is routed before the editor (`ConsumeMouse` 821–885: searchRect at 843, editorRect at 878). `highlightSearchMatch` (371–381) no longer calls `SetSelection`, but it writes `v.edit.Cursor = to` at 380 when search is focused. Focus is set on mouse press only. Refresh is missing after edits made while search is closed, on reopen, and after mouse edits.

### Milestones

- **M1 (reproduction tests, red first):** add tests before any fix.
  - PTY: open ^F, click the query field, type X, ^S; the file must hold X in the query and the document must be unchanged.
  - PTY or unit: clicks on the mode row and each control in the search bar (edit.go:848–861) reach the intended action and do not fall through.
  - Click-through: a click inside searchRect never changes the editor's cursor, selection or content.
  - Stale matches: edit with search closed, reopen, and the match count and search_next offsets match the current text.
  - Focus: hover and scroll over the search panel and the editor do not change focus.
  - Cursor: a search highlight move does not change `edit.Cursor` (edit.go:380).
  - Acceptance: commit the failing tests with the fix in M2–M4, or mark them as `t.Skip` only with a dated reason. Do not weaken assertions.
- **M2 (stale matches and tiny-panel clicks):** refresh matches when search opens and on every document modification, not only when `showSearch` is set (edit.go:667–668, 305–316, OnStateChange at 134–141). Set `searchRect` only inside the draw guard (edit.go:491 vs 493) so an invisible panel does not swallow clicks.
- **M3 (search highlight must not move editing cursor):** keep the highlight display-only (library `RichTextEdit.Highlights`, documented at docs/Widgets.md:20). Remove the cursor assignment at edit.go:380. If scroll-to-match is needed, use a view-only operation that does not touch `Cursor`.
- **M4 (popover precedence and mouse edits):** a formatting popover open from a selection currently captures clicks before searchRect (edit.go:809–818, modal_mouse.go:50–75). Probe in a PTY first. If it reproduces, fix it in the library modal contract, not in the app. Confirm which mouse paths can mutate text (edit.go:877–885) and refresh matches after any such edit.
- **M5 (final acceptance, host only):** the host runs `make test-q1`, `make install`, and the PTY acceptance list on the installed binary. The developer does not close the ticket.

### Pre-Work / Required Refinements

- Fix the file reference in the Technical Specification to `internal/editor/edit.go`; keep the old line numbers in a note only.
- Do not add a file-level workaround in the editor for a library routing gap; fix the library event contract (docs/Widgets.md: 0-based child-local coordinates, press-only focus, Tab belongs to containers).
- Event routing: the developer must follow the event-routing contract in docs/Widgets.md when editing mouse routing, and must not subtract 1 from mouse coordinates.
- Tests: no sleeps; PTY tests stay in the full tier (docs/TestTiers.md).

## Decision

- Plan review (host, 2026-10-30): the developer's plan turn (job b7za1pr4q, session loom-287-dev) confirmed the premises and found these extra items, which go into M2–M3 below: `searchModes` is never reset (stale mode hit target, edit.go:502 and 398); highlights paint while search is closed (updateSearchMatches 318–369, called from closeDocument 158 and openFile 184); `OnStateChange` fires per state transition, not per edit (richtextedit_file.go:59), so it cannot drive refresh; scroll-to-match is blocked because `richtextedit_render.go:41–58` snaps ScrollY to the cursor on every Draw.
- Baseline: `make test-q1` is red before any change. `internal/session` `TestServerLastTabExitCleansSocket` times out after 10m. `internal/editor` passes. A live `loom session -s projects` server (pid 1647677) is running and may interfere. The host owns this: it is outside this ticket, and it blocks the final gate, not the editor milestones.
- Plan review decisions:
  - **A (baseline):** the developer does not touch the session test. The host checks it with the live session server stopped and files its own ticket if it is a real failure. Milestone tests run with targeted `go test ./internal/editor`; the full `make test-q1` runs at the final gate only.
  - **B (M3 scope):** approved. Add `RichTextHighlight.Reveal`, the library change in `richtextedit_render.go` and `richtextedit.go`, and the matching update in `docs/Widgets.md`. A follow-up for scroll is not acceptable here, because the highlight would not be visible.
  - **C (ticket text):** the host records the findings in this ticket, as done above. The developer does not edit the ticket.
  - **D (commit policy):** approved. One commit per milestone, never a red commit. M1 tests stay uncommitted as red evidence until M2 lands in the same milestone commit.
  - **M4 model:** stays on Claude (`claude:sonnet:high`). Sol is not available, so the suggested `sol:med` is declined.
- Model mapping for this sprint: Searcher `claude:haiku:med`, developer `claude:sonnet:high`, architect skipped because the Searcher findings are narrow. The CLAUDE.md convention `codex:sol:med` for event-routing tickets is not applied: Sol is not available here, so the whole sprint stays on Claude models. Luna High (`luna:high`) is reserved for one optional one-shot second opinion, not for implementation.

## M2 review and M3 pre-work (host, 2026-10-30)

- **M2 (commit `34eda3d`) host-verified:** `go test ./internal/editor -count=1 -skip 'Test287SearchHighlightKeepsEditingCursor|Test287SearchRevealsOffscreenMatchWithoutMovingCursor|Test287EditPTYSearchHighlightKeepsCursorStatus'` passes. The three held cursor tests fail as expected until M3. `make test-q1` has not run since M2; it is owed at the final gate.
- **Assertions relaxed in M1 and held for M3:** the reopen and navigation tests in `edit_287_test.go` (about lines 195–196 and 216–217) drop their cursor checks. The M3 commit must restore those cursor checks.
- **Coverage caveat:** the tiny-terminal probe (about lines 301–307) skips sizes where the panel is painted and skips the third line at 4×7. Keep the `tested == 0` guard. Do not weaken it further.
- **Flaky-risk:** PTY test 8's check right after typing the query can run before the frame redraws. Use the existing wait helper, not a sleep, and keep the check after `search_next` as the red assertion.
- **Commit trailer:** commit `34eda3d` carries `Co-Authored-By: Claude Sonnet 5.5`, not the Haiku line the host asked for. Left as-is: amending is off the table. Flagged to the user.

### M3 pre-work

- Remove the cursor assignment at `internal/editor/edit.go` (highlightSearchMatch, about 379–380).
- Add `RichTextHighlight.Reveal` in `richtextedit_render.go` and `richtextedit.go`, with the draw change from the developer's plan (keep the highlight in view while the editor is unfocused, do not snap the cursor).
- Update `docs/Widgets.md:20` to describe the reveal behaviour and keep "display-only, does not change selection or cursor".
- Add a `richtextedit_test.go` test for reveal and for display-only highlights.
- Restore the cursor assertions removed in M1 (see above).
- Move the held cursor file into the M3 commit; all tests in the package must pass.

## M3–M4 review and installed-binary acceptance (host, 2026-10-30)

- **M3 (`c373a58`) host-verified:** `go test -count=1` passes for `internal/editor`, `internal/apitest`, the root package and `cmd/loom`. Assertion changes reviewed: the cursor checks are restored, and the navigation test now parks the cursor where a move is visible.
- **M4 (`452c8d9`) host-verified:** the same four packages pass. `DeferPopover` and `DrawPopover` are additive. The editor calls `DrawPopover` outside the search-panel block, so the popover still draws when search is closed. The probe loosening (a cell with no background of its own may show the background beneath) is accepted: text and foreground are still compared for every cell.
- **Decision (selection):** a click on the search panel while the popover is open clears the selection, as a backdrop click does (`Test304`, `docs/Widgets.md`). The stricter "selection unchanged" reading is rejected.
- **Decision (snapshot):** the `testdata/api/loom.txt` updates for M3 and M4 are additive and approved.
- **Gate blocked:** `make test-q1` fails in `internal/session` on `TestServerLastTabExitCleansSocket`. Reproduced in isolation: it hangs until the 90s test timeout. The test sets its own `LOOM_SOCKET_DIR`, so the live session server is an unlikely cause. Filed as [393](393-fix-hang-in-testserverlasttabexitcleanssocket-in-internal-session.md). This ticket cannot close until `make test-q1` is green.
- **Installed binary (`make install`, tmux 100×24, `loom edit --mousegrab`), host-run:**
  - Query click at row 3, column 60, typed `X` → `Find: X`. Save leaves the file unchanged (`first line`, `second line`, `third line`). PASS.
  - Mode clicks: `Regex` at row 4, column 72 switches to Regex; `Normal` at row 4, column 58 switches back. PASS.
  - Stale matches: query `line` shows `1 / 3`. After closing search and inserting `line ` at the start of line 2, reopening shows `1 / 4`. PASS.
  - Popover: a selection opens the formatting popover, drawn left of the search panel. The overlap case is not reproduced in this run; the M4 probe covers it. PARTIAL.
  - Hint-bar controls: a click on `^F Search` (row 23, column 15) closes an open search panel, and the file stays unchanged. PASS. A click on `^O Files` (column 12) opens the file panel. Both were run in a fresh PTY.
  - Focus on hover and scroll: `Test287HoverAndScrollDoNotChangeFocus` (`edit_287_test.go:449`). Unit test only; not exercised in the PTY run. PARTIAL.
  - Search-bar controls other than the mode labels (prev/next, close): no dedicated test and not exercised in the PTY run. NOT RUN.
- **Acceptance status (updated 2026-10-30, after 393 closed):** the `make test-q1` gate is no longer blocked: it passed on the host run after 393's fix (`4a19821`, `67f922a`). Acceptance is still not met in full, so the ticket stays open. Remaining: a PTY run of the popover-overlap case, focus on hover and scroll in a PTY, and the other search-bar controls (prev/next, close). Close only after those pass.

## PTY round 2 (host, 2026-10-31, installed binary from 3d04ef2)

- Hover and scroll over the search panel, focus in the editor: keys still go to the document. PASS.
- Hover and scroll over the editor, focus in search: keys still go to the query (`Find: Q`). PASS.
- Popover overlap, FAIL: search open, select on row 2 from col 40 to col 48. The formatting popover opens on row 4 and reaches the panel's mode row. The first click on Regex (row 4, col 72) only dismisses the popover; the mode stays `Normal *`. A second click toggles it to `Regex *`. The document is unchanged. Reproduces M4: the popover modal consumes the press before `searchRect`.
- Other search-bar controls: the panel has no clickable prev, next or close buttons. `Enter: next` and `S-Enter: prev` are text hints, and close is `^F`. Still to confirm in code that `edit.go:848–861` has no hit regions for these, and that clicking the hint text does nothing.

## Decision (round 2, host, 2026-10-31)

- A press on the search panel while a popover is open dismisses the popover and is then handled by the panel, so the click reaches the control it hit. Reason: the M1 acceptance says panel clicks reach their action and do not fall through. Backdrop clicks outside the panel keep the existing dismiss-and-consume behaviour (Test304). Low-risk and reversible, recorded here.
- Searcher audit (Claude Haiku, read-only) answers the hint-text and hit-region question, and confirms the modal path. Developer (Claude Sonnet) then fixes the modal contract in the library, not in the editor.
