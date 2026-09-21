package loom

type keyDefault struct {
	Action, Keys, Applies, Conflict string
}

var libraryKeyDefaults = []keyDefault{
	{"quit/close", "q, Q, ctrl-c", "View, Choice, Frame actions", "Popup intentionally closes on esc"},
	{"activate/select", "enter, space", "Choice, View", "space pages View"},
	{"navigation", "up, down, left, right", "Choice, View, Tabs, TextInput", "WASD remains opt-in/historical View only"},
	{"page navigation", "pgup, pgdown", "Choice, View", "Choice also accepts pageup/pagedown aliases"},
	{"editing", "backspace, delete, home, end", "TextInput, TextArea", "undo/redo are not provided by these widgets"},
}
