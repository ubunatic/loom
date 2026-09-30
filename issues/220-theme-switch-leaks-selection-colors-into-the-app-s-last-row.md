# 220 — Theme switch leaks selection colors into the app's last row

**Status**: Closed — fixed with tests, suite green
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: 205

---

## 1. Problem & Motivation
User feedback: after "next theme" (bottom bar), the themed active selection color also appears in the app's last row. Seen in Choice, FilePicker (selected item color), Media (dark blue) and Table (selected). Tree does not leak, although it also has a selection.

## 2. Technical Specification / Findings
Some renderers leave the selection style active past their cell range (missing reset, or style carried into the next row). Compare Tree's rendering with the leaking widgets and fix the shared cause in the library.

## 3. Implementation & Verification Plan
PTY/render test per affected widget: after a theme switch the last row carries the base theme style only.

/goal No widget leaks its selection style after a theme switch, proven by tests; or stop and report when blocked on a user decision or denied permission.
