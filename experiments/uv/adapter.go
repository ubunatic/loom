package main

import (
	"bytes"
	"ubunatic.com/loom"
	uv "github.com/charmbracelet/ultraviolet"
)

// canvasToUV copies a loom Canvas into an ultraviolet TerminalScreen.
func canvasToUV(c *loom.Canvas, out *bytes.Buffer) *uv.TerminalScreen {
	cols, rows := c.Cols(), c.Rows()

	scr := uv.NewTerminalScreen(out, nil)
	scr.Resize(cols, rows)

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			loomCell := c.Get(x, y)

			if loomCell.Continuation {
				continue
			}

			text := loomCell.Text
			if text == "" {
				text = " "
			}

			uvCell := uv.NewCell(scr.WidthMethod(), text)

			loomStyle := loomCell.Style
			var uvStyle uv.Style

			if loomStyle.Bold {
				uvStyle.Attrs |= uv.AttrBold
			}
			if loomStyle.Italic {
				uvStyle.Attrs |= uv.AttrItalic
			}
			if loomStyle.Underline {
				uvStyle.Underline = uv.UnderlineSingle
			}
			if loomStyle.Strike {
				uvStyle.Attrs |= uv.AttrStrikethrough
			}
			if loomStyle.Invert {
				uvStyle.Attrs |= uv.AttrReverse
			}

			uvCell.Style = uvStyle
			scr.SetCell(x, y, uvCell)
		}
	}

	return scr
}
