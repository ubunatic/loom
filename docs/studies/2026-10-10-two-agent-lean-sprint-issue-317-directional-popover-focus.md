# 2026-10-10 Lean Sprint Study: Directional Popover Focus and Tab/Ctrl+I Decoupling (Issue 317)

## Summary
- **Ticket**: 317 (Directional popover focus and decoupling Tab from popover focus)
- **Role**: Host Orchestrator (Antigravity), Searcher: `codex:luna:low`, Developer: `codex:sol:low`
- **Status**: In Progress

## Initial Findings
- In `richtextedit.go`, popover placement relative to selection determines whether the menu bar is above or below.
- Currently, `<tab>` was used to focus the unfocused popover.
- New design:
  - Arrow towards popover (Up if above, Down if below) focuses the menu.
  - Side/away arrows cancel the unfocused menu and move cursor.
  - `<tab>` / `^I` toggles italic when text is selected.
  - `<tab>` switches panels or handles indents when no text is selected.
  - `^<space>` reopens popover while text remains selected.
