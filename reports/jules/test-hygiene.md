## Item 480 (implement)

### Verdict
done

### Changes
pane_kitty_test.go: replaced sleep with poll for sequence arrival.
pane_pty_test.go: replaced static sleeps with inline polling for specific dimensions in resize burst tests; removed static sleeps for pulse expiration; commented kept sleeps.
pane_suspend_test.go: commented kept sleeps.

### Sleeps
| path:line | line | result |
| --------- | ---- | ------ |
| pane_kitty_test.go:151 | `time.Sleep(50 * time.Millisecond)` | replaced |
| pane_suspend_test.go:133 | `for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {` | kept: poll loop sleep |
| pane_suspend_test.go:298 | `time.Sleep(100 * time.Millisecond) // keep: let SIGWINCH arrive while suspended` | kept: let SIGWINCH arrive while suspended |
| pane_suspend_test.go:418 | `time.Sleep(50 * time.Millisecond) // keep: let the capture see the last writes` | kept: let the capture see the last writes |
| pane_pty_test.go:323 | `time.Sleep(10 * time.Millisecond) // keep: poll error wait` | kept: poll error wait |
| pane_pty_test.go:541 | `time.Sleep(5 * time.Millisecond) // keep: poll loop sleep` | kept: poll loop sleep |
| pane_pty_test.go:591 | `time.Sleep(SpeccedCursorPressPulse.Lifetime + 30*time.Millisecond)` | replaced |
| pane_pty_test.go:611 | `time.Sleep(SpeccedCursorPressPulse.Lifetime + 30*time.Millisecond)` | replaced |
| pane_pty_test.go:917 | `time.Sleep(10 * time.Millisecond) // keep: pace input for split reads` | kept: pace input for split reads |
| pane_pty_test.go:987 | `time.Sleep(5 * time.Millisecond) // keep: pace input for split reads` | kept: pace input for split reads |
| pane_pty_test.go:1091 | `time.Sleep(20 * time.Millisecond)` | replaced |
| pane_pty_test.go:1102 | `time.Sleep(15 * time.Millisecond) // keep: pace bursts` | kept: pace bursts |
| pane_pty_test.go:1106 | `time.Sleep(80 * time.Millisecond)` | replaced |
| pane_pty_test.go:1153 | `time.Sleep(30 * time.Millisecond)` | replaced |
| pane_pty_test.go:1203 | `time.Sleep(30 * time.Millisecond)` | replaced |
| pane_pty_test.go:1207 | `time.Sleep(50 * time.Millisecond)` | replaced |
| pane_pty_test.go:1247 | `time.Sleep(30 * time.Millisecond)` | replaced |
| pane_pty_test.go:1249 | `time.Sleep(30 * time.Millisecond)` | replaced |
| pane_pty_test.go:1419 | `time.Sleep(30 * time.Millisecond)` | replaced |
| pane_pty_test.go:1425 | `time.Sleep(30 * time.Millisecond)` | replaced |
| pane_pty_test.go:1433 | `time.Sleep(120 * time.Millisecond)` | replaced |
| pane_pty_test.go:1481 | `time.Sleep(30 * time.Millisecond)` | replaced |
| pane_pty_test.go:1485 | `time.Sleep(30 * time.Millisecond)` | replaced |
| pane_pty_test.go:1493 | `time.Sleep(120 * time.Millisecond)` | replaced |
| pane_pty_test.go:1537 | `time.Sleep(25 * time.Millisecond)` | replaced |
| pane_pty_test.go:1541 | `time.Sleep(25 * time.Millisecond)` | replaced |
| pane_pty_test.go:1618 | `time.Sleep(50 * time.Millisecond) // keep: wait for pty to process` | kept: wait for pty to process |
| pane_pty_test.go:1624 | `time.Sleep(50 * time.Millisecond) // keep: wait for pty to process` | kept: wait for pty to process |
| pane_pty_test.go:1662 | `time.Sleep(300 * time.Millisecond) // keep: verify negative condition (never asks)` | kept: verify negative condition (never asks) |

### Tests
`go test -race -count=3 -run 'TestPane' .` (before)
```text
ok  	ubunatic.com/loom	131.462s
```

`go test -race -count=3 -run 'TestPane' .` (after)
```text
ok  	ubunatic.com/loom	109.591s
```

