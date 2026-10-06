# 269 — Refine RichTextEdit hints, picker focus, and help

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [268](268-add-save-and-save-as-to-ricktextedit-with-a-file-menu.md)

---

## 1. Problem & Motivation
The RichTextEdit file bar still spends too much space repeating standard shortcuts, and the row can be clipped at common terminal widths. The Save as picker also switches focus without reliably moving the text cursor, stays open after clicks outside it, and visually runs the filename field into the directory list. Add concise key guidance and an F1 help popup that lists all RichTextEdit bindings.

## 2. Technical Specification / Findings
Keep the Ctrl+Space text-formatting popover separate from document actions. Move shortcut detail out of the crowded file bar and into the existing lower hotkey information area; retain a visible `F1 Help` entry there.

## 3. Implementation & Verification Plan
/goal Refine RichTextEdit hints and Save as interactions so the file bar stays readable, picker focus and dismissal behave predictably, and F1 explains every RTE key binding; stop and report if a product decision or denied permission blocks completion.

Acceptance: remove redundant `Alt+F`, `Ctrl+S`, and `Ctrl+Shift+S` text from the file bar and make the relevant shortcuts discoverable in the lower hotkey information area and F1 help popup; Tab focus changes also place the cursor in the focused FilePicker input; clicking outside the Save as popup closes it; a divider separates directory search from filename entry; F1 opens a popup listing all RichTextEdit key bindings. Verify hint layout at narrow and normal widths and add regression tests for focus, outside-click dismissal, divider, and help contents.

### Milestones

- **M1 delivered — concise hints, picker behavior, and F1 help:** Implemented the file bar cleanup, lower hotkey hints, scrollable help popup, picker cursor/focus updates, filename divider, and outside-click dismissal. Commit: `44d4464`. `go build ./...` and `make install` passed. The first `make test-q1` stopped in `go vet` before tests ran because `cmd/loom/main_test.go` called the nonexistent `loom.Canvas.Screen()` method; host confirmation is in `/tmp/loom-269-host-test-q1.log`.
- **M2 pre-work / required refinement — fix the new gallery test compile error:** Replace the unsupported `Canvas.Screen()` call in the new `cmd/loom/main_test.go` assertions with the existing canvas row access API. Then run one clean-worktree `make test-q1`, capture its full output, and check it for `--- FAIL`. Do not change assertions to weaken coverage. Preserve the pre-existing modified `docs/README.md` and untracked `test.ansi`; do not stage either.
