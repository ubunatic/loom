# 199 — Mouse input in the widget gallery; only Tab cycles tabs

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Feature
**Related**: 195 gallery

---

## 1. Problem & Motivation
In `loom widgets`, clicking a tab should select it, and clicks/wheel inside the demo should reach the widget (buttons, lists, media controls). Arrow keys currently cycle tabs, which steals them from the demo widgets; only Tab/Shift-Tab should cycle tabs.

## 2. Technical Specification / Findings
Mouse events must be translated to child-local 0-based coordinates and delivered through the dispatcher (`ConsumeMouse` returning `EventResult`, see docs/Widgets.md §8). Depends on the vertical tabs ticket for the tab hit areas.

## 3. Implementation & Verification Plan
PTY or dispatcher tests: click a tab selects it; click inside a demo reaches the widget; arrows go to the demo, Tab switches.

/goal Gallery tabs and demos respond to the mouse and arrows reach the demo widget; or stop and report when blocked on a user decision.
