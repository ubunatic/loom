# 290 — Keep the loom edit file browser usable after cancel and mouse activation

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: [PR #16](https://github.com/ubunatic/loom/pull/16), [282](282-add-f2-side-panel-with-file-browser-to-loom-edit.md), [283](283-add-f3-and-ctrl-f-search-panel-to-loom-edit.md)

---

## 1. Problem & Motivation
PR #16 embeds FilePicker without OnSelect/OnCancel callbacks. FilePicker mouse activation sets its done state without opening a document, and Esc sets done without closing/resetting the side panel. Subsequent keyboard navigation is ignored. Keyboard Enter is separately intercepted by the app, so input paths behave differently.

## 2. Technical Specification / Findings
Connect file activation and cancellation through the supported widget lifecycle. Keyboard and mouse activation must use the same unsaved-change guard, report open/save failures, and leave the browser reusable. File/link library lifecycle gaps instead of per-widget workarounds. Start on codex:sol:med.

Observed at PR head b8a59ca; this branch has not been merged locally. Before implementation, check live code and recent commits and reverify the finding.

Updated review (2026-10-07, PR head a9f1a7d): newEditView still constructs FilePicker without OnSelect/OnCancel, and browser key/mouse handling is unchanged. FilePicker's done state still suppresses keyboard events after Esc or mouse activation. The new tests cover search only, not these picker lifecycle paths. Keep open; the browser finding remains unresolved.

## 3. Implementation & Verification Plan
/goal The editor file browser opens files consistently with keyboard and mouse and remains usable after cancel or repeated activation; verify the behavior, or stop and report when blocked on a user decision or denied permission.

Acceptance: Cover actual FilePicker activation, cancel/reopen and repeated navigation, with modified documents and Save/Discard/Cancel. Verify focus and real captured mouse behavior, make test-q1 and make install.
