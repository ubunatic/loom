# Session and VT follow-up (batch 7)

Scope: loom `internal/session` attach, detach, resize and clipboard passthrough, and loom `vt.VT` kitty keyboard state, alternate screen, scrollback and OSC 52. Compared with TUIOS (`internal/session`, `vt/`, `internal/app`) and charmbracelet/x `vt/`. Plan: `followup-plan.md` section 7. Shape follows `report-widgets.md`. Query-reply behaviour (DECRQM 2026/2027/2048) is not repeated here; see [report-queries.md](report-queries.md), which covers loom VT replies.

## Pins

Loom revision for citations: `e08d201850f150ebf7d175c0895142a4d1ebb732` (on origin, Codeberg). Loom citations are `file:line` at that revision; the permalink form is `https://codeberg.org/ubunatic/loom/src/commit/e08d201850f150ebf7d175c0895142a4d1ebb732/<file>#Lx-Ly`.

| Alias | Repository | Revision used | Plan pin | State |
|---|---|---|---|---|
| T | `Gaurav-Gosain/tuios` | `6fe2c0df9922b2eec28f392fee4e196efd86be36` (HEAD, `git ls-remote` on 2026-11-10) | same | unchanged |
| X | `charmbracelet/x` | `27ecf3e4d454ff2656204b806b7ee6ef11593746` (HEAD) | same | unchanged |

The newest TUIOS tag seen in `git ls-remote --tags` is not used: the plan asks for HEAD pins. TUIOS's own release `v0.12.1` (`850335c0`) was pinned in report-queries and is older than HEAD. Clones: `b7-src/tuios` and `b7-src/x` (`git clone --depth 1`) in the scratch area, not tracked. Canary harnesses: `b7-canary` (VT) and `b7-src/loom-head` (session, a `git archive` of loom `e08d2018`, so the owner's tree and sockets are untouched).

Upstream links below use the immutable SHAs above: `https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/<file>#Lx-Ly` and `https://github.com/charmbracelet/x/blob/27ecf3e4d454ff2656204b806b7ee6ef11593746/<file>#Lx-Ly`.

## Loom code map (file:line at e08d2018)

### Attach, output and detach

- **Attach** (`internal/session/server.go:121-140`): one `conn` slot. A second attach is refused with `session already has an attached client` (`server.go:123-125`). Attach sets `ready=false` and sends an `attachEvent` to the session loop (`server.go:134`).
- **Output gate** (`server.go:88-98`): `writeOutput` returns nil and drops bytes when there is no conn, or when `!ready && !direct`. `serverOutput.Write` (`server.go:159-169`) detaches the client on a write error and reports success so PTY processing never stops.
- **Snapshot and boundary** (`internal/session/session.go:1376-1405`): the session loop applies the attach size, then sends modes (`1390`), repaints with `direct=true` (`1394-1397`) and sets `markReady` (`1401-1403`), all under `sessionMu`.
- **PTY reader** (`session.go:752-800`): takes `sessionMu` for each read (`758`), feeds the VT and gets the framed bytes (`764`), answers child queries from the VT (`767-769`), and writes the frame to the client only for the active tab with no help or search overlay (`782-784`). The unlock is at `795`.
- **Repaint** (`session.go:539-564`; `vt/vt.go:299-365`, screen rows at `:316-365`): the snapshot is `Repaint()`, which emits `1049h/l`, `7h/l`, then the 24 screen rows only. It has no scrollback and no kitty flag value.
- **Client** (`internal/session/client.go:164-166`): `msgOutput` goes to the terminal unmodified. `msgStatus` `modes` sets only booleans for alt screen and kitty (`client.go:173-177`).
- **Detach**: client sends `msgDetach` on EOF (`client.go:153`); server closes its conn (`server.go:142-157`). The child keeps running and its output is dropped while detached (`server.go:94-95`).
- **Write deadline**: none on the attached conn. `writeMessage` (`ipc.go:217-237`) has no deadline and `SetDeadline` appears only in the control request (`control.go:103`). The reader holds `sessionMu` during the write (`session.go:758-795`). A client that stops reading may block every tab. Not canaried.

