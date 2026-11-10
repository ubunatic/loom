## Item 469 (test-draft)
### Verdict
bug reproduced

### Changes
form_field_keys_test.go

### Tests
`go test -v -run TestFormCapturesDownAndEnterInTextArea`
FAIL: TestFormCapturesDownAndEnterInTextArea (0.00s)

### Trace
- `form.go:271`: `func (f *Form) ConsumeKey(e KeyEvent) EventResult {`
- `form.go:276`: `switch {`
- `form.go:290`: `case e.Is(bindings["submit"]...):`
- `form.go:296`: `f.activate()`

## Item 475 (test-draft)
### Verdict
bug reproduced

### Changes
popup_focus_return_test.go

### Tests
`go test -v -run TestPopupFocusReturn`
FAIL: TestPopupFocusReturn (0.00s)

### Trace
- `stack.go:138`: `func (s *Stack) ConsumeKey(e KeyEvent) EventResult {`
- `popup.go:77`: `func (p *Popup) ConsumeKey(e KeyEvent) (quit EventResult) {`
- `popup.go:79`: `if !p.Open || p.Inner == nil {`
- `popup.go:80`: `return Ignored()`
- `stack.go:143`: `if res := s.Children[s.focus].ConsumeKey(e); res.Consumed {`

## Item 467 (test-draft)
### Verdict
bug reproduced

### Changes
text_entry_drag_capture_test.go

### Tests
`go test -v -run TestTextEntrySelectionDragCapture`
FAIL: TestTextEntrySelectionDragCapture (0.00s)

### Trace
- `pane_dispatch.go:35`: `func (p *Pane) dispatchMouse(root Widget, me MouseEvent) EventResult {`
- `text_entry.go:1239`: `func (e *TextEntry) mouseCaptureTarget() (Widget, Rect) {`
- `text_entry.go:1240`: `if e.barDrag.active || e.pan.active {` (dragging not checked)
- `text_entry.go:338`: `func (e *TextEntry) SetFocus(focused bool) {`
- `text_entry.go:339`: `e.focused = focused`
- `text_entry.go:341`: `if e.focused {` (always false if setting to false)

Skipped auto-scroll for item 467.
