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
