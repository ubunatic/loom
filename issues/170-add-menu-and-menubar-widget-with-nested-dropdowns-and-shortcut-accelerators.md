# 170 — Add Menu and MenuBar widget with nested dropdowns and shortcut accelerators

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: popup.go, choice.go, issues/155-built-in-modal-and-dialog-overlay-primitive.md

---

## 1. Problem & Motivation
Full-featured terminal desktop applications, editors (like `ansiedit`), and multi-tool shells commonly require a top menu bar (`File`, `Edit`, `View`, `Help`) with popdown submenus, accelerator shortcuts (`Ctrl+S`, `F1`), separator lines, and checkable menu items.
While Loom has `Popup` for generic floating rects and `Choice` for flat lists, assembling a cohesive menu system with:
- Top-level horizontal menu bar navigation (`←`/`→` switching between active menus),
- Dropdown popup placement directly under the activated title,
- Keyboard accelerators / mnemonic underlines (e.g. `Alt+F`),
- Submenu nesting,
currently requires reinventing popup coordinate math, key routing, and dismissal logic.

## 2. Technical Specification / Findings
Introduce `loom.MenuBar` and `loom.Menu` (implementing `loom.Widget`, `loom.EventConsumer`, `loom.MouseConsumer`):
- **Structure**:
  ```go
  type MenuItem struct {
      Label       string
      Mnemonic    rune      // e.g. 'F' in "File"
      Shortcut    string    // e.g. "Ctrl+S"
      Disabled    bool
      Checked     *bool     // optional toggle checkmark
      Action      func()
      Submenu     []MenuItem
  }

  type MenuBar struct {
      Menus       []Menu
      ActiveMenu  int
      Open        bool
  }
  ```
- **Navigation & Interaction**:
  - `F10` or `Alt+<mnemonic>`: Focus menu bar and open target menu.
  - `←` / `→`: Navigate between top-level menus (opens adjacent menu if one is already open).
  - `↑` / `↓`: Navigate items in open dropdown.
  - `Enter` / `Space`: Trigger item action (or open submenu).
  - `Esc`: Close open menu (or unfocus menu bar).
  - Mouse hover & click to open menus and select items; clicking outside closes open menus.
- **Rendering**:
  - Single-row horizontal bar with menu titles.
  - Overlay dropdown box with borders, shortcut labels aligned to the right, and separator rules (`---`).

## 3. Implementation & Verification Plan
- Create `menu.go` and `menu_test.go`.
- Unit tests:
  - Menu bar activation, horizontal and vertical arrow navigation.
  - Dropdown bounds calculation and overlay layer rendering.
  - Action callback triggering and toggle checkmark updates.
  - Escape dismissal and outside click handling.
- Document in `docs/Widgets.md`.

## Sprint goal (roadmap 180)
/goal Ship `Menu` and `MenuBar` with one level of items (submenus are 193), accelerators via 158's KeyMap, and tests plus an example for human review; stop and report when interaction details need a user decision.
