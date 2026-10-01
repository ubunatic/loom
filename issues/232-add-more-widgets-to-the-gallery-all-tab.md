# 232 — Add more widgets to the gallery All tab

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 229 (All tab, 3-column grid)

---

/goal Show every gallery demo widget that fits a grid cell on the "All" tab of the loom widget gallery,
with tests passing, or stop and report when blocked on a user decision (e.g. which large widgets to leave out).

## 1. Problem & Motivation
The "All" tab (`newAllDemo` in `gallery/gallery.go`, issue 229) shows only 12 of the 26 gallery
demos: Button, Toggle, Checkbox, NumberInput, Badge, PillCluster, ProgressBar, Sparkline, Spinner,
Stopwatch, Timer, Paginator. The user wants more widgets there so the overview is representative.

## 2. Technical Specification / Findings
Demos not yet on the All tab (as of commit 641bd44): Chart, DatePicker, Choice, Dialog, FilePicker,
Form, KeyHelp, MenuBar, Media, PaintCanvas, Popup, Table, Tabs, TextArea, TextInput, Tree, Viewport.

- Compact candidates (one grid cell): TextInput, Choice, DatePicker, KeyHelp, MenuBar, small Chart/Table/Tree.
- Large or modal widgets (Dialog, Popup, FilePicker, Media, PaintCanvas, Form, TextArea, Viewport)
  may need a reduced size, a taller row, or be left out; record the choice in the ticket.
- Focus and mouse routing on the All tab must keep working for every added widget (see 229, 199, 204).

Re-check the live demo list in `gallery/gallery.go` before starting.

## 3. Implementation & Verification Plan
- Extend `newAllDemo` (grid may grow more rows); keep 3 columns.
- Extend `gallery/gallery_test.go`: each added widget draws inside its cell and receives focus/keys/mouse.
- Run `make test-q1`, `make install`, and check `loom widgets --show` manually.
