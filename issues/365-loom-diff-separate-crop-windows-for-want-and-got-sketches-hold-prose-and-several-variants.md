# 365 — loom diff: separate crop windows for want and got (sketches hold prose and several variants)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 362 (`loom diff`), 351 (live-check `diff` step)

---

/goal `loom diff` can compare a window of the sketch with a window of the capture at a different position
(e.g. `--want-region` and `--got-region`, same width/height), so a sketch with prose and several variants
(docs/data/colorpicker-ideas-001.ansi, variant b2) compares against a live tmux capture where the widget
sits elsewhere; tests and docs/CLI.md; or stop and report when blocked on a user decision.

## Findings (host, 2026-10-22, loom at db59497)

- Already robust: trailing spaces and empty rows compare equal; `--ignore-colour` is the plain-text pass,
  `--ignore-glyph` the colour-only pass.
- Gap: `--region` applies one window to both files. Real sketches contain prose lines and several
  variants, and the widget sits at another row/column in the capture than in the sketch.

## Decisions (host)

- No comment syntax or frontmatter in `.ansi` files: they must stay plain terminal output that
  `cat`, `loom eval` and other tools show as-is. Cropping selects the part to compare.
- Keep `--region` as shorthand for the same window in both.
- Optional, if cheap: a helper to find a crop, e.g. `--find <text>` that anchors the window at the first
  occurrence of a text in each file.
