# 197 — Complete widget names for loom widgets and --show

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 195 gallery

---

## 1. Problem & Motivation
Shell completion for `loom widgets <TAB>` and `loom widgets --show <TAB>` offers nothing useful. It should list the widget names that can actually be selected.

## 2. Technical Specification / Findings
Names come from the gallery registry (`gallery/`) and `spec/widgets.yaml`; complete only names the gallery can show. Use Cobra `ValidArgsFunction` and `RegisterFlagCompletionFunc` (the flag takes multiple names).

## 3. Implementation & Verification Plan
Test the completion functions directly and via `loom __complete widgets --show ''`.

/goal Tab completion lists the selectable widget names for both forms, with tests; or stop and report when blocked on a user decision.
