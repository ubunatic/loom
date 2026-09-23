# SPDX-FileCopyrightText: 2026 Uwe Jugel
# SPDX-License-Identifier: AGPL-3.0-or-later

# Source this file in zsh to add a widget that inserts a picked emoji at the
# current command-line cursor. Bind it with, for example:
#   bindkey '^X^M' loomoji-insert-widget
loomoji-insert-widget() {
	local selected
	zle -I
	if ! selected="$(command loomoji)"
	then zle redisplay
	     return 1
	fi
	LBUFFER+="$selected"
	zle redisplay
}

zle -N loomoji-insert-widget
