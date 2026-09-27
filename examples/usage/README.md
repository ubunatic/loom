# Usage

Run from the repository root:

```sh
go run ./examples/usage
go run ./examples/usage --help
go run ./examples/usage --view=plain
```

The example rebuilds the compact All Usage and Load watch as Loom widgets.
All Usage comes from deterministic example quota data; Load samples local CPU
and memory counters from procfs. Collection runs asynchronously, while Loom's
pane redraw cadence remains independent. Start with `--view=plain` to use the
ANSI row view; press `v` to switch between it and the Loom Frame view, and `q`
to exit.
