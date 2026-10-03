# 251 — Nested Grid widgets do not show the NumberInput edit cursor

**Status**: Closed — implemented Focusable and cursor propagation for NumberInput, Table, SubCanvas/Blit, and fixed gallery TextArea focus
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**:

---

## 1. Problem & Motivation
In `loom widgets`' All tab, pressing Enter on the NumberInput grid cell starts inline editing, but the terminal cursor is not visible while typing. In the NumberInput-only tab, the same widget shows its cursor. This makes editing in a composite layout hard to follow.

**Goal**: Make the active editor cursor visible when NumberInput is hosted in Grid, or stop and report if the behavior depends on a user decision or denied permission. Done when a regression test covers cursor placement through the library layout and the fix passes the relevant tests.

## 2. Technical Specification / Findings
NumberInput draws its inline editor with `TextInput.Draw(..., focused=true)`. In the All tab, the widget is nested in Grid inside Tabs; the standalone tab draws NumberInput directly. The cursor is therefore lost or misplaced somewhere in composite drawing/layout. The precise library-level cause remains to be identified.

## 3. Implementation & Verification Plan
Trace cursor coordinates and visibility through the composite widget draw path. Fix the library-level propagation or coordinate handling, and add a regression test for NumberInput inside Grid. Verify that the standalone and All-tab cases both expose the active cursor.
