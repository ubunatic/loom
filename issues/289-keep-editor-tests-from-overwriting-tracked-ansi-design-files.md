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

Updated review (2026-10-07, PR head a9f1a7d): normal screenshot tests now write to t.TempDir; after the full make test-q1 run, git diff in the review worktree is empty. Explicit UPDATE_GOLDEN/UPDATE_SNAPSHOTS/GENERATE_DESIGN_SCREENSHOTS modes still embed random paths/live directory contents and do not assert geometry. Keep open for deterministic explicit generation and documentation; ordinary tracked-file rewrites are fixed.

## 3. Implementation & Verification Plan
/goal Ordinary editor tests leave tracked designs unchanged and screenshot generation produces reproducible review artifacts; verify the behavior, or stop and report when blocked on a user decision or denied permission.

Acceptance: A clean make test-q1 leaves git diff empty; deliberate generation is documented and deterministic; produced assets pass loom eval, loom measure and loom check-box. Run make install before reporting.

## Integration verification (2026-10-07)

Host integration restores all approved designs and removes environment-triggered writes to tracked assets. Screenshot tests always use temporary directories; final suite leaves designs unchanged. Keep open for a documented deterministic standalone evidence generator, separate from approved mockups.
