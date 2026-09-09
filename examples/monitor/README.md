# Monitor

From the repository root:

```sh
go run ./examples/monitor
go run ./examples/monitor --watch
go run ./examples/monitor --help
```

This prints a 64-column, nine-row monochrome frame and exits. It works with
redirected stdin/stdout and does not open a terminal, wait for keys, or change
terminal modes. Dimensions come from the document, not terminal detection.

`--watch` opens an inline pane. The clock is collected at 1 Hz and redrawn at
20 Hz independently of input. Press `q` or Ctrl-C to quit. The embedded
[watch declaration](spec/watch.yaml) sets intervals, clock format, title
template, status hints, and quit keys. Nonpositive intervals are rejected.
Cobra provides CLI parsing/help, as prescribed by the Go conventions.

`Pane.RunWatch` adds owned timers and cancellation to the existing event loop.
Collection and widget callbacks execute serially; drawing only reads the latest
snapshot. Collection must be bounded/nonblocking; external I/O is future work.
Input and resize may repaint but never collect. Missed ticks coalesce. `Run`
remains event-driven. Callers defer `Close`; signals now return through the loop
instead of cleaning up the terminal concurrently. All owned timers stop on exit.

```text
Loom monitor
┌ All Usage ──────────────────┐  ┌ Load ───────────────────────┐
│                             │  │                             │
│                             │  │                             │
│                             │  │                             │
│                             │  │                             │
│                             │  │                             │
└─────────────────────────────┘  └─────────────────────────────┘
Show once
```

## Files

- [main.go](main.go) embeds the document, calls `loom.BuildWidget`, and prints
  `loom.Render` rows. It supplies no UI labels or coordinates.
- [spec/monitor.yaml](spec/monitor.yaml) declares dimensions, chrome, ordered
  boxes, sizes, spacing, and padding. Edit it and rerun to change the UI.
- [monitor.schema.json](../../spec/schemas/monitor.schema.json) covers this
  initial shell vocabulary. The legacy widget vocabulary is not yet schematized.
- Loom's [box spec](../../spec/box.yaml) supplies border glyphs and title
  decoration, with its own [schema](../../spec/schemas/box.schema.json).

The example stays in Loom's root Go module. A consuming application can embed
its own YAML and use the same `BuildWidget` / `Render` calls. Schema checks are
development tooling; the executable needs neither Python nor external files.

## Contract and limits

The frame reserves one title and one status row, placing declared box widths in
sequence with the declared gap. It clips to available space and does not wrap.
At one row, only the title is shown; at zero child area, nothing is drawn. A box
needs at least two rows and columns for a border. Padding is inside that border.
Children draw on isolated canvases, protecting border, padding, and neighbors.

The first shell accepts printable ASCII labels and status text; border glyphs
come from the library spec. General Unicode/ANSI text, color, responsive reflow,
and visibility controls are later tickets. Watch currently clips narrow layouts.

YAML loading now rejects unknown fields, multiple documents, ambiguous root
forms, duplicate IDs, bad dimensions, and unresolved grid/order/root references.
Use exactly one of `pane`, `view`, or `views`. For `views`, an explicit `app.root`
must resolve; otherwise `main`, then the first declared view, is selected.
Single `view` now renders like `pane`. Legacy element maps without a grid use
lexical key order, or an explicit `order: [second, first]` containing every key.
Legacy `on_key` remains accepted but does not gain new routing behavior.

## Verification

```sh
GOWORK=off make test
sh scripts/check-no-tty.sh
setsid --wait env GOWORK=off go run ./examples/monitor </dev/null
```

`make test` validates schemas with Python 3, PyYAML, and jsonschema, then runs Go
vet and tests. Both validator packages were available in the development
environment. `GOWORK=off` isolates Loom from a parent workspace that excludes it.
The standalone canary checks that `setsid` removes the controlling terminal
before the same mechanism is used to smoke-test the example.

The local `go.work` permits commands without the `GOWORK=off` override.
Linux PTY verification uses only Python's standard library:

```sh
python3 scripts/check-watch-pty.py
go build -o /tmp/loom-monitor-007 ./examples/monitor
python3 scripts/check-watch-pty.py /tmp/loom-monitor-007 --watch
LOOM_TEST_NO_DSR=1 python3 scripts/check-watch-pty.py /tmp/loom-monitor-007 --watch
LOOM_TEST_SIGNAL=1 python3 scripts/check-watch-pty.py /tmp/loom-monitor-007 --watch
```

The standalone probe checks raw input/restore before exercising Loom. The watch
smoke checks idle redraws, resize, quit, and terminal-mode restoration. Variants
check missing cursor-position replies and SIGTERM. Cursor queries use bounded
synchronous polling, preventing a timed-out reader from stealing later keys.
