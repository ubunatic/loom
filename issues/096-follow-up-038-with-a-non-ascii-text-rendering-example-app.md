# 096 — Follow up 038 with a non-ASCII text rendering example app

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Enhancement
**Related**: Issue 038

---

## Problem & Motivation

Follow up issue 038 with a durable visual test application for non-ASCII text.
The app should make display width, alignment, clipping, and composition issues
easy to inspect across common Loom widgets rather than relying on isolated unit
cases.

## Scope

Add `examples/textrender`, a runnable Loom example that demonstrates varied
non-ASCII text in panes, titles, buttons, and other representative widgets. Use
carefully selected samples covering accented Latin text, German umlauts,
wide/CJK characters, combining marks, symbols, and emoji where supported.

Show the same text in contexts that exercise padding, borders, alignment,
truncation, scrolling, selection, and button interaction. Label the cases so a
human can identify expected behavior and compare terminal rendering with Loom's
display-width calculations.

## Goal

/goal: Provide `examples/textrender` as a clear, runnable visual regression
surface for non-ASCII text across Loom panes, titles, buttons, and related
widgets, following up issue 038 with representative documented cases.

## Verification

Add focused checks for the example's text cases and display-width/layout
expectations. Extend the PTY tests to launch `examples/textrender`, exercise its
views and controls, and prove the non-ASCII cases render and interact correctly
through a real terminal session.

---

## Resolved Design (host decisions, binding for the sprint)

- Layout follows `examples/split` and `examples/tabs`: `examples/textrender/main.go` delegates to
  `textrender.Run(os.Args[1:])`; the library package `examples/textrender/textrender` exposes
  `NewWidget(args []string) (loom.Widget, error)` (see ticket 062) and a thin `Run`.
- A table of labeled cases: `{Label, Text, WantWidth}` covering ASCII baseline, accented Latin,
  German umlauts (precomposed and decomposed), CJK, combining marks, symbols, emoji (including a ZWJ
  sequence and a flag), and a mixed line. `WantWidth` values are literal numbers written in the test
  data and checked against `loom.StringWidth`; do not compute them with the function under test.
- One `Tabs` with four views: `Borders` (samples as box/popup titles and bodies, several widths),
  `Buttons` (samples as button labels, activatable), `Clipping` (samples in shrinking columns showing
  truncation), `Scroll` (a `Choice`/list of all samples with selection and scrolling).
- Register the example in `internal/examplesreg` with its `NewWidget`.
- No new library API; if a real width or clipping bug shows up, record it in your report with the
  smallest reproduction and keep going (the host decides about a separate ticket).

## Milestones (lean sprint, dev agent: haiku)

Host reviews only diffs and test output; this ticket is the only channel. Root-package tests
needing `/dev/tty` fail before this work; ignore them. Commit each milestone (message ends
'(issue 096 MX)'), stage only your files, never docs/README.md, keep the repo root free of stray
binaries. Evidence: frames produced by code, gated on env `LOOM_EVIDENCE=1`, written to repo-root
`docs/progress/096/` (find root by walking up to `go.mod`), run only the evidence test of this ticket
(`go test -run <Name> ./examples/textrender/...`), and check `git status` before committing. gofmt.

### M1 - Cases and Borders view
- Package skeleton, case table, width test against literal `WantWidth`, `Borders` view, standalone `Run`.
- Evidence: `M1-borders.ansi` (80x24) and `M1-borders-narrow.ansi` (40x12). Frames must not overlap
  or misalign any label; labels sit on their own rows.

### M2 - Buttons, Clipping, Scroll views and registry
- The three remaining views, key handling (tab switching, button activate, list scroll/select), and
  the `examplesreg` entry; registry test and `loom-bench` smoke still green.
- Evidence: `M2-buttons.ansi`, `M2-clipping.ansi`, `M2-scroll.ansi` (scrolled and selected state).

### M3 - PTY tests
- Extend the PTY tests (see `internal/ptytest` and existing example PTY tests) to launch
  `examples/textrender`, switch through all four views, activate a button, scroll the list, and quit;
  assert on real terminal output that non-ASCII cases appear.
- Evidence: `M3-pty-session.ansi`, a captured final screen from that real session.
