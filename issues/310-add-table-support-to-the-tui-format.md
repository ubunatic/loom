# 310 — Add table support to the .tui format

**Status**: Closed — Delivered [table] block markup in .tui format with pipe rows, column alignments, junction glyphs in spec/box.yaml, and tests
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**:

---

## 1. Problem & Motivation
Add table support to `.tui` so tabular content can be authored in readable text and rendered as aligned terminal output.

## 2. Technical Specification / Findings
This is a future feature for the `.tui` format currently under development. Table syntax and alignment behavior should be designed to fit the format's established conventions.

## 3. Implementation & Verification Plan
Define and implement table markup, then verify parsing and rendered alignment across representative cell contents and terminal widths.

**/goal** Add and verify table support in `.tui`, or stop and report if blocked on a user decision or denied permission.
