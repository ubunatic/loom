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
| implement | jules-implement-413 | 18068907042553961998 | - | running |
| implement | jules-implement-414 | 2652737723992373498 | - | running |
| implement | jules-implement-421 | 14083610897382528377 | - | running |
| audit | jules-audit-259 | 11292883786477104415 | - | running |
| batch audit | jules-batch-prio-a | 11899285583632198342 | - | running |
| batch audit | jules-batch-prio-b | 15741866014100721340 | - | running |
| ideas | jules-ideas-text-input | 14890562797936204173 | - | running |

## Notes for wave 2
- Owner tip (2026-11-06): open each prompt with a role line, "You are a <role> and use <approach> to run the <task>", e.g. "You are a Go TUI reviewer and use evidence-first code reading to audit these tickets". Compare its results with wave 1.
