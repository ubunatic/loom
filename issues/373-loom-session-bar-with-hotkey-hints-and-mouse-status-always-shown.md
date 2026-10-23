# 373 — loom session: bar with hotkey hints and mouse status, always shown

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 371 (tabs), 372 (restore, bar clicks safe mode), 374 (scrollback search),
`docs/data/loom-session-bar-design-doc.ansi` (user design)

---

/goal The `loom session` bar matches the user design in `docs/data/loom-session-bar-design-doc.ansi`:
colored tabs with spacing, hotkey hints in the `loom edit` bar style, a mouse status icon, and a prefix
help; all labels, glyphs and keys come from the spec; validated with `loom eval`/`loom measure`; PTY
tests cover the bar at one and several tabs, the mouse toggle and the help; or stop and report when
blocked on a user decision.

## Decisions (user, 2026-10-23)

- The bar is always shown, also with one tab; spec `session.bar: always` (default) or `tabs` (only with
  more than one tab, the 371 behaviour). The child gets rows-1 whenever the bar shows.
- Layout (left to right): tabs, `^B…` Tools, help hint, spacer, search hint (374), `⌃P` 📷, `^Q` Quit,
  `▏` + mouse icon. Dim text and key highlight are the `loom edit` bar's (reuse its spec colors and
  `KeyCap`), tabs keep their colors and get a space between them. Emojis save space.
- Help is prefix `?` (tmux style), not `F1`, so children keep F1. The hint shows the real binding
  from the spec. Help lists all session bindings (generated from the spec).
- Search hint shows prefix `f` (374), not `ctrl-f`; hidden until 374 exists.
- Mouse icon: text-presentation `🖱︎` (VS15) in both states, bright when loom's mouse handling is on,
  dim when off; check width with `loom measure` and docs/EmojiWidth.md rules.
- Prefix `m` toggles loom's mouse handling (from 372 "safe" mode to on for the active tab); a click on
  the icon while on turns it off. Mouse on in a plain shell means native selection needs Shift+drag.

## Verification

- Golden `.ansi` of the bar at 100 columns with 1 and 3 tabs; `loom eval` + `loom measure` clean.
- Narrow terminal (40 columns): hints drop from the right before tabs are cut; no wrap.
- PTY: prefix `?` shows help and any key closes it; prefix `m` flips the icon and the outer mouse mode;
  clicking the icon turns it off.
