# 401 — Scrolling word-wrapping transcript view with follow-end and scroll indicator

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: harnez/taia internal/tui

---

/goal Loom ships a transcript widget (word wrap by display width, scroll, follow-end that pauses when the user scrolls up, scroll indicator) adopted from taia's proven internal/tui implementation, with tests; stop and report if taia's version is not yet proven or blocked on a user decision.

Reported by the taia session while building the taia chat TUI on loom v0.5.6 (source: codeberg harnez/horg docs/studies/2026-11-01-agent-performance-log.md, "Tool Experience"). Request, not a bug. taia builds it first in its own internal/tui; port it once proven there. Draft until taia reports it stable.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| audit | jules-audit-401 | 7671810449244141841 | GH #82 | discarded: verdict unclear with no path:line refs; host confirms no transcript widget in loom (`git grep -il transcript -- *.go` empty); still waiting on taia |

## Source pointer (2026-11-03)

- Source to adopt: `~/projects/taia/internal/tui` at taia commit 45a6469. It is about 7.3k lines; taia calls it stable, but it is not a reusable widget yet.
- Deferred by the owner: do it later. The work happens in loom only; taia is read-only.
