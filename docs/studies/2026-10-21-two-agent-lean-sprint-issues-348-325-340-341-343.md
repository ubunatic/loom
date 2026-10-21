# Two-Agent Lean Sprint: Issues 348, 325, 340, 341, 343

Date: 2026-10-21

Searcher skipped for all five: each ticket named the files, and every developer started with a
read-only plan turn that rechecked live code.

## 348 Submenus clear of selection (codex:sol:med, d348)

- Plan (42s, 33k new): approved with one refinement (no silent overlap when neither side fits).
- M1 8df60ad (3m32s, 113k new): dev q1 failed on Test318 (Box submenu on the toolbar row), fixed
  unrerun. Host q1 green; live tmux: picker opens upward above a bar placed above the selection.
- Test gap: the regression test would also pass if the submenu stayed closed; the live check covered it.

## 325 Chart and pill colors in spec (codex:luna:high, d325)

- Plan (35s, 35k new), M1 514ad29 (3m50s, 100k new): theme roles in spec/themes.yaml; API diff
  additive only (six ThemeColors fields). Dev q1 green; host rerun was blocked by Quota-1 (no change
  since), which confirmed the dev run was final.

## 340, 341 Flaky PTY tests (codex:luna:high, d340)

- Plan (46s, 30k new), 916b3a6 and 5f1d8be (5m24s, 71k new): screen-state waits, no sleeps.
- `-race -count=20`, GOMAXPROCS=2, taskset 0,1: 340 15/20 -> 20/20, 341 19/20 -> 20/20; host rerun
  20/20 both. Pushed; GitHub #21 and #22 closed with the hashes.

## 343 Ctrl+G mousegrab toggle (codex:sol:med, d343)

- User chose Ctrl+G. Plan (1m10s, 63k new) confirmed grab-off disables mouse reporting, so the `m`
  indicator cannot turn it back on.
- M1 aaea9ee (5m33s, 262k new): dispatch before overlays, F1 help, droppable hint. Dev q1 green; live
  tmux: `mouse_any_flag` 1 -> 0 -> 1 with indicator M -> m -> M.

## Observations

Luna handled the narrow tickets (spec move, test sync) in one write turn each with green q1; sol
was used for the routing/layout tickets per AGENTS.md. Before/after repro counts made the flaky-test
fix reviewable without reading the diff in depth.
