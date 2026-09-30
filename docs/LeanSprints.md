# Lean Sprints

Loom field notes for lean sprints with cheap developer agents. The generic
workflow lives in the harnez-managed [AgenticLoop](AgenticLoop.md); these notes
are loom-specific and survive `harnez init`.

## 2026-09 field notes
Source: [session report](studies/2026-09-21-lean-sprint-session-report.md). 12 tickets, developers Haiku, `codex:luna:low`, `codex:sol:low`, native Sonnet.
- **Plan first**: demand a rough plan from the developer before it codes. It lets the host fix direction without exploring code.
- **Escalation ladder**: Haiku, luna:low, sol:low, Sonnet, Opus. Very low quality or no plan counts as struggling. After a stronger agent clears a step, hand the next step back down.
- **Review the diff for fakes**: the defects that recurred were stubs (squarified fell back to slice-dice), vacuous tests, tests that pin the bug, and evidence frames that prove nothing. Read the assertions, and check that new tests fail without the fix.
- **Evidence convention**: frames come from code (`LOOM_EVIDENCE=1`), land in `docs/progress/<ticket>/`, and only the ticket's own evidence test is run (a package-wide run rewrote other tickets' frames).
- **Mechanics**: `harnez agent start|resume` is synchronous, so run it as a background task. Codex developers cannot commit when `.git/index.lock` is read-only; the host commits after review. Tell developers not to leave built binaries in the repo root.
- **Goal hooks**: a `/goal` text is judged literally by a Stop hook on every turn. If the plan changes (here: "haiku dev agents" after switching developers), update or clear the goal at once, and do not answer a hook that cannot pass. Ten near-identical replies burned tokens for nothing.
- **Close the loop**: file leftovers as tickets (101) and group human-only checks in one collection ticket (102) instead of leaving them in prose.

- **Model ladder (2026-09-24)**: luna, flash37, terra, opus. Escalate after one failed fix round; the host reruns the suite whenever code changed after the developer's single quota-1 run.
- **Rate before delete**: `harnez agent rate --name <s> <1-5> "<reason>"` only works on a live session. Rate every developer session at milestone review, then delete it (2026-09-24: six ratings lost to early deletes).
- **Where to start the ladder**: luna for hygiene, docs and single-file fixes; start at flash37 for refactors that move ownership between widgets (frame, pane, `Choice`). In the 105 sprint luna failed every such step ([report](studies/2026-09-24-roadmap-now-sprint-099-101-105.md)).

- **Model ladder (2026-09-29)**: `luna` handled both #166 (annotated output mode) and #171 (right-margin callout fix) cleanly in 1 milestone each. Token cost: ~1M/session (heavy cached). Suitable for bounded, single-file feature additions and UX fixes.

## 2026-09-29 field notes

Source: session on `loom eval -a` annotated output feature (issues #166, #171) and `docs/data/` asset investigation.

- **Plan-first is reliable for `luna`**: both sprints converged in 1 milestone after a read-only plan round. The plan quality was high enough that no pre-work or M2 was needed.
- **Right-margin callout bug was in M1 itself**: #166 delivered `<-- N` on its own prepended line. Host caught it on live `loom eval -a` output — not from the diff or tests alone. **Lesson**: always run the real command after `make install` as a plausibility check, not just `make test-q1`.
- **Agent lost context across server restarts**: the wait/resume cycle re-delivered the original planning message instead of the implementation result for #171, causing the host to see the plan twice. The status check (`harnez agent status`) confirmed completion; `git log` showed the actual commit. **Lesson**: after a wait-task, always verify via `git log -n 1 --stat` before inspecting agent messages, which may be stale replays.
- **VS16 in static assets is a silent portability trap**: `loom eval` correctly identified the line-width mismatch; the root cause was that `ℹ️` (with VS16) was used in an asset authored in tilix (1-cell advance) but foot advances 2 cells for it. `loom view --plain` stripped VS16 and hid the discrepancy. See `docs/EmojiWidth.md` §VS16 Portability for the full diagnosis and invariant.
- **No `harnez rate` calls**: tool feedback protocol was not followed. Rate developer sessions at review time, before `delete`.

## 2026-09-30 field notes

Source: roadmap 180 session, ~30 tickets with `codex:luna:med` developers, host `claude:opus`.
- **luna:med scaled well**: most widget tickets (Spinner, Tree, Form, DatePicker, Chart, …) landed in one round with a shared `devprompt.txt` (TDD, gallery demo + `.ansi`, one `make test-q1` at the end, commit trailers, `make install`).
- **Developers report unverified fixes**: several ended with "fixed after the suite run, not rerun". Treat that as untested; the host reruns `make test-q1` (harnez prints the files changed since the last run).
- **Repeated rules belong in the prompt, not in review**: catalog name order, `pgdown`/`pgup` aliases, 0-based child-local mouse. Each was found once by a failure, then added to `devprompt.txt` and never recurred.
- **Two failed rounds → host debugs**: on 112 the developer twice "fixed" the PTY test (byte offset, row guess) while the real cause was a mouse-handler signature the dispatcher never calls (see [Widgets](Widgets.md) §Event Handling). Reading the dispatcher took the host three tool calls.
- **Colour-dependent tests**: after `LOOMCOLOR` (182), pin `LOOMCOLOR=truecolor` in `TestMain` of packages that assert colours.
- **Waiting on agents**: a `pgrep` waiter matches itself; wait on the PID with `kill -0` or use the background-task notification.
- **Deliverables**: plans and roadmaps as Markdown; `.ansi` only for visible widget/example changes (`docs/progress/<Widget>.ansi`).

## Widgets feedback rounds (2026-09-30, tickets 203-223)

Source: two `loom widgets` feedback rounds, host `claude:opus`, developers `codex:luna:med` and `codex:sol:med`.
- **luna:med vs. event-model work**: luna committed red suites on cross-cutting event tickets (209, 204, 223: fixed the key path but not the mouse path). `sol:med` fixed each in one short round. Rule: when luna ends red or stalls about 15 min, hand the ticket to `sol:med`.
- **One sol:med developer for a batch**: 11 small tickets (212-222) in one run of about 47 minutes, one commit and one `make test-q1` per ticket, all green on the host rerun. Cheaper for the host than 11 dispatches.
- **Review the diff for caller-side fixes**: two batch fixes landed only in `cmd/loom/widgets.go`. One was legitimate (220, host drew over its footer), one masked a library contract flaw (212, filed and fixed as 223).
- **PTY tests found real library bugs**: coalesced key bytes, keys lost after a mouse report, and stale inline bounds on the alternate screen were all found by making flaky gallery tests deterministic instead of retrying them.
