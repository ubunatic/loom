# Terminal Input

This page describes how Loom owns a terminal input stream and maps its bytes to
keyboard events. The behavior is implemented in `pane.go` and `event.go`.

## 1. Open the controlling terminal

`Pane.New` opens `/dev/tty` for reading and writing instead of using
`os.Stdin`. Standard input may be redirected or be a pipe, especially when a
pane is started inside command substitution such as `result=$(command)`. The
controlling terminal remains the user's interactive terminal in that case, so
Loom can read keys without consuming the command's input.

Opening the tty is also the point where Loom claims exclusive pane ownership,
queries the cursor position, saves the terminal state, and enters raw mode.
If `/dev/tty` is unavailable or terminal setup fails, `New` returns an error.
Call `Close` (usually with `defer`) to restore terminal state and release
ownership. `Close` is safe to call more than once.

## 2. Poll before reading; handle interruption and close

The input goroutine polls the tty descriptor with a short timeout before each
read. Calling `os.File.Fd` during setup detaches the descriptor from Go's
runtime poller and puts it in blocking mode. A bare blocking `Read` on that
descriptor may not be interrupted by closing the file, which could leave a
goroutine alive and competing with the shell for later input.

Polling lets the reader periodically check its completion channel. When
`Close` shuts down the run loop, the reader notices completion on the next
poll timeout and exits; the pane waits for that reader to finish before
returning terminal ownership. When `unix.Poll` returns `EINTR` (for example,
because `SIGWINCH` was delivered), the reader polls again. An interrupted poll
is not treated as an input event or a fatal read error.

The tty's Go `Read` is reached only after poll reports input ready. In contrast,
`queryCursor` performs a bounded synchronous poll/read while `New` waits for the
terminal's cursor-position response; its poll also retries `EINTR` and observes
a deadline. This startup query avoids a temporary reader goroutine that could
steal ordinary keys after timing out.

During input handling, when an incomplete byte sequence or lone `ESC` arrives at a
read boundary, `Pane` buffers the pending bytes and arms a timer governed by
`esc_key_timeout` from `spec/defaults.yaml` (default 50 ms). This ensures a
standalone `ESC` resolves promptly into an escape event without blocking on further
bytes. Similarly, width-guard intervals fall back to `guard_duration` from
`spec/defaults.yaml` (default 1 s) when unset on the pane.

## 3. Key decoding and shell compatibility

`DecodeKey` maps terminal byte sequences to Loom's `KeyEvent` names. Cursor
keys may arrive in normal CSI form (`ESC [ A`) or application cursor form
(`ESC O A`); both decode to `up` (and likewise for down, left, and right).
ZSH ZLE can leave DECCKM application cursor mode enabled after `zle -I`, so
supporting both forms keeps arrow keys working in that shell context. Tests
assert that both prefixes produce the same key names.

The decoder also recognizes control keys, printable UTF-8 text, common CSI
navigation and function-key sequences, and supported modifier sequences.
Unknown or invalid sequences produce an empty event rather than a guessed key.
The exact decoded names are part of the input contract used by widgets and
applications; see `KeyDefaults.md` for the library's default key actions.

### Modified navigation keys and CSI-u

- Home/End, Insert/Delete keep their xterm modifier parameter (`ESC[1;2H`, `ESC[1;5F`, `ESC[2;5~`, `ESC[3;2~`): they decode to `shift-home`, `ctrl-end`, `ctrl-insert`, `shift-delete` and so on. Plain forms (`ESC[H`, `ESC[1~`, `ESC[7~`, `ESC OH`, ...) stay `home`/`end`. Callers matching only the plain name no longer see modified presses (issues 257, 263; see `Upgrading.md`).
- `0x00` decodes to `ctrl-space`. Plain `0x09` is always `tab`; `ctrl-i` exists only in the kitty/CSI-u form (`ESC[105;5u`), which also yields `ctrl-shift-<letter>` keys. Legacy terminals send `ctrl-shift-b` as plain `0x02` (`ctrl-b`), so every CSI-u-only binding needs a function-key fallback (e.g. RichTextEdit box mode: F5).

### Terminal-reserved keys (VTE)

VTE terminals (Tilix, GNOME Terminal) keep Shift+Home/End and Shift+PgUp/PgDn for their own scrollback while the app is on the primary screen; the app never receives them. On the alternate screen they pass through (probed in Tilix 2026-10-04: `ESC[1;2H` arrives). VTE has no CSI-u. Editing widgets that need these keys must run on the alternate screen (issue 263).

After decoding, the pane handles F10 globally by default: it requests quit
before dispatching the event to any widget, including the help overlay.
Applications that need F10 within their widget tree can set
`DisableGlobalF10Quit` on the pane, or have the root widget return
`OwnsQuit: true` from its `PaneRequest`, which disables both the default quit
keys and the global F10 quit. This global behavior is separate from the
key decoder, which simply reports F10 as `f10`.
