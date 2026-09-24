// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	_ "embed"
	"fmt"

	"codeberg.org/ubunatic/loom/measure"
)

//go:embed spec/emoji.yaml
var emojiYAML []byte

func init() {
	if err := measure.LoadEmojiSpecYAML(emojiYAML); err != nil {
		panic(fmt.Sprintf("invalid spec/emoji.yaml: %v", err))
	}
}
