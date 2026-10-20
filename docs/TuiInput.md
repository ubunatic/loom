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

The input goroutine (`inputReader` in `pane_suspend.go`) polls before each
read. Calling `os.File.Fd` during setup detaches the descriptor from Go's
runtime poller and puts it in blocking mode. A bare blocking `Read` on that
descriptor may not be interrupted by closing the file, which could leave a
goroutine alive and competing with the shell for later input.

On Unix the reader blocks in `poll(2)` on the tty and a self-pipe with no
timeout (`waitInput` in `pane_wait_unix.go`), so an idle pane causes no
wakeups. Stopping it (`Close`, or `Suspend` before running a child) writes to
the pipe, which wakes it at once; bytes it read but could not deliver are kept
and replayed by the next reader. `Close` joins the reader (bounded wait)
before draining and closing the tty. Without a self-pipe (Windows, or if
`pipe(2)` fails) it falls back to a 50 ms poll timeout. When the poll returns
`EINTR` (for example, because `SIGWINCH` was delivered; see `isInterrupted`),
the reader polls again. An interrupted poll is not treated as an input event
or a fatal read error.

The tty's Go `Read` is reached only after poll reports input ready. In contrast,
`queryCursor` performs a bounded synchronous poll/read while `New` waits for the
terminal's cursor-position response; its poll also retries `EINTR` and observes
a deadline. This startup query avoids a temporary reader goroutine that could
steal ordinary keys after timing out.

On Windows (`GOOS=windows`), Unix `poll(2)` and `ioctl(TCFLSH)` are abstracted via
`waitTTY`, `drainTerminalInput`, and `notifyPaneSignals` in `pane_wait_windows.go`.
Because Windows console handles differ from Unix file descriptors, `waitTTY`
checks context/channel readiness or provides non-blocking polling, gracefully
falling back to safe defaults without leaking goroutines or panicking.

Pitfall: a reader goroutine must close a channel captured in a local before
`go func()`, never the struct field (`p.readerDone`). A later run replaces the
field, and the old goroutine then closes the new channel twice (panic
"close of closed channel", issue 242).

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
- `0x00` decodes to `ctrl-space`. In RichTextEdit, Ctrl+Space closes an open formatting popover or submenu, returning focus to the editor without changing the document or selection. When closed, it selects the smallest closed sharp or rounded box perimeter under the cursor, falls back to word selection, or opens a cursor-anchored popover on whitespace. `^D` (Ctrl+D) or F5 toggles box drawing mode; `Enter` or `Esc` exits box drawing mode without inserting text. Within the popover, "Draw" is nested under the "Box" submenu. Up/Down may place the caret past line ends or below the last line; typing or drawing there materializes padding, while navigation alone leaves the document unchanged. Plain `0x09` is always `tab`; `ctrl-i` exists only in the kitty/CSI-u form (`ESC[105;5u`), which also yields `ctrl-shift-<letter>` keys. Legacy terminals send `ctrl-shift-b` as plain `0x02` (`ctrl-b`); RichTextEdit does not bind Ctrl+Shift+B.

### Kitty keyboard protocol (opt-in)

Legacy terminals send Enter, Ctrl+J and Shift+Enter as the same byte, and
Ctrl+I as Tab. Set `Pane.KittyKeyboard` (or spec `pane.kitty_keyboard`) before
`Run` to push kitty progressive enhancement flag 1 (`ESC[>1u`) on start; Loom
pops it (`ESC[<u`) on close, on a panic exit, before a `Suspend` child runs
(pushing again on resume), and around alternate-screen switches, since each
screen keeps its own stack. Modified Enter, Esc, Backspace and Tab then decode
as `shift-enter`, `ctrl-enter`, `alt-enter`, `ctrl-esc`, `ctrl-backspace`,
`ctrl-tab` and so on, and Ctrl+I/Ctrl+J/Ctrl+M as `ctrl-i`/`ctrl-j`/`ctrl-m`.
Unmodified keys keep their legacy names (`enter`, `esc` from `ESC[27u`, `tab`,
`backspace`; letters stay text), so existing bindings work unchanged. Keypad
keys (codes 57399-57427) decode like their legacy counterparts (digits and
operators as text, `enter`, `up`, `pgdown`, ...), F13-F35 as `f13`..`f35`, and
other private-use codes as `key-<code>`. Super, Hyper and Meta add `super-`,
`hyper-` and `meta-` after `ctrl-shift-alt-`; Caps/Num Lock are ignored. Terminals
without the protocol ignore the push and behave as before (issue 321).

