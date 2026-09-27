# 159 — Set up Go release pipeline for loom and CLI tools

**Status**: Closed — Implemented in 6d10dd5 (M1): configure .goreleaser.yaml, version.yaml, spec validation, Cobra version wiring, Makefile check and release targets, and minisign key
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature

---

## 1. Problem & Motivation

Loom now includes the `loom` CLI tool (`cmd/loom`) alongside helper CLI binaries (`cmd/loom-probe`, `cmd/loom-demo`, `cmd/loom-bench`, etc.) in addition to the core TUI library. We need a standard release pipeline using `harnez release` + `goreleaser` + `minisign` + `version.yaml` adhering to `@docs/GoRelease.md` and matching peer repositories (e.g. `harnez`, `voxi`).

## 2. Technical Specification / Findings

1. `version.yaml` at repo root with version `0.2.9` matching `version.go`.
2. `spec/schemas/version.schema.json` schema and registered validation in `cmd/validate-spec`.
3. Cobra root command wiring `root.Version = loom.Version` in `cmd/loom/main.go`.
4. `.goreleaser.yaml` (GoReleaser v2) configured for `loom` (and helper tools), building linux binaries (amd64, arm64) with `CGO_ENABLED=0` and ldflags `-s -w -X codeberg.org/ubunatic/loom.Version={{.Version}}`.
5. Minisign key generated at `~/.minisign/loom.key` / `~/.minisign/loom.pub`.
6. `Makefile` updated with `check` and `release: check ⚙️` targets.

## 3. Implementation & Verification Plan

- **M1**: Create `version.yaml`, `spec/schemas/version.schema.json`, wire `cmd/validate-spec`, wire `root.Version = loom.Version` in `cmd/loom/main.go`, add `.goreleaser.yaml`, and add `check` / `release` targets to `Makefile`.
- **Verification**: `go run ./cmd/validate-spec`, `go test ./...`, `goreleaser check`, `harnez release --dry-run` or dry validation.
