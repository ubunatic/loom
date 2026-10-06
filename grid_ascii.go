// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"sort"
	"strings"
)

// ParseASCIIGrid parses the ASCII visual layout grid and maps widget keys to layout stacks.
func ParseASCIIGrid(gridStr string, widgets map[string]Widget) (Widget, error) {
	lines := strings.Split(gridStr, "\n")
	// Clean empty lines at start/end
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("empty grid layout")
	}

	// 1. Identify horizontal boundary rows
	var yCoords []int
	for y, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// A line is a horizontal border boundary if it contains only +, -, |, and spaces, and has at least one -
		isBorder := true
		hasDash := false
		for _, r := range trimmed {
			if r == '-' {
				hasDash = true
			} else if r != '+' && r != '|' && r != ' ' {
				isBorder = false
				break
			}
		}
		if isBorder && hasDash {
			yCoords = append(yCoords, y)
		}
	}

	// 2. Identify vertical boundary columns
	xCoordsMap := make(map[int]bool)
	for _, line := range lines {
		for x, r := range line {
			if r == '+' || r == '|' {
				xCoordsMap[x] = true
			}
		}
	}
	var xCoords []int
	for x := range xCoordsMap {
		xCoords = append(xCoords, x)
	}
	sort.Ints(xCoords)

	if len(yCoords) < 2 || len(xCoords) < 2 {
		return nil, fmt.Errorf("invalid grid boundaries: horizontal boundaries=%d, vertical boundaries=%d", len(yCoords), len(xCoords))
	}

	var rowWidgets []Widget

	// 3. Process each row segment
	for r := 0; r < len(yCoords)-1; r++ {
		yStart := yCoords[r]
		yEnd := yCoords[r+1]

		// Find cell columns in this row segment
		var colWidgets []Widget
		cStart := 0

		for cNext := 1; cNext < len(xCoords); cNext++ {
			xBorder := xCoords[cNext]
			// Check if xBorder has a vertical line in all content rows of this segment
			isBorder := true
			for y := yStart + 1; y < yEnd; y++ {
				if y >= len(lines) || xBorder >= len(lines[y]) {
					isBorder = false
					break
				}
				char := lines[y][xBorder]
				if char != '|' && char != '+' {
					isBorder = false
					break
				}
			}

			// If it's a border or the last column, we have a cell
			if isBorder || cNext == len(xCoords)-1 {
				xStart := xCoords[cStart]
				xEnd := xBorder

				// Find the character inside this cell
				key := ""
				for y := yStart + 1; y < yEnd; y++ {
					if y >= len(lines) {
						continue
					}
					for x := xStart + 1; x < xEnd; x++ {
						if x >= len(lines[y]) {
							continue
						}
						char := lines[y][x]
						if char != ' ' && char != '|' && char != '+' && char != '-' {
							key = string(char)
							break
						}
					}
					if key != "" {
						break
					}
				}

				if w, ok := widgets[key]; ok {
					colWidgets = append(colWidgets, w)
				} else if key != "" {
					return nil, fmt.Errorf("grid: unknown element %q at row %d", key, yStart+1)
				}
				cStart = cNext
			}
		}

		if len(colWidgets) == 1 {
			rowWidgets = append(rowWidgets, colWidgets[0])
		} else if len(colWidgets) > 1 {
			rowWidgets = append(rowWidgets, NewStack(Horizontal, colWidgets...))
		}
	}

	if len(rowWidgets) == 0 {
		return nil, fmt.Errorf("no widgets mapped in layout grid")
	}

	if len(rowWidgets) == 1 {
		return rowWidgets[0], nil
	}

	return NewStack(Vertical, rowWidgets...), nil
}