With a RichTextEdit selection, Tab and Shift+Tab cycle enabled popover actions;
Left/Right move along the bar, and Enter or Space activates the focused action.
Up/Down are consumed while the bar is open. Esc closes the popover. These keys
are consumed by the editor while the popover is available.

Space opens the focused foreground, background, or box-style submenu. Within a
submenu, Tab/Shift+Tab and the arrow keys wrap through its choices; Enter or
Space applies the focused choice. Esc closes the submenu first and returns focus
to the popover bar; another Esc closes the bar. Physical Space is decoded as
printable text and activates the menu just like Enter. These keys remain
consumed when the editor is nested in Frame, Split, or Tabs.

### Mouse Modes and Modal Capture

- **Mouse-off mode (`"m"`)**: Terminal mouse tracking sequences are fully disabled (`\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l`). Native terminal text selection and copying operate without app interference.
- **Temporary modal grab**: Opening a modal overlay (`Popup`, `Dialog`, context menu, F1 Help, `RichTextEdit` format popover) temporarily engages mouse tracking (`\x1b[?1000h\x1b[?1006h` / `1003h`). When the overlay closes or the backdrop is clicked, mouse-off mode is automatically restored.
- **Persistent mouse grab (`"M"`)**: Explicit global mouse grab maintains persistent mouse tracking across all screens and overlays.

RichTextEdit uses a bar terminal cursor for normal typing and a block cursor
while box drawing mode is active. Pane sends DECSCUSR only when the requested
shape changes and restores the terminal default on close. Ghost-cursor
navigation beyond existing text or rows is disabled by default; applications
can enable `RichTextEdit.GhostCursorEnabled` to keep a virtual position there
until typing or drawing creates the needed padding.

The `loom widgets --show` gallery leaves Esc to its widgets. Ctrl+Q and F10 quit
the gallery.

### Terminal-reserved keys (VTE)

VTE terminals (Tilix, GNOME Terminal) reserve S-Home, S-End, S-PgUp, and
S-PgDn for their own scrollback while the app is on the primary screen; the app
never receives them. On the alternate screen they pass through (probed in Tilix
2026-10-04: S-Home arrives as `ESC[1;2H`). VTE has no CSI-u, so CSI-u modified
bindings such as C-S-a/C-S-e cannot substitute. Editing widgets that need these
keys must run on the alternate screen (issue 263).

After decoding, the pane handles F10 globally by default: it requests quit
before dispatching the event to any widget, including the help overlay.
Applications that need F10 within their widget tree can set
`DisableGlobalF10Quit` on the pane, or have the root widget return
`OwnsQuit: true` from its `PaneRequest`, which disables both the default quit
keys and the global F10 quit. This global behavior is separate from the
key decoder, which simply reports F10 as `f10`.

Close defaults (spec `quit_keys`, `escape_quits`): an unconsumed `^Q` or `F10` closes
the pane, an unconsumed `Esc` never does unless `Pane.EscapeQuits` is set, and `^C`
or SIGINT/SIGTERM/SIGHUP count as an interrupt. Every one of these first calls
`Pane.OnCloseRequest(reason)`; returning `CloseVeto` keeps the pane running, and the
app calls `Pane.Quit()` later (for example from a Save/Discard/Cancel dialog).
