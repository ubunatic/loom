# Batch 9: testing, lifecycle and producer load

Ticket 458 (follow-up to 457), batch 9 of 9. Loom revision for citations:
`9d633f9b482c08831da6fd93af16d9623451479f` (on origin, Codeberg). Loom permalinks use
`https://codeberg.org/ubunatic/loom/src/commit/9d633f9b482c08831da6fd93af16d9623451479f/<file>#L<x>-L<y>`;
the `file:line` form below uses the same lines.

Scope: who restores the terminal on start and exit, `Pane.Post` queue behaviour, the headless VT and PTY
test helpers, and how the tests wait. Input fairness and producer load are measured by canaries in section 4.
The batch covers testing and lifecycle only; the other batches are linked, not repeated.

## 1. Pins (re-resolved 2026-11-10)

Latest release per library, resolved with `git ls-remote` (peeled `^{}` SHA for annotated tags). Where a
library has no release tag, the default-branch HEAD is used and the row says so.

| Library | Source used | Latest release or HEAD | Commit SHA | Plan pin in batch table |
|---|---|---|---|---|
| Loom | `origin` (Codeberg) | `9d633f9b` | `9d633f9b482c08831da6fd93af16d9623451479f` | same revision |
| Bubble Tea v2 | `charmbracelet/bubbletea` | tag `v2.1.1` | `67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7` | plan had `v2.0.10` (TUIOS dep) and `v2.1.1` |
| Bubble Tea default branch | same | HEAD (not a release, not quoted) | `e4d617e7e0c0dd595cc553a856236221c678470f` | n/a |
| charmbracelet/x (teatest, `x/vt`) | `charmbracelet/x` | no release tag for `exp/teatest/v2`; HEAD used | `27ecf3e4d454ff2656204b806b7ee6ef11593746` | plan had the same SHA for X |
| x/vt tag | same | `vt/v0.1.0` (peeled) | `43427fdbad12fde220b737fc36303291a5ce4602` | not in plan |
| gocui | `jesseduffield/gocui` | tag `v0.4` | `a319843434acb2446b85fdcaac286d45dd9e27ca` | plan had HEAD `c81715e9` (kept for reference) |
| TUIOS app | `Gaurav-Gosain/tuios` | no app release; HEAD | `6fe2c0df9922b2eec28f392fee4e196efd86be36` | same SHA |
| TUIOS VT module | same repo | tag `vt/v0.3.2` (peeled) | `fdaab4d0d8af5a409e10303bfd0782f0fe2ef899` | not in plan |

Shallow clones (`git clone --depth 1`, pinned with `git fetch --depth 1` on each SHA) live in
`scratchpad/b9-src/{x,bubbletea,gocui,tuios}`. Those paths are outside the repository.

## 2. Loom from code (revision 9d633f9b)

### 2.1 Terminal setup and restore

- **New enters raw mode and returns without a restore hook.** `New` claims the process-wide pane slot
  (`pane_terminal.go:27`), opens `/dev/tty` (`:30`), then calls `term.MakeRaw` (`:37`). After that point the
  function has no error return: it writes the cursor moves and reserve scroll with ignored write errors
  (`:83-90`), installs the signal handler (`:112`) and returns `p, nil` (`:113`). The doc string tells the caller
  to "Call Close (or defer it) to restore the terminal" (`:17-18`).
- **Run sets up the screen and modes after New.** `run` (`pane_run.go:106`) checks the context before any
  change (`:93` in `RunWatch`, `:351` in the loop), then sets the alternate screen through `changeScreen`
  (`:164`), bracketed paste on (`:170`), and kitty keyboard on (`:173`). Mouse mode is set only when a root
  widget asks for it (`:118-120`). Each of paste and kitty has a `defer` that reverses it on any return from `run`
  (`:171`, `:174`). Mouse, alternate screen and raw mode have no such defer.
- **Close is the only full restore.** `close` (`pane_terminal.go:191`) sets `postClosed`, drops queued posts
  (`:197-198`), releases ownership (`:201`), stops the reader (`:218-223`, bounded wait of 300 ms), disables
  mouse (`:226`), turns off paste and kitty (`:227-228`), clears the inline region if not in the alternate screen
  (`:230-238`), then calls `terminal.Restore` (`:240`) and closes the tty (`:242`).
- **`terminal.Restore` writes the reset set.** `Restore` writes `Reset(altScreen, pop)` (`terminal/terminal.go:64-70`).
  `ResetSequences` (`terminal/terminal.go:24-38`) disables mouse modes 9, 1000-1006, 1015, focus 1004, paste
  2004, and resets SGR, cursor visibility, cursor shape and wrap. `ResetAltScreen` is `?1049l` (`:41`) and
  `ResetKittyKeyboard` is `<u` (`:44`).

### 2.2 Who must call Close, and what restores by itself

