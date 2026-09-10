# 026 — Investigate monitor PTY smoke-test idle redraw regression

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Bug
**Related**: [007](007-clock-watch-mode-with-independent-collection-and-redraw.md), [PTY smoke checker](../scripts/check-watch-pty.py)

## 1. Problem & Motivation

The monitor PTY smoke test now fails in the responsive/toggle invocation, despite the
other sequential checks passing. This is an untracked failure relative to the current
issue tracker. Existing issue [007](007-clock-watch-mode-with-independent-collection-and-redraw.md)
is closed and records the same PTY watch smoke as previously passing.

The failing command is:

```text
env LOOM_TEST_RESPONSIVE=1 LOOM_TEST_TOGGLES=1 GOWORK=off python3 scripts/check-watch-pty.py go run ./examples/monitor
```

It exited with status 1 and reported:

```text
RuntimeError: too few idle redraws
  at scripts/check-watch-pty.py line 72
```

The available evidence does not establish whether this is a product/test regression
or an environment/command mismatch; root cause is intentionally left open.

## 2. Technical Specification / Findings

The same sequential run passed:

- `go test ./...`
- `go test -race ./...`
- `python3 scripts/validate-spec.py`
- `python3 scripts/check-geometry-replay.py`
- `bash testdata/geometry/wide.sh`
- `bash testdata/geometry/slim.sh`
- `bash scripts/check-no-tty.sh`
- `python3 scripts/check-watch-pty.py`

The investigation should compare the standalone PTY check with the failing command,
including environment variables, workspace/module settings, timing assumptions, and
the monitor's idle redraw behavior. Reproduce before changing implementation or test
logic.

## 3. Implementation & Verification Plan

- Re-run the exact failing command and capture terminal dimensions, timing, stderr,
  exit status, and the observed redraw count.
- Compare it with the passing standalone PTY invocation and identify whether the
  responsive/toggle environment changes the expected idle-redraw cadence.
- If a repository defect is confirmed, add a focused regression test and bounded fix;
  if the command or environment is the mismatch, document the corrected smoke-test
  invocation and rationale.
- Re-run the full passing sequence above, the exact monitor command, and the
  standalone PTY checker.

## Scope limits

Investigation is limited to the monitor PTY smoke-test failure and its reproduction.
Do not infer a root cause from issue 007's historical passing result, and do not
broaden this ticket into general redraw scheduling or terminal support work without
new evidence.

## 4. Investigation — 2026-09-10

- Reproduced the exact command above: checker exit 1, `too few idle redraws`.
  An in-memory diagnostic of the unchanged checker observed child exit 0, zero
  `ESC[?25l` redraw markers, 1,006 output bytes, and exit after about 0.085 s.
  The PTY starts at 24 rows × 80 columns; stdout/stderr share that PTY.
- Most likely fix: append `--watch` to the monitor command. In
  [main.go](../examples/monitor/main.go), `watch` defaults to false and dispatches
  to `runWidth`; only `--watch` selects `runWatch`. The checker treats every
  supplied command as a watch process and requires at least 20 cursor-hide
  markers (emitted by `Canvas.Flush`), so normal show-once output fails this
  assertion. The responsive/toggle variables configure the checker, not watch mode.
- Verified the corrected invocation:
  `env LOOM_TEST_RESPONSIVE=1 LOOM_TEST_TOGGLES=1 GOWORK=off python3 scripts/check-watch-pty.py go run ./examples/monitor --watch`.
  It passed redraw, responsive placement, all four toggle states, and terminal
  restoration (56,096 bytes). The standalone checker also passed (6 bytes),
  but it runs a tiny Python raw-input probe and does not test monitor redraws.
  The [monitor README](../examples/monitor/README.md) already documents `--watch`
  with a prebuilt binary.
- Uncertainty: these local runs establish an invocation mismatch, not a scheduler
  regression. The checker starts its 0.25 s toggle / 1 s resize / 2.2 s quit
  deadlines at process launch, so cold `go run` compilation could independently
  consume the observation window. Prefer the documented prebuilt-binary workflow;
  readiness-relative timing is a possible follow-up only if that flake reproduces.
  No implementation changed or full regression suite run; leave the issue Open
  pending the planned verification and resolution.
