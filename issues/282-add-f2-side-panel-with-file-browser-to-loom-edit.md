# 282 — Add F2 side panel with file browser to loom edit

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [browser design](../docs/design/loom-edit-02-file-browser.ansi)

---

## 1. Problem & Motivation
`loom edit` needs a toggleable side panel, with F2 as its default keybinding. The first panel content is a file browser for navigating and opening documents.

## 2. Technical Specification / Findings
PR #16 review findings are tracked in [288](288-compose-loom-edit-panels-with-standard-widgets-and-clipped-responsive-layout.md) and [290](290-keep-the-loom-edit-file-browser-usable-after-cancel-and-mouse-activation.md). See [Editor](../docs/Editor.md); reverify findings against the current branch before implementation.

Implementation constraint: the [approved design guidance](../docs/design/loom-edit-design-notes.md) allows the final app to differ with standard Loom widget behavior. New UI elements must use standard Loom widgets; small tweaks are acceptable, heavy hacks are not. Always file or link a library issue for larger observed Loom gaps instead of adding app workarounds.

Reuse the library's file-browser/navigation primitives. F2 shows/hides the panel and resizes the editor; hiding returns focus to editing. Support keyboard navigation, Enter to open and mouse interaction when capture is enabled. Preserve the existing Save/Discard/Cancel protection when switching away from a modified document. Put the default binding in the spec. Containers own Tab focus moves and child-local event translation; fix library flaws in the library. Start event-routing work on `codex:sol:med`.

## 3. Implementation & Verification Plan
/goal `loom edit` has a side panel toggled by F2 whose file browser navigates and opens files without losing unsaved edits; verify focus, routing and resize behavior, or stop and report when blocked on a user decision or denied permission.

Before implementation, check live code and recent commits. Acceptance: repeated toggle works; keyboard and captured mouse select/open files; focus moves predictably; editor expands when hidden; narrow terminals remain usable; switching modified files offers Save/Discard/Cancel. Verify through focused integration/PTY checks, `make test-q1` and `make install`. Mockup proposes the browser on the left; initial visibility is still a design choice.
