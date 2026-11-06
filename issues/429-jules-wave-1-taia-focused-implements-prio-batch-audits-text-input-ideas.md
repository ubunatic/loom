# 429 — Jules wave 1: taia-focused implements, prio batch audits, text-input ideas

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Chore
**Related**: 413 414 421 259, 424-428 (earlier triage batches), harnez docs/studies/2026-11-06-jules-backlog-cleanup.md

---

/goal Clean the loom backlog with Jules before it stops accepting work (2026-11-11): run the waves, check each result's evidence, close/merge/correct tickets, keep our own taia priorities. Stop and report when blocked on a user decision; on a loop or no progress, write a failure report here.

## 1. Focus
taia (codeberg harnez/taia) uses loom's root package (TUI widgets, text input), `syntax` (shell mode) and `measure`. This wave favours TUI, text input and robustness.

## 2. Wave 1 (7 runs, cap 10)
- implement 413 (ansibuffer parser), 414 (measure cluster scanner), 421 (shell lexer keywords).
- audit 259 again (the last run ended in AWAITING_USER_FEEDBACK).
- batch prio audit A (widgets, focus, editor): 014 090 095 157 161 241 259 265 266 274 275 277 288 305 312.
- batch prio audit B (session, view, tooling, rendering): 102 120 137 139 313 314 389 390 392 396 397 400 401 408 413 414 421.
- ideas: robustness of the text input widgets (TextInput, TextArea, RichTextEdit).

Left out: 401 (deferred by the owner, port of 7.3k lines), 408 and 274/275/277 implements (event routing goes to codex per AGENTS.md).

## 3. Our taia priority (Jules does not know taia)
1. Text input robustness (ideas run, 265/277 focus and cursor invariants).
2. 421 shell lexer (taia shell mode `!`), 414 measure (display width).
3. 392 mouse modes/text selection, 401 transcript (when unblocked).
4. 413 ansibuffer (tooling robustness).

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| implement | jules-implement-413 | 18068907042553961998 | GH #112 | rejected: no tests, prescan claim false (ansibuffer.go:579 untouched), CSI B unbounded vs prescan, global `:`-to-`;` rewrite; kept as lead in 413 |
| implement | jules-implement-414 | 2652737723992373498 | GH #109 | merged as 8397a50 after host check (parity test fails without fix); lone VS16 now width 2 |
| implement | jules-implement-421 | 14083610897382528377 | GH #111 | merged as 95ed419 after host check (5 cases fail without fix) |
| audit | jules-audit-259 | 11292883786477104415 | GH #108 | kept: still valid, 4/4 refs hold, OSC 8 absence re-checked; gallery URL correction |
| batch audit | jules-batch-prio-a | 11899285583632198342 | GH #110 | discarded: ~15/21 refs hold (<80%); priority list kept as lead only (top: 275 274 095 157 312) |
| batch audit | jules-batch-prio-b | 15741866014100721340 | GH #107 | kept: 7/7 code refs hold, but most evidence cites tickets; 137 "done" corrected to partly done; no merges |
| ideas | jules-ideas-text-input | 14890562797936204173 | GH #113 | kept as leads: 3 confirmed by host probe (filed 432); 2 render-safe (control chars stripped at draw, value keeps them); 1 not a crash; 4 perf only; 5 not verified |

## Notes for wave 2
- Owner tip (2026-11-06): open each prompt with a role line, "You are a <role> and use <approach> to run the <task>", e.g. "You are a Go TUI reviewer and use evidence-first code reading to audit these tickets". Compare its results with wave 1.
- Lesson: 259 was already re-audited (259b, GH #74); wave 1 repeated it. Read the whole `## Jules runs` table before picking a ticket.

## Wave 2 (7 implement runs, role preprompt)
Role line: "You are a senior Go TUI library developer. You use test-first development: write table tests against the real behaviour, show them failing, then make the smallest change that passes them."
Picks: 432 (grapheme backspace, taia), 392 (mouse modes, Jules #1 and taia), 390 (tab label), 397 (view-only attach), 400 (headless keys), 157 (golden comparator rest), 413 (retry after rejected PR 112).

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| implement | jules-implement-432 | 8759762248526563848 | GH #114 | merged reworked: no tests, stray plan.md/run_tests.sh, exported 5-value scanner; host added tests and NextCluster/PrevCluster; skin tones to 433 |
| implement | jules-implement-392 | 2798506688453647188 | GH #116 | rejected (see 392): vacuous drag test, weakened existing test, breaking spec rename, stray scripts |
| implement | jules-implement-390 | - | - | pending |
| implement | jules-implement-397 | 7541561626893540044 | - | failed: Jules "unable to complete the task", no PR; not re-run |
| implement | jules-implement-400 | - | - | pending |
| implement | jules-implement-157 | 17239518533138541790 | GH #115 | merged: bold/dim cases fail without fix; continuation claim was already true |
| implement | jules-implement-413b | 10998755120330844800 | - | failed: AWAITING_USER_FEEDBACK, stalled after 2 replies, no PR; stopped |

## Wrap 2026-11-06 (owner timer)
- Wave 1 done: 421, 414 merged and closed; 413 rejected (PR 112); prio-b kept, prio-a discarded (<80% refs); ideas run gave 432; 259 audit was a duplicate.
- New: 430 (P1, session tab PTY tests time out on v0.10.0, q1 red), 431 (lone VS16 width), 432 (grapheme backspace).
- Wave 2 running: implements 432 392 390 397 400 157 413b with role preprompt. Check each PR (tests run by host, new tests fail without fix), merge by cherry-pick with a conventional message, close ticket.
- Hand-off verdict: Jules audits and ideas are useful leads; implements need a full host check (2 of 3 usable in wave 1).
- Resume: read this table, `harnez agent list`, `harnez agent resume --model jules --id <id> wait` per running row. Jules stops new work 2026-11-11.
