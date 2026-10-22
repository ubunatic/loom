# Testing

Run checks from the repository root. `make lint-tools` installs the pinned Harnez
version for the lint gate; CI and the container image use the same target.

| Command | Purpose |
| --- | --- |
| `make lint` | gofmt check, go vet, and Harnez rules for tracked Go, shell, Markdown and Makefiles |
| `make test-pkg PKG=./loomtest` | One package or pattern, with `-race -count=1`; defaults to the root package |
| `make test-short` | `go test -short ./...`; unit tests and API snapshots, skipping PTY, external media and subprocess integration |
| `make test` | Specs, saved geometry replay, vet, and all package tests |
| `make test-q1` | Full tests with the agent Quota-1 guard |
| `make container-test` | Lint, specs, geometry replay, build, and full race tests under Podman or Docker |
| `make live-check ARGS='...'` | Run installed `loom` in an isolated tmux server; interactive live check, outside test tiers |

The short tier retains in-memory image and widget unit tests. PTY entry points,
binary-building helpers and external media tests check `testing.Short()` before
opening devices, building binaries or starting tools. Add the same guard to new
integration entry points. Full tests probe external tools and PTY availability,
skipping unavailable features with a reason. Containers exercise `/dev/ptmx` tests;
foreground `/dev/tty` tests still require an interactive controlling terminal.

`make lint` checks tracked files, so stage newly added sources before validating.
CI runs the lint gate before building and running the race suite. Container logs
are written to `/tmp/loom-container-test` by default; override `CONTAINER_OUTPUT`,
`ENGINE` or `CONTAINER_CPUS` as needed. If the client environment cannot create
container namespaces, select an existing Podman service with
`ENGINE="podman --remote --url unix://$XDG_RUNTIME_DIR/podman/podman.sock"`.
The current Quota-1 sandbox hides runtime sockets and prevents rootless Podman
startup; supported runtime access is tracked in issue #335. Keep quota sandbox
enforcement enabled when validating agent changes.

## Screen assertions

Import `ubunatic.com/loom/loomtest` to test rendered ANSI without a real terminal:

```go
screen := loomtest.NewVT(40, 10)
fmt.Fprint(screen, "\x1b[2;3H\x1b[1;31mOK")
if got := screen.Cell(2, 1); got.Rune != 'O' || !got.Style.Bold {
    t.Fatalf("expected a bold O at column 2, row 1: %v", got)
}
```

Assert the resulting cells, cursor, attributes and `AltScreen()` state rather than
requiring a particular escape sequence order. `Cells()` and `FrameCells()` return
independent snapshots. `Screen()` trims trailing spaces; use `Cells()` when blanks
or wide-character continuation cells matter. Coordinates are zero-based. The
package docs list the supported terminal sequences and limitations.

`loomtest.CompareANSI(t, reference, width, height, render)` compares every ANSI cell rendered
by a callback with a checked-in `.ansi` reference. Run the package test with
`-update` to deliberately rewrite references. Crop a
design sketch with `loom crop sketch.ansi --region row,col,width,height` before
using it as a reference; keep later design updates in a separate reference file
when the sketch records an earlier variant.
Pass `loomtest.CompareOptions{IgnoreColour: true}` as an optional final argument
for layout-only comparisons.

New concurrency tests should use channels, barriers or fake clocks instead of
wall-clock sleeps. Scale unavoidable timeout budgets for race builds. Use
`LOOM_ZWJ=join` for deterministic joined emoji, and `LOOMCOLOR=truecolor` when a test
asserts colors rather than terminal profile detection.

See [APIStability.md](APIStability.md) for the exported API review/update workflow.

## Installed CLI live checks

Use `make live-check` for a reproducible check of the installed `loom` binary. It
starts a detached tmux server on a unique socket, waits for the initial text,
executes stdin steps, prints the pane after each step, and always kills its server.
Requires `tmux` and the installed `loom` on `PATH`.

```bash
make live-check ARGS='--wait "Ln 1, Col 1" --output-dir /tmp/loom-live -- loom edit --mousegrab=false'
```

Steps are one per line: `keys <tmux key names>`, `hex <two-digit bytes...>`
(for terminal sequences such as Shift+Right), `wait <pane text>`, `pause <seconds>`,
`snap [label]`, or `diff <want.ansi>`. By default, `snap` writes the current SGR
capture into a fresh temporary directory and prints its path. Use `--output-dir`
to choose a persistent output directory. `diff` runs installed `loom diff` against the current SGR
capture and stops on a mismatch. For example, open the action bar, move to 🎨 and
open the color picker, then save and compare its screen:

```bash
printf 'keys C-Space\nkeys Right Right Right Right Enter\nsnap picker\ndiff docs/data/colorpicker-redesign-b.ansi\n' |
  scripts/live-tmux.sh -x 100 -y 24 --wait 'Ln 1, Col 1' -- loom edit --mousegrab=false
```

Use `-e` to print SGR in step output and `--rows 4:7` to restrict printed rows.
`--timeout` and `--interval` set bounded text polling. The script is a manual
live check and does not run as part of `make test`.