| Exit from `Run` | Restores by itself | Caller must call `Close` for |
|---|---|---|
| Normal return (`nil`) from quit, interrupt or key dispatch (`pane_run.go:441-443`, `:551`) | paste and kitty (defers at `:171`, `:174`) | mouse, alternate screen, raw mode, tty fd, pane slot |
| Context error (`pane_run.go:351-352`, `:439-440`) | paste and kitty | everything above |
| Collect callback error (`pane_run.go:445-448`) and input read error (`:554-556`) | paste and kitty | everything above |
| Panic in any widget or callback (`pane_run.go:136-141`) | `close` runs inside the deferred recover, then the panic is re-raised | nothing (the call is idempotent through `p.restored`, `pane_terminal.go:192-195`) |
| `New` returns an error before raw mode (`pane_terminal.go:28-41`) | the tty is closed and the pane slot released (`:32`, `:40`) | nothing |
| `New` returns a pane (after raw mode, `:113`) | nothing | always, and `Run` is not needed to reach the caller's `Close` |

`RunPane` (`driver.go:29-40`) is the only helper in the tree that pairs `New`, `Run` and `Close` for the caller.
The `cmd/loom` callers do `defer pane.Close()` after `New` (`cmd/loom/main.go:381`, `cmd/loom/widgets.go:246`,
`cmd/loom/info.go:28`). `cmd/loom-demo/main.go:163` does the same.

Signal path: SIGINT, SIGTERM and SIGHUP go to `p.interrupts` (`pane_wait_unix.go:38`). The loop returns
`nil` after `requestClose(CloseReasonInterrupt)` (`pane_run.go:441-443`), so a signal is not a restore
by itself: the caller still needs `Close`. The comment at `driver.go:22-28` says "Terminal restore on SIGINT/SIGTERM
and panic is handled by Pane itself"; the code shows that only the panic path restores.

### 2.3 Post, runPosted and Close

- **Post is a mutex plus an unbounded slice.** `Post` (`pane_run.go:50-62`) takes `postMu`, returns if the pane
  is closed (`:55-58`), appends to `p.posted` (`:59`) and calls `Invalidate` (`:61`). The slice has no cap, so the
  queue is bounded only by memory. `Invalidate` is a non-blocking send to a buffered channel of size 1
  (`:37-43`).
- **The doc comment overclaims.** `Post` says it "never drops callbacks" (`pane_run.go:45-49`). Close does drop them:
  `close` sets `p.posted = nil` (`pane_terminal.go:198`), and `runPosted` breaks out of the batch when `postClosed` is
  set (`pane_run.go:72-77`). The comment's "never drops" holds only while the pane is open.
- **runPosted takes the whole batch.** `runPosted` (`pane_run.go:66-81`) takes the slice and resets it (`:67-70`),
  then runs each callback in FIFO order (`:71-79`). Posts made during the batch go into the new slice and wait
  for the next turn (doc `:64-65`).
- **Order in the loop.** The loop checks the context (`pane_run.go:351`), then runs `runPosted` (`:354`), then
  runs a frame or waits in `select` (`:439`). `select` picks at random among ready channels (Go semantics), so
  an input read ready at the same time as an `invalidate` wake is not given priority (`:439-554`).
  Input is therefore handled only after the whole posted batch has run.
- **Callbacks run on the loop goroutine.** The `runPosted` callbacks run inside `run` on the Run goroutine, so
  a slow callback delays input and frames (section 4, canary B).
- **Panic inside a callback** is a panic inside `run` and takes the same recover path (`pane_run.go:136-141`).
  The callback's batch is not resumed.

### 2.4 Test helpers and wait style

- **loomtest** (`loomtest/doc.go:1-21`) is a headless VT: `NewVT` (`loomtest/vt.go:12`) wraps `vt.NewVT`.
  Its doc says it is "a deterministic virtual terminal ... without a PTY, subprocess ... or timing dependency",
  and that scroll regions, insert/delete and terminal queries are unsupported. "A VT is not safe for concurrent use."
- **CompareWidget / CompareANSI** (`loomtest/compare.go:23`, `:53`) compare rendered output against a reference
  screen. `compareANSIReport` (`:57`) builds a report from `parseComparisonScreen` (`:78`).
- **ptytest.Start** (`internal/ptytest/session.go:38`) skips under `testing.Short()` (`:40-41`), opens a pty
  pair, starts the child with `Setsid` and `Setctty` (`:75`), and pins `LOOM_ZWJ=join` and `LOOMCOLOR=truecolor`
  (`:73-74`). Child output feeds a `loomtest` VT through `pump` (`:80`, `:85`).
- **Waits are condition-based.** `WaitForState` (`session.go:233-265`) loops on a predicate, wakes on a `changed`
  channel, and fails on a `time.NewTimer` deadline (`:246-263`). `Wait` (`:271-291`) waits for exit and EOF with a
  timer. `WaitFor` uses `timeout*RaceFactor` (`:216`). `RaceFactor` is defined per race build
  (`internal/ptytest/racefactor_race.go`, `racefactor_norace.go`).
