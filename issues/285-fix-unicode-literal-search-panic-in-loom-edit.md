# 285 — Fix Unicode literal search panic in loom edit

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [PR #16](https://github.com/ubunatic/loom/pull/16), [282](282-add-f2-side-panel-with-file-browser-to-loom-edit.md), [283](283-add-f3-and-ctrl-f-search-panel-to-loom-edit.md)

---

## 1. Problem & Motivation
PR #16's literal search can crash the editor. Reproduced in a PTY: open a document containing `Ⱥ`, press ^F, and type `ⱥ`; `updateSearchMatches` panics with `slice bounds out of range [:3] with length 2` at cmd/loom/edit.go:310.

## 2. Technical Specification / Findings
Lowercasing changes UTF-8 byte lengths. The code computes offsets in lowercased text, then slices the original text with those byte offsets. Preserve original rune positions when matching; do not fix this with bounds checks that silently lose matches.

Observed at PR head b8a59ca; this branch has not been merged locally. Before implementation, check live code and recent commits and reverify the finding.

Updated review (2026-10-07, PR head a9f1a7d): rune-based matching fixes the reproduced panic. PTY probes for length-growing `Ⱥ` → `ⱥ` and length-shrinking `K` before a match stay alive without modifying the document. The full make test-q1 suite passes. New unit tests cover accented/CJK text but not these length-changing mappings or exact span positions. Keep open pending integration and remaining acceptance verification.

## 3. Implementation & Verification Plan
/goal Literal search handles Unicode case mappings without crashes or incorrect match positions; verify the behavior, or stop and report when blocked on a user decision or denied permission.

Acceptance: Cover characters whose case mapping grows or shrinks UTF-8 byte lengths, multibyte prefixes, next/previous navigation, and exact highlighted spans. Verify a real PTY search, make test-q1 and make install.
