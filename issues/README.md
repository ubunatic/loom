<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# loom — issues

Work remaining before the **first tagged release (v0.1.0)** of the extracted
loom module. The extraction (copy, module path, import rewrite, green build) is
already done in this repo; these issues cover the release polish and the
downstream cutover.

| # | File | Title | Status |
|---|---|---|---|
| 001 | [001-first-release-v0.1.0.md](001-first-release-v0.1.0.md) | Cut the first release (v0.1.0) | Blocked on Codeberg repo + tag (user action) |
| 002 | [002-pane-driver-helper.md](002-pane-driver-helper.md) | Ship an exported pane-driver helper | Done in loom |
| 003 | [003-public-api-audit.md](003-public-api-audit.md) | Audit and document the public API surface | Done |
| 004 | [004-migrate-uzu-to-shared-loom.md](004-migrate-uzu-to-shared-loom.md) | Migrate uzu to depend on the shared loom module | Open (after tag; uzu repo) |
| 005 | [005-license-clarification.md](005-license-clarification.md) | Confirm AGPL is the intended license for a shared library | Resolved — keep AGPL |
