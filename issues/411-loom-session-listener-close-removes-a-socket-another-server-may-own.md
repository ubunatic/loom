# 411 — loom session listener Close removes a socket another server may own

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: [393](393-fix-hang-in-testserverlasttabexitcleanssocket-in-internal-session.md)

---

## 1. Problem & Motivation
`sessionListener.Close` (`internal/session/ipc.go:132-139`) always runs `os.Remove(l.path)`. If a newer session server has already replaced the socket at that path (for example after this server hung and a client started a fresh one), the old server's Close deletes the new server's socket, and clients can no longer connect to it.

Found by Jules ideas run `session-vt` (GitHub #76, Jules ID 10548991855057164559); the reference was checked against main 14ec5a9. The takeover scenario itself is not verified.

## 2. Technical Specification / Findings
Record the socket's device and inode (`os.Lstat` + `os.SameFile`) when listening, and remove the path on Close only if it is still the same file.

## 3. Implementation & Verification Plan
/goal Close removes only its own socket: a test replaces the socket file with another listener's, closes the first listener, and checks the second socket remains. Confirm first that a takeover can actually happen in the current server start logic; close the ticket if it cannot. Stop and report when blocked on a user decision.

## Jules runs

| kind | session | Jules ID | PR | state |
| --- | --- | --- | --- | --- |
| implement | jules-implement-411 | 10538171783052097585 | GH #79 | adopted: new test fails without the fix, host make test-q1 green |

Takeover check (host): `removeStaleSocket` replaces a socket only when dialing it fails, and a live listener still accepts, so a second server cannot take over a live socket by itself. It happens only when the file is removed from outside (tmp cleaner, manual rm). The fix is kept as a cheap guard; it also turns off Go's own unlink-on-close (`SetUnlinkOnClose(false)`), which removed the path without any check.
