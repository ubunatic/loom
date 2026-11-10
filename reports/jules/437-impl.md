## Item 437 (implement)
### Verdict
done

### Changes
internal/ptytest/view/attach.go
internal/ptytest/view/attach_test.go
internal/ptytest/view/frame.go
internal/ptytest/view/frame_test.go
internal/ptytest/view/relay.go
internal/ptytest/view/relay_test.go

### Tests
`go test -race -count=3 ./internal/ptytest/...`
```
ok  	ubunatic.com/loom/internal/ptytest	2.964s
ok  	ubunatic.com/loom/internal/ptytest/view	1.310s
```

`go test -short ./...`
```
ok  	ubunatic.com/loom	2.014s
ok  	ubunatic.com/loom/ansi	0.010s
ok  	ubunatic.com/loom/cmd/loom	1.969s
ok  	ubunatic.com/loom/cmd/loom-bench	0.095s
ok  	ubunatic.com/loom/cmd/loom-demo	0.095s
ok  	ubunatic.com/loom/cmd/loom-probe	0.037s
ok  	ubunatic.com/loom/cmd/loom-repaint-probe	0.062s
?   	ubunatic.com/loom/cmd/validate-spec	[no test files]
ok  	ubunatic.com/loom/collector	0.060s
?   	ubunatic.com/loom/examples/ansicanvas_demo	[no test files]
ok  	ubunatic.com/loom/examples/ansicanvas_demo/ansicanvas_demo	0.041s
?   	ubunatic.com/loom/examples/ansiedit	[no test files]
ok  	ubunatic.com/loom/examples/ansiedit/ansiedit	0.056s
?   	ubunatic.com/loom/examples/ansiviewer	[no test files]
ok  	ubunatic.com/loom/examples/ansiviewer/ansiviewer	0.229s
?   	ubunatic.com/loom/examples/background	[no test files]
ok  	ubunatic.com/loom/examples/background/background	0.049s
?   	ubunatic.com/loom/examples/filebrowser	[no test files]
ok  	ubunatic.com/loom/examples/filebrowser/filebrowser	0.124s
ok  	ubunatic.com/loom/examples/loomoji	0.032s
ok  	ubunatic.com/loom/examples/loomoji/loomoji	0.129s
ok  	ubunatic.com/loom/examples/media	0.038s
?   	ubunatic.com/loom/examples/monitor	[no test files]
ok  	ubunatic.com/loom/examples/monitor/monitor	0.242s
?   	ubunatic.com/loom/examples/screens	[no test files]
ok  	ubunatic.com/loom/examples/screens/screens	0.073s
?   	ubunatic.com/loom/examples/splash	[no test files]
ok  	ubunatic.com/loom/examples/splash/splash	0.053s
?   	ubunatic.com/loom/examples/split	[no test files]
ok  	ubunatic.com/loom/examples/split/split	0.068s
?   	ubunatic.com/loom/examples/suspend	[no test files]
ok  	ubunatic.com/loom/examples/suspend/suspend	0.041s
?   	ubunatic.com/loom/examples/tabs	[no test files]
ok  	ubunatic.com/loom/examples/tabs/tabs	0.055s
?   	ubunatic.com/loom/examples/textedit	[no test files]
ok  	ubunatic.com/loom/examples/textedit/textedit	0.159s
?   	ubunatic.com/loom/examples/textrender	[no test files]
ok  	ubunatic.com/loom/examples/textrender/textrender	0.108s
?   	ubunatic.com/loom/examples/treemap	[no test files]
ok  	ubunatic.com/loom/examples/treemap/treemap	0.121s
?   	ubunatic.com/loom/examples/usage	[no test files]
ok  	ubunatic.com/loom/examples/usage/usage	0.136s
?   	ubunatic.com/loom/examples/winch	[no test files]
ok  	ubunatic.com/loom/examples/winch/winch	0.083s
ok  	ubunatic.com/loom/gallery	0.382s
ok  	ubunatic.com/loom/graph	0.118s
ok  	ubunatic.com/loom/internal/apitest	8.132s
ok  	ubunatic.com/loom/internal/boxcheck	0.013s
?   	ubunatic.com/loom/internal/canary/treesitter	[no test files]
ok  	ubunatic.com/loom/internal/clitest	0.042s
ok  	ubunatic.com/loom/internal/editor	0.760s
ok  	ubunatic.com/loom/internal/examplesreg	0.046s
ok  	ubunatic.com/loom/internal/ptytest	0.008s
ok  	ubunatic.com/loom/internal/ptytest/view	0.050s
ok  	ubunatic.com/loom/internal/session	0.239s
ok  	ubunatic.com/loom/internal/specdefaults	5.699s
ok  	ubunatic.com/loom/layout	0.004s
ok  	ubunatic.com/loom/loomtest	0.017s
ok  	ubunatic.com/loom/measure	0.033s
ok  	ubunatic.com/loom/media	0.191s
?   	ubunatic.com/loom/scripts/canary095	[no test files]
?   	ubunatic.com/loom/spec	[no test files]
ok  	ubunatic.com/loom/specdefaults	0.066s
ok  	ubunatic.com/loom/syntax	0.014s
ok  	ubunatic.com/loom/terminal	0.004s
ok  	ubunatic.com/loom/vt	0.024s
```