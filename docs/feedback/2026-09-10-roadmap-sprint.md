<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# Roadmap Sprint Retrospective

This sprint advanced the first three Now items without claiming to close them.

## Delivered

- Added `measure`, a renderer-independent facade for Loom's terminal-cell
  policy, with multiline control handling and truncation canaries.
- Synchronized collector history publication and snapshot reads, verified under
  the race detector.
- Kept the monitor's fixed geometry and simulated/real provenance boundaries
  unchanged while the larger contracts remain open.

## Learning

The first measurement facade imported the renderer package, which would have
prevented the renderer from consuming it later. The review gate caught this
before integration. Shared terminal policy must live below rendering, with
compatibility wrappers at the root. The next sprint should define the YAML
contract for dynamic sizing before changing `Frame.Layout`, and define the
numeric file format and stale/error semantics before wiring Issue 027 into the
display.

## Verification

`make test`, `go test -count=1 ./...`, `go test -race -count=1 ./...`, and
`go vet ./...` pass. Issues 013, 027, and 028 remain Open.
