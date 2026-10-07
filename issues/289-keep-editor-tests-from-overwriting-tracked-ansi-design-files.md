# 289 — Keep editor tests from overwriting tracked ANSI design files

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: [PR #16](https://github.com/ubunatic/loom/pull/16), [282](282-add-f2-side-panel-with-file-browser-to-loom-edit.md), [283](283-add-f3-and-ctrl-f-search-panel-to-loom-edit.md)

---

## 1. Problem & Motivation
PR #16 TestGenerateAnsiDesignScreenshots writes all five tracked docs/design/loom-edit-*.ansi files during ordinary go test. Running make test-q1 passed but changed every design file, because screenshots embed random t.TempDir paths and live directory listings.

## 2. Technical Specification / Findings
Separate explicit screenshot generation from normal tests. Tests should render to temporary outputs or compare deterministic fixtures without modifying tracked assets. Use stable fixture paths and directory contents, and assert geometry instead of merely writing files. Retain approved mockups unless intentionally updating them.

Observed at PR head b8a59ca; this branch has not been merged locally. Before implementation, check live code and recent commits and reverify the finding.

## 3. Implementation & Verification Plan
/goal Ordinary editor tests leave tracked designs unchanged and screenshot generation produces reproducible review artifacts; verify the behavior, or stop and report when blocked on a user decision or denied permission.

Acceptance: A clean make test-q1 leaves git diff empty; deliberate generation is documented and deterministic; produced assets pass loom eval, loom measure and loom check-box. Run make install before reporting.
