# Screens

A small inline TUI with a visible border that switches to full screen (the
alternate screen) and back, and that promotes itself when it is nearly full height.

```
go run ./examples/screens --height 16 --width 60 --theme plain
```

| Key | Action |
| --- | --- |
| `f` | Toggle inline / full screen. Full screen is the alternate screen (`?1049`), so the scrollback stays clean and leaving restores the shell screen and prompt |
| `n` | Toggle full screen on the primary screen (demo only; it overwrites shell rows) |
| `c` | Toggle auto full screen |
| `l` | Auto full screen uses the alternate screen |
| `+` / `-` | Wanted inline height |
| `w` / `W` | Pane width wider / narrower, up to the terminal width |
| `t` | Next theme (`--theme`) |
| `a` | Astra star field background on/off |

The bottom row of the app has clickable `[-]` `[+]` buttons for width and
height, and `[Theme]` and `[Astra]` buttons (left click).

| `m` / `M` | `margin_rows` up / down |
| `p` / `P` | `min_percent` up / down (steps of 10) |
| `q` | Quit |

## Quasi-fullscreen detection

The point is a stable screen: an inline pane that is nearly as big as the
terminal reflows badly when the terminal changes, so it switches to full
screen; a clearly smaller pane stays inline. Parameters live in
`auto_fullscreen` in `spec/resize.yaml` and are flags here. Width and height
are detected separately and either one is enough.

| Detector | Toggle | Full when | Parameters |
| --- | --- | --- | --- |
| Width | `x`, `--by-width` | `terminal_cols - wanted_cols <= margin_cols` (a pane without a width cap wants the whole width) | `k`/`K`, `--margin-cols` (default 0) |
| Height | `y`, `--by-height` | `terminal_rows - wanted_rows <= margin_rows`, or `wanted_rows` is at least `min_percent` of the terminal | `m`/`M` `--margin` (default 1: the pane always leaves the prompt row), `p`/`P` `--percent` (default 0 = off) |

`c` / `--auto` switches all detection on or off. The promotion uses the
alternate screen when `l` / `--auto-alt` is on (default), otherwise full screen
on the primary screen, which overwrites the shell's screen content.

## Leak guard and defaults

Switching to the alternate screen is loom's normal answer to resize-error prone
dimensions, so auto full screen is on by default (`auto_fullscreen.default`);
a pane opts out with `Pane.InlineOnly`. Terminals restore the auto-wrap flag
along with the cursor when leaving the alternate screen, which can make a wide
row wrap and leak into the scrollback. `leak_guard` (`r`, `--leak-guard`) re-asserts
the wrap setting and clears the pane rows on the primary screen when switching.
Turn it off to compare.
