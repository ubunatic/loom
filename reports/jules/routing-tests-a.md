## Item 462 (test-draft)
### Verdict
bug reproduced

### Changes
`pane_reply_route_test.go` (1 line: new file)

### Tests
`go test -short ./...`
--- FAIL: TestPaneReplyRouting (24.03s)
    --- FAIL: TestPaneReplyRouting/DECRPM (2.00s)
        pane_reply_route_test.go:136: got keys [<report> <empty>], want []
    --- FAIL: TestPaneReplyRouting/DA1 (2.00s)
        pane_reply_route_test.go:136: got keys [<report> <empty>], want []
    --- FAIL: TestPaneReplyRouting/CPR (2.00s)
        pane_reply_route_test.go:136: got keys [<report> <empty>], want []
    --- FAIL: TestPaneReplyRouting/kitty_flags (2.00s)
        pane_reply_route_test.go:136: got keys [<empty> <empty>], want []
    --- FAIL: TestPaneReplyRouting/resize_report (2.00s)
        pane_reply_route_test.go:136: got keys [<empty> <empty>], want []
    --- FAIL: TestPaneReplyRouting/OSC_11_BEL (2.00s)
        pane_reply_route_test.go:136: got keys [alt-] <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> ctrl-g <empty>], want []
    --- FAIL: TestPaneReplyRouting/OSC_11_ST (2.00s)
        pane_reply_route_test.go:136: got keys [alt-] <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> <empty> alt-\ <empty>], want []
    --- FAIL: TestPaneReplyRouting/focus_in (2.00s)
        pane_reply_route_test.go:136: got keys [<empty> <empty>], want []
    --- FAIL: TestPaneReplyRouting/focus_out (2.00s)
        pane_reply_route_test.go:136: got keys [<empty> <empty>], want []
    --- FAIL: TestPaneReplyRouting/997 (2.00s)
        pane_reply_route_test.go:136: got keys [<empty> <empty>], want []
    --- FAIL: TestPaneReplyRouting/interleaved_key (2.00s)
        pane_reply_route_test.go:136: got keys [<report> <empty> <empty>], want [A]
    --- FAIL: TestPaneReplyRouting/split_DECRPM (2.00s)
        pane_reply_route_test.go:136: got keys [<empty> <empty> <empty> <empty> <empty> <empty>], want []
--- FAIL: TestPaneReplyRouting_NormalKey (2.00s)
    pane_reply_route_test.go:204: got keys [<report> <empty> <empty>], want [A]

### Trace
- `pane_run.go:554`: `case rr := <-reads:`
- `pane_run.go:572`: `ke := DecodeKey(pending)`
- `pane_run.go:683`: `ke, used, ok := scanKey(raw)`
- `pane_dispatch.go:10`: `func (p *Pane) dispatchKey(root Widget, ke KeyEvent) bool {`
- `event.go:763`: `return root.ConsumeKey(ke)`

## Item 468 (test-draft)
### Verdict
bug reproduced

### Changes
`event_hwheel_test.go` (1 line: new file)

