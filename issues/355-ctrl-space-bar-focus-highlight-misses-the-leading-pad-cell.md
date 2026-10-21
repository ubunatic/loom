# 355 — Ctrl+Space bar focus highlight misses the leading pad cell

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 353 (new bar layout, 3906319, bcc267a)

---

/goal Highlight the whole cell of the focused Ctrl+Space bar item, including the pad cell before
its label, with tests by display column; or stop and report when blocked on a user decision or
denied permission.

## Problem (user, 2026-10-21, screenshot)

Bar ` B ▏I ▏U ▏S ▏…`: with S focused, only `S ` is highlighted; the space before `S` is not.

## Decision (host, revised after M1 review)

The user's layout ` B ▏I ▏U ▏S ▏🎨 ▏Frame ▏Link ▏≡ ` stays exactly as designed (353); no extra
spaces. The "character before" an item is the `▏` separator cell (for B, the leading edge pad).
An item's highlight spans that preceding cell plus its label and trailing space, e.g. `▏S `; the
`▏` glyph keeps its FG on the focus BG, giving a thin left edge. Mouse hit areas match.

## M1 review (930f803, rejected)

M1 added a space before every label (` B ▏ I ▏ U ▏…`), which changes the user's layout and broke
Test303 (its click columns). Revert that geometry; implement the decision above.
