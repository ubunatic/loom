## Verdict
does not work, reading the env via an atomic snapshot like measure does not improve ParseANSI performance and adds slight overhead of `envSnapshot.Load()`.
## Approach
- Read LOOM_FAST_ANSI once via an atomic snapshot (like measure/env.go): tried it, added `env.go` with `atomic.Pointer[envSwitches]` and `ReloadEnvSwitches()`, changed `useFastANSI()` to return `env().fastANSI`.
## Tests
- `go test -run=^$ -bench=BenchmarkParseANSI -benchmem` before changes: `BenchmarkParseANSI-4 82200 13969 ns/op`
- `go test -run=^$ -bench=BenchmarkParseANSI -benchmem` after changes: `BenchmarkParseANSI-4 80661 14410 ns/op`
## Risks
- Standard tests were not run, so there might be subtle functionality issues with env caching.
- Did not extensively test if any code actively relies on changing `LOOM_FAST_ANSI` during execution.

## Host check (2026-11-01)

Rerun locally, -count=5: experiment 4562-4640 ns/op, base 4639-4801 ns/op; difference is within noise (one Getenv per ParseANSI call, not per byte). Conclusion: no LOOM_FAST_ANSI caching ticket.
