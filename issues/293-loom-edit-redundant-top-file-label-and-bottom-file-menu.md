# 293 — loom edit: Redundant top File label and bottom File menu

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: UX
**Related**: [268](268-add-save-and-save-as-to-ricktextedit-with-a-file-menu.md), [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [288](288-compose-loom-edit-panels-with-standard-widgets-and-clipped-responsive-layout.md)

---

## 1. Problem & Motivation
`loom edit` currently displays a "File" label on the top header/status bar and a "File" menu item in the bottom menu bar simultaneously. Having both creates visual redundancy and ambiguous affordances for file actions.

Solve this problem on the library level: define clear roles and compositional patterns for top-level headers, title pills, and menu bars in compound container widgets. Leverage `spec/widgets.yaml` and layout specs to establish standardized, non-redundant header/menu layouts across all loom tools and widgets.

## 2. Technical Specification / Findings
- Inspect the top title/header bar and bottom menu bar in `loom edit` and `RichTextEdit`.
- Determine the correct division of responsibility: top bar typically conveys file path/name and state (clean/modified), while the menu bar or action bar provides interactive commands (File menu, Search, etc.).
- Update library container templates so standard layout specs avoid repeating static menu names in header title positions.
- Align layout definitions with `spec/widgets.yaml` and `spec/defaults.yaml`.

## 3. Implementation & Verification Plan
- Refactor the top header in `loom edit` and `RichTextEdit` to display the active document path/title without redundant "File" labeling.
- Ensure the bottom menu bar remains the sole interactive "File" menu trigger.
- Add/update visual layout tests and verify with `make test-q1`.

/goal Eliminate the redundant "File" label/menu duplication in `loom edit` and library editor containers with spec-aligned header and menu composition, or stop and report when blocked on a user decision or denied permission.
