# loom-repaint-probe

`loom-repaint-probe` is a standalone raw-terminal benchmark for comparing repaint strategies. Run it from the repository root; the centered demo area uses up to 160 columns by 67 rows and shrinks to fit smaller terminals.

See [the design and measurement notes](../../docs/TerminalRepaintProbe.md) for frame behavior, incremental rendering, and HUD metric semantics.

```sh
go run ./cmd/loom-repaint-probe
```

The demo starts on the alternate screen with incremental row rendering enabled and per-frame clearing disabled. The centered dark-gray area is up to 160×67 on a dark-blue background and shrinks to fit smaller terminals. Resize the terminal to add or remove the surrounding area, or use the area controls to change the rectangle. `SIGWINCH` updates the layout and clears the screen once after each terminal resize; this happens even when per-frame clearing is disabled.

The HUD reports target and measured FPS, frame count, elapsed time, frame composition and output write timings, incremental rows updated, and bytes written. The main compose and write values and the incremental row percentage are moving averages over the last second; parenthesized averages and maximum remain cumulative since the last settings change or resize. Timing values use one decimal place in milliseconds with a 0.1 ms minimum; byte counts above 1 KiB are shown as whole KiB with a `k` suffix. These timings measure work performed by the process and the terminal write call; they do not measure when the terminal finishes displaying a frame.

Changing a render setting or resizing the terminal starts a fresh measurement window and resets the frame count, elapsed time, averages, maximum, and last-frame values. HUD text ends at the last character on each line; each character's background uses the animation color beneath it at 50% brightness.

Press `a` to cycle through the registered animations: none, Astra, and a moving rainbow across the whole rectangle. Astra uses deterministic Braille placement, a 150 ms phase cadence, the rectangle's dark-gray background, and a fade into that surface. Animation names, aliases, drawing, and HUD background colors are registered together so more animations can be added without changing option parsing. The HUD spans two key-hint rows and shades each HUD cell's background from the animation beneath it to 50% brightness.

With the background-plus-clear strategy selected, `c` disables the clear operation. The demo then paints only the centered area, leaving surrounding cells as they were; the HUD calls this out so partial repaint behavior is visible.

| Key | Action |
| --- | --- |
| `q` / F10 | Quit |
| Space | Pause or resume |
| `+` / `-` | Grow or shrink width and height |
| Page Up / Page Down | Raise or lower target FPS by 5 (range 1–120) |
| `w` / `W` | Grow or shrink width |
| `h` / `H` | Grow or shrink height |
| `p` | Cycle row, span, and per-cell cursor painting |
| `b` | Cycle full row fill and background-plus-clear strategies |
| `c` | Toggle clearing the screen before each frame |
| `i` | Toggle incremental row rendering |
| `a` | Cycle animations: off, Astra, rainbow |
| `t` | Toggle the alternate screen |

Run `go run ./cmd/loom-repaint-probe --help` to list all options.

## Command-line options

```sh
go run ./cmd/loom-repaint-probe --fps 60 --anim astra --paint spans --background fill --clear 1
go run ./cmd/loom-repaint-probe --tune 2s --fps 60 --altscreen 0
go run ./cmd/loom-repaint-probe --tune 2s --out tune-results.txt
go run ./cmd/loom-repaint-probe --incremental --clear 0
go run ./cmd/loom-repaint-probe --width 120 --height 40
```

`--tune` accepts an optional duration and defaults to `10s`. The duration applies to each candidate: tuning runs the selected baseline, then each alternative paint, background, clear, and incremental mode individually unless the corresponding option was supplied. Supplied `--paint`, `--background`, `--clear`, and `--incremental` options pin those modes; `--altscreen` pins alternate screen to `0` or `1`. Unpinned modes are varied one at a time. The animation stays fixed to the selected animation for every candidate. The report records terminal sizes, requested rectangle size, target FPS, duration, animation, initial modes, pinned options, and each candidate's full mode settings and measurements. It ranks candidates by the sum of their one-second moving compose and write averages; cumulative averages and maximum write time are included for reference. Press `q` or F10 to stop early and print results collected so far.

| Option | Values |
| --- | --- |
| `--tune [duration]` | Run and rank mode candidates; duration per candidate, default `10s` |
| `--out`, `-o` | Save the `--tune` report to a file or directory (also printed in the terminal); parent directories are created. A directory gets a generated `tune-terminal-<cols>x<rows>-rect-<width>x<height>-<fps>fps.txt` filename. Existing directories, paths ending in `/`, and extensionless paths are treated as directories. |
| `--incremental`, `-i` | Write only rows that changed since the previous frame (default on; use `--incremental=false` to disable) |
| `--fps` | Target frame rate, 1–120 (default `30`) |
| `--width`, `-W` | Starting rectangle width in columns (default `160`) |
| `--height`, `-H` | Starting rectangle height in rows (default `67`) |
| `--anim`, `-A` | `astra` (default), `rainbow`, or `none` |
| `--altscreen`, `-t` | Alternate screen `0` or `1` (default `1`); locks the `t` key and pins this mode during tuning when explicitly supplied |
| `--paint`, `-p` | `rows` (default), `spans`, or `cells` (also `0`–`2`) |
| `--background`, `-b` | `fill` or `clear` (also `0` or `1`) |
| `--clear`, `-c` | Clear before each frame: `0` or `1` (default `0`) |
