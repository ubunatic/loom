# 400 — loom view and loom-probe keys need /dev/tty; add headless key input for agents

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: 129

---

/goal Headless agents can drive `loom view` and `loom-probe keys` without /dev/tty (e.g. `--keys-from` stdin/file), or the commands fail with a clear hint pointing at `view -p`; docs updated; or stop and report when blocked on a user decision.

Reported by the taia session while building the taia chat TUI on loom v0.5.6 (source: codeberg harnez/horg docs/studies/2026-11-01-agent-performance-log.md, "Tool Experience"). Both commands fail with "open /dev/tty: no such device" in agent shells; `loom view -p` works. At minimum document the limitation and the `-p` alternative.
