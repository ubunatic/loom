# 096 — Follow up 038 with a non-ASCII text rendering example app

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Enhancement
**Related**: Issue 038

---

## Problem & Motivation

Follow up issue 038 with a durable visual test application for non-ASCII text.
The app should make display width, alignment, clipping, and composition issues
easy to inspect across common Loom widgets rather than relying on isolated unit
cases.

## Scope

Add `examples/textrender`, a runnable Loom example that demonstrates varied
non-ASCII text in panes, titles, buttons, and other representative widgets. Use
carefully selected samples covering accented Latin text, German umlauts,
wide/CJK characters, combining marks, symbols, and emoji where supported.

Show the same text in contexts that exercise padding, borders, alignment,
truncation, scrolling, selection, and button interaction. Label the cases so a
human can identify expected behavior and compare terminal rendering with Loom's
display-width calculations.

## Goal

/goal: Provide `examples/textrender` as a clear, runnable visual regression
surface for non-ASCII text across Loom panes, titles, buttons, and related
widgets, following up issue 038 with representative documented cases.

## Verification

Add focused checks for the example's text cases and display-width/layout
expectations. Extend the PTY tests to launch `examples/textrender`, exercise its
views and controls, and prove the non-ASCII cases render and interact correctly
through a real terminal session.
