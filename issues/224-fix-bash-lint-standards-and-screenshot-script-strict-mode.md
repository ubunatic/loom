# 224 — Fix bash lint standards and screenshot script strict mode

**Status**: Closed — fixed bash lint standards and screenshot script strict mode
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Hygiene
**Related**: `scripts/check-no-tty.sh`, `screenshot.go`, `testdata/geometry/`, `docs/Bash.md`

## Goal

Resolve `harnez lint` violations across repository shell scripts and update `loom.ScreenshotScript()` in `screenshot.go` to generate compliant `set -euo pipefail` bash headers for ANSI replay scripts.

## Problem & Motivation

`harnez lint` flags 3 violations across the repository:
1. `scripts/check-no-tty.sh:6:34`: forbidden semicolon before `then` (`[bash-no-semicolon]`).
2. `testdata/geometry/slim.sh:4:1`: missing `set -euo pipefail` strict mode header (`[bash-strict-mode]`).
3. `testdata/geometry/wide.sh:4:1`: missing `set -euo pipefail` strict mode header (`[bash-strict-mode]`).

`testdata/geometry/slim.sh` and `wide.sh` are generated and asserted by `loom.ScreenshotScript()` in `screenshot.go`. Adding `set -euo pipefail` to `ScreenshotScript()` ensures all generated shell replay scripts automatically conform to `docs/Bash.md`.

## Acceptance

- `scripts/check-no-tty.sh` conforms to `docs/Bash.md` (proper shebang `#!/usr/bin/env bash`, `set -euo pipefail`, 3-line `if test`/conditional formatting, `printf` over `echo`).
- `loom.ScreenshotScript()` in `screenshot.go` emits `set -euo pipefail` on line 2.
- `testdata/geometry/slim.sh` and `testdata/geometry/wide.sh` include `set -euo pipefail`.
- `screenshot_test.go` and `geometry_test.go` updated and passing.
- `harnez lint $(git ls-files)` exits 0 with zero warnings/errors.
- `make test` and `make geometry-replay` pass.
