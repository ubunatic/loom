# loom session

`loom session -- <cmd>` wraps a command in a PTY, tmux style, and grabs a few hotkeys so loom can work
with the child's ANSI output. Code: `internal/session` (since 378; `cmd/loom` only holds the Cobra
command). Tickets: 369–376, 378.

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
the rendered screen, loom records the current screen, resets the VT parser, then briefly toggles the
child PTY row count to deliver `SIGWINCH` and request a redraw. After 150 ms it records the updated
screen. Each incident produces a `.log`, `-before.ansi`, and `-after.ansi` file. The log contains
the anomaly and the recent raw PTY chunks from the 64 KB ring buffer. FIFO rotation retains at most
five incident bundles (15 files). A 10-second circuit breaker limits repairs and diagnostic writes.
While a debug alert is active, the session bar shows `[!]`; it clears after the cooldown.

## Keys (all from `spec/defaults.yaml` `session`)

- Prefix `ctrl-b` (tmux habit; `ctrl-]` needs AltGr on German keyboards). Prefix twice sends it.
- Direct and swallowed: `ctrl-p` screenshot (`.ansi`, same key as `loom edit`), `ctrl-q` quit, `ctrl-n`
  new tab, `ctrl-f` search, `alt-1`…`alt-9` select tab. Ctrl+digit is unusable (legacy encodings collide).
- Prefix: `s` screenshot, `q` quit, `c` new tab, `n`/`p` next/previous, `1`–`9`, `x` close, `?` help
  (not F1, so children keep it), `m` mouse toggle, `f` search.
- Swallowed keys are lost to the child: `ctrl-f` is forward-char in shells and page-down in less/vim;
  the user chose that trade-off (376).

## Terminal modes and restore

- The `terminal` package (372) owns the reset set: mouse 9/1000/1002/1003/1005/1006/1015, focus, paste,
  kitty pop, DECSTBM, SGR, cursor, auto-wrap, alt screen. `terminal.Restorer` runs once on every exit
  path, including signals (exit 128+sig) and panic. Use it for any raw-mode program.
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
