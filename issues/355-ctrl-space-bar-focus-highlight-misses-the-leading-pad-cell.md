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

## Decision (host)

An item's highlight spans the cells from just after the previous separator up to its own separator,
so ` S ` is highlighted. The first item includes the leading edge pad, the last item (`≡`) its
trailing pad. Separators stay unhighlighted. Mouse hit areas match the highlighted cells.