- **Sleeps still exist in loom tests.** `TestTiers` says no wall-clock sleeps. Found: `pane_kitty_test.go:151`
  (`time.Sleep(50ms)`), `pane_suspend_test.go:133` (a poll loop with `time.Sleep(10ms)` until a 2 s deadline),
  `:298` (100 ms "let SIGWINCH arrive"), `:418` (50 ms "let the capture see the last writes"), and
  `pane_pty_test.go:323`, `:541`, `:591`, `:611`, `:917`, `:987`. The suspend poll at `:133` already has a deadline.
  The others use fixed budgets. This is test hygiene, not a lifecycle bug, but it is a TestTiers violation.
  Repo-wide, `grep -rn time.Sleep --include=*_test.go .` finds 119 calls in 25 test files; the list above is only the
  pane and PTY files. Each file was not checked for a justified use, so the 119 figure is a count, not a violation count.
- **PTY tests skip under Short.** `pane_pty_test.go` checks `testing.Short()` before its PTY work (for example
  `:269`, `:402`, `:467`, `:864`), which matches the full-tier rule.

### 2.5 Existing lifecycle tests (what they cover)

| Test | File:line | Covers |
|---|---|---|
| `TestPanePostFromManyGoroutinesKeepsOrderOnLoop` | `pane_post_test.go:74` | 20 goroutines x 50 posts, FIFO on the loop |
| `TestPanePostBeforeRunDeliveredOnFirstLoopTurn` | `pane_post_test.go:118` | posts before Run |
| `TestPanePostAfterCloseIsNoOp` | `pane_post_test.go:148` | post after Close does nothing |
| `TestPaneKittyKeyboardPushPop` | `pane_kitty_test.go:17` | kitty push and pop |
| `TestPaneKittyKeyboardPanicPops` | `pane_kitty_test.go:93` | kitty pop after panic |
| `TestPaneKittyKeyboardAltScreenBalanced` | `pane_kitty_test.go:142` | kitty balanced across the alt screen |
| `TestPaneSuspendRestoresAfterPanicAndError` | `pane_suspend_test.go:346` | Suspend restores after panic and error |
| `TestPaneSuspendCloseInsideFn` | `pane_suspend_test.go:385` | Close inside a Suspend callback |
| `TestPaneMouseCaptureEndedOutsideDoesNotLeakIntoFrame` | `mouse_capture_test.go:240` | drag ended outside leaves no frame residue |

Not covered by any of these (checked by grep for `Post` and `Close` in tests): Close with posts still queued,
a slow callback that delays input, a Post burst while keys arrive, an early error from `RunWatch` or `Run`
followed by a missing `Close`, and a Draw panic followed by the mode check. Those are canaries in section 4.

## 3. Feature matrix

Cell values: **proved** (code read to the executing caller, or canary), **partial** (some code path, limit
stated), **disproved** (code or canary contradicts the claim), **missing** (grep finds nothing), **n/a** (the
library has no such feature), **unknown** (source not read in this batch, or a canary failed). Loom cells use
`file:line` at 9d633f9b. Other cells use `file:line @ pin` with the pins from section 1 (gocui at v0.4 unless
marked HEAD).

