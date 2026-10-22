# 368 — loom view: Space and Ctrl+F do not page down

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: 323 (View bindings from spec: `view.bindings.page_down` lists `ctrl-f` and `" "`)

---

/goal In `loom view`, every key in `view.bindings.page_down` and `page_up` pages the view, proven by a PTY
test; reproduction test first; or stop and report when blocked on a user decision.

## Reproduction (host, 2026-10-22, live check for 323)

`seq 1 200 > n.ansi`, `loom view n.ansi` in a 40x10 tmux pane: `j` scrolls, `q` quits, but Space and
Ctrl+F leave line 1 on top. The binary from before 323 (b7d2fb7) behaves the same, so 323 did not cause it.

## Notes

- Possibly the CLI view wraps `View` without passing these keys, or `lastH` is 0 on first page; check
  live code.
