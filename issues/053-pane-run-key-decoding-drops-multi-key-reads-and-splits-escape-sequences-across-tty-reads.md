# 053 — Pane.Run key decoding drops multi-key reads and splits escape sequences across tty reads

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: [pane.go](file:///home/uwe/projects/loom/pane.go), [event.go](file:///home/uwe/projects/loom/event.go), 034, 045, 046, 042

---

## 1. Problem & Motivation

In `Pane.run` (pane.go, the `case rr := <-reads:` branch), non-mouse input is
handled as:

```go
raw := rr.data
...
ke := DecodeKey(raw)
if root.HandleKey(ke) || p.handleKeyFallback(ke) { ... }
```

`DecodeKey` consumes the whole `raw` buffer as **one** key event and returns.
Two related transport-level bugs follow, both distinct from the SGR-parsing
concern in ticket 034 and the codec coverage in tickets 045/046 (those are
about correctly decoding a *complete* escape sequence, not about how bytes
arrive from the tty):

1. **Multiple sequences in one read are dropped.** The reader goroutine does
   a single `p.tty.Read(buf)` per wakeup (up to 64 bytes) and forwards
   whatever arrived as one `readResult`. If the OS coalesces several
   keystrokes into one read (fast typing, paste, or several nav-key repeats),
   only the first decoded key is dispatched; the rest of `raw` is silently
   discarded. The mouse branch immediately above already demonstrates the
   fix pattern — it loops `scanMouse` over `raw` and redraws after each
   event — but the key path does not drain analogously.

2. **A split escape sequence is decoded as garbage instead of buffered.**
   `os.File.Read` on a tty has no framing guarantee: a fast terminal can
   deliver `\x1b` in one read and the rest of an SGR mouse report or a
   function-key CSI sequence (`[15~`, etc.) in the next. The existing
   partial-mouse test acknowledges this can happen ("wait for the rest") but
   `Pane.run` has no carry-over buffer — each `readResult.data` is decoded
   independently, so a lone leading `\x1b` from the first read is decoded as
   an `esc` keypress (or similar wrong event) instead of being held and
   prefixed onto the next read.

Both are input-transport bugs, not decoding-table gaps, so they will not show
up in `DecodeKey`/`scanMouse` unit tests built on complete byte slices — they
only reproduce against a real or PTY-simulated tty under bursty input.

## 2. Proposed Solution

- Maintain a small pending-bytes buffer on `Pane` (or in the `run` closure)
  that prior loop iterations leave unconsumed.
- On each `reads` result, prepend any pending bytes to `rr.data`, then loop
  decoding key events (mirroring the mouse-branch drain pattern) until the
  buffer is empty or ends in a byte sequence that looks like the prefix of a
  valid escape sequence — in which case retain that tail as pending for the
  next read instead of dispatching it.
- Bound the pending buffer and/or add a short timeout so a stray lone ESC
  (distinct from an ESC-prefixed sequence) still resolves to a plain `esc`
  key promptly rather than hanging indefinitely for bytes that will never
  arrive.

## 3. Verification & Acceptance

- A PTY-based test that writes two complete key sequences in a single
  `Write` (so they are likely to land in one `Read`) and asserts both
  `HandleKey` calls occur, in order.
- A PTY-based test that writes an escape sequence split across two `Write`
  calls with a small delay between them and asserts it decodes as the single
  intended key event, not as a lone `esc` plus garbage.
- Existing `DecodeKey`/`scanMouse` unit tests continue to pass unchanged.
