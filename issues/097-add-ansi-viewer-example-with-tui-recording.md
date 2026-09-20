# 097 — Add ANSI viewer example with TUI recording

**Status**: Open
**Priority**: P2
**Severity**: Moderate
**Category**: Feature
**Related**: `examples/loom-demo`, `docs/data/harnez-usage.ansi`

---

## Goal

Add an `examples/ansiviewer` TUI that browses a directory in a left-hand file
pane and renders the selected file in a right-hand viewer pane without
overflow. Text files are shown as plain text, binary and image-like files show
metadata, and `.ansi` files retain their full color and ANSI styling. Extend
the Loom widgets safely as needed and include the viewer in `loom-demo`.

The viewer must also support `ansiviewer --record <time>`: launch a TUI,
capture one bounded screenshot after the requested delay, and cleanly finish.
Recording must work for long-lived TUIs such as `go run ./examples/splash
--watch` (closing the test app after capture) and for self-exiting TUIs such as
`go run ./examples/splash`.

## Acceptance Criteria

- `ansiviewer <dir>` starts the browser rooted at `<dir>` with usable keyboard
  navigation and a left file browser/right viewer layout.
- Plain-text files, metadata-only binary/image files, and colored `.ansi`
  files are rendered according to the goal; ANSI output remains inside the
  viewer pane and does not overflow its bounds.
- The example is available through `loom-demo` and has focused automated
  coverage for file classification, rendering bounds, and key interactions.
- `ansiviewer --record <time>` captures exactly one post-delay TUI snapshot,
  handles both long-lived and self-exiting subprocesses, and terminates or
  reaps the test app cleanly.
- Existing Loom examples and tests continue to pass.

## Notes

Investigate the current widget and terminal-capture APIs before implementation;
do not duplicate spec values or introduce unsafe process handling. Resolve any
format or platform limitations discovered during implementation in the ticket
or accompanying documentation.

## Milestones (lean sprint, dev agent `codex:luna:low`)

Every milestone ends with a committed, `cat`-able evidence file under
`docs/progress/097/` (plain ANSI bytes, one frame, produced headlessly via
`loom.RenderTo` or a Go test with `-update`-style golden write; never require a
TTY). Name: `M<N>-<what>.ansi`, 100x30. Commit the `.ansi` files with the code.
Test runs: use targeted `go test ./examples/ansiviewer/... -run <Name>`; root
package `/dev/tty` failures (`TestTerminalSize`, `TestPaneStartup`) are
pre-existing headless gaps, not regressions.

### M1 — Classification and viewer widget
- Package `examples/ansiviewer`: `Classify(path) Kind` (text, ansi, binary,
  image) and a viewer widget that renders text / metadata / `.ansi` content
  clipped to its bounds (use `Canvas.WriteANSI` if present, otherwise a bounded
  SGR parser inside the example; see ticket 034).
- Tests: classification table, rendering never writes outside bounds
  (wide lines, CJK, long ANSI lines), key scrolling.
- Evidence: `M1-text.ansi`, `M1-ansi.ansi` (renders `docs/data/harnez-usage.ansi`),
  `M1-binary.ansi`.

### M1 Review (host) — delivered `312c617`, NOT accepted; finish as M2 Pre-Work

Delivered: skeleton browser, single-color ANSI writer, 2 weak tests.
Pre-Work / Required Refinements (do these first, commit as `(issue 097 M1b)`):
1. Add `Classify(path) Kind` (text, ansi, binary, image) by extension plus
   content sniff (NUL byte / invalid UTF-8 => binary); binary and image files
   show metadata (name, size, kind) only, never raw bytes. Table test.
2. Replace `writeANSI` with a real SGR handler: reset, bold, dim, underline,
   reverse, 30-37, 90-97, 40-47, 100-107, 38/48;5;n and 38/48;2;r;g;b, using
   loom.Style fields. Skip unknown CSI sequences without printing them. Use
   `Canvas.WriteANSI` if it exists (check first; ticket 034 is not done).
3. Clip by display width, not bytes: iterate runes, use the loom display-width
   helper; CJK/emoji never straddle the right bound. Right pane must not
   overwrite the left file list or exceed `r`.
4. Real bounds tests: fill a canvas with a sentinel, draw wide/long/CJK/ANSI
   lines into a sub-rect, assert every cell outside the rect is unchanged.
   Assert specific cell text/colors, not just non-empty rows.
5. Scrolling: PgUp/PgDn/j/k in the viewer with a test.
6. Add SPDX headers (see other files). Regenerate the M1 `.ansi` evidence by
   rendering the widget (loom.RenderTo, 100x30) from a test or small
   generator over `docs/data/harnez-usage.ansi`, a text file, and a binary
   file. No hand-written evidence. Files: `M1-text.ansi`, `M1-ansi.ansi`,
   `M1-binary.ansi`.

### M2 — Browser + viewer layout and `loom-demo` registration
- `ansiviewer <dir>`: left file browser, right viewer, keyboard navigation
  (up/down/enter/tab focus switch, q quit). Register in `loom-demo` and add a
  PTY smoke test consistent with ticket 085.
