# 428 — Jules batch: triage rendering and platform tickets 090 120 241 413 414

**Status**: Closed — triage applied: 120 241 413 414 still valid, 090 partly done (switch done, average fill left), no merges (Jules PR 104, 9/10 refs)
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Triage
**Related**: 090, 120, 241, 413, 414

---

/goal Get a checked Jules triage report for the listed tickets and apply it (close, merge, correct), or stop and
report when Jules fails twice or a decision needs the owner.

## Scope
Group: image backgrounds, Tree-Sitter engine, SDL3 grab package, ANSI parser gaps, grapheme scanner. Jules reads
each ticket and the current code, and reports per ticket: verdict (`already done`, `partly done`, `still valid`,
`obsolete`, `duplicate of N`), evidence with `path:line`, and which tickets in the group should be merged or
closed. Report: `reports/jules/428.md`. The host checks the evidence (jules skill, Section 4), applies the
verdicts and closes this ticket.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| triage | jules-triage-428 | 921243824981913591 | GH #104 | kept: 9/10 refs hold (examples/background/background/background.go:128 is `split.Ratio`; `bgs` is at :162). 120 241 413 414 still valid. 090 partly done: Jules says the theme switch is missing, but the host found it adopted (`bgs := []loom.Background{astra, cat}` at :162, PR 68); only the average-colour fill is left, as the ticket says. No merges, no ticket changes |
