# 363 — loom man pages from cobra (loom man, man --install)

**Status**: Closed — loom man prints loom(1); loom man --install writes 18 cobra/doc pages to ~/.local/share/man/man1 (--dir for a custom dir); host q1 green, man loom-diff renders (ab6866c)
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: docs/ManPages.md (the standard), 362 (`loom diff` skipped a hand-written man page)

---

/goal `loom man` prints the generated man page and `loom man --install` writes the full page tree
(cobra/doc `GenManTree`) to `~/.local/share/man/man1`, per docs/ManPages.md, so `man loom` and
`man loom-<cmd>` work for every subcommand; tested and documented in docs/CLI.md; or stop and report when
blocked on a user decision or denied permission.

## Notes

- Pages come only from cobra metadata (Short/Long/Example); no hand-written `.1` files.
- Check that `make install` stays fast and does not install man pages unless docs/ManPages.md says so.
