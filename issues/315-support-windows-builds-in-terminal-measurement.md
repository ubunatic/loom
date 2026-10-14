# 315 — Support Windows builds in terminal measurement

**Status**: Closed — supported Windows cross-builds in measure and probe commands
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: Windows cross-build failure reported from loom-games

---

## 1. Problem & Motivation
`ubunatic.com/loom` does not currently compile for Windows because terminal measurement
uses Unix-only polling APIs. This prevents downstream applications from cross-compiling
for Windows even when their own code is otherwise portable.

## 2. Technical Specification / Findings
A Windows amd64 cross-build of downstream `loom-games` fails in
`measure/zwj.go` on references to `unix.PollFd`, `unix.POLLIN`, `unix.Poll`, `unix.EINTR`,
and `unix.Read`.

## 3. Implementation & Verification Plan
/goal Make the Loom terminal-measurement package compile and work on Windows, with platform-appropriate input polling or a documented fallback; verify a Windows amd64 build and tests, or stop and report if required Windows behavior needs a user decision or unavailable platform access.
