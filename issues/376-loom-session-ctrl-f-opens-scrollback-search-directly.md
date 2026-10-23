# 376 — loom session: ctrl-f opens scrollback search directly

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 374 (scrollback search, prefix f), 373 (bar hints)

---

/goal `ctrl-f` in `loom session` opens scrollback search directly (swallowed like `ctrl-n`), the bar hint
shows `⌃F Search`, the binding lives in `session.bindings` in the spec, prefix `f` keeps working, and a PTY
test covers ctrl-f open/close; or stop and report when blocked on a user decision.

## Decision (user, 2026-10-23)

The user prefers `ctrl-f` over prefix `f`. The 373 note "not ctrl-f, which shells and less use" was a
host proposal; children lose ctrl-f (zsh/readline forward-char, less/vim page forward). Prefix `f` stays
as a second binding.
