# 270 — Pass demo arguments to loom widgets --show after --

**Status**: Closed — user runbook passed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: [197](197-complete-widget-names-for-loom-widgets-and-show.md), [271](271-richtextedit-save-ends-files-with-a-final-newline.md)

---

## 1. Problem & Motivation
`loom widgets --show RichTextEdit` always starts the demo with built-in sample content. To try the editor on a real file (for example to check Save behavior on `test.ansi`), the user needs to hand the demo a file: `loom widgets --show RichTextEdit -- test.ansi`.

## 2. Technical Specification / Findings
- `cmd/loom/widgets.go`: `--show` is a bool flag and all positional args are widget names passed to `showWidgetDemos(args, …)`. Cobra keeps args after `--` in `args`; split them with `cmd.ArgsLenAtDash()`.
- `gallery.New(name)` takes no options. Add a generic way for a demo to receive its args (e.g. `gallery.NewWithArgs(name, args)` or an optional interface on the demo) instead of a RichTextEdit special case in the CLI.
- First consumer: the RichTextEdit demo (`gallery/richtextedit.go`) takes `<file>`, loads it via `RichDocument.FromANSI`, and sets `RichTextEdit.FilePath` so Ctrl+S writes back to it. A missing file starts an empty document bound to that path.
- Reject demo args when more than one widget (or `all`) is shown, and for demos that take none, with a clear error.

## 3. Implementation & Verification Plan
/goal `loom widgets --show RichTextEdit -- <file>` opens the file in the RTE demo and saves back to it, with a generic demo-args path other demos can adopt; stop and report when blocked on a user decision or denied permission.

Acceptance: CLI test for the `--` split and error cases; RTE demo test that loads a file and that Ctrl+S writes to the same path; `--help` and shell completion mention the `-- <args>` form; `make install`.

---

## Delivered

- **Generic Demo Arguments**: Implemented argument splitting via `cmd.ArgsLenAtDash()` in `cmd/loom/widgets.go` and `gallery.NewWithArgs(name, args)` in `gallery/gallery.go`.
- **RichTextEdit File Loading & Save**: `RichTextEdit` demo loads files via `RichDocument.FromANSI`, sets `FilePath`, and saves back on Ctrl+S; missing paths start an empty document bound to the path.
- **Validation**: Demo arguments are rejected for multi-widget displays, `all`, or demos that do not accept arguments. `--help` usage updated to `loom widgets [name] [-- <demo-args>]`.
- **Commits**: `0c97f5e`
- **Tests**: `make test-q1` passed. CLI and demo unit tests verified in `cmd/loom/main_test.go` and `gallery/gallery_test.go`.
