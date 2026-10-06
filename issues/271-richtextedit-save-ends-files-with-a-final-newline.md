# 271 — RichTextEdit save ends files with a final newline

**Status**: Closed — user runbook passed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: [268](268-add-save-and-save-as-to-ricktextedit-with-a-file-menu.md), [270](270-pass-demo-arguments-to-loom-widgets-show-after.md)

---

## 1. Problem & Motivation
Files saved by RichTextEdit have no trailing newline, so `cat file` leaves the shell prompt on the last line (zsh shows `%`). Text files should end with `\n` (POSIX line definition), and tools like `cat`, `wc -l` and `git diff` expect it.

## 2. Technical Specification / Findings
- `RichDocument.ToANSI()` (`richtext.go`) joins lines with `\n` and ends with `\x1b[0m`, no final newline. `RichTextEdit.SaveAs` (`richtextedit_file.go`) writes that output unchanged.
- Challenge raised with the user: appending `\n` alone breaks round-trips. `RichDocument.FromANSI` turns every `\n` into a new line, so each save/load cycle would add one empty line at the end. The fix must be symmetric:
  - save writes `…\x1b[0m\n` (newline after the final reset);
  - load drops exactly one trailing `\n` (also when only SGR codes follow the last text), so a document that really ends in an empty line still keeps it (`…\n\n`).
- Decide where the rule lives: `ToANSI()` is also a general serializer; prefer adding the newline in the file save path (default serializer only) and leave custom `SerializeDocument` output untouched, unless a reason to change `ToANSI` shows up.

## 3. Implementation & Verification Plan
/goal RTE-saved files end with exactly one newline and save/load/save is byte-stable; stop and report when blocked on a user decision or denied permission.

Acceptance: tests for save output ending in `\x1b[0m\n`; load→save→load round-trip keeps the line count for documents with and without a trailing empty line; custom `SerializeDocument` output is written as returned.

---

## Delivered

- **Symmetric Trailing Newline**: Updated default `RichTextEdit.SaveAs` serialization to append `\n` to `ToANSI()`, writing `…\x1b[0m\n`. Updated `RichDocument.FromANSI` to symmetrically drop exactly one trailing newline (including when followed by SGR codes).
- **Custom Serialization Preserved**: Custom `SerializeDocument` outputs remain untouched.
- **Commits**: `be35086`
- **Tests**: `make test-q1` passed. Single-line and empty-line round-trip tests verified in `richtext_test.go` and `richtextedit_file_test.go`.
