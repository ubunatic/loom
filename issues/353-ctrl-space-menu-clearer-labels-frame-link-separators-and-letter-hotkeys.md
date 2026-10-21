# 353 — Ctrl+Space menu: clearer labels (Frame, Link, ▏ separators) and letter hotkeys

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: UX
**Related**: 345 (≡ hamburger), 349 (Up/Down entry), 342 (Ctrl+Space toggle), docs/data/cspace-menu-001.ansi (user request)

---

/goal Restyle the RichTextEdit Ctrl+Space bar to the user's layout and add letter hotkeys, all
specced in spec/defaults.yaml, tested (key and mouse), documented (docs/Widgets.md, F1 help), with
gallery and docs/progress .ansi updated and validated; run `make test-q1` and `make install`; or
stop and report when blocked on a user decision or denied permission.

## Request (user, 2026-10-21, docs/data/cspace-menu-001.ansi)

1. Labels. Current: `[ B│ I│ U│ S│ Link│ 🎨│ Box│ ≡]`.
   Wanted: ` B ▏I ▏U ▏S ▏🎨 ▏Frame ▏Link ▏≡ ` with the **F** of Frame and the **L** of Link bold
   (hotkey hint).
2. Hotkeys: `b`, `i`, `u`, `s` for bold/italic/underline/strike, `F` for Frame, `L` for Link.

## Decisions (host, 2026-10-21; all changeable in spec)

- "Box" is renamed "Frame" in the bar label and F1 help only; internal action names, Box submenu
  items (Plain/Rounded/Draw) and the `^D Box` editor hotkey stay unless the label is the only place.
- Order as in the request: B I U S 🎨 Frame Link ≡. Separator `▏` (U+258F, 1 cell; verify with
  `loom measure`), no outer brackets; focus highlight as today.
- Bold hotkey letter: rendered with the bold attribute on the first rune of Frame and Link; specced
  per item (e.g. a hotkey-index or marked label), not hardcoded.
- Hotkeys work while the bar is open (focused or not, right after Ctrl+Space) and act like
  activating the item (toggle style; F/L open their submenu/prompt). Both cases accepted (`b`/`B`,
  `f`/`F`, ...). When the popover was not opened by Ctrl+Space and the bar is unfocused, letters must
  keep typing into the text if the library auto-shows the bar on selection; the plan must check.
- 345 rule holds: no new bar items.
- Link (plan finding: disabled, no handler): stays disabled in this ticket; no URL prompt. `L` is
  wired to the same activation path and does nothing while Link is disabled (it must not type `L`
  when the bar is explicitly open). A Link prompt is a separate feature.

## Verification

Unit tests: rendered bar text and bold cells by display column; each hotkey (both cases) applies the
action; letters still type when expected; mouse hit-testing per new item geometry. Live tmux check of
the installed `loom edit`.
