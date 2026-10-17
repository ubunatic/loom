# 338 — Restore public graph and collector default constants

**Status**: Closed — resolved in a1b4372
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**:

---

## 1. Problem & Motivation
Ticket 326 moved three public defaults from constants to mutable variables.
That source-level API change appears in gorelease's compatibility report and
breaks callers that use these values in constant expressions.

## 2. Technical Specification / Findings
The declared defaults are `graph.DefaultBackgroundANSI == "100"`,
`graph.MaxWidth == 10`, and `collector.DefaultRetention == 15m`.
Go constants cannot be initialized from embedded YAML, so the exported values
remain constants and package initialization validates that `spec/defaults.yaml`
matches them.

## 3. Implementation & Verification Plan
Restore the three constants, validate their spec values at initialization, and
update the exported API snapshots. No upgrade-guide breaking-change row is
needed because the compatibility break is avoided.

Verification: `go test ./graph ./collector ./internal/apitest` and the full
`go test ./...` suite.
