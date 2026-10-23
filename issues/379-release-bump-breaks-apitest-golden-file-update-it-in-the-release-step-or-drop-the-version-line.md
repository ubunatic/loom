# 379 — Release bump breaks apitest golden file: update it in the release step or drop the version line

**Status**: Closed — normalized Version from api snapshots in internal/apitest (commit 0f81580)
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: 377 (test suite roadmap); release v0.5.0 and v0.5.1 both needed a golden-file fix

---

## 1. Problem & Motivation
Every `harnez release` bumps `version.go`, but `testdata/api/loom.txt` records the old version line
(`var Version = "0.x.y"`). `internal/apitest` then fails until someone runs `go test ./internal/apitest
-update`. This happened after v0.5.0 (fixed in `ca386ef`) and again after v0.5.1 (fixed in `0083611`).
The failing test also looks like a real API break, which wastes time.

## 2. Technical Specification / Findings
- `internal/apitest` compares the exported API against `testdata/api/loom.txt`, which includes the
  `Version` constant.
- The release step (`harnez release`) changes `version.go` and tags, but it does not touch the golden
  file.
- Option A: have the release step run `go test ./internal/apitest -update` and commit the golden file
  with the version bump.
- Option B: keep the version out of the golden file, so only real API changes fail the test.

## 3. Implementation & Verification Plan
1. Choose A or B (B is simpler and removes the release-time edit).
2. Implement it in the test or in the release step.
3. Verify: `go test -count=1 ./internal/apitest` passes right after a release bump, and still fails on a
   real exported-API change.
