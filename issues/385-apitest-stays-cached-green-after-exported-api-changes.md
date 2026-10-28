# 385 — apitest stays cached green after exported API changes

**Status**: Closed — Fixed in 2f14c19 (test cache keyed on GoFiles; verified cached->rerun after source edit) and 34ddb95 (terminal snapshot); gokay check --all: fmt/vet/test ok
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 379 (apitest golden files), 377 (test suite)

---

## 1. Problem & Motivation

`internal/apitest` `TestPublicAPI` compares `go doc -all` output against `testdata/api/*.txt`. The
`go doc` and `go list` subprocesses read the package sources, but Go's test cache keys only on
files the test process opens itself. After an exported API change the test therefore keeps
reporting a cached `ok`, and `make test` passes with a stale snapshot.

Found by `gokay check --all` on branch `session-pty-keys-fixes`: commit b4eb249 changed
`terminal.ResetSequences`, `make test-q1` stayed green, and only an uncached run failed:

```
api_test.go:84: exported API differs from testdata/api/terminal.txt
api_test.go:88: first difference at line 24: got "\t\"\\x1b[>4m\" +", want "\t\"\\x1b[r\" +"
```

## 2. Technical Specification / Findings

- `go test` records files opened via `os.Open`/`os.ReadFile` (testlog) and hashes them into the
  cache key; files read by child processes are invisible to it.
- Fix: `go list -json` already returns `Dir` and `GoFiles`; read each package's Go files in the
  subtest before running `go doc`, so a source change invalidates the cached result.

## 3. Implementation & Verification Plan

1. Add `Dir`/`GoFiles` to `listedPackage` and read the files per subtest.
2. Refresh `testdata/api/terminal.txt` for b4eb249.
3. Verify: `go test ./internal/apitest` twice reports `(cached)`; after appending a comment to
   `graph/treemap_layout.go` the next run executes again instead of `(cached)`.
4. `gokay check --all --json` test step passes.
