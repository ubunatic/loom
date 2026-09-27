# 144 — Fix examples/media pane initialization to take full terminal size

**Status**: Open
**Priority**: P1 (High)
**Severity**: Major
**Category**: Bug
**Related**: [143](143-make-the-media-example-full-width-and-play-video.md)

---

## 1. Problem & Motivation

In `examples/media/main.go`, the pane was created with `loom.New(8)` (a fixed height of 8 lines inline mode), instead of `loom.New(1 << 16)` or full-screen terminal height. Because of this, the pane bounds clamped to 8 rows and did not expand to fill the available terminal window in full-screen/TUI use.

## 2. Technical Specification / Findings

Change `examples/media/main.go` to initialize the pane with full terminal screen mode (`loom.New(1 << 16)`), consistent with `filebrowser`, `ansiviewer`, `ansiedit`, and `textedit`. Update `examples/media/media_pty_test.go` to test that the media demo utilizes full terminal rows and cols when run in larger terminals.

## 3. Implementation & Verification Plan

- **/goal**: Make `examples/media` take the full terminal height and width, verify with PTY tests on larger terminals, or stop and report when blocked on a user decision or denied permission.