| Feature | Loom | Bubble Tea v2 (+teatest) | gocui v0.4 | TUIOS |
|---|---|---|---|---|
| Early-error cleanup (before first frame) | proved negative: `Run` restores only paste and kitty (`pane_run.go:170-174`); mouse, alt screen and raw mode wait for `Close` (`pane_terminal.go:191-242`); canary C earlycancel | partial: `initTerminal` error returns before raw mode (`tea.go:1066-1069`); restore runs in `shutdown` (`tea.go:1311`); `initTerminal` body not read | unknown, not read | partial: termios captured before `Program.Run` (`cmd/tuios/signal.go:67-70`, `:134-152`); cited upstream issue not found |
| Panic restore | proved: `recover` then `close` then re-panic (`pane_run.go:136-141`); canary C panicdraw and panicdraw-mouse balanced | proved (code): `Run` recover (`tea.go:1059-1062`), `recoverFromPanic` shuts down (`tea.go:1317-1322`); not canaried | unknown, not read | partial: teardown on goroutine panic (`cmd/tuios/signal.go:70`, `:203-210`) |
| Close drops queued posts | disproved against the doc: `Post` says it "never drops callbacks" (`pane_run.go:45-49`); `close` clears the queue (`pane_terminal.go:197-198`) and the batch breaks (`pane_run.go:72-77`); canary B: 1 of 10 ran (slowclose), 4 of 10 (inclose), three runs each | n/a for Post; `Send` after exit is a no-op (`tea.go:1215-1219`) | unknown | n/a |
| Post bound / backpressure | disproved: unbounded append (`pane_run.go:59`); canary A sampled queue up to 70,823 entries with no cap or wait | proved: unbuffered `msgs` (`tea.go:620`); `Send` blocks until the loop reads or the context ends (`tea.go:1215-1219`) | partial: `userEvents` has buffer 20 (`gui.go:118` @ v0.4), `Update` spawns a goroutine per call and does not guarantee order (`gui.go:368-374` @ v0.4) | unknown, not read (uses Bubble Tea) |
| Input fairness under Post load | partial: `runPosted` runs the whole batch before `select` (`pane_run.go:354`, `:66-81`); canary A: key handled after 151 to 85,055 of 100,000 posts (5 runs), all posts done in 22 to 23 ms | unknown, not canaried | unknown | unknown |
| Headless VT harness | proved: `loomtest.NewVT` (`loomtest/vt.go:12`), doc `loomtest/doc.go:1-21` | partial: teatest `WaitFor` reads a captured output stream (`x/exp/teatest/v2/teatest.go:78-112`); not a VT | unknown | partial: `vt/ghostty_differential_test.go` (build tag); `internal/app/frame_render_test.go:113-123` compares 2D string rendering with ultraviolet, not a VT |
| Cell compare | proved: `CompareWidget`, `CompareANSI` (`loomtest/compare.go:23`, `:53`) | unknown, not read | unknown | proved (code): `TestFrameRenderMatchesUltraviolet` compares bytes with ultraviolet's trimmed output (`internal/app/frame_render_test.go:113-118`) |
| Differential VT vs reference emulator | missing in tree: no `*_test.go` uses x/vt, ghostty, libvterm or xterm as oracle; canary D (out of tree) shows one split-boundary mismatch vs x/vt | n/a (renderer, not emulator) | n/a | partial: `vt/ghostty_differential_test.go:1-8` (libghostty oracle, `//go:build ghostty`, not run here); `vt/vtgen/vtgen.go:1-9` (generator) |
| PTY end-to-end tests | proved: `ptytest.Start` (`internal/ptytest/session.go:38-80`); PTY tests skip under `-short` (`pane_pty_test.go:269` and others) | partial: teatest is not a PTY harness (not read further) | unknown | unknown, `e2e/tui` exists, not read |
| Wait by condition, not sleep | partial: `WaitForState` waits on a `changed` channel with a timer (`internal/ptytest/session.go:233-265`); pane tests still sleep (section 2.4); 119 sleeps repo-wide in 25 test files | disproved for teatest `WaitFor`: polls with `time.Sleep(CheckInterval)` (`teatest.go:109`), default 50 ms, total `Duration` 1 s (`:92-93`); `WaitFinished` uses a timer and `doneCh` (`teatest.go:186-193`) | unknown | unknown |

Note on the gocui cell: the buffer is `g.userEvents = make(chan userEvent, 20)` at `gui.go:118` in v0.4. The
enclosing constructor's name was not checked, so only the line is cited.

## 4. Canary results

All Loom canaries run against a `git archive` copy of `9d633f9b` (`scratchpad/b9-src/loom-head`), using
`internal/ptytest` for PTY children. Every wait is a condition with a deadline (`ptytest` wait, or a channel plus
`time.After`); no sleeps were added. The working tree was not changed. Raw byte logs and JSON results are in
`scratchpad/b9-canary/out/`.