`go test -short ./...`
```text
ok  	ubunatic.com/loom	1.642s
ok  	ubunatic.com/loom/ansi	0.013s
ok  	ubunatic.com/loom/cmd/loom	1.579s
ok  	ubunatic.com/loom/cmd/loom-bench	0.108s
ok  	ubunatic.com/loom/cmd/loom-demo	0.067s
ok  	ubunatic.com/loom/cmd/loom-probe	0.026s
ok  	ubunatic.com/loom/cmd/loom-repaint-probe	0.040s
?   	ubunatic.com/loom/cmd/validate-spec	[no test files]
ok  	ubunatic.com/loom/collector	0.041s
?   	ubunatic.com/loom/examples/ansicanvas_demo	[no test files]
ok  	ubunatic.com/loom/examples/ansicanvas_demo/ansicanvas_demo	0.038s
?   	ubunatic.com/loom/examples/ansiedit	[no test files]
ok  	ubunatic.com/loom/examples/ansiedit/ansiedit	0.042s
?   	ubunatic.com/loom/examples/ansiviewer	[no test files]
ok  	ubunatic.com/loom/examples/ansiviewer/ansiviewer	0.198s
?   	ubunatic.com/loom/examples/background	[no test files]
ok  	ubunatic.com/loom/examples/background/background	0.039s
?   	ubunatic.com/loom/examples/filebrowser	[no test files]
ok  	ubunatic.com/loom/examples/filebrowser/filebrowser	0.103s
ok  	ubunatic.com/loom/examples/loomoji	0.043s
ok  	ubunatic.com/loom/examples/loomoji/loomoji	0.106s
ok  	ubunatic.com/loom/examples/media	0.026s
?   	ubunatic.com/loom/examples/monitor	[no test files]
ok  	ubunatic.com/loom/examples/monitor/monitor	0.092s
?   	ubunatic.com/loom/examples/screens	[no test files]
ok  	ubunatic.com/loom/examples/screens/screens	0.046s
?   	ubunatic.com/loom/examples/splash	[no test files]
ok  	ubunatic.com/loom/examples/splash/splash	0.031s
?   	ubunatic.com/loom/examples/split	[no test files]
ok  	ubunatic.com/loom/examples/split/split	0.027s
?   	ubunatic.com/loom/examples/suspend	[no test files]
ok  	ubunatic.com/loom/examples/suspend/suspend	0.023s
?   	ubunatic.com/loom/examples/tabs	[no test files]
ok  	ubunatic.com/loom/examples/tabs/tabs	0.033s
?   	ubunatic.com/loom/examples/textedit	[no test files]
ok  	ubunatic.com/loom/examples/textedit/textedit	0.106s
?   	ubunatic.com/loom/examples/textrender	[no test files]
ok  	ubunatic.com/loom/examples/textrender/textrender	0.037s
?   	ubunatic.com/loom/examples/treemap	[no test files]
ok  	ubunatic.com/loom/examples/treemap/treemap	0.084s
?   	ubunatic.com/loom/examples/usage	[no test files]
ok  	ubunatic.com/loom/examples/usage/usage	0.071s
?   	ubunatic.com/loom/examples/winch	[no test files]
ok  	ubunatic.com/loom/examples/winch/winch	0.070s
ok  	ubunatic.com/loom/gallery	0.298s
ok  	ubunatic.com/loom/graph	0.093s
ok  	ubunatic.com/loom/internal/apitest	6.293s
ok  	ubunatic.com/loom/internal/boxcheck	0.009s
?   	ubunatic.com/loom/internal/canary/treesitter	[no test files]
ok  	ubunatic.com/loom/internal/clitest	0.031s
ok  	ubunatic.com/loom/internal/editor	0.802s
ok  	ubunatic.com/loom/internal/examplesreg	0.036s
ok  	ubunatic.com/loom/internal/ptytest	0.020s
ok  	ubunatic.com/loom/internal/session	0.191s
ok  	ubunatic.com/loom/internal/specdefaults	5.241s
ok  	ubunatic.com/loom/layout	0.004s
ok  	ubunatic.com/loom/loomtest	0.020s
ok  	ubunatic.com/loom/measure	0.020s
ok  	ubunatic.com/loom/media	0.200s
?   	ubunatic.com/loom/scripts/canary095	[no test files]
?   	ubunatic.com/loom/spec	[no test files]
ok  	ubunatic.com/loom/specdefaults	0.026s
ok  	ubunatic.com/loom/syntax	0.013s
ok  	ubunatic.com/loom/terminal	0.004s
ok  	ubunatic.com/loom/vt	0.014s
```
