# 438 — TextEntry Ctrl+U visual-line clearing and previous-line navigation

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: https://codeberg.org/harnez/taia/issues/151, https://codeberg.org/harnez/taia/issues/106

---

## 1. Problem & Motivation

Owner request: "Now see ^U ticket in ../taia and apply ^U to the TextEntry widget". Taia #151 requests clearing from cursor to current line beginning; at a line beginning Ctrl+U moves to the previous line end without deleting the newline, and at the first line beginning stays in place. Repeated presses alternate clearing and moving backward. Support visible soft-wrapped segments where feasible and define boundary behavior explicitly.

## 2. Technical Specification / Findings

Recheck live TextEntry behavior, wrapping/cursor mapping and recent history before implementation. Fix the library rather than adding gallery/app workarounds. Taia is read-only context; no external implementation authorized. Preserve selection, undo, paste, Unicode display-width and existing Ctrl+C contracts unless a documented intentional change is approved. Ctrl+L is not requested.

## 3. Implementation & Verification Plan

Use a bounded read-only Searcher, then a developer plan with exact files and visual-boundary semantics before approving writes. Add regression-first tests for line interior/start, first line, empty text, multiline, newline preservation, repeated presses, soft-wrap boundaries and display-width edge cases; validate the real `loom widgets --show TextEntry` path. Run quota verification and install before reporting. Record any caller interception that prevents the widget key from reaching TextEntry without editing that caller.

## Goal

/goal Deliver Ctrl+U clearing and previous-line navigation in loom.TextEntry matching Taia #151, including visual-line clearing where feasible, with explicit tested wrap-boundary behavior, gallery/live terminal evidence and passing repository verification; stop and report when blocked on a user decision or denied permission. On a loop, hard architectural or OS blocker, or no progress, clear the goal and record a failure report explaining why.
