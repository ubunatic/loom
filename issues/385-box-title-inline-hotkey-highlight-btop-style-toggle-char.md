# 385 — box title: inline hotkey highlight (btop-style toggle char)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: harnez usage --watch (btop-style box toggles)

---

/goal A loom `Box` title can mark one character (or a short key badge) as its hotkey and draw it in its own
style, btop-style (`¹cpu`, `menu` with a red `m`, `preset` with a red `p`), with the rest of the title in the
normal title style; verified by a per-cell canvas test (glyph, FG, BG) and a gallery entry; or stop and report on
an API decision. Circuit breaker: on a loop or no progress, clear the goal and write a failure report here.

## Motivation

harnez usage wants btop-style toggles in its box titles: the key letter inside the title word is coloured
(e.g. red `m` in `menu`), or a superscript digit precedes it (`¹cpu`). Today `Box.Draw` (`box.go:227-238`)
writes `TitlePrefix + Title + TitleSuffix` in one `Style.Title`, so a single char cannot be styled. Only the
toggle char styling is wanted, not btop's frame style.

## Notes

- Design point: a hotkey index/rune on `Box` plus a `Style.TitleKey` (theme/spec colour, semantic name), or a
  small markup in the title. Prefer the field: no escaping, measurable width.
- Hotkey letter not in the title (e.g. key `L` for "Load" works, `d` for "Fetch diagnostics" does not): fall back
  to a prefixed key badge in the key style.
- Width must stay display-width correct (superscript digits, wide runes); YAML/spec fields as for other Box
  fields.
