# 339 — Split Pane.run by lifecycle concern

**Status**: In Progress
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**:

---

## 1. Problem & Motivation
`harnez assess --hotspots` reports `Pane.run` as a 572-line event loop. Mouse
report scanning, coordinate conversion, cursor effects, dispatch, and repaint
are a self-contained event concern embedded in the loop.

## 2. Technical Specification / Findings
The mouse path must continue draining SGR reports in order, translate terminal
coordinates to pane-local coordinates once, repaint after each event, and leave
trailing key bytes for the key decoder.

## 3. Implementation & Verification Plan
Extract SGR mouse draining to `pane_run_mouse.go` without changing the public
API or routing semantics.

Verification: run `go test ./...` and `make container-test CONTAINER_CPUS=2`.
