# 288 — Compose loom edit panels with standard widgets and clipped responsive layout

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: [PR #16](https://github.com/ubunatic/loom/pull/16), [282](282-add-f2-side-panel-with-file-browser-to-loom-edit.md), [283](283-add-f3-and-ctrl-f-search-panel-to-loom-edit.md)

---

## 1. Problem & Motivation
PR #16 hand-draws frames, dividers and Normal/Regex controls across cmd/loom/edit.go:342–499 rather than composing standard Loom UI widgets as requested. Its search width is at least 24 even when the editor is narrower, producing a negative/outside panel origin; mode and navigation rows are not clipped to panel width.

## 2. Technical Specification / Findings
Use standard Loom frame/layout/control widgets with only modest tweaks. Keep search within available bounds at narrow widths and short heights. Preserve editor/FilePicker chrome rather than overwrite it with app-drawn status rows. Always file/link larger observed library gaps. Related existing library work: issues 274 and 276.

Observed at PR head b8a59ca; this branch has not been merged locally. Before implementation, check live code and recent commits and reverify the finding.

Updated review (2026-10-07, PR head a9f1a7d): the manual frame/control rendering and minimum 24-column search panel remain unchanged. All five saved 100×24 assets pass loom eval/measure/check-box, which does not verify narrow responsive layouts or standard-widget composition. This requirement remains unresolved.

## 3. Implementation & Verification Plan
/goal Editor chrome and search/browser panels use standard Loom widgets with usable, correctly clipped layouts across supported terminal sizes; verify the behavior, or stop and report when blocked on a user decision or denied permission.

Acceptance: Exercise wide/narrow and short terminals, including sidebar plus search, verify cell bounds and visible controls. Validate ANSI renders with loom eval/measure/check-box; verify make test-q1 and make install.
