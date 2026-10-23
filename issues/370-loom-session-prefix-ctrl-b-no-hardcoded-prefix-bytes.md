# 370 — loom session: prefix ctrl-b, no hardcoded prefix bytes

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 369 (`loom session`)

---

/goal `loom session` uses `ctrl-b` as its prefix (tmux habit; `]` needs AltGr on German keyboards), the
prefix and its forwarding come only from `spec/defaults.yaml`, PTY tests cover prefix+s, prefix+q,
prefix twice forwards ctrl-b, and `ctrl-]` reaches the child unchanged; or stop and report when blocked.

## Decision (user, 2026-10-23)

`ctrl-b` like tmux; loom session may grow into a tmux-like tool, so running tmux inside it is not a goal.

## Findings

- `spec/defaults.yaml` `session.prefix: "ctrl-]"` and `prefix_bindings.send_prefix: ["ctrl-]"]`.
- `cmd/loom/session.go:259` writes `0x1d` for send_prefix; forward the bytes of the key that was
  pressed instead.
- `cmd/loom/session.go:342` `scanSessionKey` special-cases bytes 16, 17, 29; derive the single-byte
  hotkeys from the spec bindings (or drop the special case if `ScanKey` already handles them), so a spec
  change needs no Go change.
- Help text in `sessionCommand` names `ctrl-]`; build it from the spec values.
- `cmd/loom/session_pty_test.go` sends byte 29; switch to ctrl-b (byte 2) and add a test that byte 29
  passes through.
