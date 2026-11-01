# 402 — TestRenderEmojiGrid fails: expects VTE Width 2 header, grid prints STANDARD Width 2

**Status**: Closed — Root cause: TestRenderEmojiGrid used env-detected RenderEmojiGrid, so TERM=foot selected the Standard render path and the grid output diverged. Fix pins the VTE path via RenderEmojiGridWithPath(..., "vte") (commit 3ab57ba). Green make test-q1 run: examples/loomoji and examples/loomoji/loomoji ok, no --- FAIL (log /tmp/loom-issue402-test-q1-2.log).
**Closed-By**: claude:haiku:med, session t402
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 398

---

/goal `make test-q1` is green again: decide whether `RenderEmojiGrid` or the test expectation is wrong (likely terminal/env detection leaking into the test), fix that side, and make the test independent of the ambient terminal; or stop and report when blocked on a user decision.

`examples/loomoji/loomoji/measure_grid_test.go:44`: `RenderEmojiGrid missing "VTE Width 2:"`; the grid prints `STANDARD Width 2:`. Pre-existing: fails on `ad12e5c` (before the 398 fix), found during the 398 lean sprint on 2026-11-01. Check for an ambient env dependency (e.g. `VTE_VERSION`, `TERM_PROGRAM`) first.
