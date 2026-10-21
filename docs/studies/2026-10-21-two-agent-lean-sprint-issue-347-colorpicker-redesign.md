# Two-Agent Lean Sprint: Issue 347 (ColorPicker redesign)

Date: 2026-10-21

## Design round (opus:med, tuidev347)

- Round 1 (2m58s, 470k incl. cached): designs A (two rows), B (one row, right-click BG), C (two bar
  items), 94bd139. User: spacer cells between swatches look disturbing.
- Web search subagent (45s, 35k): thin results; suggested gapless 2-cell swatches, cursor inside or
  below the swatch.
- Round 2 (1m43s, 590k): gapless A and B, C dropped, f114015. User chose a one-row picker with FG/BG
  modes and mode-switch keys.

Searcher skipped: the code was the 346 picker from the previous sprint, and the ticket held the full
decision set.

## Developer (codex:sol:med, d347)

- Plan (56s, 58k new): approved unchanged.
- M2 one-row mode picker, 2eb0d14/dddfacb/9d21eb3 (6m42s, 266k new): q1 failed in 4 tests, fixed unrerun.
- M3 toggle keys, 7c8c027 (2m42s, 326k new): q1 failed on test setup + API snapshot, fixed unrerun.
- M4 marker fixes, 77c089c (3m12s, 379k new): q1 failed on API snapshot only, fixed unrerun.

## Verification

Host `make test-q1` after M2 and after M4: green. Live tmux checks of the installed `loom edit`:
FG/BG modes, Down/Up, Tab/`b` toggles, `▔▔` marker on popover bg in both modes. User checked
right-click with mousegrab on and off.

## Observations

Every developer q1 run failed on stale tests or the API snapshot and was fixed without rerun; the
host rerun caught nothing broken. Design work through `.ansi` rounds plus screenshots converged in
two rounds. Follow-ups filed: 348 (picker overlaps selection when bar is above), 349 (Down enters a
bar above), 350 (default/no-colour cell).
