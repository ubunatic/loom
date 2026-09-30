# 208 — Canvas paint widget (MVP: braille lines)

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 195, 197-202 gallery

---

## 1. Problem & Motivation
User request: loom needs a Canvas widget where the user can paint with the mouse; MVP draws braille lines.

## 2. Technical Specification / Findings
Braille gives 2x4 sub-cell dots per cell. Press starts a stroke, drag extends it with a line (Bresenham between successive points so fast drags do not leave gaps); a clear key resets. Reuse existing braille code if present (Chart uses braille lines). Add to catalog and gallery.

## 3. Implementation & Verification Plan
Proof: each item gets a regression test that fails before the fix, driven through the real `loom widgets --show <Name>` binary in a PTY (`gallery/gallery_test.go`, `internal/ptytest`), plus a `docs/progress/<Widget>.ansi` capture where the look changes. Drag in the PTY and assert braille glyphs along the path.

/goal A paintable Canvas widget draws braille lines via mouse drag, in catalog and gallery, proven by PTY test; or stop and report when blocked on a user decision.
