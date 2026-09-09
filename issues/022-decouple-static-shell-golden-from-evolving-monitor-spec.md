<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# 022 — Decouple static shell golden tests from evolving example monitor spec

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Tech Debt / Testing
**Related**: [006](006-static-declarative-monitor-shell-with-a-minimal-validated-contract.md), [010](010-geometry-and-visual-evidence-milestone-before-rich-content.md), [011](011-aligned-dashboard-rows-and-ansi-safe-truncation.md), [frame_test.go](../frame_test.go), [monitor.yaml](../examples/monitor/spec/monitor.yaml)
**Roadmap stage**: 5 & 6 testing infrastructure
**Depends on**: None

## Problem and findings

In [`frame_test.go`](../frame_test.go#L12-L29), [`TestStaticShellGolden`](../frame_test.go#L21) loads [`examples/monitor/spec/monitor.yaml`](../examples/monitor/spec/monitor.yaml) as its test fixture via `shellFixture(t)`.

In commit `c19f226`, adding dummy `rows` to `monitor.yaml` caused `TestStaticShellGolden` to fail because the boxes now contained children. Rather than decoupling the test fixture, the commit added an in-memory mutation:
```go
	// Preserve the original empty-shell geometry fixture as content is added.
	for i := range w.(*Frame).Boxes {
		w.(*Frame).Boxes[i].Child = nil
	}
```

This anti-pattern creates test coupling between the evolving runnable example ([examples/monitor](../examples/monitor)) and the baseline empty-shell geometry regression tests:
1. Mutating `b.Child = nil` inside a golden test bypasses the parser/builder contract and masks potential regressions in how children interact with parent box geometry.
2. Every subsequent addition of content (graphs, status bars, actions) to `monitor.yaml` requires hacks in tests that expect an empty shell.
3. The actual output of `examples/monitor` lacks a full integration golden test reflecting its real contents.

## Acceptance criteria

- [ ] Extract an explicit, immutable empty-shell specification fixture (e.g. in `testdata/fixtures/empty-shell.yaml`) for testing the bare frame and box chrome layout.
- [ ] Update `TestStaticShellGolden` to load the dedicated empty-shell fixture and remove the `b.Child = nil` mutation.
- [ ] Add explicit golden test coverage for `examples/monitor/spec/monitor.yaml` verifying that rendered output includes declared rows, values, and titles without mutation.
- [ ] Verify `check-geometry-replay.py` and `make test` pass cleanly with clean fixture boundaries.

## Verification

Run `go test ./...` and `make test` to verify that both the empty shell regression test and the monitor example test pass independently without in-memory fixture mutation.

## Scope limits

Limited to test fixtures and assertions in `frame_test.go` and `testdata/`. Does not modify `Frame` rendering or `Canvas` logic.
