# TUIOS study: summary (ticket 457)

Sources: the seven `report-*.md` files and `features.md` in this folder. Researched 2026-11-10
by seven haiku:med searchers. The host checked the P1 rows against loom code at `163b3738`.

## What TUIOS is built on

- Bubble Tea v2, Lip Gloss v2 and ultraviolet. No Bubbles, Huh or Glamour, no forks, no `replace`.
- `tuios.dev/loom` is TUIOS's renderer for its own TML documents, not a widget SDK.
  No import clash with `ubunatic.com/loom`; say "TUIOS loom" when citing it.
- Workarounds found: Bubble Tea's single `WithFilter` slot (TUIOS sets its own last), the frame
  rate clamp at 120, mouse motion dropped unless a feature needs it, and a cleanup gap when `Run`
  fails before `Init` (waiting on upstream).
- TUIOS is modal (leader key Ctrl+B, vim copy mode). loom stays non-modal; that part does not apply.

## Quality of the libraries compared with loom

| Library | Strengths | Weaknesses | Against loom |
|---|---|---|---|
| Bubble Tea v2 | cell renderer, mode 2026, kitty flags with reported support, typed mouse and paste messages | `Send` blocks; `teatest` polls with sleeps; v1 to v2 broke most APIs | loom has a richer widget and routing model; Bubble Tea diffs frames, loom does not (cost difference not measured) |
| ultraviolet | row hash diff, writes changed cells only, grapheme width | marked as in active development, API may change | loom has no diff yet |
| Bubbles / Huh | pinned, clean sources; virtualised table; fuzzy list; accessible forms; `$EDITOR` | no undo/redo in text inputs; select-all on Ctrl+G; Tab accepts suggestions; table has no sort or filter | loom widgets are more desktop-like (undo, Ctrl+A/C/V/X, mouse) |
| Lip Gloss v2 | Z-ordered layers, hit test | no focus, click-through or modal handling | loom's claim-cell routing does more |
| tview / tcell | mature, grapheme cells, kitty query before enabling | Ctrl+Q copy (Ctrl+C quits); mouse off by default | similar scope; loom has finer routing |
| gocui | small | mouse buttons share key codes; internal clipboard only | loom is ahead |

loom's main lead: desktop-like editing, mouse capture, claim-cell composition and its test tiers.
loom's main lag: output efficiency and terminal capability detection.

Scope rule (owner, 2026-11-10): only the latest version of each library counts; API stability and
breaking changes are P3 notes while loom is alpha.

## Top gaps for loom (P1 first)

1. **Write only changed rows or cells.** loom redraws all rows on a dirty flush; idle panes write
   nothing. ultraviolet compares old and new lines from the first changed cell; Bubble Tea v1 skips
   unchanged frames and lines. TUIOS's "style diff" is an SGR cache, not a frame diff. Comparative
   cost is not measured yet (458 batch 1 rerun); do not pick an implementation from this alone.
2. **Mode 2027 grapheme width.** Ask the terminal with DECRQM and switch it on when supported,
   instead of guessing from the terminal name in `measure`. TUIOS forces 2027 on in its emulator.
3. **Ask before enabling.** Query DECRQM 2026 and `CSI ? u` (kitty) at start, as tcell does,
   instead of only opt-in switches.
4. **Drag auto-scroll.** While a captured drag is held outside the content, keep scrolling on a
   tick (TUIOS: 50 ms). Fits the existing pane capture.
5. **In-band resize (mode 2048)** in Pane, with SIGWINCH as the fallback.
6. **Kitty flags 2 and 4** (press/release/repeat, alternate keys), with a report of what the
   terminal accepted.
7. **Unchecked P1 rows**: paste sanitising, Table virtualisation, cleanup when `Run` fails early,
   OSC 52 write policy and kitty flag stack in `loom session`.
8. **P2**: dark/light detection (DSR 996, mode 2031, OSC 11), focus reporting in Pane, OSC 52 read,
   text input validation, password echo and suggestions, `$EDITOR` hand-off.

## Limits of this study

- 10 of 189 rows are unverified (kitty pages cut off, xterm ctlseqs refused, two 404 pages).
- Some line numbers are approximate (Lip Gloss rows, ±3), and many low-level rows lack line numbers.
- P2 and P3 rows were not checked against loom.
- Not researched: TUIOS session attach code, Bubble Tea issue threads, ratatui.
- No canary was run; mode 2026 and kitty behaviour in real terminals is unmeasured.
