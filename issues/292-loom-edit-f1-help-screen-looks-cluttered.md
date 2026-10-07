# 292 — loom edit: F1 help screen looks cluttered

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: UX
**Related**: [269](269-refine-richtextedit-hints-picker-focus-and-help.md), [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md)

---

## 1. Problem & Motivation
The F1 help screen in `loom edit` looks cluttered, dense, and difficult to scan. Key combinations, descriptions, and operational modes lack clean visual hierarchy, proper whitespace, and structured categorization.

This needs to be solved at the library level: provide a clean, reusable help overlay/modal component or layout engine in the library that formats keymaps into clear logical groups (e.g. Navigation, Editing, File Operations) rather than hardcoded monolithic text blocks in the command. Make good use of the spec YAML files (`spec/widgets.yaml` / keymap and theme specs) as the single source of truth for key bindings, descriptions, and styling.

## 2. Technical Specification / Findings
- Audit the current F1 help overlay in `loom edit` and `RichTextEdit`.
- Restructure the help layout to use categorized two-column or grid sections with clear headers, aligned keycaps, and consistent padding.
- Derive shortcut entries and descriptions from the spec (`spec/widgets.yaml` / defaults), ensuring the help screen dynamically reflects defined bindings without code duplication.
- Validate ANSI rendering, box geometry, and responsive sizing using `loom eval` and `loom check-box`.

## 3. Implementation & Verification Plan
- Refactor the library help modal/view to render structured, uncluttered keymap sections.
- Update `loom edit` to use the spec-backed, uncluttered help view.
- Verify with visual evaluation tests and `make test-q1`.

/goal Declutter the F1 help screen in `loom edit` and library help components using structured categorization and spec-backed key definitions, or stop and report when blocked on a user decision or denied permission.
