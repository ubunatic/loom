# Two-Agent Lean Sprint: Issues 355, 356, 358, 357

Date: 2026-10-21

Searcher skipped for all four: each ticket came from a live user observation, and developers
started with a read-only plan turn (355 excepted: a narrow draw fix dispatched directly).

## 355 Bar highlight covers the preceding separator (d355, codex:luna:high)

- 930f803 rejected: it added a space before every label, changing the user's 353 layout and breaking
  Test303 click columns. Host revised the decision: the highlight includes the `▏` before the label.
- fa2d988: dev q1 green. Live tmux: focus BG starts at the `▏` before S.

## 356 Plain Left/Right collapse the selection (d356, codex:sol:med)

- Plan (1m11s) found both causes: the unfocused bar swallowed the arrow, and `preserveSelection`
  restored the selection afterwards.
- e414ddd: dev q1 stopped at a compile error (no tests ran); host q1 red in one edited test; 691eabd
  fixed it. Live tmux: select "hel", Right, X gives "helXlo".

## 358 Every bar close clears the selection (d358, codex:sol:med)

- User decision during refinement: submenus open only with Enter/Space; arrows never open them.
- 3d75d0e (close paths, `unselect_on_cancel` default true), 0c249e4 (either vertical arrow closes).
  Host q1 red in 14 old tests that encoded "selection survives a close": the dev had read "don't
  fix untouched tests" too literally. 05c59f4 and edf4c53 aligned them; no code bug found.
- Live tmux, both bar positions: Down Down X and Up Up X insert X without replacing text.
- No Upgrading.md row for the semantics change; carried into 357 as pre-work.

## 357 ColorPicker Enter applies (d357, codex:sol:med)

- Plan (1m12s): Enter only closed the picker and left bar and selection active.
- 9858dbb, 0418828 (a test now focuses the bar before Left, since 356). Host q1 green. Live tmux:
  one Enter saves "hello" in palette colour 1, picker and bar closed.

## Observations

- Host q1 reruns caught a red suite after three of four "done" reports; developer q1 runs were
  either untested (compile stop) or followed by unverified edits.
- The "report, don't fix untouched failures" rule needs its scope in the prompt: it covers
  unrelated flakes, not tests whose expectation the ticket changes on purpose.
- Tickets that change shared semantics (358) should name the Upgrading.md row in the prompt.

## 359 ColorPicker two-row redesign b2 (d359, codex:sol:med, 2026-10-22)

- Searcher skipped: the ticket cites the sketch and the 347/350/357 picker. Plan turn 1m02s; host added
  that the BG marker must show right of the rug (tested at palette 15).
- M1 a09c9f5 + fe6b21a (11m): behaviour right (live: FG 1 on BG 2 saved), host q1 green, but the panel
  had no background and edges used the border FG. Reopened; M2 249c8d1 (4m) adds the palette-8 panel
  BG, panel-coloured edges and the sketch's row-3 layout. Host q1 green (first host run hit the
  1-minute harnez exec timeout; rerun with HTO=0). Live tmux confirms panel BG 8 and both markers.
- Observation: green tests and passing `loom check-box` did not catch a missing panel colour; only
  the live look against the sketch did.

## 360 ColorPicker look fixes (d360, codex:sol:med)

- Plan 1m01s: no shared selection colour existed; added theme role `selection: 24`, panel BG from the
  bar, inactive marker from `separator_fg`, trailing padding (36 columns).
- M1 0a60961 (8m): host q1 green, but live tmux showed edges as solid panel blocks (BG 239) and the
  selection colour leaking over the whole row-3 edge. M2 3adbbe6 (2m) fixed both with per-cell canvas
  tests. Host q1 green; live: edges fg 239 on editor BG, marker 24 only on the pick, inactive marker 8.
- Observation: for the second ticket in a row, only the host's per-cell ANSI capture found the look
  bugs; developer canvas tests asserted what was drawn, not what the sketch shows.

## 361 Picker rows share the cursor column (d361, codex:sol:med)

- Host decision: shared cursor, per-row applied value; Enter applies only the active row (359's
  "Enter applies both" would apply passed-over colours). Plan and write in one turn.
- 942aeae (column kept; q1 red in one reset-preview test, fixed but not rerun), 0ebabb1 (user addendum:
  dim markers for applied colours), aa67c35 (user correction: dim markers always visible, starting on
  the text's existing FG/BG or `∅`). Host q1 green; live: both dim markers on `∅` for plain text.
- `harnez agent send` to a running codex worker only queues for its next resume; mid-turn user
  additions reach it only via resume.
