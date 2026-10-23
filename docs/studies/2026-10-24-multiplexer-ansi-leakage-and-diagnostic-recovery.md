# Multiplexer ANSI Leakage, VT State Machines, and Diagnostic Recovery

## 1. Header & Context

- **Date:** 2026-10-24
- **Scope:** Investigation into ANSI escape code handling, escape leakage failure modes, and recovery mechanisms across classic multiplexers (`tmux`, `screen`, `dtach`, `abduco`) and modern tools (`zellij`, `ghostty`, `herdr`, `wezterm`).
- **Trigger:** Observed escape sequence leakage and display corruption during interactive testing of `loom session` with `btop`, prompting the design of Issue #381 (session debug mode, rotating diagnostics, and self-healing repair).
- **Goal:** Synthesize authoritative design patterns and concrete findings from three search audits into an architectural reference for `loom session`.

## 2. Executive Summary

- **Root Cause of "Leakage" in Terminal Wrappers**:
  - Escape codes leak into the rendered view when a tool behaves as a naive byte relay or when its parser state does not survive across PTY `read()` boundaries.
  - If a buffer cut occurs immediately after `\x1b` or midway through a CSI/OSC parameter list (e.g. `\x1b[38;2;...`), treating incomplete sequences as plain text flushes raw ANSI parameters into the printable cell buffer.
  - Another major source is input/query race conditions: terminal device queries (DA `CSI c`, DSR `CSI 6n`) or SGR mouse sequences split across chunks where initial control bytes are dropped or misrouted to stdin.
- **How Mature Multiplexers Prevent Leakage**:
  - **Full State Machine Ownership**: `tmux`, `zellij`, and `wezterm` never relay raw PTY output bytes directly to the outer terminal screen. Every byte is fed into a persistent per-pane DEC-compatible VT parser state machine.
  - **State Preserved Across Reads**: The parser retains intermediate states (`csi_param`, `osc_string`, `dcs_passthrough`) across PTY read calls without time-based flushes.
  - **Safe Discard over Leaking**: Unrecognized or malformed sequences are parsed to their valid terminator (or dropped on cancellation/length limit) and discarded. They never fallback to printing raw control bytes into the terminal grid.
- **The Limits of `SIGWINCH` as "Auto-Repair"**:
  - `SIGWINCH` on Linux informs the child process group of a window geometry change (`TIOCSWINSZ`), causing curses/TUI apps (`btop`, `htop`, `neovim`) to re-query terminal dimensions and redraw.
  - However, `SIGWINCH` **does not reset corrupted terminal emulator state** (modes, cursor positions, scrolling regions). If the wrapper's own VT parser state is desynchronized, an application redraw will still be rendered on corrupted virtual grid state.
  - Therefore, true recovery requires a two-step remedy: first reset the wrapper's internal parser state to ground, then trigger a geometry toggle/resize event to force the child to repaint.

## 3. What Worked Well (Tools & Architectures Analyzed)

- **`tmux` (`input.c`, `resize.c`)**:
  - Uses an explicit byte-by-byte VT state machine with persistent per-pane context.
  - Unknown escape sequences dispatch to no-op handlers rather than being copied into visible text cells.
  - On resize, `tmux` recalculates geometry and explicitly redraws its own virtual screen buffer, redrawing from its clean internal cell model.
- **`Ghostty` / `Herdr`**:
  - `Herdr` vendors Ghostty's high-performance VT state machine rather than rolling an ad-hoc parser.
  - Implements a strict streaming parser based on the Paul Flo Williams DEC VT state machine (handling ground, escape, csi_entry, csi_param, osc_string, dcs_entry).
- **`WezTerm` (`mux`, `vte`)**:
  - Decouples PTY reads from rendering completely. Raw bytes accumulate into a parser queue, and screen rendering only reads from an immutable snapshot of the resulting terminal grid.

## 4. Honest Post-Mortem & Identified Pitfalls

