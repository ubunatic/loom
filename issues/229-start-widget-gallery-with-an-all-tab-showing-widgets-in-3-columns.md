# 229 — Start widget gallery with an All tab showing widgets in 3 columns

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `gallery/gallery.go`, 195, 198, 200, 207

---

## 1. Problem & Motivation

Currently, the `loom widgets` gallery starts on an individual widget tab. Starting the gallery with an "All" overview tab that presents a collection of widgets arranged in 3 columns (small ones first) will demonstrate a rich, dense TUI on a single screen and allow checking that widgets render and interact well together.

## 2. Technical Specification / Findings

- Add an "All" tab in `gallery/gallery.go` set as the initial default active tab.
- Arrange smaller widgets in a 3-column layout (e.g. Button, Toggle, Checkbox, NumberInput, Badge, PillCluster, ProgressBar, Sparkline, etc.) using Loom's layout/grid capabilities.
- Ensure keyboard focus, tab cycling, and mouse event routing function seamlessly across the multi-widget composite layout.

## 3. Implementation & Verification Plan

- Construct the 3-column "All" composite demo in `gallery/gallery.go`.
- Make "All" the default tab on startup.
- Add test coverage in `gallery/gallery_test.go` and capture ANSI snapshots where applicable.
- Verify with `make test-q1`.

/goal Add the 3-column All overview tab as the default gallery starting tab and verify with tests, or stop and report when blocked on a user decision or denied permission.
