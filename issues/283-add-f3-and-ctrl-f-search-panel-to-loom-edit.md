# 283 — Add F3 and Ctrl-F search panel to loom edit

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [282](282-add-f2-side-panel-with-file-browser-to-loom-edit.md), [normal search design](../docs/design/loom-edit-03-search-normal.ansi), [regex search design](../docs/design/loom-edit-04-search-regex.ansi), [274](274-overlay-and-modal-layer-for-canvas-to-eliminate-local-popup-stacks.md)

---

## 1. Problem & Motivation
`loom edit` needs search on F3 and ^F. Search opens a small panel at the top right, with next/previous keys and a Normal/Regex switch.

## 2. Technical Specification / Findings
PR #16 review found a Unicode crash ([285](285-fix-unicode-literal-search-panic-in-loom-edit.md)), unsaved-change loss ([286](286-prevent-search-panel-quit-keys-from-losing-unsaved-editor-changes.md)), and overlay/selection defects ([287](287-fix-search-overlay-mouse-routing-and-selection-isolation-in-loom-edit.md)). See [Editor](../docs/Editor.md); reverify findings against the current branch before implementation.

Implementation constraint: the [approved design guidance](../docs/design/loom-edit-design-notes.md) allows the final app to differ with standard Loom widget behavior. New UI elements must use standard Loom widgets; small tweaks are acceptable, heavy hacks are not. Always file or link a library issue for larger observed Loom gaps instead of adding app workarounds.

Both shortcuts open/focus the same compact panel. Show query, mode and match position; highlight matches and move the viewport to the selected result. The mockup proposes Enter for next, Shift-Enter for previous, Tab to reach controls and Esc to close/restore editor focus. Store bindings in the spec. Handle empty queries, no matches and invalid regex without losing editor state. Use library-level search/overlay support where needed rather than per-app routing workarounds. Containers own Tab and coordinate translation; start event-routing work on `codex:sol:med`.

## 3. Implementation & Verification Plan
/goal F3 and ^F open a small top-right search panel in `loom edit`, with working next/previous navigation and Normal/Regex modes; verify search and focus behavior, or stop and report when blocked on a user decision or denied permission.

Before implementation, check live code and recent commits. Acceptance: both shortcuts work; literal and regex searches find expected text; next/previous move to and reveal matches; no-match and invalid-regex states are clear; Esc restores editing; resize keeps the panel usable with the F2 browser shown or hidden; searching preserves text and styling. Verify through focused search/integration/PTY checks, `make test-q1` and `make install`.
