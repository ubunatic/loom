# 484 — session: child ESC [ r resets the outer scroll region that protects the session bar

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 371, 391, 482

---

## 1. Problem & Motivation
Lead from Jules experiment GH #135 (session 5533486279476418706, not started by the 481 sweep),
not yet reproduced by the host. `loom session` writes the child's bytes to the outer terminal
as they come (`internal/session/session.go:783`, `out.Write(frame)` from `vt.WriteFramed`). The
bar row is kept out of the child's area by an outer scroll region (371). A child that sends a
bare DECSTBM reset (`ESC [ r` or `ESC [ ; r`) resets that region to the full screen. The bar is
redrawn afterwards (`drawSessionTabBar`, `:788`), but the region is not set again, so the next
scroll at the child's last row can scroll the bar up into the child area. Jules reported
duplicated bars in an integration test but dropped the test.

## 2. Technical Specification / Findings
- Jules's change (not adopted): `bytes.ReplaceAll` of `ESC [ r` and `ESC [ ; r` with
  `ESC [ 1;<childRows> r` before `out.Write(frame)`. It misses explicit full-height regions and
  the cursor home that DECSTBM implies.
- A child's own region `ESC [ t;b r` is already within child rows, since the child PTY has
  `childRows` rows. Also check what `ESC [ r` does when the bar is hidden (one tab).

## 3. Implementation & Verification Plan
/goal Reproduce with a PTY test (child prints `ESC [ r`, then enough lines to scroll, with two
tabs open), then keep the bar row fixed and out of the child's scroll area, with tests and no
sleeps; stop and report when blocked on a design decision or denied permission; on a loop or no
progress, clear the goal and write a failure report.

- Re-check live code first. Fix it where the session filters child output, not per app.
- `make test-q1`, `make install`; owner check: two tabs, run `printf '\e[r'; seq 100` in one.
