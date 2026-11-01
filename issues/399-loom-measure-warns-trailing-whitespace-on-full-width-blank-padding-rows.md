# 399 — loom measure warns trailing whitespace on full-width blank padding rows

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Trivial
**Category**: Bug
**Related**: 129

---

/goal `loom measure` stops warning "trailing whitespace detected" for blank padding rows that span the declared sketch width, while still flagging real trailing whitespace, proven by a test; or stop and report when blocked on a user decision.

Reported by the taia session while building the taia chat TUI on loom v0.5.6 (source: codeberg harnez/horg docs/studies/2026-11-01-agent-performance-log.md, "Tool Experience"). Blank, full-width padding rows in .ansi sketches trigger the warning. It is cosmetic but trains agents to ignore it. Decide whether intentional padding (row width == sketch width) is exempt.
