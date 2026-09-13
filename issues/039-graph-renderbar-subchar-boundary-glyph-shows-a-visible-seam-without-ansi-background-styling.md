# 039 — `graph.RenderBar` `SubChar` boundary glyph shows a visible seam without ANSI background styling

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Cosmetic
**Category**: Bug/Documentation
**Related**: `graph/options.go` (`RenderBar`, `subCharacterFill`, `eighthBlockGlyphs`), `graph/bracketed.go` (`BrailleSubCharacterGlyphs`), `examples/monitor/main.go` (`applySnapshot`'s `graph.RenderBar(..., SubChar: true)` usage), voxi issue 110 (`examples/miclevel`, the report that surfaced this)

---

## 1. Problem & Motivation

Filed while building voxi's `examples/miclevel` (voxi issue 110), a real-mic
level meter rendered via a `loom` pane, using
`graph.RenderBar(value, graph.BarOptions{Width: 20, SubChar: true})` for a
plain monochrome bar (no `ANSI`/`BackgroundANSI`) — the same pattern
`examples/monitor`'s own `applySnapshot` already uses for its usage bars.

A user reviewing a live screenshot of the rendered bar
(`[████████░░████░░░░]`-shaped, ~47%) reported a visible gap/seam exactly at
the fill/empty boundary cell, distinct from both the solid fill and the
stippled empty region — see the attached description in voxi issue 110's
conversation (not reproduced here, but described as "a little gap between
the bar and the background of the bar" at "the middle symbol").

Root-caused (not a logic bug in `graph`'s glyph-selection math — verified
the boundary character selected, e.g. U+258D `▍` for a specific
percentage/width combination, is correct): most terminal emulators
(VTE-based ones, e.g. GNOME Terminal, included) special-case the *basic*
block/shade characters (`█ ▓ ▒ ░`, U+2588/2593/2592/2591) for pixel-perfect
procedural cell-filling that tiles seamlessly against its neighbors, but do
not extend that same procedural handling to the finer "eighth block"
fractional characters `▏▎▍▌▋▊▉` (U+258F down to U+2589) `subCharacterFill`
uses for the boundary cell. Those fall through to the actual font glyph
outline, which carries its own left/right bearing baked in — even when the
font *has* full coverage of that Unicode range (confirmed via `fontTools`
against Adwaita Mono, GNOME's own default terminal font: all seven
eighth-block code points are present in its `cmap`). The mismatch between a
terminal's procedurally-drawn solid cell and a font-glyph-rendered
fractional cell immediately next to it is the visible seam.

`options.go`'s own code already hints at this being a known shape: the
`opts.ANSI` branch in `RenderBar`'s `SubChar` handling deliberately forces
`emptyRune = ' '` "so a flat space... reads as pure panel-bg... instead of
'░''s own stipple pattern layering a third tone on top of it" — i.e., `ANSI`
mode's continuous `BackgroundANSI` wash is the intended way to absorb
exactly this kind of glyph-rendering inconsistency. But `SubChar` is
documented and offered as usable standalone (no mention that plain/
non-ANSI use has this terminal-dependent seam risk), and `examples/monitor`
itself demonstrates the standalone (non-ANSI) combination without comment.

## 2. Scope

### In scope
- Document (in `BarOptions.SubChar`'s doc comment, and/or a short note in
  `graph`'s package doc) that `SubChar` without `ANSI`+`BackgroundANSI` can
  show a visible seam at the boundary cell on terminals that procedurally
  render the basic block-element glyphs but not the eighth-block fractional
  ones — recommend pairing `SubChar` with `ANSI: true` and a
  `BackgroundANSI` wash, or omitting `SubChar` for guaranteed
  terminal-independent rendering.
- Decide whether `examples/monitor`'s own bars (`applySnapshot`, `SubChar:
  true`, no `ANSI`) should be updated to demonstrate the recommended
  ANSI-paired usage, left as-is with an added comment explaining the
  tradeoff, or switched to plain (non-`SubChar`) rendering — whichever this
  project's maintainer prefers as the example's own best-practice
  demonstration.

### Out of scope
- Any change to voxi or `examples/miclevel` — that consumer already worked
  around this by dropping `SubChar` (whole-cell snapping has no such
  terminal dependency).
- Auto-detecting terminal procedural-rendering support for eighth-block
  glyphs — no portable, general mechanism for this exists; this is a
  documentation/default-guidance issue, not something `graph` can reliably
  detect and route around at render time.

## 3. Acceptance Criteria

- [ ] `BarOptions.SubChar`'s doc comment (and/or `graph`'s package doc)
      explains the seam risk and the `ANSI`+`BackgroundANSI` mitigation.
- [ ] A decision is recorded (and applied, if changed) for
      `examples/monitor`'s own `SubChar` usage per §2.

## 4. Verification

Documentation-only unless `examples/monitor` changes: `go build`/`go
vet`/`go test` stay green either way. Visual confirmation of the seam (or
its absence with `ANSI`+`BackgroundANSI`) is manual, in a real terminal —
same category of check as voxi's own mic-level verification, not something
a headless test can assert.
