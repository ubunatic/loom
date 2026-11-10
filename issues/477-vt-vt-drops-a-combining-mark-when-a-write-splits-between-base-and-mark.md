# 477 — vt.VT drops a combining mark when a write splits between base and mark

**Status**: Closed — fixed by PR 91 (c81475bf); autoWrap-off gap noted
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: 458, 433, 382

---

## 1. Problem & Motivation
`vt.VT.Write` (`vt/vt.go:507`) buffers an incomplete UTF-8 sequence across writes, but it commits a
complete base character before the next write can bring its combining mark. If a chunk ends with `e`
and the next one starts with U+0301 (0xCC 0x81), the screen shows `Qe` instead of `Qé`. Real PTY reads
split anywhere, so this hits `loom session` tabs (`WriteFramed` calls the same `Write`) and every
`loomtest.NewVT` test (`loomtest.VT` is an alias of `vt.VT`, `loomtest/vt.go:8`).

458 batch 9 (`docs/studies/tuios/report-testing.md`, canary D, scratchpad `b9-canary/replay_test.go`
`TestB9Replay`): the `sgr-cup-wide` stream split at every byte gives 1 mismatch of 48 in loom (split 22)
and 0 of 48 in charmbracelet/x vt. Not run end to end through a session. Distinct from 433 (skin-tone
width in `measure`); closed 382 covered streaming but not this boundary.

## 2. Technical Specification / Findings
- A cell's grapheme must stay open until the next non-extending rune arrives: a combining mark, ZWJ
  sequence, variation selector or skin-tone modifier in the next write extends the previous cell.
- Check the cursor and wrap state when the extension arrives after an auto-wrap at the last column.
- Related split case from fuzzing (Jules #118 seed, checked by the host): `Write("0\xc2")` leaves
  the cursor at 0,0, but `Write("0")` then `Write("\xc2")` leaves it at 1,0. The pending UTF-8
  byte holds back the previous cluster; whole and split writes must agree here too.

## 3. Implementation & Verification Plan
/goal Every split of a mixed stream (SGR, CUP, wide chars, combining marks, ZWJ emoji, scroll) gives the
same screen text and cursor as the unsplit write, in a split-every-byte test in `vt`; stop and report
when blocked on a design decision or denied permission; on a loop or no progress, clear the goal and
write a failure report.

- Re-check live code first; port `TestB9Replay` into `vt` tests with the streams it uses plus ZWJ and
  variation-selector cases.
- `make test-q1`, then `make install`.

## Jules runs

| kind | session name | Jules ID | PR | state |
|---|---|---|---|---|
| implement | jules-implement-477 | 7791606068217650291 | GH #119 | late delivery: PR opened 2026-11-10 20:40 after the run was recorded as stalled; checked by the host and adopted (see below) |
| implement | owner restart in the Jules UI | unknown | none | not running on 2026-11-10 (owner), no PR |
| implement | jules-implement-477b | 3034362514808774344 | none | failed: FAILED "Jules was unable to complete the task" after about 4 min, no message, no PR |

The first run delivered late as GH #119 (`putCluster` appends a zero-width cluster to the previous
base cell, stepping back over continuation cells, with the pending-wrap case). Host check: 3 of its
4 tests fail on main (the VS16 one passes there); a split-every-byte test over 9 streams passes on
the PR except `❤` + U+FE0F split between the two (cursor 3 vs 2): the width change that VS16 brings
is not applied when VS16 arrives in a later write.

Lead from the failed run, checked by the host: a mark that starts a new write forms its own zero-width
cluster, and `putCluster` returns early for `w <= 0` (`vt/vt.go:600-603`), so the mark is dropped.
Fix direction: append an extending cluster to the cell before the cursor (or the last column while a
wrap is pending).

## Resolution (2026-11-10)
Done by Codeberg PR 91 (merge `c81475bf`): Jules commit from GH #119 plus host commit `88dc8563`
(`extendCluster` rewrites the cell when the joined cluster changes width, drops it at column 0) and
`b9c349b2` (byte-by-byte variant). `vt/split_every_byte_test.go` passes for all 9 streams (6 failed
on the old main); `make test-q1` passed. Known gap from the agy:flash38 review, not fixed: with
autoWrap off, a mark written after the last column attaches to the cell before it.
