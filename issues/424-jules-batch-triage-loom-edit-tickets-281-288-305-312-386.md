# 424 — Jules batch: triage loom edit tickets 281 288 305 312 386

**Status**: Closed — triage applied: 281 386 closed, 288 corrected, 305 312 still valid (Jules PR 105, 5/6 refs)
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Triage
**Related**: 281, 288, 305, 312, 386

---

/goal Get a checked Jules triage report for the listed tickets and apply it (close, merge, correct), or stop and
report when Jules fails twice or a decision needs the owner.

## Scope
Group: `loom edit` features and the .tui/.ansi round trip. Jules reads each ticket and the current code, and
reports per ticket: verdict (`already done`, `partly done`, `still valid`, `obsolete`, `duplicate of N`), evidence
with `path:line`, and which tickets in the group should be merged or closed. Report: `reports/jules/424.md`.
The host checks the evidence (jules skill, Section 4), applies the verdicts and closes this ticket.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| triage | jules-triage-424 | 8266562899155890663 | GH #105 | kept: 5/6 refs hold (markup flags at cmd/loom/markup.go:65-66, not :50). 281 closed (LoadEditorConfig, spec/schemas/editor.schema.json, flag override), 386 closed (-o and --in-place, docs/TuiMarkup.md; optional .tui diff not done), 288 path corrected, 305 312 still valid; no merges |
