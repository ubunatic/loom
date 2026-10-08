# 306 — Avoid redundant ANSI reset prefixes when loom edit saves files

**Status**: Open — In progress
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**:

---

## 1. Problem & Motivation
Files created with `loom edit` appear to gain `ESC[0m ESC[39m ESC[49m` at the start of every line. These reset/default-color codes may be redundant and make ANSI source files noisier to inspect or diff.

## 2. Technical Specification / Findings
The behavior was observed in `docs/data/simple.ansi`. Remove the forced prefix only if it is safe: saved content must preserve the same visual result and required style boundaries. If those codes are needed in some cases, avoid removing them there.

## 3. Implementation & Verification Plan
Inspect `loom edit` serialization and determine when these codes are necessary. Safely omit redundant per-line prefixes, document any retained cases, and verify save/reopen preserves appearance and ANSI semantics.

**/goal** Remove redundant forced ANSI prefixes from `loom edit` saves where safe and verify saved files preserve rendering, or stop and report if safety depends on a user decision or denied permission.
