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
