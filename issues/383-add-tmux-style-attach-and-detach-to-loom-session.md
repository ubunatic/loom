# 383 — Add tmux-style attach and detach to loom session

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: `internal/session`, `cmd/loom`, 371, 372, 376, `docs/Session.md`

---

/goal Implement daemonized session persistence with UNIX domain socket IPC, allowing sessions to detach (prefix d / hotkey) and reattach (loom session attach [name]) with full tab and terminal state restored, verified with PTY tests; or stop and report when blocked on a user decision or denied permission.

## 1. Problem & Motivation
Currently, `loom session` runs strictly in the foreground. If the user's terminal window or SSH connection closes, or if the user wants to detach and resume their work in another terminal, all tabs and running child processes are terminated.

Adding tmux-style attach and detach capabilities allows users to keep long-running builds, agents, debuggers, or monitoring TUIs running continuously in the background, detaching with a hotkey and reattaching seamlessly from any terminal.

## 2. Technical Specification & Architecture

### 2.1 Daemonized Session Server & UNIX Socket IPC
- **Socket Path**: Listen on a UNIX domain socket in `$XDG_RUNTIME_DIR/loom/session-<name>.sock` (fallback: `~/.local/state/loom/session-<name>.sock`).
- **Session Naming**: Support `-s, --session <name>` (defaulting to `default` or auto-numbered when omitted).
- **Background Persistence**:
  - When detached, the server process continues owning child PTY masters, updating each tab's `vt.VT` screen and scrollback buffer.
  - SIGHUP or abrupt client connection drop automatically transitions the session into detached state rather than killing child processes.

### 2.2 Client Attach / Detach Workflow
- **Detach**:
  - Hotkey: `prefix d` (matching tmux convention).
  - CLI: `loom session detach [-s <name>]`.
  - When detaching, the client restores outer terminal raw modes via `terminal.Restorer` and exits with code 0, leaving the server running.
- **Attach**:
  - CLI: `loom session attach [-s <name>]` (or `loom session a`).
  - Client connects to the server UNIX socket, transmits outer terminal dimensions (`cols x rows`), and synchronizes raw I/O.
  - Server sets child PTY dimensions, sends `SIGWINCH` if changed, and immediately pushes `t.vt.Repaint()` for the active tab to render the screen cleanly.

### 2.3 Session Management CLI
- `loom session list` (or `ls`): lists active sessions, active tabs, creation time, and attached status.
- `loom session kill [-s <name>]`: sends shutdown request to terminate the server daemon and children.

## 3. Implementation & Verification Plan
- **M1 (UNIX Socket IPC & Server Daemon)**: Implement session server process, socket protocol, and connection handshake.
- **M2 (Client Attach/Detach Loop & Hotkey)**: Implement `prefix d` detach, client connection handler, resize synchronization, and active tab repainting on reattach.
- **M3 (CLI Commands)**: Add `attach`, `detach`, `list`, and `kill` subcommands to `cmd/loom/session.go`.
- **M4 (PTY Integration Tests)**: Add automated PTY tests in `internal/session/session_pty_test.go` verifying background persistence while detached, clean reattach, and child process survival.
