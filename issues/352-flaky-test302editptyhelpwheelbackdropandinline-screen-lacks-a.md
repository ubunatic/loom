# 352 — Flaky Test302EditPTYHelpWheelBackdropAndInline: screen lacks " A "

**Status**: Closed — Test waited for help dismissal but accepted the cleared screen before the status bar's " A " returned; now waits for that click target plus fresh-output waits around each input, raceFactor timeouts (b3eef58). VT replay of that redraw fails with the old check; original flake not reproduced by stress; host -race -count=20 on 2 CPUs and q1 green
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Bug
**Related**: 340, 341 (same PTY sync pattern), 337 (a1b4372)

---

/goal Make `Test302EditPTYHelpWheelBackdropAndInline` (cmd/loom) deterministic by waiting on screen
state before and after each input, timeouts scaled by `raceFactor`, no sleeps; pass
`-race -count=20` on 2 CPUs; or stop and report when blocked on a user decision or denied permission.

## Problem

The 350 developer's `make test-q1` (2026-10-21, after 534843e, which does not touch this test)
failed once: the test could not find `" A "` in the rendered screen. Host reruns passed 5/5 plain,
20/20 with `-race -count=20 GOMAXPROCS=2 taskset -c 0,1`, and a full q1 run was green.

## Fix

Same pattern as 340 (916b3a6) and 341 (5f1d8be): fresh-output screen waits around each key and
wheel event in `cmd/loom/edit_302*_test.go`.
