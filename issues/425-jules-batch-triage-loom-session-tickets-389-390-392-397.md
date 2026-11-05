# 425 — Jules batch: triage loom session tickets 389 390 392 397

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Triage
**Related**: 389, 390, 392, 397

---

/goal Get a checked Jules triage report for the listed tickets and apply it (close, merge, correct), or stop and
report when Jules fails twice or a decision needs the owner.

## Scope
Group: `loom session` (tab labels, mouse modes, view-only attach, modifier keys). Jules reads each ticket and the
current code, and reports per ticket: verdict (`already done`, `partly done`, `still valid`, `obsolete`,
`duplicate of N`), evidence with `path:line`, and which tickets in the group should be merged or closed.
Report: `reports/jules/425.md`. The host checks the evidence (jules skill, Section 4), applies the verdicts and
closes this ticket.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| triage | jules-triage-425 | | | pending |
