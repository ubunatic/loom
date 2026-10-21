# Two-Agent Lean Sprint: Issues 349, 344, 345, 350

Date: 2026-10-21

Searcher skipped for all four: each ticket named files and line ranges, and every developer
(codex:sol:med, routing/key tickets per AGENTS.md) started with a read-only plan turn on live code.

## 349 Up/Down enter the bar (d349)

- Plan (1m29s, 49k new) named the one test whose expectation changes (Test317, by the user decision).
- c8d1d8d (5m13s, 176k new): specced `menu_entry`; dev q1 green. Live tmux: Down focuses a bar above.

## 344 Selection movements (d344)

- Plan (1m46s, 75k new): decoder already emitted every sequence; word/document/page selection and
  select all were missing.
- dcfe1e0 (5m, 193k new): dev q1 stopped at schema validation (no tests ran), fixed unrerun; host q1
  green. Live tmux: Ctrl+Shift+Right word steps, Alt+A (legacy and CSI-u) selects all.

## 345 Hamburger submenu (d345)

- Host decision: glyph `≡` (measures 1 cell). Plan (58s, 48k new) found the "last item is Box"
  assumption and byte-length visibility checks.
- df86d50 (8m41s, 210k new): dev q1 green. Live tmux: `≡` -> Select all selects the document.

## 350 Default-colour cell (tuidev350 design, d350)

- Design (claude:opus:med, 58s, 210k incl. cached): variants A/B, 080ecd5; user chose B.
- Plan (1m32s, 52k new). Host decisions: `∅` previews the result like other swatches; no `-1` in
  the public `Selected` field (callers may index Palette), additive `ResetSelected` instead.
- 534843e (7m22s, 194k new): dev q1 failed only in untouched Test302 (flaky: host 5/5, race 20/20,
  full q1 green) -> ticket 352. Live tmux: red pick, then `∅` restores default; saved file clean.

## Observations

The user asked the host to decide reversible choices itself and report them (harnez issue 840);
after that, 345 glyph and 350 API/preview decisions cost no round trips. Developer q1 runs were green
on first try in 2 of 4 tickets; the other two failed on a schema gap and an unrelated flake.
