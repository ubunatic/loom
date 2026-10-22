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

## Where bindings live

Widget default bindings are spec data: `spec/defaults.yaml` (`<widget>.bindings`, `tabs.keys`,
`media.bindings` + `labels`) and the table above (`key_conventions`), validated by
`spec/schemas/defaults.schema.json` and loaded once into `SpeccedDefaults`. A missing action panics at
init. Widgets match keys via `KeyMap` or `KeyEvent.Is` on those slices; do not add Go key tables or
`case "key":` defaults (issues 323, 367). `key_088_test.go` keeps this table equal to the spec; the
`internal/specdefaults` probes prove a YAML edit changes key handling.

Terminal limitations: Ctrl-I is delivered as Tab, and Ctrl-M as Enter; terminals
cannot distinguish those pairs at the byte-input layer unless the pane opts in to
the kitty keyboard protocol (`Pane.KittyKeyboard`, see TuiInput.md). Function-key and modified
cursor support depends on the terminal's xterm-compatible sequences.
