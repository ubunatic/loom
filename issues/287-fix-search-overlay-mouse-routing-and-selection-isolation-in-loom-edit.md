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
cmd/loom/edit.go:623–660 does not route searchRect before editorRect. Synchronize widget focus, prevent mouse hover/scroll from stealing focus, and render search highlights separately from editing selections/popovers. Refresh matches after document edits and restore usable editor focus when closing search. Use standard Loom widgets and container-owned routing; file/link library gaps rather than app hacks. Start on codex:sol:med.

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

- Model mapping for this sprint: Searcher `claude:haiku:med`, developer `claude:sonnet:high`, architect skipped because the Searcher findings are narrow. The CLAUDE.md convention `codex:sol:med` for event-routing tickets is not applied: Sol is not available here, so the whole sprint stays on Claude models. Luna High (`luna:high`) is reserved for one optional one-shot second opinion, not for implementation.
