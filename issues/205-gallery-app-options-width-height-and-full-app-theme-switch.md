# 205 — Gallery app options: --width/--height and full-app theme switch

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 195, 197-202 gallery

---

## 1. Problem & Motivation
User feedback: `loom widgets` should accept `--width|-W` and `--height|-H` to test fixed app/panel sizes. F2 theme switching only recolours the tab bar, split bar and theme label; the demo panels keep their old colours.

## 2. Technical Specification / Findings
`themedGallery.applyTheme` themes only the root Tabs; it must reach every demo (Themeable children, and the panel background). Width/height: run the pane with fixed size (MaxCols/rows) instead of the terminal size.

## 3. Implementation & Verification Plan
Proof: each item gets a regression test that fails before the fix, driven through the real `loom widgets --show <Name>` binary in a PTY (`gallery/gallery_test.go`, `internal/ptytest`), plus a `docs/progress/<Widget>.ansi` capture where the look changes. Assert that a demo cell's colour changes after F2 and that the drawn area matches -W/-H.

/goal Fixed-size flags work and F2 recolours the whole app, proven by PTY tests; or stop and report when blocked on a user decision.
