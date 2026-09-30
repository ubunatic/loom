# 178 — Research: compare loom to ncurses and list feature gaps

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Research
**Related**: issues/180 (roadmap), issues/177-179

---

## 1. Problem & Motivation
ncurses defines the classic terminal capability baseline (terminfo, panels, forms, menus, color pairs, mouse, input modes). We want a list of features ncurses offers that loom lacks, so the roadmap (180) can close the gaps that matter.

## 2. Scope & Rules
- Stay true to loom's design (read AGENTS.md, docs/Widgets.md, README, `loom widgets`): a gap is only a gap if it fits loom's model. Features that contradict the design go in a "Deliberately not adopted" list with one-line reasons.
- Check live code before claiming a gap; cite loom file/symbol for what exists.
- Cross-reference existing open tickets (`harnez find -d . issues -a status:open`, e.g. 155 modal, 158 keymap, 167-173 widgets) instead of re-listing them as new.
- Output: append a `## Findings` section to this ticket: a table `Feature | <other> | loom today | Gap? | Fits design? | Existing ticket | Size (S/M/L)`, then the not-adopted list. No code changes.

## 3. Implementation & Verification Plan
/goal The Findings section lists every relevant ncurses feature with loom's status and a design-fit verdict, committed as `docs(issues): findings for 178`. Stop and report when blocked on a user decision or denied permission.

## Findings

ncurses is a terminal-independent screen and input library built around terminfo. The comparison below uses the ncurses overview and terminfo reference in the [ncurses(3x) manual](https://man7.org/linux/man-pages/man3/ncurses.3x.html) and [terminfo(5) manual](https://man7.org/linux/man-pages/man5/terminfo.5.html). Loom's design is an inline, dependency-light Go widget library; its current screen modes and composed layouts extend beyond the README's typical inline use.

| Feature | ncurses | loom today | Gap? | Fits design? | Existing ticket | Size (S/M/L) |
|---|---|---|---|---|---|---|
| Terminal capability database | Uses terminfo to select terminal-specific control sequences and capabilities. | `pane.go` (`Pane`, `termSize`, `flush`) and `style.go` (`Color` sequences) use `x/term` sizing and direct ANSI/SGR sequences; no terminfo lookup. | Yes: no capability-based fallback. | Partial: useful for portability, but a database-backed terminal abstraction expands the small inline runtime. | None found | L |
| Screen, windows, and pads | Maintains a virtual screen, independent windows, and off-screen pads for large scrollable regions. | `canvas.go` (`Canvas`) is a cell buffer; `Frame`, `Box`, `Grid`, `Split`, and `View` provide composed regions and scrolling. There is no general pad abstraction. | Partial: no independent off-screen pad API; layout regions are present. | Yes: a bounded scrollable surface could compose as a widget. | 156 (canvas/subcanvas/blit layering) | M |
| Overlapping panels | `panel` adds ordered, overlapping windows and hides/shows/reorders them. | `Stack` and `Popup` compose layered widgets; `Canvas` supports foreground, surface, and decoration painting (`canvas.go`). | No core gap; panel ordering is handled by loom composition. | Yes; already represented in loom's widget model. | 155 (modal/dialog overlay) | S |
| Forms and fields | `form` provides field placement, traversal, validation hooks, and field display behavior. | `TextInput`, `TextArea`, and `Settings` cover individual and settings inputs; no general form/field-group validation-summary composite is listed by `loom widgets`. | Yes: no general form composite. | Yes: composable input widgets fit the library model. | 169 | M |
| Menus | `menu` provides item selection, menu posting, and menu navigation. | `Choice` selects/filter lists and `Tabs` switches panels; no nested menu/menu-bar widget is listed by `loom widgets`. | Yes: nested menus, menu bars, and accelerators are missing. | Yes: reusable input widgets fit the model. | 170 | M |
| Mouse events | Supports mouse reports where the terminal provides them, with configurable event masks and coordinates. | `Pane.EnableMouse` enables SGR tracking (press, release, hover, scroll); widget/container dispatch and local coordinates are established (`pane.go`, `docs/Widgets.md` §8). | No common interaction gap; support is specifically SGR-based. | Yes; already implemented. | None found | S |
| Colors and attributes | Uses terminal capabilities, color availability and color pairs; supports terminal-defined attributes and color configuration. | `style.go` (`ColorIndex`, `ColorRGB`, `Style`) emits 256-color and true-color SGR; styles include bold, underline, and dim. It has no color-pair allocator or terminfo-derived attribute limits. | Partial: no pair allocation or capability negotiation; direct styles cover ordinary use. | Yes: capability-aware fallback could improve terminal portability. | None found | M |
| Keyboard and input modes | Uses raw/cbreak/echo/keypad modes; translates terminal sequences to named keys and supports ncurses input timing options. | `Pane` owns/restores terminal state via `x/term`; the decoder handles supported key sequences, and `docs/KeyDefaults.md` documents xterm-compatible function/modified cursor limitations. No configurable ncurses-style keypad/timing API. | Partial: terminal-specific key decoding and richer binding aliases remain limited. | Yes: key binding configuration suits widgets and hosts. | 158 (KeyMap/action aliases) | M |
| Terminal resize | Handles size changes and reflows windows in response to resize signals. | `Pane.applyWinch` re-queries terminal size, adjusts pane/canvas bounds, and redraws after SIGWINCH (`pane.go`). | No. | Yes; implemented. | None found | S |
| Alternate character set and wide characters | Provides ACS line-drawing symbols and wide-character APIs subject to terminal support. | Loom renders Unicode/specced box glyphs and tracks display-width cells (`canvas.go`, `spec/box.yaml`, `measure/`); ANSI parsing and drawing preserve styled cells. | No common gap; ACS-specific legacy symbol APIs are absent. | Yes; Unicode and spec-driven glyphs fit loom. | None found | S |

### Deliberately not adopted

- **Full terminfo/termcap compatibility as the default rendering contract** — loom's documented baseline is a lightweight inline widget library using ANSI-capable terminals; supporting every legacy capability would substantially broaden that contract. A focused fallback remains a possible portability improvement (see the terminfo row).
- **A curses-compatible procedural API and its global screen state** — loom's `Widget`/`Canvas` composition is the intended programming model; importing curses' global-window lifecycle would not improve widget composition.
- **ncurses forms/menu C APIs and ABI compatibility** — the reusable behavior is tracked as native loom widgets in tickets 169 and 170, rather than binding the C library or copying its API surface.