| Canary | Command | Expected | Actual | Status |
|---|---|---|---|---|
| A burst, 8 producers x 12,500 `Post` | `go test -run TestB9ABurst -count=5 -v .` in `loom-head` | all 100,000 run; key handled promptly; frames bounded | 100,000 run in 22 to 23 ms per run; max queue sampled 44,821 to 70,823; enqueue-to-run p50 9.0 to 9.8 ms, p99 14.8 to 15.8 ms, max 15.0 to 15.9 ms; frames drawn 5 to 7; key handled after 4,178, 898, 85,055, 1,043, 151 posts | partial: all posts run, input waits behind batches |
| B1 slow callback, close from another goroutine | `go test -run TestB9BSlowCloseExternal -count=3 -v .` | all 10 run (doc says "never drops") | ran 1, dropped 9, `Run` returned, all three runs | disproved (doc claim) |
| B2 close inside callback 3 | `go test -run TestB9BCloseInCallback -count=3 -v .` | all 10 run (doc) | ran 4, dropped 6, all three runs | disproved (doc claim) |
| C1 Draw panic, alt screen | `go test -run TestB9CPanicDraw -v .` | every enable has a disable | `?1049h` 1, `?1049l` 1; `?2004h` 1, `?2004l` 2 (defer plus reset set); `?1002` 0 | proved |
| C2 Draw panic, alt screen and mouse 1002 | `go test -run TestB9CPanicDrawMouse -v .` | every enable has a disable | `?1049` 1/1; `?2004` 1/2; `?1002h` 1, `?1002l` 2 | proved |
| C3 `RunWatch` with cancelled context, no Close | `go test -run TestB9CEarlyCancel -v .` | ? | returns `context canceled`; 27 bytes, no mode sequence; termios `raw` at exit | proved: Run leaves raw mode for the caller's Close |
| C4 same, then Close | `go test -run TestB9CEarlyCancelClose -v .` | restore after Close | 169 bytes: `?2004l` 1, `?1002l` 1, `?25h` 1 (reset set); termios after close not observable (fd closed by `close`) | partial: bytes prove the reset, termios cannot be read |
| C5 SIGINT, no Close | `go test -run TestB9CSigint -v .` | per `driver.go:22-28`: "handled by Pane itself" | `Run` returned nil; `?1049h` 1, `?1049l` 0; `?25l` 1, `?25h` 0; termios `raw` | disproved for `Pane.Run` alone (see discrepancy 2) |
| C6 SIGINT, then Close | `go test -run TestB9CSigintClose -v .` | balanced | `?1049` 1/1; `?25l` 1, `?25h` 1; `?2004` 1/1 | proved with Close |
| C7 write or read failure injected at startup | not run | a failing tty write or read makes `New` or `Run` fail | `tty` writes ignore errors (`pane_terminal.go:83-90`; `pane_modes.go:136` is in `writeDirect`, which has no caller); the only reader error return is `pane_run.go:554-556`, not reached by an injection that does not touch the code | unknown: no injection point without a code change |
| D replay, split at every byte | `go test -run TestB9Replay -v .` in `scratchpad/b9-canary` | 0 mismatches for both emulators | `sgr-cup-wide` (16x6, 48 splits): x/vt 0 mismatches; loom 1 mismatch at split 22; `scroll` (12x4, 50 splits): both 0 | loom: discrepancy 3 |
| D cross-emulator, unsplit | same | same screen and cursor | `sgr-cup-wide`: 0 of 6 rows differ, cursor (10,0) both; `scroll`: 0 of 4 rows differ, cursor (2,3) both | proved on these streams |

Canary D notes: the first stream I used scrolled its SGR and wide-character content off a 4-row screen, so it tested
little. It was replaced with the two streams above. `go mod tidy` in `b9-canary` added dependencies; that module is
disposable.

Canary C byte logs use the mode counts from `grep -obUaP` on `out/*.raw`. Counts, not offsets, are reported.

### Discrepancies (stop rule, recorded in full)

1. **Post callbacks dropped on Close (canary B1, B2).** Steps: post 10 callbacks; callback 0 blocks on a channel; another
   goroutine calls `Close`, then releases callback 0. Result: 1 ran, 9 dropped, `Run` returned. The second run
   (close inside callback 3) gives 4 ran, 6 dropped. The cause is `close` clearing `p.posted` (`pane_terminal.go:197-198`)
   and `runPosted` breaking on `postClosed` (`pane_run.go:72-77`). The doc says the opposite (`pane_run.go:45-49`).
   Reproduced three times each. Not fixed.
2. **SIGINT without Close (canary C5).** Steps: `New`, `Run` with a root that draws `READY`, parent sends SIGINT.
   Result: `Run` returns nil, the alt screen is entered and never left (`?1049h` 1, `?1049l` 0), the cursor stays
   hidden (`?25l` 1, `?25h` 0), termios stays raw. The comment at `driver.go:22-28` says "Terminal restore on SIGINT/SIGTERM
   and panic is handled by Pane itself". For `RunPane` this holds, because `RunPane` calls `Close` (`driver.go:29-40`).
   For direct `Pane` users it does not. Not fixed.
3. **loomtest VT: combining mark lost at a write boundary (canary D, `sgr-cup-wide`).** Steps: split the stream at
   byte 22. Bytes 20 and 21 are `e` (0x65) at index 21 in the stream; the combining acute U+0301 is bytes 0xCC 0x81 at
   22 and 23. Chunk 1 ends with `e`, chunk 2 starts with the mark. Result: loom's screen text has `Qe` where the
   unsplit write gives `Qé`; x/vt matches the unsplit result for every split. The other 47 splits pass in loom.

Host check (codex:terra:med, reran B and D; agy:flash38:med checked the matrix and quotes):
- 1: dropping on Close is reasonable; the bug is the doc. `Post` (`pane_run.go:45`) promises "never drops
  callbacks" and separately "Post after Close is a no-op"; `close` discards pending work (`pane_terminal.go:196`).
  Contract to document: accepted callbacks run only while the pane is open; Close discards the rest.
- 2: follows the caller-owned Close contract (`New`, `RunWatch`, examples, `RunPane` calls Close). Only the
  comment at `driver.go:27` is wrong; panic restore inside `Run` is separate and works.
- Canary A does not measure key send-to-handler time, only how many posts ran before the key; the fairness
  problem is plausible (whole snapshot runs before input, `pane_run.go:66-81`) but not quantified in ms.
