# 457 — Research TUIOS and the Charm stack against loom

**Status**: In Progress — owner said go; 7 haiku:med searchers started
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Research
**Related**: 417, 454, 456

---

## 1. Problem & Motivation
Owner request: study TUIOS (tuios.dev), a top-notch Go TUI app, and the libraries it uses
(Bubble Tea, Bubbles, Lip Gloss, Huh, Glamour, Charm x/* and ultraviolet), plus other TUI libs,
to learn what loom should do next. Goals:
- a list of all features the libs provide, with where loom stands;
- how TUIOS uses the libs and where it needs workarounds;
- best practices for modern Linux terminal app programming (VTE, kitty, foot, ...);
- a comparison and quality assessment of each lib against loom.

loom aims to be a mouse-enabled GUI-like SDK with Ctrl/Alt keys for features, not vim/tmux-like.
Plan drafted by the host and refined by argon:med (advisor). Not yet signed off by the owner.

## 2. Technical Specification / Findings

### Searchers (7 x haiku:med, web search, read-only, parallel)
Each prompt carries a ~15-line loom baseline (37 widgets from `spec/widgets.yaml`, Ctrl+C/V/X/Z,
Shift-select, kitty keyboard, OSC 52 write, `Pane.Post`/`Invalidate`/`Suspend`, `ModalTarget`,
pane mouse capture, wheel curve, `loom session` + `vt.VT`, `measure` ZWJ probes), so searchers
spend their budget on differences.

Output in `docs/studies/tuios/` (uncommitted until host review):
- `report-<topic>.md`: max ~300 lines, code examples where useful.
- `features-<topic>.yaml`: one file per searcher; the host merges them into `features.md`.

```yaml
- id: textinput.word-jump
  area: input|widgets|render|events|desktop|theme|compose|layout|perf|test|a11y|session|other
  feature: Ctrl+Left/Right word jump in text input
  lib: bubbles/textinput
  version: v1|v2|ultraviolet|main@<sha>       # never mix Bubble Tea v1 and v2
  symbol: textinput.Model.KeyMap.WordForward  # exact Go symbol or escape bytes
  source_url: https://github.com/.../file.go#L10-L30   # file+lines or issue/PR, never a repo root
  evidence_quote: "WordForward key.Binding"   # verbatim, max 120 chars, from the fetched page
  verified: true|false
  loom: unknown                               # host fills has|partial|missing
  priority: P1|P2|P3
  note: one line
```

### Topics
1. **TUIOS** (tuios.dev/loom first: what it is, name clash?): architecture, window manager,
   terminal emulation, PTY multiplexing, attach/detach, scrollback, copy mode, mouse passthrough,
   drags leaving bounds, use of each Charm lib, workarounds (TODOs, forks, `replace` in go.mod).
   Compare with `loom session`.
2. **Bubble Tea v1/v2**: Elm model, key/mouse model, kitty keyboard, focus, paste, resize,
   renderer, suspend, high-frequency update backpressure (vs `Pane.Post`), pointer capture, `teatest`.
3. **Bubbles + Huh**: all widgets and knobs vs loom widgets (TextEntry, Form, FilePicker, Table,
   Choice, Tree, DatePicker).
4. **Lip Gloss v2 + Glamour**: layers, Z-order, popovers, borders, styles, theming, Markdown;
   vs loom RootOverlays and astra.
5. **Charm low-level** (ultraviolet, x/ansi, x/term, x/input, x/vt, colorprofile, cellbuf):
   cell buffer and diff flush, width (mode 2027 vs `measure.DetectZWJMode`), capability and
   color detection, SIGWINCH, split-CSI decoding, golden/VT testing.
6. **Modern terminal best practices**: kitty keyboard, modes 2026/2027/2031, OSC 52 read/write,
   primary selection, OSC 8/9/10/11/133/777, focus reporting, SGR-pixel mouse, DA1/XTVERSION,
   kitty graphics/sixel, foot/kitty/ghostty/wezterm/VTE/alacritty/tmux differences, crash restore.
7. **Contrast**: tview/tcell, gocui; ratatui or textual only for features no Go lib has.

### Priorities for loom
- P1: desktop-like non-modal input in every input widget; widget knobs and polish (virtualised
  Table/Choice, fuzzy ranking, inline cell edit); robustness (SIGWINCH, cascading event capture,
  emoji width across terminals, crash restore, split escapes, huge paste); composition ("claim
  cell"/astra, Z-layers, popovers without boilerplate, mode 2026, minimal diff); testability;
  `loom session` parity with TUIOS.
- P2: theming and dark/light detection; a11y and focus; performance with large data.
- P2-P3: desktop integration (terminal/font/theme detection, clipboard read, notifications, links).
- P3: images, animation, bell, i18n/RTL.

Out of scope: code in other repos, loom code changes, tickets filed by searchers.

## 3. Implementation & Verification Plan
- M0: owner signs off this plan (Status Draft until then).
- M1: run the 7 searchers in the background (`harnez agent start --model haiku:med`).
- M2 host: reject rows with repo-root URLs, no symbol or no quote; fetch all P1 URLs and check
  quotes; fill `loom` from docs/, spec/ and code, not from memory.
- M3 host: merge into `docs/studies/tuios/features.md` (priority order, gaps first) and write
  `summary.md` (quality of each lib vs loom: API, tests, docs, backlog, release cadence;
  top 10 gaps; proposed tickets). Show the owner; file tickets only after sign-off.

/goal Deliver the verified reports, merged feature list and summary in `docs/studies/tuios/` and
propose tickets, then stop; or stop and report when blocked on owner sign-off or a denied
permission. On a loop or no progress, clear the goal and write a failure report here.
