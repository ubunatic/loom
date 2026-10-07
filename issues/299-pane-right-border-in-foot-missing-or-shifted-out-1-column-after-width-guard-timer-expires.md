# 299 — pane: right border in foot missing or shifted out 1 column after width guard timer expires

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [074](074-use-measured-winch-speed-for-adaptive-resize-width-guard.md), [136](136-pane-misses-a-resize-that-happens-during-startup-foot-app-draws-at-initial-size.md), [288](288-compose-loom-edit-panels-with-standard-widgets-and-clipped-responsive-layout.md)

---

## 1. Problem & Motivation
In terminal emulators like `foot`, the right border of full-screen / boxed TUIs (such as `loom edit` or boxed layouts) is missing or cut off. During active window resizing, the border is temporarily visible (due to width-guard clamping). However, approximately one second after resizing stops (when the window settles), the right border shifts out by one character to the right and disappears or causes unwanted line-wrapping.

Solve this problem on the library level: ensure `Pane` window resizing, `guardedCols()`, and terminal width synchronization in `pane.go` respect terminal column boundaries and right-edge margin constraints defined in `spec/defaults.yaml` (`pane.guard_duration`, `pane.max_cols`, and resize specifications).

## 2. Technical Specification / Findings
- **Width Guard Expiry & Column Calculation (`pane.go`)**:
  - During resize, `p.widthGuardActive = true` and `guardedCols(p.cols)` clamps canvas columns by `p.adaptiveN`.
  - When resizing pauses, `guardTimer` expires (`pane.go:1248-1256`):
    ```go
    case <-guardTimerC:
        guardTimerC = nil
        p.widthGuardActive = false
        p.adaptiveN = 0
        if full := p.guardedCols(p.cols); full != cols {
            cols = full
            canvas = NewCanvas(cols, p.rows)
            dirty = true
        }
    ```
  - When `p.widthGuardActive` becomes `false`, `guardedCols(p.cols)` expands `cols` to the raw `termSize(p.fd)` value (`p.cols`).
  - In `foot` (and Wayland terminal emulators with sub-cell window margins or auto-wrap on the rightmost column), rendering a character at column `p.cols` without newline suppression or with an off-by-one boundary calculation pushes the right border cell past the visible window edge or triggers premature cursor wrapping.
  - Audit terminal size querying (`termSize`), SIGWINCH handling, and full-screen canvas allocation against `spec/defaults.yaml`.

## 3. Implementation & Verification Plan
- Investigate `guardedCols` and raw column sizing in `pane.go` during and after resize settling.
- Ensure terminal canvas width never exceeds the actual addressable character cells and properly handles the rightmost border column across terminals (`foot`, `xterm`, `alacritty`).
- Add regression tests verifying canvas dimensions before, during, and after `guardTimer` expiration.
- Verify with `make test-q1` and `make install`.

/goal Fix the right border clipping and off-by-one column expansion in foot after the resize guard timer expires, ensuring rock-solid right-border rendering across all terminal emulators, or stop and report when blocked on a user decision or denied permission.
