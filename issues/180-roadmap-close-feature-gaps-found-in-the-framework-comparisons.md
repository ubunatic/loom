# 180 — Roadmap: close feature gaps found in the framework comparisons

**Status**: Open
**Priority**: P1 (High)
**Severity**: Moderate
**Category**: Planning
**Related**: issues/177, issues/178, issues/179

---

## 1. Problem & Motivation
177-179 list features Bubble Tea, ncurses, tview/Ratatui/Textual have and loom lacks. We need one ordered plan to close the gaps that fit loom's design.

## 2. Scope & Rules
- Input: the Findings sections of 177-179. Merge duplicates; drop items marked "does not fit design".
- Group into phases by dependency (foundations like input/terminal capabilities before widgets that need them) and value.
- Every roadmap item maps to exactly one ticket: reuse an existing open ticket or file a new one (`harnez issues new`, lean, with a `/goal` and exit clause). Each ticket must be small enough for one lean sprint.
- Write the roadmap as `## Roadmap` in this ticket: phase, ticket number, title, size, depends-on.

## 3. Implementation & Verification Plan
/goal A committed roadmap in this ticket where every item points to a lean-sprint-sized ticket, in execution order. Stop and report when blocked on a user decision or denied permission.
