# 388 — box title: styled title segments with inline hotkey highlight

**Status**: In Progress
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: harnez usage --watch (btop-style box toggles)

---

/goal A loom `Box` title supports ordered, independently styled segments, including a hotkey character (or
short key badge), bold badge text, normal title text, and dim host/source metadata; verified by per-cell canvas
tests (glyph, FG, BG, bold) and a gallery entry; or stop and report on an API decision. Circuit breaker: on a
loop or no progress, clear the goal and write a failure report here.

## Motivation

harnez usage wants btop-style toggles in its box titles: the key letter inside the title word is coloured
(e.g. red `m` in `menu`), or a superscript digit precedes it (`¹cpu`). It also needs independently styled
segments for bold key badges (for example `[L]`) and dim host/source metadata (for example `(@host · source)`).
Today `Box.Draw` (`box.go:227-238`) writes `TitlePrefix + Title + TitleSuffix` in one `Style.Title`, so it
cannot represent these title runs. Preserve display-width clipping across segment boundaries.

## Notes

- A hotkey character is one independently styled title segment; arbitrary ordered title segments are also
  needed for badges and metadata. Prefer structured fields over markup: no escaping, measurable width.
- Hotkey letter not in the title (e.g. key `L` for "Load" works, `d` for "Fetch diagnostics" does not): fall back
  to a prefixed key badge in the key style.
- Width and clipping must stay display-width correct across segment boundaries (superscript digits, wide runes);
  YAML/spec fields as for other Box fields.
