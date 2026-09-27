# Usage

Run from the repository root:

```sh
go run ./examples/usage
go run ./examples/usage --help
```

The example rebuilds the compact All Usage and Load watch as Loom widgets.
All Usage comes from deterministic example quota data; Load samples local CPU
and memory counters from procfs. Collection runs asynchronously, while Loom's
pane redraw cadence remains independent. Press `q` to exit.