### Resize ownership

- The single attached client owns the size. Attach sets PTY and VT size from the client (`session.go:1378-1386`). Later sizes arrive as `msgResize` from SIGWINCH (`client.go:113-138`, `server.go:290-295`) and are applied in the same loop.
- A second attach fails (`server.go:125`), so there is no concurrent owner.

### Kitty keyboard state

- **VT** (`vt/vt.go:40`, `:58`): one `KittyKeyboard bool` and one `kittyFlags int`. No stack, no per-screen state.
- **Push and set** (`vt/vt.go:809-812`): `CSI > N u` overwrites `kittyFlags` with N. `CSI = N u` has no branch: the `u` case handles `?` (query, `800-808`), `>` (push-like, `809-813`) and `<` (pop, `814-818`). Any other `u`, including `=31u`, restores the saved cursor (`vt.go:819`). Verified by canary a.
- **Alternate screen** (`vt/vt.go:74-87`): `alternate` saves and restores cells and cursor only. The kitty flags are not touched and do not switch with the screen.
- **Session re-emit** (`session.go:520-532`, `:527-529`): `applyTabModes` writes the fixed `ESC[>1u` or `ESC[<u`, not the child's value, when the tab's boolean changes or on attach. Live child pushes are forwarded raw from the stream (`session.go:764`, `783`).
- **Pane screen switch** (`pane_screen.go:237-243`): pane pops and re-pushes its own flag (`kittyKeyboardPush` = `ESC[>1u`, `pane_modes.go:147-148`) around its own alternate screen. This is Pane, not the session VT.

### Alternate screen and scrollback

- **Alternate screen** (`vt/vt.go:902-903`, `:74-87`): a single saved `main` slot. Entering alt while already in alt is a no-op (`alternate`, first check).
- **Scrollback bound** (`vt/vt.go:139-156` `SetScrollbackLimit`, `:159-170` `ScrollbackCells`, `:579-594` `lineFeed`): ring of `scrollbackLimit` rows, main screen only (`!AltScreen()` at `:584`). Session sets the limit from spec `ScrollbackLines` (`session.go:749`).

### OSC 52

- **VT** (`vt/vt.go:639-657`): `ESC ]` is consumed up to BEL, ST or CAN and dropped without a handler. Nothing is stored and nothing is replied. The raw bytes still appear in `WriteFramed` output (`vt.go:220-234`), which returns the input slice.
- **Session** (`session.go:764`, `:783`): forwards the framed bytes unchanged. There is no per-pane, per-focus or per-policy gate. Only the active tab streams at all (`:782`).
- **Loom's own outbound write** (`clipboard.go:11-16` `WriteOSC52`; `pane_run.go:111` writes to the pane tty). This is Pane's copy path, tracked by 465; it is not involved in session PTY forwarding.

## Matrix

Status key: has, partial, missing, n/a (not applicable to that component), unknown (see reason). "Proved" and "disproved" are the canary results in the next section.

