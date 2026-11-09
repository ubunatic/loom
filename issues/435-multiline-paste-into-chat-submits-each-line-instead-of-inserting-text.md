# 435 — Multiline paste into chat submits each line instead of inserting text

**Status**: Closed — not loom: Tilix terminal settings caused per-line submission (owner)
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: issues/181, issues/417, issues/434

---

## Problem

The owner reports that pasting multiline text into taia chat triggers multiple submissions, one per line, instead of inserting the paste into the input. The responsible layer is unknown; bracketed paste was previously implemented in loom. This is distinct from the display-clearing report in 434.

## Goal

/goal Re-check current loom and taia behavior and recent history, reproduce the reported multiline paste, and identify the responsible layer. Fix any confirmed loom defect and add regression coverage proving multiline paste inserts as one operation without submitting any lines, including the actual pane/host event path. Verify and report any taia-side follow-up needed. Stop and report if blocked on a user decision or denied permission; on a loop, hard architectural or OS blocker, or no progress, clear the goal and record why in this ticket.
