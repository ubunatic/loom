# 426 — Jules batch: triage focus, overlay and RichTextEdit tickets 265 266 274 275 277

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Triage
**Related**: 265, 266, 274, 275, 277

---

/goal Get a checked Jules triage report for the listed tickets and apply it (close, merge, correct), or stop and
report when Jules fails twice or a decision needs the owner.

## Scope
Group: focus and popup/overlay routing, and RichTextEdit extras. Jules reads each ticket and the current code, and
reports per ticket: verdict (`already done`, `partly done`, `still valid`, `obsolete`, `duplicate of N`), evidence
with `path:line`, and which tickets in the group should be merged or closed. Report: `reports/jules/426.md`.
Any follow-up implementation here touches event routing and starts on `codex:sol:med` (CLAUDE.md).
The host checks the evidence (jules skill, Section 4), applies the verdicts and closes this ticket.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| triage | jules-triage-426 | | | pending |
