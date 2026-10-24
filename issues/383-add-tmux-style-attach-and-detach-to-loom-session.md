# 383 — Add tmux-style attach and detach to loom session

**Status**: Closed — Tmux-style attach/detach with UNIX domain socket IPC, background daemon persistence, prefix d hotkey, and CLI commands verified with PTY tests
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: `internal/session`, `cmd/loom`, `specdefaults`, 371, 372, 376, `docs/Session.md`

---

/goal Implement daemonized session persistence with UNIX domain socket IPC, allowing sessions to detach (prefix d / hotkey) and reattach (loom session attach [name]) with full tab and terminal state restored, verified with PTY tests; or stop and report when blocked on a user decision or denied permission.

## 1. Problem & Motivation
Currently, `loom session` runs strictly in the foreground. If the user's terminal window or SSH connection closes, or if the user wants to detach and resume their work in another terminal, all tabs and running child processes are terminated.

Adding tmux-style attach and detach capabilities allows users to keep long-running builds, agents, debuggers, or monitoring TUIs running continuously in the background, detaching with a hotkey and reattaching seamlessly from any terminal.

## 2. Technical Architecture & Constraints

### 2.1 Process & Connection Model
- **Client / Server over UNIX Domain Socket**:
  - **Server Daemon**: Runs in the background (via `Setsid: true` or fork), owns all `sessionTab` instances, child PTY masters, and `vt.VT` terminal state machines. It continuously reads child output, updates the screen model, buffers scrollback, and tracks active modes.
  - **Client Process**: Runs in the user's foreground terminal. Sets raw mode via `terminal.Restorer`, catches `SIGWINCH` and `SIGHUP`, and relays raw stdin bytes and resize messages to the server over the UNIX domain socket.
- **Socket Path & Resolution**:
  - Path: `LOOM_SOCKET_DIR` → `$XDG_RUNTIME_DIR/loom/session-<name>.sock` → `~/.local/state/loom/session-<name>.sock`.
  - Stale Socket Detection: Probe connect on startup; if `ECONNREFUSED`, clean up the dead socket file automatically.
- **Session Naming & Collision Avoidance**:
  - Flag `-s, --session <name>`.
  - When `-s` is omitted and an explicit command is provided (`loom session -- <cmd>`), generate a distinct session name to prevent collisions across concurrent runs or tests.

### 2.2 Client Attach / Detach Mechanics
- **Detach**:
  - Hotkey: `prefix d` (configured in `spec/defaults.yaml` under `session.prefix_bindings.detach: ["d"]`).
  - Signal: Abrupt client disconnect (e.g. `SIGHUP` on SSH drop) automatically transitions the session into detached state without killing child processes.
  - CLI: `loom session detach [-s <name>]`.
  - On detach, client restores outer terminal cooked mode and exits 0. The server enters detached state (`out = io.Discard`) and keeps all child PTYs running.
- **Attach**:
  - CLI: `loom session attach [-s <name>]` (alias `a`).
  - Client connects to the socket, sends `msgAttach(cols, rows)`.
  - Server sets child PTY dimensions, sends `SIGWINCH` if size changed, and immediately pushes `t.vt.Repaint()` for the active tab to render the screen cleanly.

### 2.3 Spec Defaults & API Snapshot Discipline
- Spec updates in `spec/defaults.yaml` and `specdefaults/defaults.go` require updating `testdata/api/specdefaults.txt` and `testdata/api/loom.txt` via `go test ./internal/apitest -update`.

## 3. Milestones & Delivery

### M1 Delivered: Spec Update & UNIX Socket IPC Foundation
- **Status**: Delivered & verified in commit `cb2a914`.
- Added `detach: ["d"]` under `session.prefix_bindings` and `detach: "Detach"` under `session.bar_labels` in `spec/defaults.yaml`, updated schema and `specdefaults/defaults.go`.
- Refreshed API snapshots in `testdata/api/specdefaults.txt`.
- Implemented `internal/session/ipc.go` with length-prefixed framed message protocol, socket path resolution, stale socket cleanup, and private permissions.
- Verified with unit tests in `internal/session/ipc_test.go` and `go test ./internal/apitest`.

### M2 Delivered: Daemon Server & Client Attach/Detach Engine
- **Status**: Delivered & verified in commit `d544475`.
- Implemented background session server in `internal/session/server.go`: manages detached/attached states, runs PTY loops continuously, handles `prefix d` in key dispatch, and repaints active tab via `t.vt.Repaint()` on client connection.
- Implemented client connection handler in `internal/session/client.go`: sets raw terminal modes via `terminal.Restorer`, traps `SIGWINCH` (`msgResize`) and `SIGHUP` (clean detach), and restores outer screen on detach/exit.
- Preserved `Run` and `RunWithOptions` backward compatibility.
- Verified with lifecycle unit tests in `internal/session/server_test.go`.

### M3 Delivered: CLI Subcommands & Flag Routing
- **Status**: Delivered & verified in commit `47389a2`.
- Added `-s, --session <name>` flag and subcommands `attach` (alias `a`), `detach`, `list` (alias `ls`), and `kill` to `cmd/loom/session.go`.
- Added socket-backed control functions and auto-generated unique session names for explicit commands.
- Verified with unit tests in `cmd/loom/session_test.go` and `internal/session/server_test.go`.

### M4: Automated PTY Canary Tests & Documentation
- **Pre-Work / Required Refinements**:
  - Address startup tabs read/write race in `internal/session/session.go` observed during `-race` run so PTY tests and full repository tests run cleanly.
- **Scope**:
  - Add comprehensive PTY integration tests in `internal/session/session_pty_test.go`:
    1. Start named session, verify detach with `prefix d` exits 0 and restores cooked terminal mode.
    2. Verify `loom session list` reports detached status.
    3. Reattach via `loom session attach`, verify `t.vt.Repaint()` redraws screen, tabs and cursor are preserved, and child input works.
    4. Child output emitted while detached is displayed upon reattach.
    5. Client SIGHUP auto-detach handling.
    6. CLI detach and kill commands.
  - Update `docs/Session.md` and help strings.
  - Run `HTO=0 make test-q1`.
