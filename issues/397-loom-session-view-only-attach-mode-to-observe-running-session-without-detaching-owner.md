# 397 — loom session: view-only attach mode to observe running session without detaching owner

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [issues/383-add-tmux-style-attach-and-detach-to-loom-session.md](file:///home/alfiz/projects/loom/issues/383-add-tmux-style-attach-and-detach-to-loom-session.md), [internal/session/server.go](file:///home/alfiz/projects/loom/internal/session/server.go), [cmd/loom/session.go](file:///home/alfiz/projects/loom/cmd/loom/session.go)

---

## 1. Problem & Motivation

Currently, `loom session attach` assumes a single exclusive client per session. In `internal/session/server.go`:
```go
if s.conn != nil {
    return errors.New("session already has an attached client")
}
```
To inspect or monitor an existing session, an observer must detach the primary owner client (`loom session detach` or tmux-style detach) and take over the terminal session, interrupting the active user.

There is no way to connect concurrently in read-only / view-only mode to observe what the session is currently displaying without disrupting or detaching the owner.

## 2. Technical Specification

`/goal Implement view-only attach mode for loom session allowing concurrent read-only observers without detaching the active owner, and verify with tests, or stop and report when blocked on a user decision or denied permission.`

### Requirements
1. **CLI Interface**:
   - Support view-only mode on attach: e.g. `loom session attach --read-only` (or `-r` / `--view`), and/or `loom session view [name]`.
2. **Server Concurrent Viewers**:
   - `sessionServer` should support one interactive owner connection plus zero or more view-only / observer connections.
   - When attaching in view-only mode, the server does not reject the connection if an owner is already attached, nor does it detach the owner.
3. **Stream Handling**:
   - **Output**: Terminal output and screen state (or initial redraw on attach) are broadcast to the primary client and all attached viewers.
   - **Input**: User input from view-only clients is discarded / suppressed, except for local detach commands (e.g., prefix key detach or quit hotkey) so the viewer can disconnect without affecting the running session.
   - **Window Size**: Viewers do not force terminal resize events on the underlying PTY/session unless configured, or should adapt display / clip to the session's active dimensions.

## 3. Implementation & Verification Plan

1. **IPC & Protocol**:
   - Extend IPC attach message (`msgAttach`) with a `readOnly`/`viewOnly` flag in `internal/session/ipc.go`.
2. **Server Support**:
   - Update `sessionServer` in `internal/session/server.go` to maintain observer connections (`viewers []net.Conn`).
   - Fan out output writes across owner and viewers; remove viewer on disconnect without triggering session detach/hang.
3. **CLI**:
   - Add `--read-only` / `-r` flag to `loom session attach` in `cmd/loom/session.go`.
4. **Verification**:
   - Unit and PTY tests in `internal/session` asserting that:
     - An active owner session accepts a view-only attach without detaching the owner.
     - View-only client receives live output.
     - Keystrokes sent by view-only client do not reach the shell/session.
     - View-only client disconnects cleanly without affecting the owner.
   - Run `make test-q1`.
