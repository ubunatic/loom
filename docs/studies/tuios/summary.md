# TUIOS study: summary (ticket 457)

Sources: the seven `report-*.md` files and `features.md` in this folder. Researched 2026-11-10
by seven haiku:med searchers. The host checked the P1 rows against loom code at `163b3738`.

## What TUIOS is built on

- Bubble Tea v2, Lip Gloss v2 and ultraviolet. No Bubbles, Huh or Glamour, no `replace` in
  `go.mod`. TUIOS's own `vt/` package is a derived, heavily extended copy of charmbracelet/x `vt`
  (its NOTICE says so; `emulator.go` 2173 lines vs 503), so TUIOS VT findings are not x/vt
  findings (458 batch 7).
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
| Bubbles / Huh | pinned, clean sources; virtualised table; fuzzy list; accessible forms; `$EDITOR` | no undo/redo in text inputs; select-all on Ctrl+G; Tab accepts suggestions; table has no sort or filter | loom's TextEntry and RichTextEdit are more desktop-like (undo, Ctrl+A/C/V/X, mouse); TextInput and TextArea have no user selection, undo or cut (458 batch 6) |
| Lip Gloss v2 | Z-ordered layers, hit test returning the top layer's ID | no focus, click-through or modal handling | loom's claim-cell routing does more; loom has no z-index (draw order). v1 and v2 can be required together because their module paths differ (v2 is `charm.land/lipgloss/v2`); a joint build was not tried (458 batch 8) |
| tview / tcell | mature, grapheme cells, kitty query before enabling | Ctrl+Q copy (Ctrl+C quits); mouse off by default | similar scope; loom has finer routing |
| gocui | small | mouse buttons share key codes; internal clipboard only | loom is ahead |

loom's main lead: desktop-like editing, mouse capture, claim-cell composition and its test tiers.
loom's main lag: output efficiency and terminal capability detection.

Scope rule (owner, 2026-11-10): only the latest version of each library counts; API stability and
breaking changes are P3 notes while loom is alpha.

## Top gaps for loom (P1 first)

1. **Write only changed rows or cells.** loom redraws all rows on a dirty flush; idle panes write
   nothing. ultraviolet compares old and new cells per touched line (its line hash is only a
   fullscreen scroll optimisation). TUIOS's "style diff" is an SGR cache, not a frame diff.
   Measured (458 batch 1, 80×24, one run, boundaries differ): one changed cell costs loom Canvas
   and Pane 2269 bytes and Bubble Tea v2.1.1 9 bytes; a bold change 2301 vs 11; a wide cell
   2270 vs 25. Not measured: Pane write calls, replay into a terminal model, real workloads.
2. **Mode 2027 grapheme width.** Ask the terminal with DECRQM and switch it on when supported,
   as ultraviolet does at start, in addition to the ZWJ cursor probe in `measure`. One width
   policy must then drive measure, canvas, renderer and loom's VT (which answers 2027 with 0
   today). TUIOS's emulator reports 2027 as permanently set. Real-terminal replies are unmeasured
   (owner check in report-width.md). Separately, loom splits Indic conjuncts (`क्ष`) that current
   Unicode keeps as one cluster.
3. **Ask before enabling, and keep replies out of key input.** Query DECRQM 2026, 2027 and
   `CSI ? u` (kitty) at start in one batch ended by DA1, as tcell (1 s backstop) and TUIOS
   (300 ms) do. First Pane must consume replies itself: today it forwards them to widgets as
   `KeyEvent{Report}`, some reply forms are not even marked, and a late fragment becomes text
   keys (458 batch 3, ticket 462; 2027 part in 461).
4. **Drag ownership and auto-scroll.** TextEntry selection drags do not capture the pointer,
   release outside leaves the drag on, and focus loss never ends a capture (an inverted check in
   `SetFocus`). No auto-scroll while held outside (TUIOS: 50 ms tick). Ticket 467. Horizontal
   wheel reports arrive as button presses (ticket 468). Nested coordinates and modal capture are
   correct (458 batch 5).
5. **In-band resize (mode 2048)** in Pane, with SIGWINCH as the fallback.
6. **Kitty flags 2 and 4** (press/release/repeat, alternate keys), with a report of what the
   terminal accepted. Today `decodeCSIu` drops the extra fields and `KeyEvent` has no event
   type; not ticketed while Pane pushes flag 1 only (458 batch 4).
7. **Paste sanitising** differs per widget: only TextEntry strips controls, RichTextEdit ignores
   paste (ticket 464). Canvas output drops controls, so the terminal is safe.
8. **Widgets (458 batch 6)**: Form takes Enter, Down and PgDn before a focused TextArea, so
   multi-line editing in a Form breaks; TextInput and TextArea lack user selection and undo.
   Table draws only visible rows (about 90 µs at 100 and 100,000 rows), but each filter
   keystroke scans all rows (6.7 ms at 100,000).
9. **Session and VT (458 batch 7)**: `vt.VT` keeps one kitty flag value, not a per-screen stack,
   and the session re-sends a fixed flag 1 on reattach (ticket 472); child OSC 52 writes and
   queries pass to the host unfiltered (ticket 473); a client that stops reading may block all
   tabs of its session (ticket 474, not yet canaried). Reattach repaints the screen only, so
   lines that scrolled off while detached stay only in the session's VT history (not ticketed).
10. **Desktop (458 batch 8)**: closing a Popup inside a container leaves focus on the closed
   popup, so keys are lost until Tab (ticket 475). An OSC 11 reply is typed into a focused text
   field as `11;rgb:…` (ticket 462). No dark/light detection (ticket 476, after 462). Overlap is
   draw order only, which is safe because Frame and Stack tile; modal popups take clicks first.
   loom renders no Markdown (Glamour and TUIOS do); with no `TERM` set, loom picks true colour
   where colorprofile says "no TTY" (not ticketed).
11. **Testing and lifecycle (458 batch 9)**: `vt.VT` (also `loomtest.VT` and session tabs) loses
   a combining mark when a write splits between base and mark (ticket 477). `Pane.Post` never
   blocks but has no bound, and keys wait behind whole batches (up to 85,055 posts; latency in ms
   not yet measured, ticket 478). `Run` restores the terminal on panic; after an error or a signal
   the caller must call `Close`, and `Close` drops queued posts: both comments say otherwise
   (ticket 479). Pane PTY tests still use timing sleeps (ticket 480). TUIOS checks its VT against
   libghostty behind a build tag; teatest polls with a 50 ms sleep.
12. **P2**: focus reporting in Pane, OSC 52 read, text input validation, password echo and
   suggestions, `$EDITOR` hand-off.

## Limits of this study

- 10 of 189 rows are unverified (kitty pages cut off, xterm ctlseqs refused, two 404 pages).
- Some line numbers are approximate (Lip Gloss rows, ±3), and many low-level rows lack line numbers.
- P2 and P3 rows were not checked against loom.
- Not researched: Bubble Tea issue threads, ratatui. TUIOS session attach was compared in 458
  batch 7 (report-session.md).
- Only the headless render canary was run (report-render.md); mode 2026 and kitty behaviour in real terminals is unmeasured.
