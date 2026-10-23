# 373 — loom session: bar with hotkey hints and mouse status, always shown

**Status**: Closed — Session bar always shown (spec session.bar), spaced tabs, ⌃B… Tools, ? Help, ⌃P 📷, ⌃Q Quit, ▏🖱︎ icon; prefix ? help from spec, prefix m per-tab mouse toggle, icon click off; narrow widths drop hints keep icon (a1b5be6, 85d150a, b4ae530). Host q1 green; live tmux 100 and 40 cols bar correct, help shows, m sets mouse flags on
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

## Handoff (d373, 2026-10-23)

### Relevant Files and Line Ranges
- `cmd/loom/session.go`:
  - `drawSessionTabBar` (L121–145): Currently checks `if len(tabs) <= 1 { return }` and renders a basic contiguous tab bar with DEC cursor save/restore (`\x1b7` / `\x1b8`). Needs refactoring to draw spaced tabs, hotkey hints in `loom edit` style, dynamic spacer, and mouse status icon; must support `session.bar: always` (always drawing even with 1 tab).
  - `tabIndexAtCol` and bar hit detection (L147–159, L390–445): Needs updating for spaces between tabs and right-hand clicks on the mouse toggle icon (`🖱︎`).
  - Terminal rows & child PTY sizing (L250–270, L330–360, L832–838): Child terminal must get `rows - 1` whenever the bar is displayed (`session.bar: always` default means child gets `rows - 1` even with 1 tab; `session.bar: tabs` means child gets `rows - 1` only when `len(tabs) > 1`). Erase bar logic on tab close must align.
  - Prefix state machine and key dispatch (L490–650): Handle prefix `?` (session help) and prefix `m` (toggle mouse mode between safe/off and active).
  - Help overlay / display: Implement prefix `?` help view rendering all specced session bindings, dismissing on any keypress and cleanly redrawing the child screen.
- `spec/defaults.yaml` (L715–735): Add `session.bar`, prefix bindings for `help` and `toggle_mouse`, and bar labels/glyphs.
- `defaults.go` (L56, L637–680): Update `SessionDefaults`, `SessionPrefixBindings`, and `SessionDefaults.validate()` to parse and validate new spec keys.
- `spec/schemas/defaults.schema.json`: Update JSON schema to define new session keys and prefix bindings.
- `hintbar.go` (L12–80), `keycap.go` (L18–50), `theme.go` (L127–133): Reusable helpers and styling primitives (`KeyCap`, `HintBarStyle`, `fit`, `DropPriority`).
- `cmd/loom/session_pty_test.go`: PTY tests covering single-tab and multi-tab bar rendering, narrow terminal (40-column) drop behavior, prefix `?` help popup, prefix `m` mouse toggle, and mouse click toggle.

### Loom Edit Bar Helpers and Spec Colors
- Keycaps: `loom.KeyCap(binding)` converts bindings (e.g. `ctrl-p` -> `⌃P`, `ctrl-q` -> `^Q` or `⌃Q`) according to `SpeccedDefaults.KeyCaps.Modifiers`.
- Hint bar layout & drop priority: `loom.HintBar` and `loom.HintEntry` in `hintbar.go` provide responsive layout (`fit`), hit detection, and drop prioritization (`DropPriority`).
- Spec colors:
  - `theme.HintBarStyle()`:
    - Keycaps: `Cap: Style{FG: t.KeyCapFG.Color(), BG: t.KeyCapBG.Color(), Bold: true}`
    - Labels: `Label: Style{FG: t.NormalFG.Color(), BG: t.NormalBG.Color()}` (dim text relative to bold keycaps).
  - Tabs: `loom.DefaultTabsStyle()` (`TabsStyle{Active, Inactive}`) with 1 space separator between tabs.
  - Mouse icon: Text-presentation `🖱︎` (VS15 `\U0001F5B1\uFE0E`), bright style when mouse mode is on, dim style when off, preceded by separator `▏` (`\u258f`).

### Proposed Spec Keys
- `session.bar`: String enum, `"always"` (default) or `"tabs"`.
- `session.prefix_bindings`:
  - `help`: `["?"]`
  - `toggle_mouse`: `["m"]`
- `session.bar_labels` / `session.bar_glyphs` (or corresponding `session` spec fields):
  - `tools_label`: `"Tools"`
  - `help_label`: `"Help"`
  - `quit_label`: `"Quit"`
  - `screenshot_glyph`: `"📷"`
  - `mouse_icon`: `"🖱︎"`
  - `separator`: `"▏"`

### Open Risks
- Emoji display width & VS15 portability across terminals: `🖱︎` (`\U0001F5B1\uFE0E`) and `📷` (`\U0001F4F7`) cell measurements. Must conform to `docs/EmojiWidth.md` rules and pass `loom measure` / `loom eval` without ragged rows or cursor/click-target desynchronization.
- Child PTY geometry & SIGWINCH: When `session.bar: always`, child rows must be `rows - 1` from the start, even with only 1 tab. Existing single-tab behavior gave the child all `rows`. Any mismatch between child terminal window size and terminal erase/scroll margins (`DECSTBM`) will cause flicker or line overwrite on row `rows`.
- Click hit targets vs child mouse event forwarding: When mouse mode is active, clicks on row `rows` belong to the session bar (tab switching, mouse toggle click), while clicks in rows `1..rows-1` must be passed to the child if the child requested mouse tracking. When mouse mode is toggled off (prefix `m` or click on icon), loom stops consuming mouse events and enters safe/off mode.
- Help screen overlay lifecycle: Prefix `?` needs to display session bindings cleanly without disrupting the running child PTY screen. Must cleanly redraw child screen (or use alt screen / save & restore VT buffer) when exiting help on any key.
- Narrow terminal degradation (40 columns): Right-hand hints must be dropped in priority order (e.g. spacer shrinks, search/screenshot/quit drop before tabs are clipped) to prevent ANSI wrap-around and line corruption.

## Approved plan (host, 2026-10-23)

d373s plan approved: spec keys (bar, help, toggle_mouse, labels/glyphs/drop priorities) + schema; session bar composed from HintBar/KeyCap/Tabs styles with layout shared by hit testing; one geometry function for child rows; help overlay from spec, child VT keeps updating, repaint on dismiss; per-tab mouse override shared by prefix `m` and icon click. Milestones: M1 spec/bar/geometry, M2 help, M3 mouse. Pre-work: the mouse icon stays visible at 40 columns (drop hints first); icon click only exists while loom's mouse is on (in safe mode with a plain shell no tracking is enabled); the bar must never wrap at any width ≥ 10.
