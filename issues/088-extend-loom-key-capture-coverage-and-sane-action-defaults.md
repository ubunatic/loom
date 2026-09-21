# 088 — Extend Loom key capture coverage and sane action defaults

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Enhancement
**Related**: None

---

## Problem & Motivation

Loom's key-capture system should be dependable for the full range of keys users
expect in terminal applications, while widgets and example apps should share
clear, predictable defaults for common actions. Existing choices are useful
starting points, but coverage and consistency need to be made explicit before
the input model grows further.

## Scope

Support capture and matching for letters (`a-z`, `A-Z`), digits, punctuation,
German umlauts, and modified keys using Shift, Ctrl, and Alt. Review and extend
common action defaults across Loom widgets and example apps, including:

- quit/close: `q`, `Q`, `F10`, `Ctrl-C`
- activate/toggle/select: `Space`, `Insert`
- undo/redo: `u`, `U`, `Ctrl-Z`, `Ctrl-R`
- edit/view: `F4`/`e`, `F3`/`v`
- navigation: arrows, Page Up, Page Down, with WASD only as an opt-in

Preserve intentional application-specific bindings, document conflicts and
platform/terminal limitations, and keep the resulting defaults sane and
consistent.

## Goal

/goal: Make Loom reliably capture and distinguish the specified printable,
international, function, navigation, and modified keys, and establish tested,
documented default bindings for common actions across widgets and example apps
without breaking existing intentional choices.

## Verification

Add focused coverage for key decoding/matching and representative widget/app
bindings, including modifier combinations, umlauts, punctuation, and conflict
handling. Verify the examples remain usable with their documented defaults.
Extend the PTY tests to exercise these key sequences and prove the resulting
actions work through a real terminal input path.

---

## Resolved Design (host decisions, binding for the sprint)

- Do not rename existing key names or change existing bindings. Only add coverage and new defaults where nothing
  conflicts; every conflict is documented, not resolved by removal.
- Audit first: a table-driven test lists every required key with the byte sequence(s) a common terminal sends
  (xterm-style: plain, Shift/Ctrl/Alt variants of letters, digits, punctuation, umlauts `ä ö ü Ä Ö Ü ß` as UTF-8,
  `Alt+<key>` as ESC-prefix, `Ctrl+<letter>`, F1-F12, arrows, Home/End, PgUp/PgDn, Insert/Delete, Shift/Ctrl/Alt
  arrows as `CSI 1;<mod>` sequences) and asserts the decoded `KeyEvent` name. Keys a terminal cannot distinguish
  (for example Ctrl+I vs Tab, Ctrl+M vs Enter, Ctrl+Shift+letter) are listed as a documented limitation with a test
  asserting the actual, documented outcome.
- Defaults are decided for the LIBRARY widgets only (Choice, View, Tabs, Popup, Frame actions, Input widgets if any),
  listed in one evergreen doc `docs/KeyDefaults.md` (PascalCase): action, default keys, where it applies, conflicts,
  terminal limitations. A test parses the table or a Go table shared with the doc so the doc cannot drift.
  Examples keep their own intentional bindings; the doc lists them and flags conflicts, no example is rewritten.
- WASD stays opt-in only (an option, off by default) and only if it needs no new public API beyond an option field
  on the navigation helper that already exists; otherwise document it as not provided.
- Test names and the doc must not claim support that the audit did not prove.

## Milestones (lean sprint, developer: luna:low)

Host reviews only diffs, test output and evidence frames; this ticket is the only channel. Root-package tests needing
`/dev/tty` fail before this work; ignore them. Commit each milestone (message ends '(issue 088 MX)'), stage only your
files, never docs/README.md, no stray binaries in the repo root; if `.git/index.lock` blocks the commit, stage your files
and say so (the host commits). Evidence: frames produced by code, gated on env `LOOM_EVIDENCE=1`, written to repo-root
`docs/progress/088/` (find the root by walking up to `go.mod`); run only this ticket's evidence test; view frames
ANSI-stripped; labels on their own rows. gofmt. No dead code.

### M1 - Decode audit
- The audit test and its rendered table. No decoder changes yet; failing rows are recorded as `known gap` entries in the
  test with a one-line reason, so the suite is green and the gaps are explicit.
- Evidence: `M1-decode-table.ansi` (columns: key, bytes as hex, decoded name, status OK / gap / terminal limitation).

### M2 - Close the gaps
- Fix each `known gap` that is a decoder bug (test first, remove the gap entry); leave true terminal limitations.
- Evidence: `M2-decode-table.ansi` (same table after the fixes; the gap count must drop).

### M3 - Defaults for library widgets and the doc
- `docs/KeyDefaults.md` and the shared table plus tests that check the widgets really react to those defaults
  (quit, activate/toggle, undo/redo where an editing widget exists, navigation, page keys). New defaults only where
  no conflict; conflicts listed.
- Evidence: `M3-defaults.ansi` (the defaults table rendered from the same Go table the doc is checked against).

### M4 - PTY
- With `internal/ptytest` and `SendRaw`: send a set of the key sequences through a real PTY to a widget that reports
  the received key (an existing example or a tiny test program), assert the reported names, including umlauts and one
  Alt and one Ctrl combination. Evidence: `M4-pty-keys.ansi` (final screen).
