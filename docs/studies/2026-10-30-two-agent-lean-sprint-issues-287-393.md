# Two-Agent Lean Sprint Study: Issues 287 and 393

**Date**: 2026-10-30
**Issues**: [287](../../issues/287-fix-search-overlay-mouse-routing-and-selection-isolation-in-loom-edit.md) (open, partial acceptance), [393](../../issues/393-fix-hang-in-testserverlasttabexitcleanssocket-in-internal-session.md) (closed), [394](../../issues/394-cover-waitforattach-wiring-in-servesession-and-remove-sleep-in-control-status-test.md) (open, follow-up)
**Roles**:
- Host Orchestrator: Claude (this session), with no code written by the host
- Searcher (287): `claude:haiku:med`, read-only
- Developer (287): `claude:sonnet:high`, sessions `loom-287-dev`, plan-first then M1–M4
- Searcher (393): `claude:haiku:med`, read-only
- Developer (393): `claude:sonnet:high`, session `loom-393-dev`, plan-first then fix and test rewrite
- Reviewer (393): `agy:flash38:med` failed to start (authentication); no review ran

---

## 1. Constraints

- Codex Sol was not available to the host, and the user restricted the sprint to Claude models. The Luna second opinion was reserved and never run.
- `codex:sol:med` is the event-routing rule in AGENTS.md. It was not applied, and the conflict is recorded in [Editor](../Editor.md).

## 2. Issue 287: search overlay

- The Searcher verified the premise. The live editor is `internal/editor/edit.go`, and the ticket's `cmd/loom/edit.go` reference is stale. Search is routed first in `ConsumeMouse`.
- The developer planned first and found three extra defects. The host added them to the ticket as pre-work.
- Milestones, each host-reviewed before the next was granted:
  - M1: red tests, with the cursor tests held until M3.
  - M2: refresh matches on open and after every consumed key or mouse edit; reset search modes; set hit regions only inside the draw guard.
  - M3: display-only `Reveal` for highlights, and no cursor movement. The library change was approved.
  - M4: `DeferPopover`/`DrawPopover`, so the formatting popover paints above the search panel when search is closed. The probe was loosened for background cells; text and foreground are still compared.
- Host PTY run of the installed binary: query click, mode clicks, stale-match refresh, and the hint-bar `^F Search` control all PASS. Popover overlap, hover and scroll focus, and the other search-bar controls were not verified. The ticket stays open.

## 3. Issue 393: test hang

- The Searcher found that `LOOM_SESSION_WAIT_ATTACH=1` is inherited from a loom session shell, and that `serveSession` read it. The host confirmed the cause with `env -u`, and the test passed in 0.011s.
- The developer's plan chose an explicit option (`Options.WaitForAttach`) over reading the environment in the library. The daemon branch reads the variable once and clears it.
- Commits: `4a19821` (the fix), `67f922a` (the no-sleep test rewrite after host review). The rewrite was needed because the first version used `time.Sleep` polling and fixed budgets, against `docs/TestTiers.md`.
- Gate: `make test-q1` exited 0 with 43 packages `ok` and no failure markers. The ticket closed with that evidence.

## 4. Process observations

- **Plan-first worked twice.** Both developers corrected stale premises before writing code.
- **Host review caught a rule breach.** The test-tier sleep was caught in diff review, not by tests.
- **Environment before code.** A hang that looked like a code bug was caused by the shell environment.
- **Reviewer failure was reported, not substituted.** The lean rules require this, and the 393 close reason records the gap.
- **Quota reading was unclear.** The user's "2% left" and the harness's "98%" could not be reconciled before dispatch.

## 5. Open items

- 287: the remaining PTY checks listed above.
- 394: `WaitForAttach` wiring test and the sleep at `server_test.go:246`.
- Gemini review of `4a19821` and `67f922a`: not run, pending authentication.
- Local `main` is ahead of `origin/main`. Nothing has been pushed.
