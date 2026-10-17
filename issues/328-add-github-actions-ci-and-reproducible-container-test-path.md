# 328 — Add GitHub Actions CI and reproducible container test path

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**: `Makefile`, `go.mod`, `.github/workflows/ci.yaml`

---

## 1. Problem & Motivation
The repository has no GitHub Actions workflow or containerized test command. Contributors cannot run the same validation locally that CI will enforce, and terminal integration tests can require a real PTY or foreground TTY.

## 2. Technical Specification / Findings
Use Go 1.25 from `go.mod`, run formatting, vet, build, and tests in CI, and run the race detector on Linux. The local container path must run the same checks with only its output directory writable on the host. Container-mode tests must skip tests that need a real terminal.

PTY dependent test files: `gallery/all_parity_test.go`, `gallery/gallery_test.go`, `cmd/loom/info_pty_test.go`, `cmd/loom/edit_pty_test.go`, `cmd/loom/edit_303_test.go`, `cmd/loom/edit_302_test.go`, `examples/media/media_pty_test.go`, `pane_pty_test.go`, `examples/suspend/suspend/suspend_pty_test.go`, `examples/ansicanvas_demo/ansicanvas_demo/ansicanvas_demo_pty_test.go`, `examples/ansiviewer/ansiviewer/viewer_pty_test.go`, `examples/loomoji/loomoji_category_width_pty_test.go`, `overlay_mouse_pty_test.go`, `examples/loomoji/loomoji_107_pty_test.go`, `examples/split/split/split_pty_test.go`, `examples/winch/winch/winch_pty_test.go`, `examples/screens/screens/screens_pty_test.go`, `examples/monitor/monitor/monitor_test.go`, `examples/textrender/textrender/textrender_test.go`, `internal/examplesreg/smoke_test.go`, `examples/filebrowser/filebrowser/browser_093_pty_test.go`, `examples/ansiedit/ansiedit/ansiedit_pty_test.go`, and `examples/usage/usage/usage_test.go`. The foreground TTY tests in `startup_test.go` (`TestStartupConfigValidation`, `TestStartupLifecycleProgressionAndHandover`, `TestStartupTransitionError`, `TestStartupDismissalImmediateHandover`) also skip in container mode. The tracked-path test keeps `.git` in the test image. No audio-device, GPU, or network dependent tests were found.

## 3. Implementation & Verification Plan
**Goal**: Add reliable GitHub Actions CI and a reproducible local Podman/Docker test path with matching checks; verify both, or stop and report when blocked on a user decision or denied permission.

The gofmt gate also requires formatting six existing files: `syntax/syntax.go`, `syntax/engine.go`, `cmd/loom/main.go`, `measure/spec.go`, `examples/ansiviewer/ansiviewer/viewer.go`, and `ansibuffer.go`.

Acceptance: CI checks out LFS when applicable, configures and caches Go from `go.mod`, runs gofmt, vet, build, and tests with Linux race detection, and cancels superseded runs per ref. A container recipe and `make container-test` use Podman when available and otherwise Docker, isolate host writes to an output directory, and set a flag that skips the listed PTY/TTY tests. Verify with `make container-test` using Podman, validate YAML with actionlint or PyYAML, and commit without pushing.
