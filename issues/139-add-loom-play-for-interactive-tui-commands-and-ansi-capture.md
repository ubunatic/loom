# 139 — Add `loom play` for interactive TUI commands and ANSI capture

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [097 — ANSI viewer recording](097-add-ansi-viewer-example-with-tui-recording.md), [099 — ANSI recording fidelity](099-fix-ansiviewer-recording-of-ansi-output-and-terminal-width.md), [129 — Loom CLI](129-build-loom-cli-tool-for-tui-asset-validation-measurement-and-interactive-viewing.md), `../wayreel/reelang`

---

## 1. Problem & Motivation

Add `loom play -- <cmd> <args>` to run an arbitrary TUI, let the user see and interact with it in their terminal, send basic scripted key sequences, and export `.ansi` captures of its output. This makes reproducible interactive TUI captures available without giving up live terminal use.

## 2. Technical Specification / Findings

Compare with existing adjacent tools before settling the interface or implementation: Loom's `ansiviewer --record` already launches and captures child TUIs; Reelang in `../wayreel` provides key sequences; tmux provides `send-keys`/`capture-pane`; asciinema records interactive terminal sessions as timed casts; VHS scripts keys and records terminal demos. They do not all provide this exact combination, so identify reusable pieces and remaining gaps. Start with a small native key-sequence mechanism if reuse would over-scope the first version; do not make Reelang adoption a prerequisite.

## 3. Implementation & Verification Plan

/goal Deliver `loom play -- <cmd> <args>` with live interactive terminal passthrough, basic key-sequence injection, and `.ansi` capture; verify a real TUI can be seen and operated while scripted keys are delivered and output is captured, and stop and report if blocked on a user decision or denied permission.
