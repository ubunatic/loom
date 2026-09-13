# 035 — Extend loom.View Navigation Keybindings and Export TerminalSize Helper

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feature
**Related**: [view.go](file:///home/uwe/projects/loom/view.go), [pane.go](file:///home/uwe/projects/loom/pane.go)

---

## 1. Problem & Motivation
In `loom`:
1. `loom.View.HandleKey` only supports basic single-line step navigation (`up`, `down`, `j`, `k`) and quit (`esc`, `ctrl-c`, `q`). Common pager ergonomics such as page navigation (`pgdown`, `pgup`, `space`, `b`), half-page navigation (`ctrl-d`, `ctrl-u`), and jump-to-boundary (`home`, `end`, `g`, `G`) are missing, forcing developers to rewrite custom scroll controllers for standard viewing tasks.
2. `loom.Pane` internally determines terminal dimensions via `termSize(fd)`, but `loom` does not export a public `TerminalSize()` helper function. Applications that need to inspect terminal dimensions (e.g. to configure initial view parameters or layout without immediately allocating a `Pane`) must import `golang.org/x/term` or `golang.org/x/sys/unix` directly.

## 2. Proposed Solution
1. **Extend `View.HandleKey`**:
   - Add support for `pgdown`, `ctrl-f`, `space` (scroll forward by visible page height `lastH`).
   - Add support for `pgup`, `ctrl-b`, `b` (scroll backward by visible page height).
   - Add support for `ctrl-d`, `ctrl-u` (half-page scroll).
   - Add support for `home`, `g` (scroll to top), and `end`, `G` (scroll to bottom).
2. **Export Terminal Size Helpers**:
   - Provide `loom.TerminalSize() (cols, rows int, err error)` or `loom.TerminalWidth()` / `loom.TerminalHeight()`.

## 3. Verification & Acceptance
- Unit tests in `view_test.go` verifying that `pgup`, `pgdown`, `space`, `home`, `end`, `ctrl-d`, `ctrl-u` update `View.Scroll` appropriately.
- Unit tests verifying exported `TerminalSize` helpers.
