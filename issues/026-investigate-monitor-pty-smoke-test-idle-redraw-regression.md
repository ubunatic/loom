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