| Feature | loom session | loom vt | TUIOS | x/vt |
|---|---|---|---|---|
| Attach snapshot | partial: `Repaint` 24 screen rows (`session.go:1394-1397`, `vt/vt.go:316-365`); no scrollback | has: `Repaint` (`vt.go:299`) | has: snapshot plus ring, scrollback seeded on attach ([REHYDRATION](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/docs/REHYDRATION.md#L154-L157)) | partial: `Render()` screen only (`emulator.go:139`); no attach transport |
| Stream boundary | has: reader and attach share `sessionMu`; no live gap in canary d1 (`session.go:758-795`, `:1376-1405`) | n/a: `WriteFramed` returns complete bytes (`vt.go:220-234`) | has: byte offset `S`, resume from `ptyResume` (REHYDRATION [L204-L207](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/docs/REHYDRATION.md#L202-L207)); not canaried | missing: no transport |
| Lines lost to a reattaching client | has (by design): lines that scroll off while detached are never sent; canary d1 shows 11 lines | n/a | unknown: not traced in detail | n/a |
| Resize owner | has: attached client only, second attach refused (`session.go:1378-1386`, `server.go:125`; canary d4) | partial: `Resize` (`vt.go:115`) | partial: resize is a stream point ([REHYDRATION L535-L541](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/docs/REHYDRATION.md#L535-L541)); owner not traced | n/a |
| Kitty stack nesting | n/a | missing: one value (`vt.go:58`); `>` overwrites, `<` clears (`vt.go:809-816`); canary a | has: max 8 entries, push and pop ([kitty_keyboard.go L18](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/kitty_keyboard.go#L18), [L116-L122](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/kitty_keyboard.go#L116-L122)) | missing: no kitty handling in `csi*.go`, `mode.go`, `emulator.go` (grep) |
| Kitty per screen | missing: session re-emits fixed `ESC[>1u` (`session.go:527`); canary kitty reattach | missing: flags survive `?1049` (canary a) | has: separate `main` and `alt` stacks, alt starts empty ([L20-L26](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/kitty_keyboard.go#L20-L26), [L88-L96](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/kitty_keyboard.go#L88-L96)) | missing |
| Kitty flags replayed on attach | partial: boolean only in `modes` status (`session.go:1390`), fixed `>1u` (`session.go:527`); canary kitty reattach | missing: `Repaint` emits no kitty (`vt.go:299-365`) | partial: `MainStack()` exists for snapshots ([L70-L74](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/kitty_keyboard.go#L70-L74)); replay not canaried | missing |
| OSC 52 write policy | missing: no gate; child bytes forwarded raw (`session.go:764`, `:783`); canary d2 | n/a: VT stores nothing (canary c) | has: `osc52_write` off, on, ask, focused (default) ([clipboard_osc52.go L21-L23](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/internal/app/clipboard_osc52.go#L21-L23), [L63-L78](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/internal/app/clipboard_osc52.go#L63-L78)) | missing: no `52` in `osc.go` (grep) |
| OSC 52 query | partial: query forwarded raw; whether the outer terminal answers is unknown (no live outer terminal) | missing: no reply, no query handling (canary c) | has: query returns empty content ([L13-L16](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/internal/app/clipboard_osc52.go#L13-L16)) | missing |
| Scrollback bound | has at spec: 5000 rows (`session.go:749`); canary b | has: 5000 limit, ring (`vt.go:139-170`, `:579-594`); canary b | has: `DefaultScrollbackSize = 10000` ([scrollback.go L17](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/scrollback.go#L17)) | has: `DefaultScrollbackSize = 10000` ([scrollback.go L10](https://github.com/charmbracelet/x/blob/27ecf3e4d454ff2656204b806b7ee6ef11593746/vt/scrollback.go#L10)); `SetScrollbackSize` ([emulator.go L487-L488](https://github.com/charmbracelet/x/blob/27ecf3e4d454ff2656204b806b7ee6ef11593746/vt/emulator.go#L487-L488)) |
| Mode-report truth (DECRQM) | see report-queries (not repeated) | see report-queries | has: unimplemented modes not stored ([csi_mode.go L22-L24](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/csi_mode.go#L22-L24)) | unknown: x/vt `csi_mode.go` differs; not runtime-checked |

Second attach: loom refuses it at the session (`server.go:125`). A session attach to the same loom session from another terminal fails rather than taking over.

## Canary results

Harness: `b7-canary` (module `b7canary`, `replace` to the working tree; VT only, no sockets). `b7-src/loom-head` (copy of `e08d2018` via `git archive`) for the session tests; they run in-process over `net.Pipe`, as `server_test.go` does. No socket, no `XDG_RUNTIME_DIR` override, owner sessions untouched. No sleeps; every read has a 5 s deadline. Each run used `-count=1`.

### (a) kitty stack and alternate screen

Command: `cd b7-canary && GOFLAGS=-mod=mod go test -v -count=1 -run KittyStack ./...`

Expected (kitty protocol: push, pop with count, set, query returns current flags, per-screen stacks): push1 gives `?1u`; push3 gives `?3u`; pop gives `?1u`; pop gives `?0u`; set 31 gives `?31u`; alt starts empty.

| Step | Input | Expected | Actual | Verdict |
|---|---|---|---|---|
| initial query | `ESC[?u` | `ESC[?0u` | `ESC[?0u` | has |
| push 1 | `ESC[>1u` | `?1u` | `?1u` | has |
| push 3 (nested) | `ESC[>3u` | `?3u` | `?3u` | has |
| pop 1 | `ESC[<u` | `?1u` (stack) | `?0u` | disproved: stack cleared |
| pop 1 | `ESC[<u` | `?0u` | `?0u` | has |
| set 31 | `ESC[=31u` | `?31u` | `?0u` | disproved: `=` form ignored |
| push 1, enter alt | `ESC[>1u` then `ESC[?1049h` | `?1u` main; alt `?0u` | `?1u`, alt `?1u` | disproved: not per screen |
| push 5 in alt | `ESC[>5u` | `?5u` | `?5u` | has |
| leave alt | `ESC[?1049l` | main `?1u` | main `?5u` | disproved: leaks to main |
| main pop | `ESC[<u` | `?0u` | `?0u` | has |

Conclusion: **disproved** for nesting (single value) and for per-screen state. Confirmed by code: `vt.go:58`, `:809-816`; `alternate` at `vt.go:74-87` leaves kitty alone.

### (b) scrollback bound

Command: `cd b7-canary && GOFLAGS=-mod=mod go test -v -count=1 -run ScrollbackBound ./...`

Expected: 100000 lines with limit 5000 leave 5000 history rows plus 24 screen rows; no history without a limit.

Actual: `ScrollbackCells=5024, screen rows=24, history=5000`; first stored `L094978`, last history `L099977`; heap delta after fill about 40.5 MiB (one run, GC before and after, 80 columns, includes the rest of the VT); with no limit set, 1000 lines give 24 cells (no history).

Verdict: **proved** (bound holds). Memory per pane is a single measurement and a hypothesis: about 8 KiB per row at 80 columns, to be confirmed by repeating the measurement.

### (c) OSC 52 write and query

Command: `cd b7-canary && GOFLAGS=-mod=mod go test -v -count=1 -run OSC52 ./...`

Expected for a stored-but-not-forwarded policy: nothing stored (`Hello` absent), no reply to the query. Expected from the session side: bytes forwarded.

Actual, with BEL and ST terminators, for write and query:

| Input | Framed output (what the session forwards) | Replies | Screen has Hello |
|---|---|---|---|
| `ESC]52;c;SGVsbG8=BEL` | identical to input | none | no |
| `ESC]52;c;SGVsbG8=ESC\` | identical to input | none | no |
| `ESC]52;c;?BEL` | identical to input | none | no |
| `ESC]52;c;?ESC\` | identical to input | none | no |

Verdict: **proved**: the VT stores nothing and replies nothing. The forwarding is shown by canary d2 below.

### (d) session attach, detach and reattach

Commands: `cd b7-src/loom-head && go test -count=1 -run 'B7' -v ./internal/session/` (the four tests are `TestB7AttachBoundary`, `TestB7OSC52AndKittyForwarded`, `TestB7KittyReattach`, `TestB7SecondAttachRefused`).

**d1, attach boundary (`TestB7AttachBoundary`).** The child prints `T000001…` and blocks on stdin every 500 lines. Attach 1 reads to `T000500`; then a goroutine sends `\n` and a detach. Attach 2 reads until `T001000`.

Expected: no line missing from the live stream; snapshot and live bytes meet with no duplicate beyond screen redraw.

Actual (two runs; the timing of the detach differs between runs):

| Run | Attach 1 numbers | Attach 2 first line | Attach 2 last | Lines never delivered | Overlap on attach 2 |
|---|---|---|---|---|---|
| 1 | 1..501 | 499 (snapshot) | 1000 | 0 | 3 (screen redraw) |
| 2 | 1..500 | 512 (snapshot) | 1000 | 11 (501..511) | 0 |

Attach 2 begins with raw `ESC[?1049l ESC[?7h ESC[0m ESC[H ESC[2J ESC[1;1HT000512 ESC[2;1HT000513 …`, i.e. the screen repaint. Live lines 536..1000 follow with no gap. The 11 lines in run 2 scrolled off the 24-row screen while detached and exist only in server history.

Verdict: **proved** for no gap in the live stream and that the snapshot is screen-only. **Lost-history behaviour proved** (lines scrolled off while detached never reach a reattaching client; the session has no replay).

**d2, OSC 52 and kitty forwarding (`TestB7OSC52AndKittyForwarded`).** The child prints `ESC]52;c;SGVsbG8=BEL`, then `ESC[>31u`, then `MARK`.

Actual: the attached client's bytes contain the OSC 52 sequence raw (true) and the raw push `ESC[>31u` (true). Verdict: **proved** that the session forwards both unmodified while attached. Gating by focus or policy is absent (no branch in `session.go:764-784`).

**d3, kitty reattach (`TestB7KittyReattach`).** The child pushes `ESC[>31u` while attached, then detach, then reattach.

Expected if the session replayed state: `ESC[>31u`. Actual: attach 1 forwarded the raw push (true); attach 2 contains `ESC[>1u` (true) and not `ESC[>31u` (true). Verdict: **proved**: the reattach re-emits the fixed value 1 (`session.go:527`), not the child's 31.

**d4, second attach (`TestB7SecondAttachRefused`).** After client 1 is attached (its status message seen), client 2 attaches. Actual reply: `type="error" error="session already has an attached client"`. Verdict: **proved** (single owner).

Not run (budget and scope): a live socket attach through `XDG_RUNTIME_DIR`, a resize during attach, tab-switch gating, and the stalled-client hypothesis (no write deadline, `sessionMu` held during writes). These are code reads, marked as such above; the stalled-client case is unknown until canaried.

## Upstream mechanisms

Quotes are from the pinned trees. Each is a short excerpt and an annotation.

### T: VT provenance (`vt/NOTICE`)

[NOTICE#L3-L4](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/NOTICE#L3-L4):

> It started as a copy of the vt package of github.com/charmbracelet/x, and
> parser.go is a modified copy of the Parser of github.com/charmbracelet/x/ansi.

Scoped finding: TUIOS's `vt/` is derived from x/vt and its parser is a modified copy of x/ansi's parser. `vt/LICENSE` is MIT, Copyright 2025 Gaurav Gosain, and NOTICE carries the Charmbracelet MIT notice. The package is not a verbatim fork: line counts diverge (table below). Whether TUIOS tracks x/vt upstream is not checked.

### T: kitty stack per screen

[kitty_keyboard.go#L20-L26](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/kitty_keyboard.go#L20-L26):

> The protocol uses a stack of flag sets that can be pushed/popped by
> applications, and the main and alternate screens each have their own: a
> full-screen program that pushes flags on the alternate screen and exits
> without popping them leaves the shell's keyboard as it found it.
> type kittyKeyboardState struct { main, alt []int // flag stacks, oldest first; never empty

Annotation: per-screen stacks, which is what canary a shows loom lacks.

[kitty_keyboard.go#L116-L122](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/kitty_keyboard.go#L116-L122):

> // Push pushes a new set of flags onto the stack. A full stack drops its
> // oldest entry first.
> func (k *kittyKeyboardState) Push(flags int) {
> 	st := k.active()
> 	if len(*st) >= maxKittyKeyboardStack {
> 		*st = append((*st)[:0], (*st)[1:]...)

Annotation: bounded at 8 (`kitty_keyboard.go#L18`) with eviction of the oldest entry. Limit: the cap is a design choice, not a spec requirement.

### T: alternate screen starts empty

[kitty_keyboard.go#L88-L93](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/kitty_keyboard.go#L88-L93):

> // Entering the alternate screen starts it with an empty stack, as it starts
> // with an empty screen: flags a previous program left there are not the next
> // one's. The main screen's stack is kept as it was.
> func (k *kittyKeyboardState) SetAltScreen(on bool) {

Annotation: a program's alt-screen flags do not leak to the main screen.

### T: OSC 52 policy

[clipboard_osc52.go#L13-L16](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/internal/app/clipboard_osc52.go#L13-L16):

> // A program in a pane sets the clipboard with OSC 52. The pane keeps no copy:
> // an OSC 52 query gets empty content back, as in xterm, because the answer is
> // written to the pane's input. Whether the host clipboard gets the write is
> // appearance.selection.osc52_write:

[L21-L23](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/internal/app/clipboard_osc52.go#L21-L23):

> //   - focused (the default): the focused pane writes and the dock says so. A
> //     write from any other pane asks.
> //   - on: every pane writes.

Annotation: the host-side write is gated by policy and focus. The query answer is empty content. Loom has neither gate nor empty-answer policy (canary d2).

### T: snapshot and stream boundary

[REHYDRATION.md#L202-L207](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/docs/REHYDRATION.md#L202-L207):

> ## Why the snapshot and the stream cannot both be applied
>
> The snapshot is the daemon emulator's state after consuming bytes `0..S`. The
> ring replay used to hand the client bytes `R..S` for some `R <= S`. Writing those
> bytes into an emulator that already holds the state at `S` re-runs output the
> client already has.

Annotation: the boundary is a byte offset, and the document describes the duplication bug it had to fix. Loom avoids the issue differently, with a mutex and no byte offset (canary d1 shows no live gap).

[REHYDRATION.md#L154-L157](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/docs/REHYDRATION.md#L154-L157):

> **M1: `primePaneFromDaemon`.** Snapshot, then subscribe with whatever
> `ptyResume` the daemon still holds for this pane.

Annotation: a resume offset exists for a pane. Loom has no equivalent and no resume.

### T: resize is a stream point

[REHYDRATION.md#L535-L541](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/docs/REHYDRATION.md#L535-L541):

> Where a line wraps is not a property of the line. It is decided once, by the
> width the emulator had when it consumed the bytes, and never revisited …
> So two emulators fed the same bytes hold the same history only if they change
> width at the same byte.

Annotation: loom applies resize to the VT out of stream order (`session.go:1378-1386`), which can change how later bytes wrap.

### T: scrollback default

[scrollback.go#L17](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/scrollback.go#L17): `const DefaultScrollbackSize = 10000`. Annotation: the default is larger than loom's spec value of 5000. Loom's value is configurable through `ScrollbackLines` (`session.go:749`).

### T: unimplemented modes

[csi_mode.go#L22-L24](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/csi_mode.go#L22-L24):

> // A mode the emulator does not implement is not stored. Storing it
> // made DECRQM report it set, and a guest that probes before it
> // enables a feature believed a mode nothing acts on.

Annotation: TUIOS records a fix for a DECRQM truth bug. Loom's DECRQM truth is in report-queries and is not repeated here.

### X: scrollback default and API

[scrollback.go#L10](https://github.com/charmbracelet/x/blob/27ecf3e4d454ff2656204b806b7ee6ef11593746/vt/scrollback.go#L10): `const DefaultScrollbackSize = 10000`. [emulator.go#L487-L488](https://github.com/charmbracelet/x/blob/27ecf3e4d454ff2656204b806b7ee6ef11593746/vt/emulator.go#L487-L488): `// SetScrollbackSize sets the maximum number of lines in the scrollback buff[er]` with `SetScrollbackSize(maxLines int)`. Annotation: x/vt exposes a size API; loom's `SetScrollbackLimit` is the equivalent.

### X: missing features

No kitty keyboard stack (grep of `vt/csi*.go`, `vt/mode.go`, `vt/emulator.go` finds none). No `52` in `vt/osc.go`. Annotation: both gaps are absent in x/vt, so TUIOS's kitty and OSC 52 work is its own, not inherited.

## Loom vs upstream

Provenance table (line counts and changed lines, `diff` of files with the same name, `x` vs `tuios`):

| File | x/vt lines | TUIOS vt lines | Changed lines |
|---|---|---|---|
| callbacks.go | 63 | 107 | 52 |
| cc.go | 60 | 154 | 100 |
| csi_mode.go | 113 | 276 | 197 |
| mode.go | 33 | 92 | 105 |
| scrollback.go | 136 | 976 | 984 |
| emulator.go | 503 | 2173 | 2032 |

Reading: TUIOS's `vt/` is a derived and heavily extended copy of x/vt, not a verbatim fork, and not an unrelated implementation. Loom has no VT dependency; its `vt` package is its own (`vt/vt.go`, 1092 lines).

Comparison by feature:

- **Attach:** TUIOS seeds scrollback and resumes from byte offsets; loom sends 24 screen rows and no history. Loom's behaviour is simpler and matches its single-client model.
- **Kitty:** TUIOS has per-screen bounded stacks; loom has one value. Loom's fixed `>1u` replay on attach is lossy. x/vt has no kitty at all.
- **OSC 52:** TUIOS gates the host write by policy and focus and answers queries with empty content. Loom forwards child OSC 52 and queries raw, with no policy. This matters if the outer terminal honours OSC 52 writes or queries; that outer behaviour is unknown here.
- **Scrollback:** all three bound history (loom 5000 by spec, TUIOS and x/vt 10000 by default). Loom's bound holds (canary b).

## Stage 4 proposals

Proposals only: this batch does not edit `features.md`, `summary.md` or tickets. The features rows are given as replacement cells.

### features.md (session rows)

- **Row 52, "Daemon with attach/detach", `tuios.session-daemon`.** Keep `partial`. Evidence: attach sends a 24-row screen snapshot and then a live stream serialised under `sessionMu` (`session.go:758-795`, `:1376-1405`; `vt/vt.go:316-365`); canary d1: no live gap, lines scrolled off while detached are never sent; canary d4: second attach refused. Replace `attach compare open` with `attach compared: no scrollback, no byte-offset resume`.
- **Row 53, "OSC 52 write policy per pane", `tuios.osc52-write-policy`.** Replace `unchecked` with `missing` (loom). Evidence: `session.go:764`, `:783` forward child bytes unmodified (canary d2); TUIOS has the policy (`clipboard_osc52.go#L21-L23`). Outer-terminal behaviour: unknown.
- **Row 54, "Kitty flag stack per screen", `tuios.kitty-keyboard-stack`.** Replace `unchecked` with `disproved` for loom. Evidence: `vt/vt.go:58`, `:809-816`, canary a: nested push and pop give 0, alt flags leak to main. TUIOS: `has` (`kitty_keyboard.go#L20-L26`, `#L116-L122`). x/vt: `missing`.
- **Row 55, "VT scrollback API", `vt.scrollback-api (UNVER)`.** Replace `UNVER` with `proved`. Correct the evidence pointer: the features.md evidence says "`Scrollback` in `pane_screen.go`", but `pane_screen.go` has no `Scrollback`. The API is `SetScrollbackLimit` and `ScrollbackCells` in `vt/vt.go:139-170`, with `lineFeed` at `vt/vt.go:579-594`. Canary b: 5000 rows bound.

### summary.md

Line 8 reads: "Bubble Tea v2, Lip Gloss v2 and ultraviolet. No Bubbles, Huh or Glamour, no forks, no `replace`." Proposed scoped replacement for the forks clause: "no `replace` directive in `go.mod`; TUIOS's own VT package is a derived copy of charmbracelet/x `vt` (its NOTICE says it started as a copy and its parser is a modified copy of x/ansi's parser), with large divergence (for example `emulator.go` 2173 lines against 503), so the TUIOS findings are not those of x/vt." Loom itself has no VT dependency.

### Ticket candidates (no tickets created)

| Candidate | Evidence | Overlap with existing tickets |
|---|---|---|
| VT kitty flags as a per-screen bounded stack, with `CSI = N u` set form | canary a; `vt/vt.go:58`, `:809-816`; `=` falls to cursor restore | new. Related 321 and 384 (Pane kitty keyboard, closed); those cover Pane flag 1, not the VT stack |
| Session forwards child OSC 52 and queries to the host unfiltered | canary d2; `session.go:764`, `:783` | related to 465 (loom clipboard writes); 465 covers loom's own Ctrl+C/V/X writes, not child passthrough. Fold in or file new |
| Attach replay: scrollback and child kitty flags are not re-sent on reattach | canary d1 (11 lines lost), canary d3 (`ESC[>1u` not `>31u`) | new. Related 372 (terminal restore, closed); 374 (scrollback search) is the live VT only |
| Session output write has no deadline, and `sessionMu` is held during writes | code read `ipc.go:217-237`, `session.go:758-795`; no canary | new, unverified. Canary first (stalled client, measure whether PTY reads stop). Related 381 (session debug mode) is a different stall detector |
| Correct features.md scrollback evidence pointer | `pane_screen.go` has no `Scrollback` | docs only |
| 462 overlap (replies as keys) | the reply-as-key overlap was not re-verified in this batch | existing 462 covers this; no new ticket |

## Unknowns

- Outer-terminal reply to an OSC 52 query forwarded by loom: depends on the user's terminal. Needs a live check in foot or kitty (manual, owner-run). Not canaried.
- TUIOS resize owner: not traced beyond the stream-order comment.
- x/vt DECRQM behaviour: grep-only; not runtime-checked.
- Stalled-client stall: hypothesis only, from the code (no write deadline, lock held during write).
- Memory per pane: one measurement (40.5 MiB for 5000 rows at 80 columns, including the rest of the VT). Repeat before quoting.
- Canary d1 timing: the detach races the release; the exact split of lines between attaches varied between runs (run 1: attach 1 ended at 501, run 2 at 500). The no-gap result held in both runs.

## Flow notes

- The `harnez read -L` call takes one range per invocation; multi-range `sed` also failed. Single-range reads and `sed -n` with several ranges worked.
- A Bash `printf >` into `go.mod` was flagged as a scripted edit; the file was written with the Write tool instead.
- The session package is `internal`, so a separate module cannot import it. The canary ran in a `git archive` copy of `e08d2018` rather than the working tree.
- In-process session tests over `net.Pipe` deadlock if the test writes before reading: the server holds the writer lock while blocked on output. The canary writes from a goroutine and reads in the test.
- The second-attach test raced: two attach handlers compete for the single slot. Waiting for the first client's status message fixed it.
- `go test` added indirect requirements to the canary `go.mod` on first run; this is the toolchain, not a loom change.
- No `harnez rate` calls were made during this batch.
- Host review (agy:flash38:med, plus host reruns of canaries a-c and reads of `vt.go:795-820`, `server.go:76-160`, `session.go:510-530`, `:750-790`): claims 1-7 confirmed. Fixed line refs (`vt.go:819`, Repaint rows `:316-365`). Scope of the stall: `sessionMu` is per session process, so a stalled client can block that session's tabs, not other sessions. The `net.Pipe` deadlock in the flow note above is consistent with the stall hypothesis, since `net.Pipe` has no buffer.
