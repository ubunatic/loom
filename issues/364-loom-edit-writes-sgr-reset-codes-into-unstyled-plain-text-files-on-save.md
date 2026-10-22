# 364 — loom edit writes SGR reset codes into unstyled plain-text files on save

**Status**: Closed — Unstyled documents serialize and save as plain text, styled ones keep SGR and the final reset (e1f00b8); Upgrading row added; host live save of a .txt is byte-exact 'qx\n', q1 green
**Priority**: P2 (Medium)
**Severity**: Major
**Category**: Bug
**Related**: 279 (`loom edit`), 271 (final newline), 354 (found during its live check)

---

/goal Saving a document with no styling writes plain text only (no escape sequences), and styled
documents still roundtrip losslessly; reproduction test first; or stop and report when blocked on a user
decision or denied permission.

## Reproduction (host, 2026-10-22, installed loom at 783bd30)

File `x\n`, `loom edit f.txt`, type `q`, Ctrl+S. The file becomes
`ESC[0m ESC[39m ESC[49m qx ESC[0m \n` instead of `qx\n`. Every save of a plain file (source code, notes)
corrupts it for other tools.

## Notes

- Likely `SerializeAnsiBuffer` emits resets for default-styled cells; emit SGR only where a cell's style
  differs from the current one, starting from default.
