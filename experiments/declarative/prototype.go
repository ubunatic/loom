// SPDX-FileCopyrightText: 2026 Jules
// SPDX-License-Identifier: AGPL-3.0-or-later

package declarative

import (
	"encoding/json"
	"fmt"

	"ubunatic.com/loom"
)

// Node represents a declarative UI node.
type Node struct {
	Type     string         `json:"type"`
	Props    map[string]any `json:"props"`
	Children []Node         `json:"children,omitempty"`
}

// Build creates a loom widget tree from a declarative Node.
func Build(n Node) (loom.Widget, error) {
	switch n.Type {
	case "view":
		lines := []string{}
		if v, ok := n.Props["lines"].([]any); ok {
			for _, line := range v {
				lines = append(lines, fmt.Sprint(line))
			}
		}
		view := loom.NewView(lines)
		if dim, ok := n.Props["dim"].(bool); ok && dim {
			view.Style.Dim = true
		}
		if bold, ok := n.Props["bold_focus"].(bool); ok && bold {
			view.FocusStyle.Bold = true
		}
		return view, nil
	case "split":
		if len(n.Children) != 2 {
			return nil, fmt.Errorf("split requires exactly 2 children")
		}
		first, err := Build(n.Children[0])
		if err != nil {
			return nil, err
		}
		second, err := Build(n.Children[1])
		if err != nil {
			return nil, err
		}
		split := loom.NewSplit(first, second)
		if v, ok := n.Props["orientation"].(string); ok {
			if v == "vertical" {
				split.Orientation = loom.Vertical
			} else {
				split.Orientation = loom.Horizontal
			}
		}
		if v, ok := n.Props["ratio"].(float64); ok {
			split.Ratio = v
		}
		if v, ok := n.Props["min_first"].(float64); ok {
			split.MinFirst = int(v)
		}
		if v, ok := n.Props["min_second"].(float64); ok {
			split.MinSecond = int(v)
		}
		if v, ok := n.Props["divider"].(string); ok {
			split.Divider.Glyph = v
			split.Divider.Style.Dim = true // Simplify for prototype
		}
		return split, nil
	case "tabs":
		var tabs []loom.Tab
		for _, child := range n.Children {
			if child.Type != "tab" {
				return nil, fmt.Errorf("tabs child must be a tab")
			}
			if len(child.Children) != 1 {
				return nil, fmt.Errorf("tab requires exactly 1 child widget")
			}
			w, err := Build(child.Children[0])
			if err != nil {
				return nil, err
			}
			title := ""
			if v, ok := child.Props["title"].(string); ok {
				title = v
			}
			tabs = append(tabs, loom.Tab{Title: title, Widget: w})
		}
		t := loom.NewTabs(tabs...)
		if v, ok := n.Props["keys"].(map[string]any); ok {
			if k, ok := v["previous"].(string); ok {
				t.Keys.Previous = k
			}
			if k, ok := v["next"].(string); ok {
				t.Keys.Next = k
			}
			if k, ok := v["cycle"].(string); ok {
				t.Keys.Cycle = k
			}
			if sel, ok := v["select"].([]any); ok {
				for _, s := range sel {
					t.Keys.Select = append(t.Keys.Select, fmt.Sprint(s))
				}
			}
		}
		return t, nil
	case "choice":
		var items []loom.Item
		if v, ok := n.Props["items"].([]any); ok {
			for _, itemAny := range v {
				if itemMap, ok := itemAny.(map[string]any); ok {
					name, _ := itemMap["name"].(string)
					desc, _ := itemMap["desc"].(string)
					items = append(items, loom.Item{Name: name, Desc: desc})
				}
			}
		}
		return loom.NewChoice(items), nil
	case "table":
		var cols []loom.Column
		var rows []loom.Row
		if v, ok := n.Props["cols"].([]any); ok {
			for _, colAny := range v {
				if colMap, ok := colAny.(map[string]any); ok {
					header, _ := colMap["header"].(string)
					cols = append(cols, loom.Column{Header: header})
				}
			}
		}
		if v, ok := n.Props["rows"].([]any); ok {
			for _, rowAny := range v {
				if rowMap, ok := rowAny.(map[string]any); ok {
					if cellsAny, ok := rowMap["cells"].([]any); ok {
						var cells []string
						for _, c := range cellsAny {
							cells = append(cells, fmt.Sprint(c))
						}
						rows = append(rows, loom.Row{Cells: cells})
					}
				}
			}
		}
		return loom.NewTable(cols, rows), nil
	default:
		return nil, fmt.Errorf("unknown widget type: %s", n.Type)
	}
}

// ParseJSON parses a JSON declarative tree.
func ParseJSON(data []byte) (Node, error) {
	var n Node
	err := json.Unmarshal(data, &n)
	return n, err
}
