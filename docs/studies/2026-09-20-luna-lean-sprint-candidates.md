# Study: Roadmap Items Suited to `codex:luna:low` Lean Sprints

**Date**: 2026-09-20
**Status**: Proposal (first sprint, 097, running)
**Sources**: `docs/Roadmap.md`, `docs/studies/2026-09-20-harnez-agent-lean-sprint-orchestration-and-pitfalls.md`

## Selection Criteria

An item qualifies when all of these hold:

1. **No human input**: acceptance criteria are testable, with no design taste call, no
   external dependency choice, and no "decide whether to close/keep".
2. **Bounded**: one to three milestones, each a 60-150 s luna job.
3. **Visible**: each milestone can emit a headless `.ansi` frame (via `loom.RenderTo`
   or a golden-write test) that can be `cat`-ed.

## Evidence Convention

Each sprint commits frames to `docs/progress/<ticket>/M<N>-<what>.ansi`. View
everything the next morning with:

```text
for f in docs/progress/*/*.ansi; do echo "== $f"; cat "$f"; done
```

The convention is written into each ticket's milestone section, since the ticket is the
only host-to-developer channel.

## Candidates, in Proposed Run Order

| Order | Ticket | Why luna can do it alone | `.ansi` evidence |
|---|---|---|---|
| 0 | 097 ansiviewer | running now (M1-M3) | text, ansi, layout, record frames |
| 1 | 034 ParseANSI / WriteANSI | pure function, table tests, unblocks 065 | styled sample rows, round trip |
| 2 | 061 Themeable | interface plus propagation, test per composite | same widget tree in each theme |
| 3 | 062 NewWidget split/tabs | mechanical factory extraction, 057/058 already shipped | hosted vs standalone frame |
| 4 | 051 Frame fills height | layout arithmetic with tests | frames at 3 terminal heights |
| 5 | 037 DrawBorder/DrawBox | consolidation, golden tests | BoxStyle gallery |
| 6 | 096 textrender example | new example, no design choice beyond the ticket | CJK, umlaut, emoji, combining frames |
| 7 | 081 TreemapCell / ColorScale | pure geometry, property tests | treemap at 3 sizes |
| 8 | 092 scrollbar drag | pointer delta to offset, synthetic mouse events | before/mid/after drag |
| 9 | 093 filebrowser mouse hit-test | bug with reproduction test | hit-region overlay frame |
| 10 | 060 Ticker / Invalidate | concurrency; needs `-race` and a tight prompt | frames at tick 0/1/2 |
| 11 | 088 key capture coverage | decode tables, PTY-independent tests | rendered decode table |

### Second Wave, After Prerequisites Land

| Ticket | Gate | Note |
|---|---|---|
| 063 filebrowser as widget | 061, 062 | split into M1 conversion and M2 host test |
| 064 splash + monitor as widgets | 060, 063 | one milestone per example |
| 065 treemap widget | 034, 081 | remove `/dev/tty` loop in M1 |

## Not Suited (Needs a Human)

| Ticket | Reason |
|---|---|
| 095 human-observable PTY view | purpose is human feedback flow |
| 089 mouse proximity effects | visual design taste |
| 090 image backgrounds | external `cati` dependency choice |
| 091 double-click | timing threshold and UX decision (feasible later with a stated default) |
| 052 ChoiceStyle.Border | remove-or-implement is an API decision |
| 042 TuiInput.md | pure prose, no `.ansi` evidence |
| 048, 039, 016 | close/reclassify decisions belong to the maintainer |
| 014-019, 056 | parked |

## Overlap Warning

097 M1 needs ANSI rendering inside a bounded viewer, and 034 provides exactly that as a
library primitive. 097 was already dispatched, so M1 may carry a private parser. When
034 runs, its ticket must include the pre-work "replace the ansiviewer-local parser
with `Canvas.WriteANSI`". Writing that in the 034 ticket costs nothing and avoids a
second parser.

## Operating Rules (from the pitfalls study)

- One writer per workspace: run sprints **sequentially**, one ticket at a time.
- Escape prompts: use single quotes, no backticks or `$(`.
- Verify `git status` after every worker; commit on its behalf if `index.lock` blocked it.
- Tell every worker that root-package `/dev/tty` test failures are pre-existing.
- `harnez agent delete` at the end of each sprint (zero zombies).
