# 446 — loom widgets: add controls and headless snapshot mode for agent verification

**Status**: Closed — M1 delivered --render mode, Settings demo, state inspection flags, and full shell completion (commit 16b1c6d)
**Closed-By**: session 25f2a8e0-914e-47e2-9d37-740f360d3ca2
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: issues/400, issues/139, issues/195, issues/270

---

/goal Add static rendering and inspection controls (`--render`) to `loom widgets` with full shell completion so that agents and users can inspect widgets in representative states without interactive TTY loops, or stop and report when blocked on a user decision or denied permission.

## 1. Problem & Motivation
Currently, `loom widgets --show [Widget]` launches an interactive full-screen TUI loop attached to terminal I/O.
1. Interactive sessions require an attached interactive terminal / PTY. In agent, CI, or headless workflows, interactive commands either block or require heavy harness scaffolding (`ptytest`, `tmux`).
2. Agents cannot easily take a deterministic, static snapshot of a widget to verify what it looks like in resting and active states.
3. Tab completion for widget commands and options should allow users to tap Tab and immediately see available widgets, states, and options.

## 2. Technical Specification / Findings
- **Static Render Mode (`--render`)**:
  - Add `--render` flag to `loom widgets [name]`.
  - When `--render` is set, instantiates the widget, draws it to a `Canvas` with the specified `--width` (or default) and `--height` (or `ContentHeight`), writes the rendered text/ANSI to stdout, and exits immediately with code 0.
- **Inherent Widget State Flags (First Trial: `Settings`)**:
  - Instead of synthetic state names, expose the widget's real properties:
    - `--searching`: activates the `searching` state (renders the `SearchBar`).
    - `--query <text>`: pre-populates `searchBar.Query` and filters items.
    - `--editing`: puts the selected item into inline editing mode (`beginEdit()` / `beginNumberEdit()`).
    - `--sel <index>`: sets the active selected item index.
- **Cobra Shell Autocompletion**:
  - Ensure `ValidArgsFunction` and `RegisterFlagCompletionFunc` provide seamless bash/zsh shell completion:
    - Widget names (`Settings`, `Choice`, etc.).
    - Theme names for `--theme`.
    - Flag completions for all relevant flags (`--render`, `--show`, `--searching`, `--editing`, etc.).

## 3. Implementation & Verification Plan
- **M1: Headless Static Rendering & Settings Demo States**:
  - Register `Settings` demo in `gallery/gallery.go`.
  - Add `--render` mode in `cmd/loom/widgets.go`.
  - Support `--searching`, `--query`, `--editing`, and `--sel` arguments for configuring widget state before render.
  - Implement shell completion functions for arguments and flags.
  - Add tests in `cmd/loom/main_test.go` and `gallery/gallery_test.go` asserting rendered output for default, searching, and editing states.
- Verification:
  - `make test-q1` passes.
  - `go test -race ./...` passes.
  - `make install`.
  - Verify `loom widgets Settings --render` directly in terminal.
