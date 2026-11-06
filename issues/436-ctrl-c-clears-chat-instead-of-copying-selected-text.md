# 436 — Ctrl+C clears chat instead of copying selected text

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: issues/194, issues/417

---

## Problem

The owner reports that Ctrl+C in taia chat clears the chat even when text is selected, rather than copying the selection. The responsible layer and selection type need verification against current loom and taia behavior.

## Goal

/goal Reproduce and identify the responsible layer. When text is selected, Ctrl+C must copy the selected text rather than clear the chat; clearing the selection after copying is acceptable feedback. With no selection, preserve the existing clear-chat action. Fix any confirmed loom defect, add regression coverage for both cases and verify the real clipboard contract, and report any taia-side follow-up needed. Stop and report if blocked on a user decision or denied permission; on a loop, hard architectural or OS blocker, or no progress, clear the goal and record why in this ticket.
