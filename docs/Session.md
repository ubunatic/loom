# loom session

`loom session -- <cmd>` wraps a command in a PTY, tmux style, and grabs a few hotkeys so loom can work
with the child's ANSI output. Code: `internal/session` (since 378; `cmd/loom` only holds the Cobra
command). Tickets: 369–376, 378.

## Persistent sessions

`loom session` starts or attaches to the interactive `default` session. Starting a command with
`loom session -- <cmd> [args...]` creates a uniquely named session; add `-s <name>` (or
`--session <name>`) to choose a stable name. Sessions run in a background daemon, so closing the
foreground client does not stop tabs or their child processes.

- Prefix `ctrl-b`, then `d` detaches the current client and restores its terminal modes. The session
  remains available in the background.
- `loom session attach [-s <name>]` (alias `a`) attaches to a session. Without a name, Loom uses the
  `default` session or the only active session; if several sessions are active, specify one.
- `loom session detach [-s <name>]` disconnects an attached client. `loom session list` (alias `ls`)
  shows each active session's name, tab count, and attached/detached state.
- `loom session kill [-s <name>]` stops the daemon and its child processes. Control commands default
  to `default` or the only active session when `-s` is omitted.

Session sockets are stored under `$LOOM_SOCKET_DIR` when set, otherwise under `$XDG_RUNTIME_DIR/loom`,
then `~/.local/state/loom`, and finally a per-user directory under the system temporary directory.
The socket directory is private to the user. Loom removes socket files when a daemon exits and prunes
stale sockets after confirming that no daemon is listening.

## Model

- Each tab has its own PTY and `vt.VT` (screen model, bounded scrollback). Only the active tab writes to
  the outer terminal; inactive tabs update their VT only. Switching repaints from `VT.Repaint()`, not by
  replaying bytes.
- The bar is the last row. The child gets rows-1 whenever the bar shows; DECSTBM keeps the child's
  scrolling and clears off the bar, and the bar is redrawn after each output chunk.
- `session.bar: always` (default) or `tabs` (only with more than one tab).
- With more than one tab, loom runs on the outer alt screen so tab text never stays on the main screen.

## Session debug mode

Run `loom session --debug -- <cmd>` to enable per-tab raw-stream tracking and ANSI leakage checks.
Set `LOOM_SESSION_DEBUG=1` to enable it by default; an explicit `--debug` or `--debug=false` flag takes
precedence. Choose a diagnostics directory with `--debug-dir <path>` or `LOOM_DEBUG_DIR=<path>`.
Without either directory setting, captures go to `$XDG_STATE_HOME/loom/debug` or
`~/.local/state/loom/debug`.

When the detector finds a stalled or overlong control sequence, or a suspicious ANSI fragment in
the rendered screen, loom records the current screen, resets the VT parser, then sends one `SIGWINCH`
to the child's foreground process group to request a redraw. The PTY size is not changed, so the child
sees exactly one signal per repair. After 2 s it records the updated screen. Each incident produces a `.log`, `-before.ansi`, and `-after.ansi` file,
written in that order with the `.log` last, so a complete `.log` means both captures exist. The log contains
the anomaly and the recent raw PTY chunks from the 64 KB ring buffer. FIFO rotation retains at most
five incident bundles (15 files). A 10-second circuit breaker limits repairs and diagnostic writes.
While a debug alert is active, the session bar shows `[!]`; it clears after the cooldown.

## Keys (all from `spec/defaults.yaml` `session`)

- Prefix `ctrl-b` (tmux habit; `ctrl-]` needs AltGr on German keyboards). Prefix twice sends it.
- Direct and swallowed: `ctrl-p` screenshot (`.ansi`, same key as `loom edit`), `ctrl-q` quit, `ctrl-n`
  new tab, `ctrl-f` search, `alt-1`…`alt-9` select tab. Ctrl+digit is unusable (legacy encodings collide).
- Prefix: `s` screenshot, `q` quit, `c`/`N` new tab, `n`/`p` next/previous, `1`–`9`, `x` close, `?` help
  (not F1, so children keep it), `m` mouse toggle, `f` search.
- Swallowed keys are lost to the child: `ctrl-f` is forward-char in shells and page-down in less/vim;
  the user chose that trade-off (376).

## Terminal modes and restore

- The `terminal` package (372) owns the reset set: mouse 9/1000/1002/1003/1005/1006/1015, focus, paste,
  kitty pop, modifyOtherKeys (`>4m`), DECSTBM, SGR, cursor, auto-wrap, alt screen. `terminal.Restorer`
  runs once on every exit path, including signals (exit 128+sig) and panic. Use it for any raw-mode program.
- DECSTBM (`CSI r`) homes the cursor, so the reset wraps it in DECSC/DECRC (`ESC 7 … ESC 8`); a bare
  `CSI r` on exit leaves the shell prompt at the top left.
- The attach client clears the screen (`CSI H CSI J`) after restore only once the session sent output:
  leaving the outer alt screen brings back a stale screen, but a failed attach must keep its error visible.
- A killed child cannot reset its modes, so loom must. Each tab's VT tracks the modes its child set;
  on switch, loom resets the outer modes and applies the new tab's.
