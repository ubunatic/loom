# 280 — Add mouse capture flag to loom edit

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: [279](279-add-loom-edit-command-to-open-a-file-in-richtextedit.md), [editor design](../docs/design/loom-edit-01-editor.ansi)

---

## 1. Problem & Motivation
`loom edit` does not capture mouse input, and its CLI has no flag to enable capture. Users need mouse selection, clicks and scrolling in the editor.

## 2. Technical Specification / Findings
Add a boolean `--mousegrab` flag to enable mouse capture; allow explicit `--mousegrab=false`. Connect it through the library's supported capture lifecycle, restoring terminal mouse state on exit. Coordinate precedence with the editor settings ticket: explicit CLI values override config. Put defaults in the spec; fix library flaws in the library, without widget workarounds. Event-routing work starts on `codex:sol:med`.

## 3. Implementation & Verification Plan
/goal `loom edit --mousegrab <file>` captures mouse input and disabling capture restores normal terminal behavior; verify CLI parsing and a PTY interaction, or stop and report when blocked on a user decision or denied permission.

Before implementation, check live code and recent commits. Acceptance: flag documented in help; enabled clicks/selection/scroll reach the editor with correct child-local coordinates; disabled capture emits no enable sequence; exit restores terminal state. Run `make test-q1` and `make install` before closure.
