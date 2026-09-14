---
title: Terminal Safety Hardening and Treemap Theme Iteration
---

# Terminal Safety Hardening and Treemap Theme Iteration

## Header & Context

Date: 2026-09-14 (continuation of the same-day session covered by
[`2026-09-14-x-term-coverage-gap-in-pane-termsize.md`](2026-09-14-x-term-coverage-gap-in-pane-termsize.md)).
Scope: harden `examples/treemap`/`graph.RenderTreemap` against terminal
auto-wrap corruption, extract that hardening into reusable `loom`
primitives, and iterate `graph.RenderTreemap`'s visual theming (issue
040) through three live-review rounds to a shipped design.

## Executive Summary

1. **Root-caused and fixed terminal auto-wrap corruption** in
   `examples/treemap --watch`: a too-wide row wraps onto the next
   physical line, and the program's own explicit newline then advances
   *again* — one bad row cascades into a scrambled whole screen. Fixed
   with two layers: software clipping to the live terminal width
   (`ClipRow`) and disabling the terminal's own auto-wrap (`DECAWM`) as a
   second line of defense.
2. **Extracted the fix into `loom.RawScreen`/`loom.WriteRows`/`loom.ClipRow`**
   (`rawscreen.go`) so every raw-ANSI consumer gets it for free, then
   caught and fixed a real regression in the extraction itself: applying
   `RawScreen`'s full-screen absolute positioning (correct for a
   persistent `--watch` redraw loop) to a one-shot "print and exit" path
   clobbered prior shell scrollback — caught only by PTY testing with
   pre-existing terminal content, not by unit tests. `loom.WriteRows` is
   the corrected one-shot primitive.
3. **Also fixed the `x/term` coverage gap** from the earlier study:
   `pane.go`'s `termSize` now uses `term.GetSize` instead of a hand-rolled
   `syscall.Syscall(SYS_IOCTL, ...)`.
4. **Iterated `graph.RenderTreemap`'s theming (issue 040) through three
   live-review rounds**, each of which passed its own tests and looked
   correct in a doc comment, but only broke down once rendered against
   real process-tree data: neighbor-blended half-block corners looked
   messy → reworked to a per-box-local rule; a filled-corner-number
   variant (theme 4) looked worse than the floating-corner-number one
   (theme 3) → removed entirely; a truncated-fragment regression slipped
   past every test written alongside the feature that introduced it →
   caught by looking at real output, not by the test suite. Full account
   in issue 040 §7.
5. **A concurrent-session file collision was observed live during this
   work** (not caused by it) — see "Process Finding" below.

## Key Learnings & Architecture Decisions

### Terminal auto-wrap corruption is a two-part failure, and needs a two-part fix

The failure mode isn't just "row too wide" — it's specifically the
*combination* of (a) a multi-row grid where each row's absolute screen
position is load-bearing, and (b) sequential-newline writing, where one
wrapped row silently shifts every subsequent row down by one line with no
error, no crash, just visually scrambled output. Neither half alone is
sufficient to fix:

- **Software clipping alone** (`ClipRow`) fixes the common case (the
  detected width was wrong) but does nothing if the clip itself has a
  bug, or if something writes past it.
- **Disabling auto-wrap alone** (`DECAWM` off) contains the *failure
  mode* (an over-wide row can no longer wrap and cascade) but doesn't
  stop a wrong-width row from reaching the terminal in the first place.
- **Absolute per-row positioning** (`\x1b[<row>;1H`) is the third,
  independent layer: even if a row is somehow still wrong, the *next*
  row's position doesn't depend on it, so corruption can't accumulate
  across a redraw loop.

All three ship together in `loom.RawScreen.Draw` (`rawscreen.go`).

### A "final render loop" abstraction needs at least two shapes, not one

`loom.RawScreen`'s absolute-positioning, full-screen-takeover style is
*only* correct for a persistent `--watch`-style redraw loop (the same
convention `watch(1)`/`htop` use) — it always draws relative to the
physical top-left of the terminal. Applying it to a one-shot "print and
exit" command is a real bug, not a stricter form of safety: it overwrites
whatever the shell already had on screen above it, since absolute
row-1 positioning has no idea anything was there. `loom.WriteRows` is the
one-shot counterpart: plain sequential output at the current cursor
position, clipped for width but never touching cursor state. Caught only
by a PTY test that populated prior scrollback before running the command
— a unit test with a blank terminal cannot see this class of bug.

