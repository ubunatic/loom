# 2026-10-10 Lean Sprint Study: Configurable Popover Menu Semantics in loom edit (Issue 316)

## Summary
- **Ticket**: 316 (Configurable popup menu focus, navigation, and dismissal semantics in loom edit)
- **Role**: Host Orchestrator (Antigravity), Searcher: `codex:luna:low`, Developer: `codex:sol:low`
- **Status**: In Progress

## Searcher Audit Findings
- **Defaults**: `spec/defaults.yaml` (lines 96+), `spec/schemas/defaults.schema.json` (lines 77+), `defaults.go` (struct `RichTextEditDefaults` and validation).
- **Popover Focus & Key Dispatch**: `richtextedit.go` around lines 499–559 (`consumePopoverKey`, `popoverFocusSet`, `clearPopoverKeyboardState`).
- **Editor Forwarding**: `cmd/loom/edit.go` forwards modal popover keys at line 602 and normal editor keys at line 707, with outer Tab navigation at lines 630–649.

## Milestones
- **M1**: Spec and defaults schema definition in `spec/defaults.yaml`, `spec/schemas/defaults.schema.json`, and `defaults.go`.
- **M2**: Implementation of focus gating (`menu_focus_key`), unfocused arrow cancellation (`arrows_cancel_unfocused_menu`), reopening (`popup_open_key`), action closing and unselect flags in `richtextedit.go` and `cmd/loom/edit.go`.
- **M3**: Unit and PTY test verification, `make test-q1`, and ticket close.
