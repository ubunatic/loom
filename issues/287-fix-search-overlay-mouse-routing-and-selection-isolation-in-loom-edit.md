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
/goal Search UI receives its own input and highlights matches without changing editing selection or accidentally editing the document; verify the behavior, or stop and report when blocked on a user decision or denied permission.

Acceptance: Cover query/mode/control clicks, overlay occlusion, focus changes, stale matches after edits, and saved document integrity. Verify real PTY interaction, make test-q1 and make install.
