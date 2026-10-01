# 239 — ProgressBar fill uses the theme's selected colors

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug

---

/goal ProgressBar fill uses a theme color meant for progress/accent, not the selection colors, with
a test; or stop and report when blocked on a user decision (which theme color to use).

## 1. Problem & Motivation
In the gallery the progress bar looks selected: its fill uses the "selected" color. User: "this looks
like a bug". A bar next to a selected list item is indistinguishable from a selection.

## 2. Technical Specification / Findings
- `progressbar.go:290` (`ApplyTheme`): `StyleFill = {FG: theme.SelectedFG, BG: theme.SelectedBG, Bold: theme.SelectedBold}`.
- Use an existing accent/progress color from `ThemeColors` (or add one to `spec/themes.yaml` for all
  themes) instead of the selection pair. Check Spinner/Sparkline/PillCluster for the same pattern.

## 3. Implementation & Verification Plan
- Test that the themed fill style differs from the selection style in every theme.
- `make test-q1`, `make install`, PTY check of `loom widgets --show ProgressBar` in each theme (F9).
