// SPDX-FileCopyrightText: 2026 Jules
// SPDX-License-Identifier: AGPL-3.0-or-later

package declarative

import (
	"fmt"
	"strings"
	"testing"

	"ubunatic.com/loom"
)

func lines(label string) []string {
	rows := make([]string, 40)
	for i := range rows {
		rows[i] = fmt.Sprintf("%s line %02d", label, i+1)
	}
	return rows
}

func viewLines() []string {
	rows := make([]string, 30)
	for i := range rows {
		rows[i] = fmt.Sprintf("log line %02d", i+1)
	}
	return rows
}

func choiceItems() []loom.Item {
	return []loom.Item{
		{Name: "alpha", Desc: "first item"},
		{Name: "beta", Desc: "second item"},
		{Name: "gamma", Desc: "third item"},
	}
}

func tableData() ([]loom.Column, []loom.Row) {
	cols := []loom.Column{{Header: "Name"}, {Header: "Size"}}
	rows := []loom.Row{
		{Cells: []string{"main.go", "1.2K"}},
		{Cells: []string{"tabs.go", "3.4K"}},
		{Cells: []string{"README.md", "512B"}},
	}
	return cols, rows
}

func canvasToString(c *loom.Canvas) string {
	var sb strings.Builder
	for y := 0; y < c.Rows(); y++ {
		sb.WriteString(c.Row(y))
		sb.WriteString("\n")
	}
	return sb.String()
}

func TestDeclarativeSplit(t *testing.T) {
	// Hand-written tree
	left := loom.NewView(lines("Left"))
	left.Style = loom.Style{Dim: true}
	left.FocusStyle = loom.Style{Bold: true}

	topRight := loom.NewView(lines("Top-Right"))
	topRight.Style = loom.Style{Dim: true}
	topRight.FocusStyle = loom.Style{Bold: true}

	bottomRight := loom.NewView(lines("Bottom-Right"))
	bottomRight.Style = loom.Style{Dim: true}
	bottomRight.FocusStyle = loom.Style{Bold: true}

	vSplit := loom.NewSplit(topRight, bottomRight)
	vSplit.Orientation = loom.Vertical
	vSplit.Ratio = 0.5
	vSplit.MinFirst = 3
	vSplit.MinSecond = 3
	vSplit.Divider = loom.DividerStyle{Glyph: "─", Style: loom.Style{Dim: true}}

	hSplit := loom.NewSplit(left, vSplit)
	hSplit.Orientation = loom.Horizontal
	hSplit.Ratio = 0.4
	hSplit.MinFirst = 12
	hSplit.MinSecond = 16
	hSplit.Divider = loom.DividerStyle{Glyph: "│", Style: loom.Style{Dim: true}}

	// Declarative tree (JSON)
	jsonTree := `
{
  "type": "split",
  "props": {
    "orientation": "horizontal",
    "ratio": 0.4,
    "min_first": 12,
    "min_second": 16,
    "divider": "│"
  },
  "children": [
    {
      "type": "view",
      "props": {
        "lines": ` + toJSONList(lines("Left")) + `,
        "dim": true,
        "bold_focus": true
      }
    },
    {
      "type": "split",
      "props": {
        "orientation": "vertical",
        "ratio": 0.5,
        "min_first": 3,
        "min_second": 3,
        "divider": "─"
      },
      "children": [
        {
          "type": "view",
          "props": {
            "lines": ` + toJSONList(lines("Top-Right")) + `,
            "dim": true,
            "bold_focus": true
          }
        },
        {
          "type": "view",
          "props": {
            "lines": ` + toJSONList(lines("Bottom-Right")) + `,
            "dim": true,
            "bold_focus": true
          }
        }
      ]
    }
  ]
}
`
	node, err := ParseJSON([]byte(jsonTree))
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}
	declTree, err := Build(node)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Compare outputs
	w, h := 80, 24
	r := loom.Rect{W: w, H: h}

	c1 := loom.NewCanvas(w, h)
	hSplit.Draw(c1, r)

	c2 := loom.NewCanvas(w, h)
	declTree.Draw(c2, r)

	out1 := canvasToString(c1)
	out2 := canvasToString(c2)

	if out1 != out2 {
		t.Errorf("Outputs mismatch.\nExpected:\n%s\n\nGot:\n%s", out1, out2)
	}
}

func TestDeclarativeTabs(t *testing.T) {
	// Hand-written tree
	cols, rows := tableData()
	root := loom.NewTabs(
		loom.Tab{Title: "Log", Widget: loom.NewView(viewLines())},
		loom.Tab{Title: "Choices", Widget: loom.NewChoice(choiceItems())},
		loom.Tab{Title: "Files", Widget: loom.NewTable(cols, rows)},
	)
	root.Keys = loom.TabsKeys{
		Previous: "left",
		Next:     "right",
		Cycle:    "ctrl-t",
		Select:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
	}

	// Declarative tree (JSON)
	jsonTree := `
{
  "type": "tabs",
  "props": {
    "keys": {
      "previous": "left",
      "next": "right",
      "cycle": "ctrl-t",
      "select": ["1", "2", "3", "4", "5", "6", "7", "8", "9"]
    }
  },
  "children": [
    {
      "type": "tab",
      "props": {"title": "Log"},
      "children": [
        {
          "type": "view",
          "props": {"lines": ` + toJSONList(viewLines()) + `}
        }
      ]
    },
    {
      "type": "tab",
      "props": {"title": "Choices"},
      "children": [
        {
          "type": "choice",
          "props": {
            "items": [
              {"name": "alpha", "desc": "first item"},
              {"name": "beta", "desc": "second item"},
              {"name": "gamma", "desc": "third item"}
            ]
          }
        }
      ]
    },
    {
      "type": "tab",
      "props": {"title": "Files"},
      "children": [
        {
          "type": "table",
          "props": {
            "cols": [
              {"header": "Name"},
              {"header": "Size"}
            ],
            "rows": [
              {"cells": ["main.go", "1.2K"]},
              {"cells": ["tabs.go", "3.4K"]},
              {"cells": ["README.md", "512B"]}
            ]
          }
        }
      ]
    }
  ]
}
`
	node, err := ParseJSON([]byte(jsonTree))
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}
	declTree, err := Build(node)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	w, h := 80, 24
	r := loom.Rect{W: w, H: h}

	c1 := loom.NewCanvas(w, h)
	root.Draw(c1, r)

	c2 := loom.NewCanvas(w, h)
	declTree.Draw(c2, r)

	out1 := canvasToString(c1)
	out2 := canvasToString(c2)

	if out1 != out2 {
		t.Errorf("Outputs mismatch.\nExpected:\n%s\n\nGot:\n%s", out1, out2)
	}
}

func toJSONList(lines []string) string {
	var parts []string
	for _, l := range lines {
		parts = append(parts, fmt.Sprintf("%q", l))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
