# 168 — Add Tree widget with hierarchical nodes, expand/collapse, and keyboard navigation

**Status**: Open
**Priority**: P1 (High)
**Severity**: Moderate
**Category**: Feature
**Related**: choice.go, examples/filebrowser/filebrowser/navigation.go, syntax/outline.go

---

## 1. Problem & Motivation
Loom supports flat selectable lists via `loom.Choice` and file navigation via `NavigationPane` (which swaps directory contents into a flat list).
However, many developer tools and TUIs require a true multi-level hierarchical tree view:
- AST / syntax symbol hierarchies (e.g. `syntax.Symbol` outline in issue 120 / docs/Widgets.md §9).
- Project tree views (nested directory trees with expanded/collapsed subfolders).
- JSON / YAML / object inspector trees.
Currently, consumers must flatten tree data manually and track indentation, fold state, and selection indices, leading to redundant boilerplate across apps.

## 2. Technical Specification / Findings
Introduce `loom.Tree` (implementing `loom.Widget`, `loom.EventConsumer`, `loom.MouseConsumer`, `loom.Focusable`, `loom.Themeable`):
- **Node Structure**:
  ```go
  type TreeNode struct {
      ID       string
      Label    string
      Data     any
      Children []*TreeNode
      Expanded bool
      Icon     string
  }
  ```
- **Navigation & Interaction**:
  - `↑` / `↓` (`k` / `j`): Navigate visible rows.
  - `→` / `l`: Expand node (or navigate to first child if already expanded).
  - `←` / `h`: Collapse node (or navigate to parent node if already collapsed).
  - `Enter` / `Space`: Toggle expand/collapse or trigger `OnActivate(node)`.
  - Mouse click on disclosure arrow (`▶` / `▼`) toggles collapse; click on row selects node.
- **Rendering**:
  - Guide lines (`│`, `├─`, `└─`) or indent indentation.
  - Disclosure glyphs (`▶`, `▼`).
  - Active selection highlighting matching `ChoiceStyle`.
  - Visible row virtualization (scroll offset) when tree exceeds viewport height.

## 3. Implementation & Verification Plan
- Create `tree.go` and `tree_test.go`.
- Table-driven unit tests:
  - Navigation across nested nodes, skipping collapsed children.
  - Deep expand/collapse state toggling.
  - Keyboard and mouse selection events.
  - Viewport scrolling with large tree structures.
- Document in `docs/Widgets.md`.
