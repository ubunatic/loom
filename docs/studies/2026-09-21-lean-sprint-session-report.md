# Lean Sprint Session Report: Roadmap Items with Cheap Developer Agents

**Date**: 2026-09-21
**Status**: In progress (updated when items finish)
**Goal**: finish 11 roadmap tickets in `/lean-sprint` runs with cheap developer agents and leave `.ansi` evidence in the repo.
**Sources**: [candidate study](2026-09-20-luna-lean-sprint-candidates.md), tickets under `issues/`

## View the Evidence

```text
for f in docs/progress/*/*.ansi; do echo "== $f"; cat "$f"; done
```

## Status

| # | Ticket | State | Developer | Evidence |
|---|---|---|---|---|
| 0 | 097 ansiviewer | closed | luna:low | `docs/progress/097/` |
| 1 | 034 ParseANSI / WriteANSI | closed (M1-M5) | haiku | `docs/progress/034/` |
| 2 | 061 Themeable | closed (M1-M3) | haiku | `docs/progress/061/` |
| 3 | 062 NewWidget split/tabs | closed (M1-M4) | haiku | `docs/progress/062/` |
| 4 | 051 Frame fills height | closed (M1-M3) | haiku | `docs/progress/051/` |
| 5 | 037 DrawBorder / DrawBox | closed (M1-M4) | haiku, luna:low fixed evidence | `docs/progress/037/` |
| 6 | 096 textrender | closed (M1-M5) | haiku M1-M3, luna:low M4-M5 | `docs/progress/096/` |
| 7 | 081 TreemapCell / ColorScale | closed (M1-M5) | luna:low M1-M4, sol:low M5 | `docs/progress/081/` |
| 8 | 092 scrollbar drag | closed (M1-M6) | luna:low M1-M4, sol:low M5, sonnet M6 | `docs/progress/092/` |
| 9 | 093 filebrowser mouse hit-test | closed (M1-M5) | luna:low | `docs/progress/093/` |
| 10 | 060 Ticker / Invalidate | running (plan step) | luna:low | `docs/progress/060/` |
| 11 | 088 key capture coverage | not started | | |

Done: 10 of 12 counting 097 (9 of the 11 goal tickets). Second wave 063, 064, 065 is not part of the goal.

## How the Sprints Run

- The host reviews only diffs, test output and evidence frames, and communicates through the ticket.
- Every developer sends a rough plan before coding. The host corrects the direction from the plan alone.
- Escalation ladder: haiku, `codex:luna:low`, `codex:sol:low`, native Sonnet, native Opus. After a stronger agent
  gets past a step, the next step goes back down. Since 096 the default developer is `luna:low`.
- Evidence rules: frames come from code, are gated on `LOOM_EVIDENCE=1`, are written to the repo-root
  `docs/progress/<ticket>/`, and only the ticket's own evidence test is run.

## Findings so Far

- Haiku delivered every ticket, each with one to four fix rounds written into the ticket as pre-work.
- Typical defects a review caught: vacuous tests (062 single-tab meta-test), a missing half of the design
  (051 stacked fill), evidence frames that show nothing (034 M3, 037 popup, 096), wrong evidence path (034 M5),
  evidence runs that dirtied other tickets' frames (037).
- Escalations: 037 popup evidence went to `luna:low` after two failed Haiku tries (fixed in one run);
  096 views went to `luna:low` after Haiku produced text lists instead of boxes, buttons and clipping.
- 096 rework by luna:low: the luna commit failed on index.lock both times, the host committed on its behalf.
- Open library question (from 096, confirmed as `knownDivergences` in the example): `loom.StringWidth` may report ZWJ sequences and flags wider than a terminal
  shows (family 6, flag 4). The 096 rework records each divergence in `knownDivergences`; the host decides
  about a separate ticket after the report.
- 097 known gaps: `--record` writes `ansiviewer.ansi` in the cwd by default; the ansiviewer keeps a local SGR
  parser because it is entangled with cursor replay (decision recorded in 034 M4).

- 081: luna:low built a squarified stub that fell back to slice-dice (aspect 2.893 vs 2.893) and left the real algorithm in a comment; sol:low, given the float-squarify plus cumulative-rounding architecture, delivered it (3.375 vs 1.892). Evidence frames s1 and s2 are identical for both layouts; only s3 differs.
- 092 used the whole ladder: luna:low built the mapping, drag state and evidence; sol:low found two library bugs (Split dropped drags leaving the child, View edge column) but not why the PTY test was red; native Sonnet found that the remaining causes were in the test itself (byte index instead of rune column, an assertion on a constant, a stray press) and that Split treated 0-based mouse coordinates as 1-based. It also moved the View scrollbar hit column to the drawn column and updated tests that had pinned the wrong column.
- Lessons now written into tickets: mouse events reaching widgets are 0-based, PTY SGR is 1-based, locate screen text by runes or display width.
- 093 by luna:low alone (no escalation), four rounds: the review caught `Choice.Draw` truncation changing for all users (now pinned by a test), duplicated row formatting (now shared), and an evidence frame that proved nothing (whitespace click on the already-selected row; now a click on another row).

## Open Items

- 081 leftover: `graph/treemap.go` still holds a dead commented block after the `return` in `layoutTreemapSquarified` (about 110 lines). The ticket asked to delete it; the next sprint touching `graph/` should.
- Two old `splash --watch` processes (PIDs 920646 and 920686) predate the sprints and were not touched.
- A peer session (lmcoder-41) asked for a review of the lmcoder roadmap; not acted on, it is outside this goal.
- `gofmt -l` still lists a few files that predate or were touched during the sprints (for example
  `cmd/loom-bench/main_test.go`).
