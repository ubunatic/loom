// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansiedit

import "codeberg.org/ubunatic/loom"

// EditMode represents character insertion vs overtype behavior.
type EditMode = loom.AnsiEditMode

const (
	ModeOvertype = loom.AnsiModeOvertype
	ModeInsert   = loom.AnsiModeInsert
)

// BufferCell represents a single styled character cell in the ANSI canvas buffer.
type BufferCell = loom.AnsiCell

// BlankCell returns a default empty cell.
func BlankCell() BufferCell {
	return loom.BlankAnsiCell()
}

// Buffer is an in-memory 2D cell grid for ANSI art editing.
type Buffer = loom.AnsiBuffer

// NewBuffer creates an empty ANSI buffer with the specified columns and rows.
func NewBuffer(cols, rows int) *Buffer {
	return loom.NewAnsiBuffer(cols, rows)
}

// LoadBuffer reads an ANSI file and parses it into a 2D Buffer.
func LoadBuffer(path string, defaultCols, defaultRows int) (*Buffer, error) {
	return loom.LoadAnsiBuffer(path, defaultCols, defaultRows)
}

// ParseBuffer parses an ANSI string into a 2D Buffer.
func ParseBuffer(content string, minCols, minRows int) (*Buffer, error) {
	return loom.ParseAnsiBuffer(content, minCols, minRows)
}
