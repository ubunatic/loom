# 368 — loom view: Space and Ctrl+F do not page down

**Status**: Closed — space/ctrl-f page down via spec bindings (2ff965c); PTY test fails without fix; make test-q1 green
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

## Audit (Jules, 2026-11-01, checked against 5c3c603)

Still valid. `ansiView.ConsumeKey` in `cmd/loom/main.go:450-453` hardcodes `pgup` and
`pgdown`/`pgdn`/`pagedown` instead of reading `view.bindings.page_up`/`page_down`
(`spec/defaults.yaml:356-357`, which add `ctrl-f`, `" "`, `pageup`, `ctrl-b`, `b`). Step is
`Pane.ViewPanStep` (`main.go:430`), not `lastH`. Report: `reports/jules/368.md`.

## Jules runs

| kind | session | Jules ID | PR | state |
|---|---|---|---|---|
| audit | jules-audit-368 | 1133238704641422918 | github #40 | captured; owner prompted Jules to open the PR |
