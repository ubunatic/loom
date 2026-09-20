# 085 — Add robust smoke PTY tests for all example apps

**Status**: Closed — completed in lean sprint 2026-09-20
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Infrastructure
**Related**: `examples/`, `internal/examplesreg/`, `internal/testpty/`

---

## 1. Problem & Motivation

While several individual packages contain headless unit tests, end-to-end interactive lifecycle coverage across all example applications in real pseudo-terminals (PTYs) is inconsistent. Some examples lack PTY verification altogether, allowing regressions in TTY initialization, terminal dimension negotiation, key routing, and clean exit behavior to go unnoticed.

## 2. Desired Behavior & Goal

`/goal`: Implement robust, end-to-end smoke PTY tests for all example applications in `examples/` that launch the app in a pseudo-terminal, verify main screen entry and width bounds, exercise at least one key action, and confirm clean exit.

- **Coverage**: Cover all registered reference applications (e.g. `split`, `tabs`, `monitor`, `screens`, `filebrowser`, `winch`, `background`, `splash`, `treemap`).
- **Standard Smoke Cycle**:
  1. Spawn the example inside a PTY session with standard dimensions (e.g. 80x24).
  2. Verify initial screen draw and check width / terminal boundary adherence.
  3. Send at least one interactive key event (e.g. navigation, toggle, or action).
  4. Send exit key sequence (`q`, `Esc`, or `Ctrl-C`) and assert clean process termination with exit code 0.
- **Robustness Against Visual Changes**:
  - Tests must **not** rely on rigid, pixel-perfect visual golden diffs that break on minor cosmetic changes.
  - Assertions should check structural invariants: expected core substrings, lack of line overflow / wrap corruption, and clean exit.
