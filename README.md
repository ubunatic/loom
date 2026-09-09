<!--
SPDX-FileCopyrightText: 2026 Uwe Jugel
SPDX-License-Identifier: AGPL-3.0-or-later
-->

# loom

`loom` is a small, dependency-light **inline TUI widget library** for Go. It
renders in-place on the current terminal line (no alternate screen, no
full-screen takeover) and pipes the selected result back to the cursor.

It was extracted from [`uzu`](https://codeberg.org/ubunatic/uzu) so that other
tools — starting with [`uman`](https://codeberg.org/ubunatic/uman) — can share
the same widgets and look/feel.

## Widgets

| Widget | File | Purpose |
|---|---|---|
| `Choice` | `choice.go` | single- or multi-select list (typed filtering) |
| `Table` | `table.go` | sortable multi-column data |
| `Confirm` | `confirm.go` | yes/no guard |
| `TextInput` | `textinput.go` | single-line input |
| `TextArea` | `textarea.go` | multi-line input |
| `Settings` | `settings.go` | key/value settings pane |
| `Pane` / `Canvas` | `pane.go`, `canvas.go` | layout + rendering primitives |

Declarative layouts (`.loom.yaml`) are supported via `yaml.go`
(`BuildWidget` / `ParseYAML`, Router, ASCII grids).

A [static monitor example](examples/monitor/README.md) declares a title bar,
two empty boxes, and a status bar in embedded YAML. Run it without a TTY:

```sh
GOWORK=off go run ./examples/monitor
```

This is the first dashboard slice; watch and responsive wrapping follow later.

## Usage

```go
import "codeberg.org/ubunatic/loom"

choice := loom.NewChoice(items)
choice.Prompt = ":pick> "

// RunPane opens a pane sized to the widget, runs it, and always closes the
// pane (restoring the terminal) before returning the selection.
item, ok, _, err := loom.RunPane(choice)
if err != nil {
	return err
}
if ok {
	fmt.Println(item.Name)
}
```

If you need finer control, drive the pane yourself — always `Close` before
printing to stdout:

```go
pane, err := loom.New(choice.ContentHeight())
if err != nil {
	return err
}
err = pane.Run(choice)
pane.Close() // always close before printing to stdout
item, ok := choice.Selected()
```

## Dependencies

Intentionally minimal:
`golang.org/x/sys`, `golang.org/x/term`, `gopkg.in/yaml.v3`.

## Status

Released as `v0.1.0` and consumed by [`uzu`](https://codeberg.org/ubunatic/uzu)
through the public module proxy. Pre-1.0: the API may still move. See
[`issues/`](issues/) for what is open.

## License

AGPL-3.0-or-later © Uwe Jugel. Every source file carries an SPDX header; see
`REUSE.toml`.
