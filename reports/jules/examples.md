## Item examples (implement)

### Verdict
done

### Counts
KeyEvent: 280
Style: 193
Rect: 133
EventResult: 77
Canvas: 67
MouseEvent: 67
NewCanvas: 65
Widget: 52
New: 23
PaneRequest: 23

### Changes
example_canvas_test.go
example_eventresult_test.go
example_keyevent_test.go
example_mouseevent_test.go
example_new_test.go
example_newcanvas_test.go
example_panerequest_test.go
example_rect_test.go
example_style_test.go
example_widget_test.go
reports/jules/examples.md

### Suspected bugs
none

### Tests
`go test -run Example -v .`
```
=== RUN   TestMonitorExampleGolden
--- PASS: TestMonitorExampleGolden (0.00s)
=== RUN   ExampleCanvas
--- PASS: ExampleCanvas (0.00s)
=== RUN   ExampleEventResult
--- PASS: ExampleEventResult (0.00s)
=== RUN   ExampleKeyEvent
--- PASS: ExampleKeyEvent (0.00s)
=== RUN   ExampleMouseEvent
--- PASS: ExampleMouseEvent (0.00s)
=== RUN   ExampleNew
--- PASS: ExampleNew (0.00s)
=== RUN   ExampleNewCanvas
--- PASS: ExampleNewCanvas (0.00s)
=== RUN   ExamplePaneRequest
--- PASS: ExamplePaneRequest (0.00s)
=== RUN   ExampleRect
--- PASS: ExampleRect (0.00s)
=== RUN   ExampleStyle
--- PASS: ExampleStyle (0.00s)
=== RUN   ExampleWidget
--- PASS: ExampleWidget (0.00s)
PASS
ok  	ubunatic.com/loom	(cached)
```
`go test -short ./...`
```
All packages passed.
```
