# 406 — Flaky TestSessionDebugDetectsLeakRepairsViaSIGWINCHAndWritesDiagnostics under full test-q1 load

**Status**: Closed — single SIGWINCH repair and bundle write order in f304619; 20/20 under -race load
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**:
**Related**: 391 (session reply handling, merged 42039f8)

---

/goal Find why the test times out under full-suite load and make it deterministic without sleeps, or stop and report when blocked.

## 1. Problem
`make test-q1` on 2026-11-02 (log `.git/harnez/quota-1-logs/test-20261102T082048*.log`) failed `TestSessionDebugDetectsLeakRepairsViaSIGWINCHAndWritesDiagnostics` (`internal/session`, 5.2s timeout waiting for screen state; screen showed `REPAIRED_ON_WINCH:1`, `:2`, `LEAK_SENT`). Three isolated reruns passed in ~0.46s. First seen right after 391 changed session stdin/VT reply handling; not yet known whether 391 is involved.

## 2. Plan
1. Rerun under load (`go test -count=20 -race ./internal/session -run <name>` with `-p` parallel packages) to reproduce.
2. Check whether the debug detector's stall/repair timing (`internal/session/debug.go` constants) or the new `DrainReplies` write races the expected screen state.
3. Replace any timing dependence with an event the test can `WaitFor`.
