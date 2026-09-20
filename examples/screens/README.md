# Screens

A small inline TUI that can switch to full screen and to the alternate screen,
and that promotes itself to full screen when it is nearly full height.

```
go run ./examples/screens -height 8
```

| Key | Action |
| --- | --- |
| `f` | Toggle inline / full screen (primary screen, from row 1) |
| `b` | Toggle the alternate screen (`?1049`); leaving restores the shell screen and prompt |
| `c` | Toggle auto full screen |
| `l` | Auto full screen uses the alternate screen |
| `+` / `-` | Wanted inline height |
| `m` / `M` | `margin_rows` up / down |
| `p` / `P` | `min_percent` up / down (steps of 10) |
| `q` | Quit |

## Quasi-fullscreen detection

Parameters live in `auto_fullscreen` in `spec/resize.yaml` and are flags here
(`-margin`, `-percent`, `-auto`, `-auto-alt`). The pane is treated as full
screen when its **wanted** height (not the clamped one) satisfies:

- `terminal_rows - wanted_rows <= margin_rows` (default 1: full height counts
  as full screen; the pane always leaves the prompt row, so 0 needs every row), or
- `wanted_rows * 100 >= terminal_rows * min_percent` (when `min_percent > 0`).

Shrinking the terminal can promote the pane; growing it demotes it again. The
promotion uses the alternate screen when `alt` is on, otherwise full screen on
the primary screen (which overwrites the shell's screen content).
