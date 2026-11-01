# 403 — measure: StringWidth reads LOOM_FAST_MEASURE via os.Getenv on every call

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Performance
**Related**: 402 (ambient env leaking into tests)

---

/goal `measure.StringWidth` and other hot measure paths do no `os.Getenv` per call: env switches are read once (sync.Once/init) with a test setter, a benchmark shows the gain, and behaviour is unchanged; or stop and report when blocked on a user decision.

Reported by taia (harnez/taia `internal/tui/draw_bench_test.go`, commit 22e236c): on loom v0.5.6, 1000 transcript entries at 120x40, `os.Getenv` from `useFastMeasure` (`measure/measure.go:16-18`) was ~22% of Draw frame time.

Also audit the other per-call env reads: `measure/zwj.go:25,33,46` (`LOOM_ZWJ`) and `measure/spec.go:190-211` (`DetectRenderPath`: LOOM_RENDER_PATH, TERM, TERM_PROGRAM, VTE_VERSION, FOOT_*). Check whether they run per width call. A cached, settable env snapshot may also fix 402, which probably comes from VTE env detection leaking into the test.

## Decision

- Env switches (`LOOM_FAST_MEASURE`, `LOOM_ZWJ`, render path detection) are cached once per process in `measure/env.go`. In-process tests that change them call `measure.ReloadEnvSwitches()`, registered with `t.Cleanup` before `t.Setenv`.
- Loomoji test edits are allowed: `examples/loomoji/loomoji/measure_widget_test.go` and `examples/loomoji/loomoji_category_width_pty_test.go` both use the reload pattern.
- Breaking-change row for the read-once semantics goes in `docs/Upgrading.md` as v0.5.8. `version.yaml` is not bumped in this change.
