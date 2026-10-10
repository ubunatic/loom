# 2026-10-10 Lean Sprint Study: Cancel Focused Popover on Arrow Away from Menu Bar (Issue 318)

## Summary
- **Ticket**: 318 (Cancel focused popover on arrow away from menu bar)
- **Role**: Host Orchestrator (Antigravity), Developer: `codex:sol:low`
- **Status**: In Progress

## Goals & Acceptance Criteria
- When popover is focused on main menu bar (no submenu open):
  - Menu below + Up -> cancel popover.
  - Menu above + Down -> cancel popover.
- Submenus (FG/BG palette, Box) continue to use arrows for choices.
- Verify with unit/PTY tests and `make test-q1`.
