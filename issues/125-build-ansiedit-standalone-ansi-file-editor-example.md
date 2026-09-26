# 125 — Build ansiedit standalone ANSI file editor example

**Status**: Open
**Priority**: P1 (High)
**Severity**: Normal
**Category**: Feature
**Related**: `examples/ansiedit/`, `examples/ansiviewer/`, `docs/data/ansiedit-design-004.ansi`, `docs/data/ansiedit-design-005.ansi`

---

## Goal

`/goal`: Implement a standalone, fully-functional ANSI graphic and text art editor CLI example in `examples/ansiedit` (`cmd/ansiedit` / `examples/ansiedit/ansiedit/`) matching the visual mockups in `docs/data/ansiedit-design-004.ansi` and `docs/data/ansiedit-design-005.ansi`. Use Loom standard UI components and widgets. Fix small SDK issues directly, and file dedicated issues if larger SDK deficiencies are identified.

## 1. Functional Requirements & Key Specifications

- **Command Line**: `$ ansiedit <file.ansi>` (creates if not existing, or loads existing ANSI file).
- **Core Canvas Editing**:
  - Overtype and character insertion modes.
  - Cell editing with FG and BG colors and attributes (bold, dim, underline, invert).
  - Navigation: Arrow keys for cell-by-cell navigation; `Ctrl-Left` / `Ctrl-Right` jump word boundaries; `Ctrl-Up` / `Ctrl-Down` jump vertical space / object boundaries.
  - Character deletion: `Del` deletes current cell char (replaces with space), `Backspace` deletes preceding char.
  - Character clipboard: `Ctrl-c` copies cell char+styling, `Ctrl-x` cuts cell, `Ctrl-v` pastes cell.
  - Save & Quit: `Ctrl-s` saves file buffer, `F10` universally quits from anywhere.
- **Unified Side Panel (Function Key Views)**:
  - `Tab`: Toggles active focus between Canvas and Side Panel.
  - `F1`: Info Panel (File name, size, grid dimensions, cursor inspector with (X, Y), rune, code point, FG, BG, attributes).
  - `F2`: 2D Color Palette (Hue × Luminance dark-to-bright ramp matrix + greyscale ramp). `Arrow` keys move picker, `Enter` / click applies FG color, `Shift-Enter` / `Ctrl-Enter` applies BG color.
  - `F8`: Keybindings reference cheat sheet.
  - `F9`: Color theme switcher.
  - `F10`: Universal quit.
  - Only one side panel view is visible at a time. Clean canvas drawing area with no overlapping help text.

---

## Milestones

- **M1 (Core Model & Canvas Buffer)**: ANSI file loader/saver, in-memory cell grid with runes, FG/BG color codes, styling attributes, and core edit operations (put, erase, cut, copy, paste).
- **M2 (TUI App & Viewport Rendering)**: Main Loom application loop, layout frame matching design mockups (header, 22-wide left side panel, right canvas buffer, status footer), cursor rendering, and keyboard navigation (`Arrows`, `Ctrl-Arrows`).
- **M3 (Unified Side Panel & Color Palette)**: `F1` (Info), `F2` (2D Hue × Brightness palette with FG/BG assignment via `Enter` / `S-Enter`), `F8` (Keybindings), `F9` (Theme), and `Tab` focus switching.
- **M4 (Integration, Tests & make install)**: Full unit tests, PTY test coverage, Makefile entry (`make build`, `make install`), and validation.