- Mouse `safe` (default): loom enables tracking only when the active child asked for it, then handles
  bar-row clicks and forwards the rest in the child's encoding. A plain shell keeps native selection.
  Prefix `m` forces loom's mouse on for the tab; the 🖱︎ icon (VS15) shows the state.

## Search (374)

Scrollback is a ring per VT (`session.scrollback_lines`, 5000). Alt-screen children have none, so
search covers their visible screen. Smart case. `Enter`/`↑` older, `↓` newer, `Esc` closes and repaints
the child's current screen; child output during search only updates the VT.

## Pitfalls seen

- Open `/dev/ptmx` and the slave with `O_NOCTTY`. Without it the slave becomes the loom server's
  controlling TTY, the child's `setsid`+`TIOCSCTTY` fails, and fish exits with "No TTY for interactive
  shell".
- Terminal key support varies: run `loom-probe keys` to see whether the terminal speaks kitty keyboard or
  modifyOtherKeys and which bytes each key sends. Gram (Zed fork) supports neither; Ctrl-digits send
  nothing and Alt-digits arrive only via its `Terminal` keymap (384).

- PTY tests that wait for "the bar" can match the one-tab bar before tab 2 exists; wait for the new
  tab's label.
- Overlays (help) must fit the height: 12 rows cut the help's first lines (375).
- Live checks in tmux (`#{mouse_any_flag} #{mouse_sgr_flag} #{alternate_on}`) found bugs green tests
  missed: stale bar after quit, alt screen left on, doubled key names, swapped search direction.
- **PTY Output Stream Framing & Interleaving (382)**: High-frequency TUI monitors like `btop` emit
  truecolor SGR and private modes that cross PTY `read()` chunk boundaries. If raw chunks are forwarded
  to `out` before loom draws the tab bar or mouse overrides (`\x1b7\x1b[...H`), the outer terminal treats
  the UI sequence as an abort of the child's incomplete CSI sequence. When the completing chunk arrives
  (e.g. `20;30m`), it lands in Ground state and prints literally onto the screen. To prevent this,
  `runSession` parses via `t.vt.WriteFramed()` and forwards only complete frames to `out`, withholding
  incomplete escape and UTF-8 tails until completed or safely discarded.
- **VT Parser Continuity & Safe Discard (382)**: `vt.VT` maintains streaming state across chunk boundaries:
  - Multi-byte ESC sequences consume intermediate bytes (`0x20..0x2F`) through final (`0x30..0x7E`) so
    charset commands (e.g. `ESC ( B`) never leak their final character.
  - Cancel (`CAN`/`SUB`, `0x18`/`0x1A`) safely aborts sequences to Ground.
  - Embedded `ESC` inside an in-flight CSI or string control (OSC/DCS/APC/PM/SOS) aborts the prior sequence
    and restarts immediately at the new `ESC`.
  - ITU T.416 colon subparameters (`38:2::R:G:B`) and truncated SGR tails advance safely without misparsing
    color operands as top-level attributes.
  - A 1024-byte safety cap on incomplete controls resets to Ground to prevent unbounded memory growth on
    unterminated streams.
- **Terminal Queries & Cursor Position Reports (CPR) Leaking to Input (391)**: Interactive prompt
  applications (such as `agy` in vim mode, readline, or prompt-toolkit) query cursor position (`\x1b[6n`)
  or device attributes (`\x1b[c`) to align inline popups and completion menus. When forwarded to the
  outer terminal, responses (`\x1b[row;colR`) return on the multiplexer's `stdin`. If not intercepted
  or recognized by `loom.ScanKey`, CPR responses are parsed as plain keys and trailing characters
  (such as `R`, semicolons, and coordinate digits) leak back into the child PTY master, causing stray
  characters on the prompt, duplicated lines, and popup offset displacement.
- **Mouse Motion Tracking vs Native Selection (392)**: When child applications request button-event
  tracking (`\x1b[?1002h`) or any-event tracking (`\x1b[?1003h`), the outer terminal reports drag
  motion events (`\x1b[<...M`), which prevents the outer terminal's native drag-to-select text selection
  unless the user holds a terminal-specific modifier key (such as `Shift`). To prevent hosted applications
  from hijacking text selection, a 3-way mouse model (`off` | `basic` | `full`) clamps outer terminal
  tracking to normal tracking (`\x1b[?1000h` + SGR `\x1b[?1006h`) or filters drag motion sequences in
  Loom, allowing clicks and scrolling to reach the child app while preserving drag selection at the
  session / outer terminal level.
- **Attach gate inherited from the shell (393)**: `loom session` sets `LOOM_SESSION_WAIT_ATTACH=1` for the
  daemon it spawns, and every shell inside a loom session inherits it. The library used to read that
  variable in `serveSession`, so any `go test` or `make test-q1` run from a loom shell held its child
  until a client attached and hung (`TestServerLastTabExitCleansSocket`, and `Tabs:0` in
  `TestServerControlStatusListDetachAndKill`). The gate is now `Options.WaitForAttach`: the library never
  reads the environment, and `cmd/loom` reads the variable once in the daemon branch and clears it so tab
  children do not inherit it. If a session test hangs or reports `Tabs:0` in a loom shell, check
  `env | grep LOOM_` before changing code, and reproduce with `env -u LOOM_SESSION_WAIT_ATTACH`.

