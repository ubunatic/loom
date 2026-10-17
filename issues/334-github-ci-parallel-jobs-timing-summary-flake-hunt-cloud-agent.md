# GitHub CI: parallel jobs, timing summary, flake hunt

Status: In Progress — owned by the Claude cloud agent (ccr branch)

Scope: .github/ only — same design as harnez issue 809.

## Changes
- `.github/gotest-report.go` (shared verbatim with harnez, cati, voxi):
  failures, test time, slowest 10 packages, `-flaky` mode. Replaces
  `scripts/summarize-go-test-json.py` in CI (removal left to the local agent).
- `ci.yaml`: parallel `lint` (gofmt-check, vet, build, validate-spec,
  geometry-replay), `unit` (`go test -short`), `race`, `api`; `report` writes
  the CI wall time. Per-job module and build caches.
- `api` job: `gorelease -base=<latest tag>` into the summary; incompatible
  changes raise a warning (pre-1.0: minor bump plus a docs/Upgrading.md row).
  First run against v0.3.10 found `graph.DefaultBackgroundANSI`,
  `graph.MaxWidth` and `collector.DefaultRetention` changed from const to var.
- `flake.yaml`: nightly and dispatch `go test -race -count=5`; one GitHub issue
  per intermittent test (label `flaky`).

## Verification
- actionlint clean; CI run on the ccr branch shows the summaries.
- flake: first dispatch after merge.
