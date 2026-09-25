# 117 — Polish textedit TUI matching visual design mockup

**Status**: Open
**Priority**: P2
**Severity**: Minor
**Category**: Enhancement
**Related**: `examples/textedit/textedit/textedit.go`, `docs/data/textedit-design-001.ansi`

---

## Goal

`/goal`: Enhance the visual presentation, styling hierarchy, and UX of `examples/textedit` to match the design specified in `docs/data/textedit-design-001.ansi`.

## 1. Problem & Motivation

The initial prototype of `examples/textedit` functions well mechanically (split panes, file browser, editor, terminal, MRU list, keybindings), but its visual styling is minimal and lacks polish:
- Editor pane lacks line numbers, gutter separation, and string/type syntax coloring.
- File explorer uses simple text prefixes (`>`) without full-row selection highlight bars or directory indicators.
- Workspace title bar lacks clean badges/pills for file modified status (`[demo.go ●]`) and active focus (`[Focus: Editor]`).
- Embedded terminal header and prompt (`❯`) lack distinct styling.
- Bottom status bar displays plain text rather than structured keycap pills (`[C-s] Save`, `[F10] Quit`) with status indicator pills (`● Ready`).

## 2. Technical Specification

Follow the approved ANSI mockup in `docs/data/textedit-design-001.ansi`:

1. **Header & Window Frame**:
   - Frame title and top bar with clean breadcrumbs/pills: file status pill (`[<filename> ●]`), focus indicator pill (`[Focus: <Area>]`), and version tag.
   - Clean border framing matching Loom theme conventions.

2. **Editor Pane**:
   - Left line number gutter (` 1 │ `, ` 2 │ `, etc.) in muted/dimmed style (`Style{Dim: true, FG: Gray}`).
   - Expanded syntax highlighting: keywords (cyan/bold), comments (gray/dim), strings (`"..."` in green), and function identifiers.
   - Caret / block cursor representation.

3. **File Explorer Sidebar**:
   - Selected item highlighted with active row background bar (`BG_SELECT` / cyan theme).
   - Folder indicator prefixes (`▾` for expanded, `▸` for collapsed or directories).
   - Clean MRU list section.

4. **Terminal Pane**:
   - Clear title header (`💻 Terminal`).
   - Styled prompt symbol (e.g. `❯`) with green/cyan accent.
   - Distinct styling for commands (`$ ...`) and output messages.

5. **Status & Hotkeys Bar**:
   - Left-aligned status pill (`● Ready` in green, `● Saved <file>`).
   - Keycap badges for shortcuts: `[C-s] Save`, `[C-o] Open`, `[C-c/v] Clip`, `[Tab] Focus`, `[C-b] Tree`, `[F10] Quit`.

## 3. Implementation & Verification Plan

1. **Implement in `examples/textedit/textedit/textedit.go`**:
   - Update `Draw` methods for `editorWidget`, `fileBrowserWidget`, `terminalWidget`, and `TextEditApp`.
   - Ensure all layout measurements and borders cleanly fit within standard terminal bounds (80x24+).
2. **Update Tests**:
   - Update existing unit tests in `examples/textedit/textedit/textedit_test.go` to match new render outputs.
   - Ensure PTY smoke tests in `internal/examplesreg/smoke_test.go` pass.
3. **Verification**:
   - Run `make test-q1` and `make install`.
   - Run `git diff --check`.
