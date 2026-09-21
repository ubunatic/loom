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

### M1-M3 Review (host)
Delivered: 440c478, ffad21c, ecde18e. Tests and vet green, but the example does not deliver the Resolved
Design; the evidence proves little:
- `Borders` prints a text list with widths; there are no boxes, frames or popups with samples as titles.
- `Buttons` is a plain Choice list, not buttons. `Clipping` prints the full text ('Clip4: "Hello"'), nothing is clipped.
- The `WantWidth` values look copied from the function under test: ZWJ family 6 and flag 4 are wrong for a
  terminal (both 2); the 'decomposed' umlaut case is indistinguishable from the precomposed one in source.
- `%q` leaks `‍` into labels. A stray `textrender` binary was left in the repo root (deleted by the host).
- Two of the three views are hard to review as `.ansi`. Quality is too low for this ticket, so the rework
  moves to a stronger developer (escalation ladder).

### M4 - Rework (replaces the weak views; keep case table, registry, PTY skeleton)
1. Cases in source use `\u` escapes for combining marks, ZWJ, flags and the decomposed umlaut
   (`u` + U+0308), so the difference to the precomposed case is real. `WantWidth` values are the width a
   modern terminal shows (write them by hand: ASCII 5, Café 4, decomposed Müller 6, CJK 4, e+ring 1,
   symbols 4, emoji 2, ZWJ family 2, flag 2, mixed 11). If `loom.StringWidth` disagrees for a case,
   do NOT change the expectation: move that case into a `knownDivergences` table, assert loom's
   current value there with a comment, and list each divergence in your report (smallest reproduction).
2. `Borders`: real bordered boxes (DrawBox/Frame/Popup, existing library only), each sample as title and
   as body text, at two box widths (fitting and too narrow so the title truncates). Labels on their own rows.
3. `Buttons`: use the library's real button widget if one exists; otherwise report that and use the closest
   activatable widget, honestly named. Sample as label, activation visibly changes a status line.
4. `Clipping`: render each sample into real width-limited areas (widths 8, 6, 4, 2) so clipping and
   cluster preservation are visible; the clipped text must differ from the full text where it does not fit.
5. Labels show the raw text plus its codepoints (`U+0065 U+030A`), never `%q`.
6. Regenerate `M1-borders*.ansi`, `M2-*.ansi`, `M3-pty-session.ansi` (only this ticket's evidence tests);
   view them with ANSI stripped and make sure each is readable before committing.
7. Keep the PTY test green. Commit '(issue 096 M4)', stage only your files, no stray binaries.

### M4 Review (host)
Committed by the host (index.lock) as the M4 commit. Accepted: codepoint labels, honest `WantWidth` with
`knownDivergences` (ZWJ family: loom 6, terminal 2; flag: loom 4, terminal 2), real clipping frame.
Defect: in `Borders` each box is only 2 rows high, so the body text is drawn on the bottom border
('└ASCII: Hello ...┘') and the box has no interior. The buttons view is still a plain list.

### M5 - Pre-Work
1. Borders: boxes get at least 3 rows (top border, one body row, bottom border) so the body sits inside;
   test asserts the body row starts and ends with the side border glyphs. Regenerate `M1-borders*.ansi`.
2. Buttons: if the library has a real button widget, use it; otherwise the tab title and the top status
   line must say 'Choice list (no Button widget yet)' and your report names that gap. Regenerate `M2-buttons.ansi`.
3. Commit '(issue 096 M5)' (if index.lock blocks you, say so, the host commits).
