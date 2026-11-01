---
kind: sprint
date: 2026-11-01
ticket: 403
---

# Lean Sprint 403: read measure env switches once

- **Setup**: one developer, `d403` on haiku:med. Searcher skipped: the ticket named the files and lines; a read-only plan turn verified them instead.
- **Plan turn**: found hot reads in `useFastMeasure`, `ZWJClusterWidth` (LOOM_ZWJ) and `ActiveRenderPath` (DetectRenderPath); listed ~10 in-process `t.Setenv` tests needing a reload.
- **M1 (env snapshot)**: `measure/env.go` atomic snapshot + exported `ReloadEnvSwitches`; Upgrading v0.5.8 row. First `make test-q1` red: `TestLoomojiCategoryLineWidthsPTY` set LOOM_ZWJ without reload, so the parent's PTY emulator measured with stale widths. Dev stopped for approval instead of editing an unlisted file; fix was the same Cleanup-before-Setenv pattern. Second run green; commit 1e44167, closed in 213bfbb.
- **Result**: StringWidth ASCII 36.8 → 18.9 ns/op; ActiveRenderPath 64.5 → 2.3 ns/op; ZWJ family 406 → 357 ns/op; 0 allocs. Host reran measure, loomoji, textrender with `-count=1`: ok.
- **Cost**: three dev turns; the implementation turn used ~2.3M tokens incl. cache in 5m40s.
- **Learning**: caching env makes `t.Setenv` silently ineffective; the safeguard is the documented pattern in `docs/Testing.md`. `LOOM_FAST_ANSI` (`parse_ansi.go:17`) still reads per call — a candidate ticket.
