package session

import (
	"testing"
	"ubunatic.com/loom/specdefaults"
)

func TestFormatTabLabel(t *testing.T) {
	defs := specdefaults.SessionDefaults{
		Emoji: map[string]string{
			"bash": "🐚",
			"loom": "🧶",
		},
		BasenameWidth: 15,
	}

	tests := []struct {
		tabIndex int
		baseCmd  string
		fgCmd    string
		cwd      string
		expected string
	}{
		{0, "bash", "bash", "/home/user", "¹🐚user"},
		{1, "bash", "make", "/home/user/project", "²make·project"},
		{9, "bash", "loom", "/home/user/workspace/loom", "¹⁰🧶loom"},
		{0, "bash", "", "", "¹🐚bash"},
	}

	for _, test := range tests {
		result := formatTabLabel(test.tabIndex, test.baseCmd, test.fgCmd, test.cwd, defs)
		if result != test.expected {
			t.Errorf("formatTabLabel(%d, %q, %q, %q) = %q, expected %q", test.tabIndex, test.baseCmd, test.fgCmd, test.cwd, result, test.expected)
		}
	}
}
