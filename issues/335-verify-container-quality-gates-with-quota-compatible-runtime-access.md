# 335 — Verify container quality gates with quota-compatible runtime access

**Status**: Closed — host ran make container-test under podman: green after lint-tools/.harnez fix (beee70a)
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: #328, #333; Harnez feedback 358b753b

---

## 1. Problem & Motivation
The expanded container gate could not be executed inside the required Harnez Quota-1 sandbox. Host checks are independently verifiable, but container startup fails before image building or tests.

## 2. Technical Specification / Findings
Local Podman fails with `cannot clone: Operation not permitted`. The existing Podman service responds outside Quota-1, but its socket is hidden inside the quota sandbox (`--tmpfs $XDG_RUNTIME_DIR`). Both container-test attempts stopped at runtime startup. No quota/sandbox bypass was used. Harnez feedback 358b753b requests a supported container pathway.

## 3. Implementation & Verification Plan
Once a supported runtime is available, run `make container-test` and verify lint, spec validation, geometry replay, build, API snapshots, and full race tests including PTYs. The new Containerfile bootstraps Python and pinned Harnez via make lint-tools. Confirm the Go 1.25 toolchain and runtime behavior, then close with the verification commit or CI result.