Lesson for any future "safe terminal writer" work in this codebase:
**one-shot output and persistent-redraw output are not the same
capability with different config — they need genuinely different cursor
semantics**, and conflating them (even with good intentions, even after
passing tests) reintroduces exactly the class of bug the safety work was
meant to prevent.

### A design that passes its own tests can still look wrong

Issue 040's three-round iteration (full account in the issue) is the
concrete evidence: every version had passing unit tests and a coherent
doc comment before being revised or dropped after a live render against
real data. The recurring pattern: unit tests validate the *contract*
(dimensions, no double-rendering, no panics) but not the *gestalt* of
what many small, real-world-shaped boxes look like next to each other —
messy corners, ugly filled variants, and truncated fragments were all
things no assertion in the test suite was positioned to catch, because
each test looked at one box or one property in isolation.

**Process takeaway**: for terminal-visual work specifically, budget for
at least one live-render pass against real, messy data (not synthetic
2-3-segment fixtures) per meaningfully different visual change — not just
once before the first ship of a feature.

### Removing unwanted surface area beats leaving it "just in case"

When theme 4 (`TreemapThemeNumberedFilled`) turned out to look worse than
theme 3, the fix was deletion (`5fc60d3`), not deprecation or a
`Deprecated:` comment — matching the project's stated preference against
backwards-compatibility shims for code with no external consumers yet.
The same commit also deleted `truncateLabel`, found to be genuinely dead
code once traced through the pre-pass invariant that made it unreachable
under every theme, not just the one being fixed at the time.

## Process Finding: a concurrent session was editing this repo live

While drafting this document, system reminders repeatedly reported that
`graph/treemap.go`, `graph/treemap_test.go`, and `examples/treemap/*.go`
had "changed on disk" with content this session never wrote —
`TreemapLegendPosition`, `TreemapLegendRight`/`Bottom`, `LegendWidth`, a
`buildOptions` dark-text-on-light-background palette, and a
`TestTreemapPaletteUsesDarkTextOnLightBackgrounds` test all appeared
mid-session. This session also found `issues/041-v0-split-pane-focus...md`
already present, untracked, and unrelated to any file this session
touched (`frame.go`, `choice.go`, `view.go`).

Both are consistent with a second Claude Code session (or another agent)
working the same `loom` working tree concurrently, without a git worktree
or other isolation. This session responded by never staging files it
hadn't itself just edited (explicit `git add <file>...` per commit,
never `git add -A`/`-u`), so the two sessions' work stayed separated in
git history — but this was defensive discipline, not a guarantee: a
genuine two-writer collision on the *same* file/region at the same
moment is exactly [`docs/AgenticLoop.md`](../AgenticLoop.md) Invariant
1's "Parallel Read, Sequential Write" concern, observed in practice
rather than theory. See the session's closing recap for a concrete
suggestion (run concurrent sessions in separate `git worktree`s).

## Verification Performed

- `go build/vet/test ./...` clean after every commit in this session
  (`78ba5c5`, `95bcaf1`, `930558a`, `d8ef5f2`, `5fc60d3`).
- PTY-based repros (Python `pty.openpty()`) for: (a) a terminal that
  under-reports its own width mid-session, confirming `RawScreen`'s
  clip+autowrap-disable+absolute-positioning holds; (b) a one-shot
  `run()` invocation with 5 lines of pre-existing shell scrollback,
  confirming the `RawScreen`→`WriteRows` fix stopped it from being
  overwritten.
- Live renders (real `ps` snapshot via `examples/treemap`) at each theme
  iteration round, decoded via `sed 's/\x1b\[[0-9;]*m//g'` for
  glyph-shape verification and raw ANSI inspection for SGR-code
  correctness (foreground-only vs. filled corner numbers).
- `gofmt -l` clean on every touched file; pre-existing drift in untouched
  files (`layout/layout.go`, `splash.go`, etc.) explicitly left alone.

## Related

- [`2026-09-14-x-term-coverage-gap-in-pane-termsize.md`](2026-09-14-x-term-coverage-gap-in-pane-termsize.md) — the earlier `x/term` finding from the same day, fixed as part of this arc.
- Issue 040 — full design-iteration history for `graph.RenderTreemap` theming.
- [`../Graph.md`](../Graph.md) — evergreen architecture doc, updated alongside this study with a Treemap section.
