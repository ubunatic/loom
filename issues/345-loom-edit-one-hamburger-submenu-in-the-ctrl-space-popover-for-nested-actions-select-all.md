# 345 — loom edit: one hamburger submenu in the Ctrl+Space popover for nested actions (Select all)

**Status**: Closed — Done in df86d50: one specced ≡ entry after Box opens a submenu with Select all (reuses 344 select_all), keyboard and mouse, rule for new actions in docs/Widgets.md, F1 help, gallery/.ansi updated; q1 green; live tmux check
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 344 (select-all keys), 342 (Ctrl+Space toggle)

---

/goal Add exactly one hamburger (`≡`) entry to the RichTextEdit popover bar that opens a submenu of
less common actions, starting with "Select all"; spec, test and document it; run `make test-q1` and
`make install`, or stop and report when blocked on a user decision or denied permission.

## Decision (user, 2026-10-20)

Keep the top-level Ctrl+Space popover lean: today `B I U S Link #FG #BG Box`
(spec/defaults.yaml `popover_labels`, ~l.306). New actions go into the single hamburger submenu,
not onto the bar. Record this rule in docs/Widgets.md (popover section, ~l.365–390).

## Plan

- Add the hamburger label/glyph and its submenu items to the spec (no hardcoded labels in Go).
  First item: "Select all", sharing the action behind the 344 keys (`ctrl-alt-a`, `alt-a`, `super-a`);
  if 344 is not done yet, add the select-all action here and let 344 bind keys to it.
- Reuse the existing submenu mechanics of "Box" (keyboard: arrows/Tab/Enter/Esc; mouse; Ctrl+Space
  closes per 342). Fix in the library widget, not `cmd/loom`. Event routing: start on `codex:sol:med`.
- Check popover width (`popover_min_width`) and narrow terminals; validate any changed ANSI
  rendering with `loom eval`, `loom measure`, `loom check-box`.
- Tests: open hamburger via keyboard and mouse, Select all selects the whole document, Esc returns to
  the bar, Ctrl+Space closes all. Update F1 help and docs/Widgets.md.

Open: which glyph (`≡` or `☰`; check display width) — pick `≡` unless it measures wrong.
