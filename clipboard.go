// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"encoding/base64"
	"io"
)

// WriteOSC52 writes an OSC 52 clipboard request containing text to w.
func WriteOSC52(w io.Writer, text string) error {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	_, err := io.WriteString(w, "\x1b]52;c;"+encoded+"\x07")
	return err
}

type clipboardAware interface{ SetClipboardWriter(func(string)) }

func bindClipboardTree(root Widget, write func(string)) {
	if aware, ok := root.(clipboardAware); ok {
		aware.SetClipboardWriter(write)
	}
	switch node := root.(type) {
	case *Tabs:
		for _, tab := range node.Tabs {
			bindClipboardTree(tab.Widget, write)
		}
	case *Stack:
		for _, child := range node.Children {
			bindClipboardTree(child, write)
		}
	case *Grid:
		for _, child := range node.Children {
			bindClipboardTree(child, write)
		}
	case *Frame:
		for _, box := range node.Boxes {
			bindClipboardTree(box.Child, write)
		}
	}
}
