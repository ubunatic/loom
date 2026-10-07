# 295 — loom edit: Add ⌃⌥S as Save as and use ⇧⌃⌥ as modifier hints

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Feature
**Category**: Keybindings
**Related**: [268](268-add-save-and-save-as-to-ricktextedit-with-a-file-menu.md), [269](269-refine-richtextedit-hints-picker-focus-and-help.md), [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md)

---

## 1. Problem & Motivation
Save-as is bound to `ctrl-shift-s`, which most terminals cannot send as a distinct key, so Save-as is effectively unreachable by keyboard. Shortcut labels are inconsistent: `^S`, `^Shift+S`, `Ctrl+B`, `Alt+F` and `Ctrl+Shift+A/E` all appear.

## 2. Technical Specification / Findings
- `spec/defaults.yaml` `pane.hotkey_save_as_key: "^Shift+S"`, `hotkey_save_as_binding: "ctrl-shift-s"`. Also hardcoded in `richtextedit_filebar.go:26` (`"^Shift+S"`) and `richtextedit_help.go:18`.
- Keycap strings (`*_key`) are written by hand per entry in the spec, separate from the `*_binding` they describe, so they drift.
- Fix in the library:
  - Change the Save-as binding to `ctrl-alt-s` (ESC + `^S` on legacy terminals; check that the key decoder emits `ctrl-alt-s` before relying on it, per docs/Canary.md).
  - Add one formatter, `KeyCap(binding string) string`, that turns a binding into glyphs in a fixed order: `⇧` Shift, `⌃` Ctrl, `⌥` Alt, then the key (`ctrl-alt-s` → `⌃⌥S`, `ctrl-shift-z` → `⇧⌃Z`, `f10` → `F10`). Glyphs and order live in the spec (`keycaps.modifiers`).
  - `HintBar`, `MenuBar` shortcuts, `KeyHelp` and the help text (292) use `KeyCap(binding)`. The spec `*_key` fields are removed where they only duplicate the binding. An explicit override stays possible for odd cases like `F3/^F`.

## 3. Implementation & Verification Plan
- Canary: confirm in a PTY that the decoder reports `ctrl-alt-s` for `ESC ^S` and for the CSI-u/modifyOtherKeys form.
- Add `KeyCap` with the spec glyph table and a table test; migrate hint bar, menu and help callers; delete the duplicated `*_key` values and their Go fields.
- Rebind Save-as; update the goldens; PTY test: `⌃⌥S` opens the save picker.
- Add a row to `docs/Upgrading.md` if exported fields or the binding change break callers.
- `make test-q1`, `make install`.

/goal Rebind Save as to ctrl-alt-s and add one spec-driven KeyCap formatter (⇧⌃⌥ glyphs) used by hint bars, menus and help so keycaps derive from bindings, verified by canary, unit and PTY tests, or stop and report when blocked on a user decision or denied permission.
