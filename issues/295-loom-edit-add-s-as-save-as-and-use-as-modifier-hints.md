# 295 — loom edit: Add ⌃⌥S as Save as and use ⇧⌃⌥ as modifier hints

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Feature
**Category**: Keybindings
**Related**: [268](268-add-save-and-save-as-to-ricktextedit-with-a-file-menu.md), [269](269-refine-richtextedit-hints-picker-focus-and-help.md), [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md)

---

## 1. Problem & Motivation
In `loom edit`, "Save as" should be accessible via the standard keyboard shortcut `⌃⌥S` (Ctrl+Alt+S / Option+Ctrl+S), and modifier keycap hints throughout the UI should consistently use standard Unicode modifier symbols (`⇧` for Shift, `⌃` for Ctrl, `⌥` for Alt/Option).

Solve this at the library level: provide centralized modifier glyph formatting and key binding resolution in the keyboard/keycap library package. Ensure all shortcut definitions and modifier presentation rules are specified in `spec/` YAML files (`spec/widgets.yaml`, `spec/defaults.yaml`), eliminating ad-hoc string formatting in individual widgets.

## 2. Technical Specification / Findings
- Implement/update key event matching in the library to support `⌃⌥S` (Ctrl+Alt+S) for "Save as".
- Standardize modifier symbols across the library: `⇧` (Shift), `⌃` (Ctrl), `⌥` (Alt/Option).
- Update the keycap / hint renderer in the library to format modifier combinations using `⇧⌃⌥` glyphs.
- Keep spec files (`spec/widgets.yaml`, `spec/defaults.yaml`) as the single source of truth for modifier formatting and keymap definitions.

## 3. Implementation & Verification Plan
- Implement `⌃⌥S` binding for Save As in `RichTextEdit` and verify modifier parsing.
- Update hint bars and help overlays to render `⇧⌃⌥` symbols.
- Add unit tests for modifier glyph rendering and shortcut dispatching.
- Verify with `make test-q1` and `make install`.

/goal Add `⌃⌥S` for Save As and standardize modifier keycap hints to `⇧⌃⌥` across the library and `loom edit` backed by spec definitions, or stop and report when blocked on a user decision or denied permission.
