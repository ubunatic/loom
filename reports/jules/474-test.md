## Item 474 (test-draft)
### Verdict
reproduced: The test confirms that a client that stops reading (via `net.Pipe` buffering) causes `writeOutput` to block on `writeMessage`, holding `writeMu`, which subsequently stalls other operations like `detach` that require the same lock.
### Trace
`internal/session/server.go:97`: `return writeMessage(conn, message)`
### Changes
`internal/session/backpressure_test.go`: added failing test reproducing the stall
### Tests
`go test -race -run Backpressure -count=1 ./internal/session/`:
```
--- FAIL: TestBackpressure (8.04s)
    backpressure_test.go:85: server.detach() blocked
FAIL
FAIL	ubunatic.com/loom/internal/session	8.211s
FAIL
```
`go test -short ./...`:
```
ok  	ubunatic.com/loom	1.578s
ok  	ubunatic.com/loom/ansi	0.009s
ok  	ubunatic.com/loom/cmd/loom	1.601s
ok  	ubunatic.com/loom/cmd/loom-bench	0.082s
ok  	ubunatic.com/loom/cmd/loom-demo	0.059s
ok  	ubunatic.com/loom/cmd/loom-probe	0.048s
ok  	ubunatic.com/loom/cmd/loom-repaint-probe	0.035s
?   	ubunatic.com/loom/cmd/validate-spec	[no test files]
ok  	ubunatic.com/loom/collector	0.033s
?   	ubunatic.com/loom/examples/ansicanvas_demo	[no test files]
ok  	ubunatic.com/loom/examples/ansicanvas_demo/ansicanvas_demo	0.048s
?   	ubunatic.com/loom/examples/ansiedit	[no test files]
ok  	ubunatic.com/loom/examples/ansiedit/ansiedit	0.042s
?   	ubunatic.com/loom/examples/ansiviewer	[no test files]
ok  	ubunatic.com/loom/examples/ansiviewer/ansiviewer	0.247s
?   	ubunatic.com/loom/examples/background	[no test files]
ok  	ubunatic.com/loom/examples/background/background	0.040s
?   	ubunatic.com/loom/examples/filebrowser	[no test files]
ok  	ubunatic.com/loom/examples/filebrowser/filebrowser	0.089s
ok  	ubunatic.com/loom/examples/loomoji	0.033s
ok  	ubunatic.com/loom/examples/loomoji/loomoji	0.085s
ok  	ubunatic.com/loom/examples/media	0.037s
?   	ubunatic.com/loom/examples/monitor	[no test files]
ok  	ubunatic.com/loom/examples/monitor/monitor	0.097s
?   	ubunatic.com/loom/examples/screens	[no test files]
ok  	ubunatic.com/loom/examples/screens/screens	0.030s
?   	ubunatic.com/loom/examples/splash	[no test files]
ok  	ubunatic.com/loom/examples/splash/splash	0.027s
?   	ubunatic.com/loom/examples/split	[no test files]
ok  	ubunatic.com/loom/examples/split/split	0.028s
?   	ubunatic.com/loom/examples/suspend	[no test files]
ok  	ubunatic.com/loom/examples/suspend/suspend	0.025s
?   	ubunatic.com/loom/examples/tabs	[no test files]
ok  	ubunatic.com/loom/examples/tabs/tabs	0.026s
?   	ubunatic.com/loom/examples/textedit	[no test files]
ok  	ubunatic.com/loom/examples/textedit/textedit	0.176s
?   	ubunatic.com/loom/examples/textrender	[no test files]
ok  	ubunatic.com/loom/examples/textrender/textrender	0.052s
?   	ubunatic.com/loom/examples/treemap	[no test files]
ok  	ubunatic.com/loom/examples/treemap/treemap	0.080s
?   	ubunatic.com/loom/examples/usage	[no test files]
ok  	ubunatic.com/loom/examples/usage/usage	0.068s
?   	ubunatic.com/loom/examples/winch	[no test files]
ok  	ubunatic.com/loom/examples/winch/winch	0.087s
ok  	ubunatic.com/loom/gallery	0.307s
ok  	ubunatic.com/loom/graph	0.087s
ok  	ubunatic.com/loom/internal/apitest	6.342s
ok  	ubunatic.com/loom/internal/boxcheck	0.008s
?   	ubunatic.com/loom/internal/canary/treesitter	[no test files]
ok  	ubunatic.com/loom/internal/clitest	0.033s
ok  	ubunatic.com/loom/internal/editor	0.659s
ok  	ubunatic.com/loom/internal/examplesreg	0.051s
ok  	ubunatic.com/loom/internal/ptytest	0.011s
--- FAIL: TestBackpressure (2.03s)
    backpressure_test.go:85: server.detach() blocked
FAIL
FAIL	ubunatic.com/loom/internal/session	2.203s
ok  	ubunatic.com/loom/internal/specdefaults	5.115s
ok  	ubunatic.com/loom/layout	0.004s
ok  	ubunatic.com/loom/loomtest	0.009s
ok  	ubunatic.com/loom/measure	0.022s
ok  	ubunatic.com/loom/media	0.182s
?   	ubunatic.com/loom/scripts/canary095	[no test files]
?   	ubunatic.com/loom/spec	[no test files]
ok  	ubunatic.com/loom/specdefaults	0.015s
ok  	ubunatic.com/loom/syntax	0.012s
ok  	ubunatic.com/loom/terminal	0.004s
ok  	ubunatic.com/loom/vt	0.049s
FAIL
```
