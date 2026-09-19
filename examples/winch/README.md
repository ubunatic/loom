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
| `r` / `R` | **Reset** all modes to spec defaults |
| `m` / `M` | Toggle **Reduce Motion** (Astra background animation) |
| `t` / `T` | **Cycle Color Theme** |
| `q` / `Esc` / `F10` | **Quit** |

## Running

```bash
# Run interactively
go run ./examples/winch

# Or select via loom-demo
loom-demo winch
```
