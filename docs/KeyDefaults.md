---
title: Key Defaults
---

# Key Defaults

The following defaults apply to Loom library widgets. Existing application-specific
bindings remain unchanged; conflicts are intentional and listed here.

| Action | Default keys | Applies | Conflicts / limitations |
|---|---|---|---|
| quit/close | `q`, `Q`, `ctrl-c` | View, Choice, Frame actions | Popup intentionally closes on `esc` |
| activate/select | `enter`, `space` | Choice, View | `space` pages View |
| navigation | `up`, `down`, `left`, `right` | Choice, View, Tabs, TextInput | WASD remains opt-in/historical View only |
| page navigation | `pgup`, `pgdown` | Choice, View | Choice also accepts pageup/pagedown aliases |
| editing | `backspace`, `delete`, `home`, `end` | TextInput, TextArea | undo/redo are not provided by these widgets |
| quit/close function | `F10` | Frame actions | not provided as a universal widget default; applications may bind it |
| view | `F3`, `v` | View | not provided as a universal default; v is historical/example-specific |
| edit | `F4`, `e` | TextInput, TextArea | not provided as a universal host action |
| undo/redo | `u`, `U`, `ctrl-z`, `ctrl-r` | TextInput, TextArea | not provided: editors have no undo stack |
| activate/toggle | `insert` | Frame actions | not provided as a universal default; action specs decide conflicts |

Terminal limitations: Ctrl-I is delivered as Tab, and Ctrl-M as Enter; terminals
cannot distinguish those pairs at the byte-input layer. Function-key and modified
cursor support depends on the terminal's xterm-compatible sequences.
