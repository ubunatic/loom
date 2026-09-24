# 099 — Fix ansiviewer recording of ANSI output and terminal width

**Status**: Closed — already fixed in ce464f9/a0ba07d/17def88: repro at width 115 yields byte-identical copy of docs/data/harnez-usage.ansi; regression test testRecordFixture exists
**Priority**: P1
**Severity**: Major
**Category**: Bug
**Related**: `examples/ansiviewer`, `docs/data/harnez-usage.ansi`

---

## 1. Problem & Motivation

`ansiviewer --record` does not faithfully capture an ANSI file. Given a terminal
width of 115, running:

```text
go run ./examples/ansiviewer -o /tmp/hu.ansi --record 1s -- cat docs/data/harnez-usage.ansi
```

produces plain text without colors and wraps box lines incorrectly, including
broken `─┐` sequences on new lines. The recording path uses a fixed capture
geometry and renders through the viewer's plain snapshot path, so it does not
preserve the source terminal state.

## 2. Goal

Make ansiviewer recording preserve ANSI styling and use the hosting terminal's
actual geometry, so the reproduced output matches
`docs/data/harnez-usage.ansi` for the stated reproduction.

## 3. Implementation & Verification Plan

- Probe the recording path and terminal-size handling for command captures.
- Preserve ANSI styles in the recorded screen output.
- Derive capture width and height from the hosting terminal, with a safe
  headless fallback for tests.
- Add a regression fixture/test asserting colors, line boundaries, and geometry
  against `docs/data/harnez-usage.ansi`.