### Tests
`go test -short ./...`
--- FAIL: TestDecodeMouse_SGRHorizontalWheel (0.00s)
    --- FAIL: TestDecodeMouse_SGRHorizontalWheel/66_(left) (0.00s)
        event_hwheel_test.go:44: DecodeMouse("\x1b[<66;1;1M") = {0 2 false false false 1 1 1 1}; want not a MousePress (should be wheel/scroll action)
        event_hwheel_test.go:47: DecodeMouse("\x1b[<66;1;1M") = {0 2 false false false 1 1 1 1}; want a wheel/scroll action
    --- FAIL: TestDecodeMouse_SGRHorizontalWheel/66_+_shift (0.00s)
        event_hwheel_test.go:44: DecodeMouse("\x1b[<70;1;1M") = {0 2 true false false 1 1 1 1}; want not a MousePress (should be wheel/scroll action)
        event_hwheel_test.go:47: DecodeMouse("\x1b[<70;1;1M") = {0 2 true false false 1 1 1 1}; want a wheel/scroll action
    --- FAIL: TestDecodeMouse_SGRHorizontalWheel/66_+_alt (0.00s)
        event_hwheel_test.go:44: DecodeMouse("\x1b[<74;1;1M") = {0 2 false true false 1 1 1 1}; want not a MousePress (should be wheel/scroll action)
        event_hwheel_test.go:47: DecodeMouse("\x1b[<74;1;1M") = {0 2 false true false 1 1 1 1}; want a wheel/scroll action
    --- FAIL: TestDecodeMouse_SGRHorizontalWheel/66_+_ctrl (0.00s)
        event_hwheel_test.go:44: DecodeMouse("\x1b[<82;1;1M") = {0 2 false false true 1 1 1 1}; want not a MousePress (should be wheel/scroll action)
        event_hwheel_test.go:47: DecodeMouse("\x1b[<82;1;1M") = {0 2 false false true 1 1 1 1}; want a wheel/scroll action
    --- FAIL: TestDecodeMouse_SGRHorizontalWheel/67_(right) (0.00s)
        event_hwheel_test.go:44: DecodeMouse("\x1b[<67;1;1M") = {0 3 false false false 1 1 1 1}; want not a MousePress (should be wheel/scroll action)
        event_hwheel_test.go:47: DecodeMouse("\x1b[<67;1;1M") = {0 3 false false false 1 1 1 1}; want a wheel/scroll action
    --- FAIL: TestDecodeMouse_SGRHorizontalWheel/67_+_shift (0.00s)
        event_hwheel_test.go:44: DecodeMouse("\x1b[<71;1;1M") = {0 3 true false false 1 1 1 1}; want not a MousePress (should be wheel/scroll action)
        event_hwheel_test.go:47: DecodeMouse("\x1b[<71;1;1M") = {0 3 true false false 1 1 1 1}; want a wheel/scroll action
    --- FAIL: TestDecodeMouse_SGRHorizontalWheel/67_+_alt (0.00s)
        event_hwheel_test.go:44: DecodeMouse("\x1b[<75;1;1M") = {0 3 false true false 1 1 1 1}; want not a MousePress (should be wheel/scroll action)
        event_hwheel_test.go:47: DecodeMouse("\x1b[<75;1;1M") = {0 3 false true false 1 1 1 1}; want a wheel/scroll action
    --- FAIL: TestDecodeMouse_SGRHorizontalWheel/67_+_ctrl (0.00s)
        event_hwheel_test.go:44: DecodeMouse("\x1b[<83;1;1M") = {0 3 false false true 1 1 1 1}; want not a MousePress (should be wheel/scroll action)
        event_hwheel_test.go:47: DecodeMouse("\x1b[<83;1;1M") = {0 3 false false true 1 1 1 1}; want a wheel/scroll action
--- FAIL: TestPaneHorizontalWheelRoute (0.00s)
    event_hwheel_test.go:114: horizontal wheel reached widget as press: [{0 2 false false false 44 14 45 15}]

The fix needs `MouseScrollLeft` and `MouseScrollRight` constants added to `MouseAction`.

### Trace
- `pane_run.go:554`: `case rr := <-reads:`
- `pane_run.go:611`: `raw, quit = p.dispatchMouseRead(root, canvas, raw, redraw)`
- `pane_run_mouse.go:13`: `me, used, ok := scanMouse(raw)`
- `event.go:565`: `ev, ok := DecodeMouse(b[:end+1])`
- `event.go:551`: `action = MousePress`
- `event.go:556`: `return makeEvent(action, button), true`
- `pane_run_mouse.go:52`: `res := DispatchMouseEvent(root, me)`
- `event.go:771`: `return root.ConsumeMouse(me)`