- 3: `loomtest.VT` is an alias of `vt.VT` (`loomtest/vt.go:8`); `vt.VT.Write` (`vt/vt.go:507`) buffers
  incomplete UTF-8 but commits a complete base character before the next write brings its combining mark.
  `loom session` uses `WriteFramed`, which calls the same `Write`, so session tabs share the defect (not run
  end to end). Distinct from 433 (skin tone in `measure`).
- Sleeps: a sample of five (`pane_kitty_test.go:151`, `pane_suspend_test.go:133`, `:298`, `:418`,
  `pane_pty_test.go:323`) are timing waits that `docs/TestTiers.md:44` forbids; the 119 count is not 119
  violations.

## 5. Upstream mechanisms (quoted, pinned)

Permalinks use the pins from section 1. Quotes are verbatim and short. The gocui quote is from v0.4; the
v0.4 tree has no `UpdateAsync` (grep count 0), so the plan's `UpdateAsync` row applies to gocui HEAD only, not to v0.4.

**U1. teatest `WaitFor` polls with a sleep** (`x/exp/teatest/v2/teatest.go`,
[L102-L109](https://github.com/charmbracelet/x/blob/27ecf3e4d454ff2656204b806b7ee6ef11593746/exp/teatest/v2/teatest.go#L102-L109)):

```go
for time.Since(start) <= wf.Duration {
	if _, err := io.ReadAll(io.TeeReader(r, &b)); err != nil {
		return fmt.Errorf("WaitFor: %w", err)
	}
	if condition(b.Bytes()) {
		return nil
	}
	time.Sleep(wf.CheckInterval)
```

Annotation: the condition is checked every `CheckInterval` (50 ms default, `:92-93`). The `io.ReadAll` reads until EOF, so the
deadline is checked only between full reads. Limit: a live reader that never reaches EOF is not bounded by the deadline in this code.

**U2. teatest final wait has a timer and a done channel** ([L186-L193](https://github.com/charmbracelet/x/blob/27ecf3e4d454ff2656204b806b7ee6ef11593746/exp/teatest/v2/teatest.go#L186-L193)):

```go
if fopts.timeout > 0 {
	select {
	case <-time.After(fopts.timeout):
	...
	case <-tm.doneCh:
```

Annotation: this is event-based (program finished or timer), unlike U1.

**U3. Bubble Tea `Send` blocks on an unbuffered queue** (`tea.go`,
[L1215-L1220](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/tea.go#L1215-L1220);
the channel is created at [L620](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/tea.go#L620) as `msgs: make(chan Msg),`):

```go
func (p *Program) Send(msg Msg) {
	select {
	case <-p.ctx.Done():
	case p.msgs <- msg:
	}
}
```

Annotation: a sender waits until the loop receives the message or the program context ends. This is backpressure by
design, and it is not a Post-style queue.

**U4. Bubble Tea recovers panics and restores the terminal through shutdown**
([`recoverFromPanic` L1317-L1322](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/tea.go#L1317-L1322),
[`shutdown` restore L1311](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/tea.go#L1289-L1311)):

```go
func (p *Program) recoverFromPanic(r interface{}) {
	...
	p.shutdown(true) // Ok to call here, p.Run() cannot do it anymore.
```

```go
		_ = p.restoreTerminalState()
	})
```

Annotation: the panic path runs `shutdown`, and `shutdown` restores the terminal (once, through `shutdownOnce`). The
restore is therefore on the panic path, not on the `initTerminal` error path (U5).

**U5. Bubble Tea returns from `Run` when `initTerminal` fails, before raw mode** ([L1066-L1069](https://github.com/charmbracelet/bubbletea/blob/67b4c7a8a2b34a0b1ce15261b0c705bc7e1a04b7/tea.go#L1066-L1069)):

```go
	// Check if output is a TTY before entering raw mode, hiding the cursor and
	// so on.
	if err := p.initTerminal(); err != nil {
		return p.initialModel, err
```

Annotation: the early return is before the shutdown path. `initTerminal`'s own body is in another file and was not read,
so whether it undoes a partial change is unknown.

**U6. gocui `Update` spawns a goroutine and `UpdateAsync` is absent at v0.4** ([gui.go L370-L374 at v0.4](https://github.com/jesseduffield/gocui/blob/a319843434acb2446b85fdcaac286d45dd9e27ca/gui.go#L370-L374),
queue created at [L118](https://github.com/jesseduffield/gocui/blob/a319843434acb2446b85fdcaac286d45dd9e27ca/gui.go#L118)):

```go
// the user events queue. Given that Update spawns a goroutine, the order in
// which the user events will be handled is not guaranteed.
func (g *Gui) Update(f func(*Gui) error) {
	go func() { g.userEvents <- userEvent{f: f} }()
}
```

Annotation: each call starts a goroutine that blocks on a channel of size 20, so producers can pile up goroutines
without bound, and ordering is not guaranteed. The HEAD tree (`c81715e9`) has `UpdateAsync` and a task-based
`Update` (`gui.go:662-679`); it is not the pinned release.

**U7. TUIOS differential harness needs a libghostty build** ([vt/ghostty_differential_test.go L1-L8](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/ghostty_differential_test.go#L1-L8)):

```go
//go:build ghostty

package vt

// The differential harness: the same bytes go to the pure emulator and the
// libghostty-backed one, and the observable surface must agree. This is the
// only place both implementations exist in one process; the shipped binary
// compiles exactly one.
```

Annotation: the oracle is libghostty, compiled only under the `ghostty` build tag. Default `go test` skips the file. Not run here.

**U8. TUIOS frame renderer test defines correctness as ultraviolet's bytes** ([internal/app/frame_render_test.go L113-L115](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/internal/app/frame_render_test.go#L113-L115)):

```go
// TestFrameRenderMatchesUltraviolet is the renderer's definition: over random
// frames its bytes are ultraviolet's trimmed bytes.
func TestFrameRenderMatchesUltraviolet(t *testing.T) {
```

Annotation: a byte-level oracle, not a semantic VT one. It tests the renderer's output against a library.

**U9. vtgen generates well-formed sequences** ([vt/vtgen/vtgen.go L1-L9](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/vt/vtgen/vtgen.go#L1-L9)):

```go
// Package vtgen generates terminal input that looks like something a program
// would actually emit.
//
// Random bytes are a poor fuzzer for a VT emulator. The parser rejects almost
// all of them in its first state, so a campaign spends its budget proving that
// garbage is garbage and never reaches the code that moves the cursor, sizes a
```

Annotation: a generator for valid, parameter-stressing sequences. Its oracle is in `ghostty_differential_test.go` (U7).

**U10. TUIOS early-start wrapper** ([cmd/tuios/signal.go L59-L70](https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/cmd/tuios/signal.go#L59-L70)):

```go
// Call it before Program.Run: it records the terminal's state before Bubble
// Tea puts it in raw mode, which is what a forced exit restores. Call
// runReturned as soon as Program.Run returns: ...
func armSignalQuit(p *tea.Program) (runReturned, finish func()) {
	restore := captureTerminalState()
	var returned atomic.Bool
	crashexit.SetTeardown(func() { resetTerminalForExit(restore, returned.Load()) })
```

Annotation: this is TUIOS's own wrapper. It captures termios before Bubble Tea enters raw mode, and it restores on a
forced exit or a goroutine panic (`crashexit`). The comment cites no upstream issue, and a search of `.go` and `.md` files
for `issues/` and Bubble Tea issue links found no match (`grep -rn 'issues/[0-9]'`, `bubbletea/issues`). The cited upstream
issue is therefore **unknown**.

## 6. Loom vs upstream comparison

- **Restore on early error:** Loom needs the caller to call `Close` after any `Run` error (section 2.2). Canary C3 and
  C4 show the bytes: `Run` writes nothing on the cancelled path, and `Close` writes the reset set. Bubble Tea restores in
  `shutdown` (U4), which the panic path calls; its early `initTerminal` error (U5) is not restored in the code shown.
  TUIOS captures termios before `Program.Run` (U10), so it can restore on a forced exit.
- **Panic restore:** Loom and Bubble Tea both restore on panic (canary C1, C2; U4). Loom's restore is one `close`
  call; Bubble Tea's is `shutdown` after the error is recorded.
- **Post vs Send:** Bubble Tea's `Send` blocks until the loop reads it (U3), so a producer is slowed by the loop. Loom's
  `Post` never blocks but grows without a cap (canary A reached about 70,000 queued). gocui has a bounded channel of 20
  but one goroutine per `Update` (U6), so its bound applies to the loop, not to the producers.
- **Input fairness:** Loom runs a whole batch before input (`pane_run.go:354`); canary A showed keys waiting behind
  up to 85,055 posts. Bubble Tea's input and message order was not canaried, so this comparison is one-sided.
- **Waits in tests:** Loom's `WaitForState` is condition-based (section 3), but its pane tests still sleep in 9 places.
  teatest's `WaitFor` polls with a sleep (U1), and its final wait is event-based (U2). So neither is fully sleep-free.
- **Differential oracle:** Loom has no in-tree reference emulator. TUIOS has a libghostty oracle behind a build tag (U7).
  Canary D used x/vt out of tree and found one Loom mismatch (discrepancy 3). That result is not a full differential.
- **Cell compare:** Loom's `CompareWidget` and TUIOS's `TestFrameRenderMatchesUltraviolet` (U8) both compare output
  against a reference. Their oracles differ in kind: Loom compares screen cells, TUIOS compares renderer bytes.

Same-harness rule: Bubble Tea, gocui and TUIOS were not run through canaries A to D. Their cells above are code-read
only, and the comparison says so where it applies.

## 7. Stage 4 proposals

Nothing here is applied. The integrator applies these serially.

### 7.1 features.md rows (proposed)

| Row (proposed) | Loom status | Evidence |
|---|---|---|
| High-rate producers: Send/Post without blocking | partial | `Post` never blocks on the loop (`pane_run.go:50-62`); unbounded (`:59`); canary A: 100,000 posts in 22 to 23 ms, keys handled after 151 to 85,055 posts |
| Cleanup when start fails before first frame | **keep unchecked** | `Run` does not restore alt screen, mouse or raw mode on its early returns; `Close` required (`pane_terminal.go:191-242`); canary C3 and C4 |
| Panic recovery restores terminal | proved | `recover`, `close`, re-panic (`pane_run.go:136-141`); canary C1 and C2, byte pairs balanced |
| Headless model test harness | proved (headless VT); partial (model) | `loomtest.NewVT` (`loomtest/vt.go:12`), `CompareWidget` (`loomtest/compare.go:23`); no model-level harness read |

### 7.2 summary.md wording (proposed)

"Batch 9 (testing and lifecycle). Loom restores the terminal on panic, but not on a `Run` error or a signal: the caller must
call `Close`. `Pane.Post` never blocks but has no bound, and `Close` drops queued callbacks, which contradicts its doc.
Input waits behind whole `Post` batches (canary: up to 85,055 posts). A VT split-boundary mismatch exists in `loomtest` for a
combining mark (canary D). Bubble Tea's `Send` blocks by design; teatest polls with a sleep. TUIOS has a libghostty
differential oracle behind a build tag."

### 7.3 Ticket candidates

Filed by the host after the check: 477 (P-C, `vt.VT`, P2), 478 (fairness and bound from P-A, canary
first, P2), 479 (P-A drop-on-Close and P-B, both as doc fixes, P3), 480 (P-D, P3).

- **P-A, Post contract.** Choose one: keep "never drops" and make `Close` run the rest, or document drop-on-close. Add a bound
  or a documented unbounded policy, and a batch cap for fairness. Evidence: canary A (fairness, queue) and B1, B2 (drops).
  Overlap: issue 320 (Closed, "Pane Post to run widget mutations on the event loop from any goroutine"), which designed `Post`.
- **P-B, Run-without-Close on signal.** Either make `Run` restore on SIGINT/SIGTERM/SIGHUP, or change the `driver.go` comment.
  Evidence: C5 bytes (alt screen and cursor left), C6 balanced. Overlap: issue 372 (Closed, idempotent terminal restore),
  issue 319 (Closed, Suspend restores modes), issue 002 (Done, `RunPane` helper, which already closes).
- **P-C, VT combining mark at a write boundary.** Fix in the VT's grapheme handling, then add a split-every-byte test.
  Evidence: canary D, `sgr-cup-wide` split 22. Overlap: no tracker match for "combining mark split"; the width and grapheme work
  belongs to batch 2.
- **P-D, sleeps in pane PTY tests.** Replace the 9 pane sleeps with condition waits (`WaitForState`). Overlap: issue 377
  (Closed, "session sleeps eliminated"), which covers the session tests, not the pane tests.

## 8. Flow notes

- Pins: the plan's Bubble Tea v2.0.10 and x pins were replaced by latest (v2.1.1 and x HEAD). The TUIOS VT tag is
  `vt/v0.3.2` (`fdaab4d0`); the app is HEAD. gocui's latest tag is v0.4 (`a3198434`), not the plan's HEAD. The gocui quotes are
  from v0.4, so the plan's `UpdateAsync` row does not apply to that tag.
- The harnez tracker search with the requested keyword list returned no rows. Topic searches found 320, 372, 319, 002, 060,
  287 and 377. Overlap cells use those.
- Canary A's fairness numbers vary run to run (151 to 85,055). Five runs are reported; the spread is the result, not an average.
- Canaries C3 and C4 cannot read termios after `Close`: `close` closes the fd, so the check reports `bad file descriptor`. The
  restore is therefore shown by bytes only.
- Canary C7 (startup write or read failure) was not run: there is no injection point without a code change, and I did not change the working tree.
- Canary D: `go mod tidy` in the disposable harness changed its go.mod; the harness is outside the repository.
- No canary ran against Bubble Tea, gocui or TUIOS. Those cells are code-read and marked as such.
- The TUIOS early-start issue number is unknown (section 5, U10).
- Host review fixed line refs: `pane_terminal.go:27`/`:30` (was `:28`/`:33`), teatest default
  `CheckInterval` 50 ms (was 1 s), gocui quote spans `gui.go:371-375` (cited `370-374`), TUIOS
  `frame_render_test.go` is a string-render test, not a VT harness.
- Live terminal use: none. Every PTY child ran under `ptytest` in the archive copy, not on the owner's terminal.