1. **The Fallacy of Time-Based Escape Flushes**:
   - In terminal *input* handling, an ambiguous lone `\x1b` must be resolved with a short timeout to distinguish between the Escape key and an Alt key prefix (`ESC + char`).
   - *Pitfall*: Applying this timeout logic to the PTY *output* stream is a primary cause of escape leakage. PTY output chunks can be split arbitrarily by OS scheduling under high CPU load (e.g. `btop` updating charts). If an incomplete output sequence times out and flushes, the remaining parameters render as visible garbage on the next chunk.
   - *Rule*: Never apply standalone Escape key timeouts to child PTY output parsing.
2. **Same-Dimension `SIGWINCH` Dropping**:
   - Sending `SIGWINCH` without altering PTY window dimensions (`ws_row`, `ws_col`) is often ignored or coalesced by kernel TTY drivers and application signal handlers.
   - *Fix*: To force a repaint via resize, momentarily toggle dimensions (e.g. `rows - 1` then restore, or send an explicit terminal redraw request `\x1b[2J` if supported).
3. **Synchronized Output Mode (DEC 2026)**:
   - Modern TUI apps (including modern top/btop builds) use Mode 2026 (`CSI ? 2026 h` / `l`) for flicker-free atomic updates.
   - If a wrapper does not understand Mode 2026 or an app crashes inside a synchronized block, the wrapper can hang waiting for the end marker.

## 5. Architectural Recommendations for `Loom Session` (#381)

| Area | Recommended Design for `loom session` |
|---|---|
| **VT Parser Continuity** | Maintain a persistent VT state machine per session tab in `internal/session/` (using `vt.Terminal`). Output from PTY `read()` must feed the parser incrementally; never treat read buffers as standalone frames. |
| **Unknown Control Handling** | Discard unsupported or malformed CSI/OSC/DCS sequences upon reaching their terminator (`m`, `h`, `l`, ST `\x1b\\`, or BEL `\x07`) or on reaching a bounded parameter limit (e.g. 64 parameters or 1024 bytes). Never append unparsed control characters to printable text cells. |
| **Leakage Detection Heuristics** | Detect: (1) parser remaining in non-ground state with no incoming bytes for >1s, (2) printable cells containing literal `\x1b` or suspicious prefix fragments (e.g. `[38;2;`), or (3) control byte rate spikes during redraw. |
| **Two-Tier Self-Healing** | 1. Reset wrapper parser state machine to Ground.<br>2. Force a child redraw: nudge PTY geometry via `TIOCSWINSZ` (e.g. delta toggle) to deliver `SIGWINCH`. |
| **Diagnostic Ring-Buffer** | Maintain a fixed-size in-memory ring buffer (e.g. last 64 KB of raw stream chunks). When an anomaly triggers, flush the buffer, metadata, and before/after `.ansi` captures to a rotating log folder capped at 5 files per session. |
| **Circuit Breaker** | Minimum 10-second cooldown between auto-repair attempts; flag repeated triggers in the status bar (`[!] ERROR: ANSI LEAK`) without re-spawning disk captures. |

## 6. Key Sources & References

- [tmux `input.c` state machine](https://github.com/tmux/tmux/blob/master/input.c)
- [Ghostty terminal sequences & VT architecture](https://ghostty.org/docs/vt/concepts/sequences)
- [Paul Flo Williams DEC ANSI State Machine](https://vt100.net/emu/dec_ansi_parser)
- [Rust `vte` streaming parser crate](https://docs.rs/vte/latest/vte/)
- [Zellij terminal architecture & passthrough](https://github.com/zellij-org/zellij/blob/main/docs/ARCHITECTURE.md)
- [DEC 2026 Synchronized Output specification](https://vtdn.dev/docs/decset/mode2026-sync-output/)
- [Herdr issue #1202: SGR sequence split across reads](https://github.com/herdrdev/herdr/issues/1202)
