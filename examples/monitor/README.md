# Static Monitor

From the repository root:

```sh
GOWORK=off go run ./examples/monitor
```

This prints a 64-column, nine-row monochrome frame and exits. It works with
redirected stdin/stdout and does not open a terminal, wait for keys, or change
terminal modes. Dimensions come from the document, not terminal detection.

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
come from the library spec. General Unicode/ANSI text, color, watch, resize
reflow, and controls are later tickets. There are no command-line options yet.

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
