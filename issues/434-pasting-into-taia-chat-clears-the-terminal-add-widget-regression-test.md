# 434 — Pasting into taia chat clears the terminal: add widget regression test

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: issues/181, issues/417, issues/418

---

## Problem

The user reports that copying and pasting into the chat input in taia clears the terminal. The cause is unknown and may lie in loom; do not assume it does. Bracketed paste support was previously delivered in issue 181.

## Goal

/goal Re-verify the report against current taia and loom code and recent history; reproduce and identify the responsible layer. Add a loom chat-input widget regression test exercising paste and asserting the text is inserted without clearing the terminal or losing the surrounding display, and fix any confirmed loom defect. Verify the fix and report any taia-side follow-up needed. Stop and report if blocked by a user decision or denied permission; on a loop, hard architectural or OS blocker, or no progress, clear the goal and write a failure report explaining why.
