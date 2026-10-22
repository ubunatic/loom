# 354 — Editor swallows the first key after Escape

**Status**: Closed — Escape followed by a key in a later read now dispatches the lone Escape first (da30f65), plus the expired-Escape race (783bd30); repro had no real pause, the host wait helper 'timeout 0.1 true' does not wait. Host live check with a real 0.8 s pause and the no-pause repro both save on the first Ctrl+S; q1 green
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 353 (found during its live check), 344 (Escape + `\ea` merged into one event)

---

/goal Make the first key after a lone Escape reach `loom edit` as itself (e.g. Ctrl+S saves), with a
PTY reproduction test first; or stop and report when blocked on a user decision or denied permission.

## Reproduction (host, 2026-10-21, installed loom at bcc267a, tmux escape-time 0)

`loom edit f.txt`, type `q`, Escape, wait 0.8s, Ctrl+S: status stays "Modified". A second Ctrl+S
saves. The same happens after Escape closes the Ctrl+Space bar. A wait of 0.8s should rule out a
timing merge, so the decoder probably holds Escape as an Alt prefix until the next key arrives.
