# 278 — RichTextEdit toolbar redesign: key caps and ^ short form

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [269](269-refine-richtextedit-hints-picker-focus-and-help.md), [277](277-standardize-click-only-focus-and-cursor-invariants-across-compound-widgets.md)

---

## 1. Problem & Motivation
The bar below the RichTextEdit File menu looks clumsy: keys and labels share one style, the editor row separates with `·` while the gallery row uses `|`, and `Ctrl+Shift+S` takes a lot of space. The user chose design 001 (`docs/data/richtext-toolbar-design-001-keycaps.ansi`; view with `cat`) and asked for the `^S` short form for Ctrl keys and for clickable keys.

## 2. Technical Specification / Findings
Target layout, from design 001 (80 columns; the 40-column version is in the same file):

```
 File  notes.ansi · Modified
  F1  Help  ^S  Save  ^Shift+S  Save as  F7  View
  F8  BG plain  F9  Theme julia256  F10  Quit
```

- **Key caps:** each key is drawn on its own cap background (bold), followed by its label; pairs are separated by spaces, no `·` or `|`. The gallery controls row (`cmd/loom/widgets.go`) uses the same style, so both rows match.
- **Status colour:** the file status in the File row is coloured by state (Unsaved dim, Modified orange, Saved green, Error red in the mockup). Take the colours from the theme/spec, not hardcoded Go values (docs/Spec.md).
- **`^` short form:** Ctrl keys are written `^S`; Ctrl+Shift+S is `^Shift+S` (`⇧` is avoided because its terminal width is ambiguous). Use the same form everywhere the RTE names keys: hotkey row, File menu shortcut column, F1 help.
- **Clickable keys:** clicking a key cap (or its label) runs the same action as pressing the key, e.g. clicking `F7` toggles View/Edit, `^S` saves, `F10` quits the gallery. Clicks only; hover changes nothing (see 277).
- **Narrow widths:** drop whole key/label pairs from the end (Save as first), never clip a cap; theme/BG names drop before keys. Keep the gallery's active theme name visible at normal width (269 M3).
- **Library, not caller:** `HotkeyHint(width) string` cannot carry styles or click regions. Add a reusable library hint bar (structured key/label/action entries, rendering, hit-testing, atomic fitting) used by both RichTextEdit and the gallery, instead of styling strings in each caller. If `HotkeyHint` changes or goes away, add a row to `docs/Upgrading.md`.
- Clicks feed key actions into the widget, so this touches event routing: per AGENTS.md, start on `codex:sol:med`.

## 3. Implementation & Verification Plan
/goal The RTE bottom bar matches design 001 with `^` Ctrl short forms, coloured status and clickable key caps, built on a reusable library hint bar shared with the gallery; stop and report when blocked on a user decision or denied permission.

Acceptance: render tests at 80 and 40 columns compared against the design (no clipped caps at any width 1–120); click tests for each key cap, including the gallery row; hover does nothing; menu and F1 help use the `^` form; installed-binary check of `loom widgets --show RichTextEdit` at normal and narrow widths; `make install`.
