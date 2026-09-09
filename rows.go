// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "fmt"

// RowColumn reserves a fixed number of terminal cells for each value.
type RowColumn struct {
	Width int    `yaml:"width"`
	Align string `yaml:"align"`
	Bold  bool   `yaml:"bold"`
}

// Rows renders fixed-width, declared columns. Leftmost columns take priority
// when space is short; trailing columns are clipped or omitted. Values do not
// influence column positions. Ellipsis is supplied by the declaration.
type Rows struct {
	Columns  []RowColumn `yaml:"columns"`
	Values   [][]string  `yaml:"values"`
	Gap      int         `yaml:"gap"`
	Ellipsis string      `yaml:"ellipsis"`
}

// Draw clips all values to their column and the child's isolated canvas.
func (rows *Rows) Draw(c *Canvas, r Rect) {
	paintClipped(c, r, func(local *Canvas) {
		for y, values := range rows.Values {
			if y >= local.Rows() {
				break
			}
			x := 0
			for i, col := range rows.Columns {
				width := min(max(0, col.Width), local.Cols()-x)
				if width <= 0 {
					break
				}
				value := ""
				if i < len(values) {
					value = values[i]
				}
				text := ""
				offset := 0
				if col.Align == "right" {
					text = TruncateTextLeft(value, width, rows.Ellipsis)
					offset = width - StringWidth(text)
				} else {
					text = TruncateText(value, width, rows.Ellipsis)
				}
				local.Write(x+offset, y, text, Style{Bold: col.Bold})
				x += width
				x += min(max(0, rows.Gap), local.Cols()-x)
			}
		}
	})
}

// SetValues updates row values dynamically from Go application code.
func (rows *Rows) SetValues(values [][]string) {
	rows.Values = values
}

// GetValues returns the active row values.
func (rows *Rows) GetValues() [][]string {
	return rows.Values
}

// HandleKey keeps static rows inert.
func (*Rows) HandleKey(KeyEvent) bool { return false }

// HandleMouse keeps static rows inert.
func (*Rows) HandleMouse(MouseEvent) bool { return false }

func (rows *Rows) validate() error {
	if len(rows.Columns) == 0 || rows.Gap < 0 {
		return fmt.Errorf("rows: columns required and gap must be nonnegative")
	}
	for _, col := range rows.Columns {
		if col.Width < 1 || (col.Align != "" && col.Align != "left" && col.Align != "right") {
			return fmt.Errorf("rows: positive column width and left/right alignment required")
		}
	}
	for i, values := range rows.Values {
		if len(values) != len(rows.Columns) {
			return fmt.Errorf("rows.values[%d]: expected %d columns", i, len(rows.Columns))
		}
	}
	return nil
}