- Evidence: `M2-layout.ansi` (rooted at `docs/data`), `M2-narrow.ansi` (60x20).

### M2 Review (host) — committed `a91893d` by host (index.lock); accepted with Pre-Work for M3

Pre-Work / Required Refinements (do first, commit as `(issue 097 M2b)`):
1. `TestANSIWriteClipsToBounds` still only asserts a non-empty row. Rewrite it:
   fill a canvas with a sentinel rune, draw wide, CJK and long SGR lines into
   an inner rect, assert every cell outside the rect is unchanged and that the
   in-rect cells hold the expected text and colors.
2. Add the PTY smoke test for `ansiviewer` like the other examples (ticket 085):
   starts, shows the file list, responds to down/j, quits with q.
3. Git note: if `git commit` fails with index.lock, leave files uncommitted and
   report; the host commits.

### M3 — `--record <time>`
- Launch a subprocess TUI in a PTY, capture exactly one screen snapshot after
  the delay, write it as `.ansi`, then terminate and reap the child. Must work
  for `go run ./examples/splash --watch` (long-lived) and
  `go run ./examples/splash` (self-exiting). No leaked processes (test it).
- Evidence: `M3-record-splash.ansi`, `M3-record-watch.ansi`.

### M3 Review (host) — NOT accepted; finish as M3b

Delivered (committed by host): M2b tests and PTY smoke test. Rejected:
1. `Record` renders a directory snapshot; it must run a *subprocess TUI* (e.g.
   `go run ./examples/splash --watch`, `go run ./examples/splash`) in a PTY,
   wait `<time>`, capture exactly one screen snapshot (feed PTY output into a
   virtual screen, e.g. loom `ParseANSI`/rawscreen if suitable), write it as
   `.ansi`, then terminate and reap the child (process group, no orphans).
2. `main.go` has no `--record <time>` flag. Add it: `ansiviewer --record 2s
   -- <cmd> [args...]` writes the frame to stdout or `-o file`.
3. Tests: self-exiting child (short-lived helper command) and long-lived child
   (`sleep`-like helper); assert exactly one snapshot, child gone afterwards
   (`syscall.Kill(pid, 0)` returns ESRCH), and timeout does not hang.
4. Evidence must be real: `M3-record-splash.ansi` from recording
   `go run ./examples/splash`, `M3-record-watch.ansi` from
   `go run ./examples/splash --watch`. Regenerate; they currently show the
   viewer file list. Verify no `splash` process remains afterwards.
5. Pre-existing `splash --watch` processes (PIDs 920646, 920686) may be
   leaked test children; do NOT kill them, but find what leaked them if it
   is in this repo's tests and fix it.

### M3b Review (host) — PTY lifecycle accepted, snapshot and CLI NOT accepted; finish as M3c

Accepted (committed by host): private PTY, child process-group reap, lifecycle tests.
Required (design given, do not improvise):
1. Snapshot must be a *rendered screen*, not the byte stream. Feed PTY output
   into a virtual screen model (check root `rawscreen.go` first; otherwise write
   a small model in `examples/ansiviewer/ansiviewer` handling: CSI H/f cursor
   move, K/J erase, 2K, m (SGR into cells), ?25/?7/?2026 (ignore), 6n (ignore
   or answer with cursor pos), S scroll, alt screen ?1049). Snapshot = rows of
   cells rendered as SGR text like `Canvas.Row`.
2. Timing: take the snapshot at `<time>` after start. If the child exits
   earlier, use the last screen state *before* its teardown (teardown =
   clearing lines, leaving alt screen, or ?25h at exit): keep a copy of the
   screen at each end-of-frame (`ESC[?2026l`, or 30 ms of quiet) and use the
   latest non-blank one. The splash self-exit capture must therefore show
   "harnez usage", the progress bar and the pill line, not a blank screen.
3. Add the CLI in `main.go`: `ansiviewer --record 2s [-o file] -- <cmd> [args]`.
   Write a test that runs `main` logic (extract a `run(args, stdout)`) for it.
4. Regenerate `M3-record-splash.ansi` (`go run ./examples/splash`) and
   `M3-record-watch.ansi` (`go run ./examples/splash --watch`); after
   stripping SGR, each file must contain the text `harnez usage` and have no
   raw ESC sequences other than SGR. Add a test asserting exactly that on
   the recorded output of a scripted child.

### M3c Review (host) — splash capture accepted; finish as M3d (last two gaps)

1. `M3-record-watch.ansi` is blank after SGR stripping. For a long-lived child
   the snapshot at `<time>` must show the splash frame ("harnez usage"), i.e.
   snapshot the live screen at the deadline, not after teardown. Add a test
   using a scripted long-lived child that paints text then sleeps; assert the
   text is in the snapshot. Regenerate the file and check for `harnez usage`.
2. `main.go` still has no `--record` flag (item 3 of M3b was skipped). Add
   `ansiviewer --record 2s [-o file] -- <cmd> [args...]`, extract
   `run(args []string, stdout io.Writer) error`, and test it with a scripted
   child. Also update `--help` text and the ticket Notes with the usage line.
