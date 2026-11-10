# 433 — measure: skin-tone modifier splits from its emoji (👍🏽 = 2 clusters, width 4)

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 432 (cluster edits leave skin tones out), 414, 431, taia chat input

---

/goal The measure cluster scanner joins an emoji modifier (U+1F3FB..U+1F3FF) to its base, so `ClustersNew`, `StringWidth` and the text widgets treat `👍🏽` as one cluster, with the width checked in foot and kitty; stop and report when blocked on a user decision.

## 1. Problem
Host probe on f720cad: `StringWidthNew("👍🏽")` = 4, `ClustersNew` = `["👍" "🏽"]`; `"🤷🏽‍♂️"` = 4, `["🤷" "🏽‍♂️"]`. No code in `measure/` mentions the modifier range. Backspace in `TextInput` after 432 leaves `👍`.

## 2. Open question
458 batch 6 (2026-11-10, `docs/studies/tuios/report-widgets.md`) confirms the split at
`c96fbf66`: `measure.Clusters("a👍🏽b")` gives 4 units `a 👍 🏽 b`, and the TextInput caret
after two Left presses from the end sits at rune 2, not 1, so Backspace removes only the
modifier.

Some terminals draw a modifier sequence as two glyphs (width 4). Whether loom should say 2 or 4 (or follow the emoji spec mode) is not verified; probe foot and kitty first.

## 3. Verification
Add `👍🏽` and `🤷🏽‍♂️` to `measure/measure_parity_test.go` and to `clusterEditCases` in `text_cluster_edit_test.go`.
