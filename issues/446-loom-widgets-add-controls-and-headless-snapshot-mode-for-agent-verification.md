# 446 — loom widgets: add controls and headless snapshot mode for agent verification

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: issues/400, issues/139, issues/195, issues/270

---

/goal Add controls to `loom widgets` to inspect and drive widgets across various states in headless/agent environments (e.g. driving key events, capturing text/ANSI snapshots, and inspecting widget demo states), or stop and report when blocked on a user decision or denied permission.

## 1. Problem & Motivation
Currently, `loom widgets --show [Widget]` launches an interactive full-screen TUI loop attached to terminal I/O. In agent and headless workflows:
1. Interactive sessions require a fully attached TUI / PTY. When agents run without interactive TTY or in automated validation runs, interactive commands either block or require heavy harness scaffolding (`ptytest`, `tmux`).
2. Agents cannot easily drive a widget through different internal states (e.g. toggled, focused, typing a search query, expanded dropdowns) or take a clean snapshot of the rendered output.
3. Verification of visual appearance, widget states, and edge cases becomes difficult for coding agents without an explicit non-interactive inspection route.

## 2. Technical Specification / Findings
- **Headless Snapshot / Render Option**:
  - Add flags to `loom widgets` (such as `--snapshot`, `--headless`, or `--render`) to output the rendered widget directly to stdout (plain text or ANSI) at a specified geometry (`--width`, `--height`).
- **Key Injection / Scripting Controls**:
  - Allow passing an initial key sequence or script (e.g. `--keys "...,..."` or `--keys-from FILE/stdin`, related to issue 400) to transition the widget into a target state (such as activating search `/`, entering text, stepping values, opening popups) before rendering/snapshotting.
- **State Selection / Inspection**:
  - Allow demos to expose known state variants or accept demo flags/arguments (`-- <args>`) to initialize into specific states.

## 3. Implementation & Verification Plan
- Extend `cmd/loom/widgets.go` and gallery integration to support headless rendering and key injection.
- Add tests verifying headless capture and state driving for representative widgets (e.g. `Settings`, `Choice`, `TextEntry`).
- Verification:
  - Unit & CLI command tests pass.
  - `make test-q1` passes.
  - Test `loom widgets <Widget> --snapshot` from non-interactive subshells.
