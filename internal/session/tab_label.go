package session

import (
	"os"
	"path/filepath"

	"ubunatic.com/loom"
	"ubunatic.com/loom/specdefaults"
)

var superscriptDigits = []rune{'⁰', '¹', '²', '³', '⁴', '⁵', '⁶', '⁷', '⁸', '⁹'}

func formatTabLabel(tabIndex int, baseCmd string, fgCmd, cwd string, defs specdefaults.SessionDefaults) string {
	tabNum := tabIndex + 1
	var numStr string
	for n := tabNum; n > 0; n /= 10 {
		numStr = string(superscriptDigits[n%10]) + numStr
	}

	cmd := baseCmd
	if fgCmd != "" {
		cmd = fgCmd
	}

	emoji, ok := defs.Emoji[cmd]
	var cmdStr string
	if ok {
		cmdStr = emoji
	} else {
		cmdStr = cmd + "·"
	}

	cwdStr := ""
	if cwd != "" {
		home, err := os.UserHomeDir()
		if err == nil && home != "" && cwd == home {
			cwdStr = "~"
		} else {
			base := filepath.Base(cwd)
			cwdStr = loom.TruncateText(base, defs.BasenameWidth, "…")
		}
	} else {
		cwdStr = baseCmd
	}

	return numStr + cmdStr + cwdStr
}
