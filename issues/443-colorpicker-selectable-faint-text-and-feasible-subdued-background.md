# 443 — ColorPicker selectable faint text and feasible subdued background

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: issues/442, issues/346, issues/347

---

## Problem

After reviewing the comparison in issue 442, the owner wants faint text to be selectable in ColorPicker, and a subdued background if it can be supported without hacks. The existing comparison is otherwise approved. Faint is a text attribute; the theme-background sample is not currently a semantic dim-background color.

## Goal

/goal Recheck the live ColorPicker, style and persistence contracts; provide a selectable faint-text option that is applied and preserved correctly. Investigate whether a subdued background can be offered cleanly without hacks; implement it only if a sound representation and behavior are available, otherwise document the limitation and report it. Verify selection, rendering, persistence and existing color behavior with appropriate tests. Stop and report when blocked on a user decision or denied permission; on a loop, hard architectural or OS blocker, or no progress, clear the goal and record why in this ticket.
