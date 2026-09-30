// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"codeberg.org/ubunatic/loom/measure"
)

// TreeNode is one entry in a Tree. Children are drawn only while Expanded.
type TreeNode struct {
	ID       string
	Label    string
	Data     any
	Children []*TreeNode
	Expanded bool
	Icon     string
}

// TreeStyle controls the normal and selected row appearance.
type TreeStyle struct {
	Normal   Style
	Selected Style
}

// Tree displays a keyboard and mouse navigable hierarchy.
type Tree struct {
	Nodes      []*TreeNode
	Style      TreeStyle
	OnActivate func(*TreeNode)
	ScrollY    int

	selected int
	lastRect Rect
	drawn    bool
	focused  bool
}

// NewTree creates a tree with the first root selected.
func NewTree(nodes []*TreeNode) *Tree {
	return &Tree{Nodes: nodes, Style: TreeStyle{Normal: DefaultChoiceStyle().Normal, Selected: DefaultChoiceStyle().Selected}, focused: true}
}

// VisibleNodes returns nodes in their current depth-first display order.
func (t *Tree) VisibleNodes() []*TreeNode {
	var nodes []*TreeNode
	var visit func([]*TreeNode)
	visit = func(children []*TreeNode) {
		for _, node := range children {
			if node == nil {
				continue
			}
			nodes = append(nodes, node)
			if node.Expanded {
				visit(node.Children)
			}
		}
	}
	visit(t.Nodes)
	return nodes
}

// SelectedNode returns the currently selected visible node, or nil.
func (t *Tree) SelectedNode() *TreeNode {
	nodes := t.VisibleNodes()
	if len(nodes) == 0 {
		return nil
	}
	t.selected = min(max(0, t.selected), len(nodes)-1)
	return nodes[t.selected]
}

// Measure reports the natural width and height of visible content.
func (t *Tree) Measure(width int) measure.Size {
	nodes := t.VisibleNodes()
	w := 1
	for _, node := range nodes {
		w = max(w, len([]rune(node.Label))+4)
	}
	return measure.Size{Width: min(w, max(1, width)), Height: len(nodes)}
}

// Draw renders visible rows, scrolling the selection into the assigned region.
func (t *Tree) Draw(c *Canvas, r Rect) {
	t.lastRect, t.drawn = r, true
	if r.W <= 0 || r.H <= 0 {
		return
	}
	nodes := t.VisibleNodes()
	if len(nodes) == 0 {
		return
	}
	t.selected = min(max(0, t.selected), len(nodes)-1)
	if t.selected < t.ScrollY {
		t.ScrollY = t.selected
	}
	if t.selected >= t.ScrollY+r.H {
		t.ScrollY = t.selected - r.H + 1
	}
	t.ScrollY = min(max(0, t.ScrollY), max(0, len(nodes)-r.H))
	for row := 0; row < r.H && t.ScrollY+row < len(nodes); row++ {
		i := t.ScrollY + row
		node := nodes[i]
		style := t.Style.Normal
		if i == t.selected && t.focused {
			style = t.Style.Selected
		}
		c.PaintSurface(Rect{X: r.X, Y: r.Y + row, W: r.W, H: 1}, style)
		depth, _ := t.nodeInfo(node)
		marker := "  "
		if len(node.Children) > 0 {
			marker = "▶ "
			if node.Expanded {
				marker = "▼ "
			}
		}
		label := node.Icon
		if label != "" {
			label += " "
		}
		label += node.Label
		c.Write(r.X, r.Y+row, repeatTreeIndent(depth)+marker+label, style)
	}
}

func repeatTreeIndent(n int) string {
	out := ""
	for range n {
		out += "  "
	}
	return out
}

func (t *Tree) nodeInfo(want *TreeNode) (depth int, parent *TreeNode) {
	var visit func([]*TreeNode, int, *TreeNode) bool
	visit = func(nodes []*TreeNode, d int, p *TreeNode) bool {
		for _, n := range nodes {
			if n == want {
				depth, parent = d, p
				return true
			}
			if n.Expanded && visit(n.Children, d+1, n) {
				return true
			}
		}
		return false
	}
	visit(t.Nodes, 0, nil)
	return
}

func (t *Tree) ConsumeKey(e KeyEvent) EventResult {
	key := e.Name()
	nodes := t.VisibleNodes()
	if len(nodes) == 0 {
		return Ignored()
	}
	t.selected = min(max(0, t.selected), len(nodes)-1)
	node := nodes[t.selected]
	switch key {
	case "up", "k":
		t.selected = max(0, t.selected-1)
	case "down", "j":
		t.selected = min(len(nodes)-1, t.selected+1)
	case "right", "l":
		if len(node.Children) > 0 && !node.Expanded {
			node.Expanded = true
		} else if node.Expanded && len(node.Children) > 0 {
			t.selected++
		}
	case "left", "h":
		if node.Expanded {
			node.Expanded = false
		} else if _, parent := t.nodeInfo(node); parent != nil {
			for i, n := range nodes {
				if n == parent {
					t.selected = i
					break
				}
			}
		}
	case "enter", " ":
		if len(node.Children) > 0 {
			node.Expanded = !node.Expanded
		} else if t.OnActivate != nil {
			t.OnActivate(node)
		}
	case "pgdown", "pgdn", "pagedown":
		t.selected = min(len(nodes)-1, t.selected+max(1, t.lastRect.H))
	case "pgup", "pageup":
		t.selected = max(0, t.selected-max(1, t.lastRect.H))
	case "home":
		t.selected = 0
	case "end":
		t.selected = len(nodes) - 1
	default:
		return Ignored()
	}
	t.ensureSelectionVisible()
	return Handled()
}

func (t *Tree) ensureSelectionVisible() {
	h := max(1, t.lastRect.H)
	if t.selected < t.ScrollY {
		t.ScrollY = t.selected
	}
	if t.selected >= t.ScrollY+h {
		t.ScrollY = t.selected - h + 1
	}
	t.ScrollY = min(max(0, t.ScrollY), max(0, len(t.VisibleNodes())-h))
}

// ConsumeMouse selects a row or toggles a node through its disclosure marker.
func (t *Tree) ConsumeMouse(e MouseEvent) EventResult {
	if !t.drawn || e.Action != MousePress || e.Button != MouseLeft || e.X < 0 || e.Y < 0 || e.X >= t.lastRect.W || e.Y >= t.lastRect.H {
		return Ignored()
	}
	i := t.ScrollY + e.Y
	nodes := t.VisibleNodes()
	if i >= len(nodes) {
		return Ignored()
	}
	t.selected = i
	depth, _ := t.nodeInfo(nodes[i])
	if e.X >= depth*2 && e.X < depth*2+2 && len(nodes[i].Children) > 0 {
		nodes[i].Expanded = !nodes[i].Expanded
	}
	return Handled()
}

// Focused reports whether this tree is focused.
func (t *Tree) Focused() bool { return t.focused }

// SetFocus updates focus.
func (t *Tree) SetFocus(focused bool) { t.focused = focused }

// ApplyTheme updates row styles from a theme.
func (t *Tree) ApplyTheme(theme ThemeColors) {
	style := theme.ChoiceStyle()
	t.Style = TreeStyle{Normal: style.Normal, Selected: style.Selected}
}
