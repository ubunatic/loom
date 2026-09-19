# Winch — Resize Diagnostics Application

`examples/winch` is an interactive diagnostic application that exposes Loom's terminal resize rendering modes as runtime switches.

## Purpose & Overview

Terminal resize handling involves several interacting mitigations:
- **Resize-event coalescing**: Draining intermediate `SIGWINCH` notifications so intermediate jitter is coalesced and only the latest terminal size is painted.
- **Atomic buffered frame flushes**: Rendering complete frames in memory before a single terminal write, preventing visual tearing and row-by-row flicker.
- **Per-row `CSI K` clearing**: Emitting erase-to-end-of-line (`\x1b[K`) after each rendered line to eliminate stale characters from wider previous frames.
- **Synchronized output mode (`CSI ?2026h` / `CSI ?2026l`)**: Wrapping frame updates in terminal synchronized output sequences where supported.
- **Terminal auto-wrap handling (`CSI ?7l` / `CSI ?7h`)**: Disabling terminal auto-wrap while the pane is active and restoring it on exit to prevent right-margin reflow corruption.
- **Out-of-band resize clearing**: An obsolete/diagnostic-only comparison mode that clears lines immediately inside the `SIGWINCH` handler.

## Keyboard Controls

| Key | Action |
| --- | --- |
| `1` | Toggle **Resize-event coalescing** |
| `2` | Toggle **Atomic buffered frame flushes** |
| `3` | Toggle **Per-row CSI K clearing** |
| `4` | Toggle **Synchronized output mode** |
| `5` | Toggle **Terminal auto-wrap handling** |
| `6` | Toggle **Out-of-band resize clearing** *(diagnostic only)* |
| `7` | Toggle **Full-screen buffer** *(diagnostic only)* |
| `8` | Toggle **Resize handling** |
| `9` | Toggle **Resize width guard** |
| `a` | Toggle **Use WINCH speed** (adaptive width guard) |
| `+` / `=` | Increase **Width guard columns (n)** |
| `-` / `_` | Decrease **Width guard columns (n)** |
| `r` / `R` | **Reset** all modes to spec defaults |
| `m` / `M` | Toggle **Reduce Motion** (Astra background animation) |
| `t` / `T` | **Cycle Color Theme** |
| `q` / `Esc` / `F10` | **Quit** |

## Adaptive width guard

With `a` on, the guard width comes from the measured `SIGWINCH` rate instead of
the manual `n`. All constants live in `spec/resize.yaml` (`adaptive_guard`):
the rate is the number of events in the last `window_ms` divided by the window
(the window is the smoothing), and `n = min_n + floor(rate / rate_step)`,
clamped to `max_n`. With fewer than two events in the window (startup, a single
resize) there is no measurable speed and the manual `n` applies. The width is
latched at each `SIGWINCH`, and the usual one-second settle restores full width.
The status line shows the effective `n`, whether it is manual or adaptive, the
manual `n`, and the measured `WINCH: <rate>/s`.

Known limitation: the kernel coalesces pending `SIGWINCH` signals, so a very
fast drag reports a lower event rate than the terminal's raw size changes.

## Running

```bash
# Run interactively
go run ./examples/winch

# Or select via loom-demo
loom-demo winch
```
