# 300 — loom edit: replace top title bar with compact bottom status icons

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Enhancement
**Category**: UX
**Related**: [288](288-compose-loom-edit-panels-with-standard-widgets-and-clipped-responsive-layout.md), [293](293-loom-edit-redundant-top-file-label-and-bottom-file-menu.md)

---

## 1. Problem & Motivation
`loom edit` currently displays a separate top title bar solely for displaying verbose status text (`theme: <theme> mouse: <mouse>`). Having a dedicated top bar consumes valuable vertical screen real estate without providing sufficient utility.

Replace the top title bar by integrating compact, iconic status indicators (for theme, mouse grab, and alt-screen mode) directly into the bottom bar alongside the existing status indicators.

Solve this on the library level: provide reusable, spec-configurable status pill / badge components in the library (or compound bottom bar), defining iconic indicator glyphs and colors in `spec/defaults.yaml` (`editor.status_icons` / `spec/widgets.yaml`).

## 2. Technical Specification / Findings
- **Remove Top Header Bar**:
  - In `cmd/loom/edit.go`, remove the top header layout row from `editView.Draw()` / layout computation, recovering +1 row of vertical space for the text editor and side panel.
- **Compact Bottom Status Indicators**:
  - Move app-level status into the bottom bar (or file bar / status pill cluster):
    - Theme indicator (e.g. `🎨 <theme>` or icon).
    - Mouse grab status (e.g. `🖱️` / `[M]` icon).
    - Alternate screen status (e.g. `⛶` / `[A]` icon).
  - Use short iconic glyphs instead of full words like `"theme: plain mouse: false altscreen: true"`.
  - Declare status glyphs and pill styling in `spec/defaults.yaml` and `spec/widgets.yaml`.

## 3. Implementation & Verification Plan
- Remove top header row allocation in `cmd/loom/edit.go`.
- Add compact iconic indicator rendering in the bottom status line / file bar.
- Update layout calculation tests and PTY snapshot tests in `cmd/loom/edit_test.go` and `cmd/loom/edit_pty_test.go`.
- Validate layout geometry with `loom eval` and `loom measure`.
- Verify with `make test-q1` and `make install`.

/goal Remove the top title bar in loom edit and relocate theme, mouse, and alt-screen status to compact iconic indicators in the bottom bar, backed by spec definitions, or stop and report when blocked on a user decision or denied permission.
