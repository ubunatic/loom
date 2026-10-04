# 267 — README release status is stale

**Status**: Closed — Updated release status to v0.3.0, canonical module path, and current adoption; make check/install passed
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Documentation
**Related**: `README.md`, `version.yaml`, release `v0.3.0`

---

## 1. Problem & Motivation

The README says the project is released as `v0.1.0`, while the repository's version spec is `v0.3.0` and the release history includes `v0.3.0` (2026-10-03). The README also describes the library as consumed by `uzu`, although the recent module migration records uzu as archived and eight other repositories switched to the new module path. This presents outdated adoption and release information to users evaluating loom.

## 2. Technical Specification / Findings

- `version.yaml` sets `version: 0.3.0`.
- Commit `e07b518` bumps the release to `v0.3.0`.
- `docs/studies/2026-10-01-loom-module-migration.md` records the `ubunatic.com/loom` module path migration and says uzu was archived.
- `README.md` still names `v0.1.0` and uzu in its Status section.

## 3. Implementation & Verification Plan

Update the README status to match the current release and active module path, and describe current adoption without implying that archived uzu tracks this release. Review nearby getting-started claims for the same stale snapshot. Verify the references against `version.yaml`, release history, and the migration report.
