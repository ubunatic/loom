// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestTreeNavigationAndExpansion(t *testing.T) {
	leaf := &loom.TreeNode{ID: "leaf", Label: "leaf"}
	child := &loom.TreeNode{ID: "child", Label: "child", Children: []*loom.TreeNode{leaf}}
	root := &loom.TreeNode{ID: "root", Label: "root", Children: []*loom.TreeNode{child}}
	tree := loom.NewTree([]*loom.TreeNode{root})
	if got := tree.VisibleNodes(); len(got) != 1 || got[0] != root {
		t.Fatalf("initial visible nodes = %v", got)
	}
	tree.HandleKey(loom.KeyEvent{Key: "right"})
	if got := tree.VisibleNodes(); len(got) != 2 || got[1] != child {
		t.Fatalf("after expansion = %v", got)
	}
	tree.HandleKey(loom.KeyEvent{Key: "j"})
	tree.HandleKey(loom.KeyEvent{Key: "right"})
	tree.HandleKey(loom.KeyEvent{Key: "down"})
	if got := tree.SelectedNode(); got != leaf {
		t.Fatalf("selection after nested expansion = %v, want leaf", got)
	}
	tree.HandleKey(loom.KeyEvent{Key: "left"})
	if got := tree.SelectedNode(); got != child {
		t.Fatalf("left from collapsed leaf = %v, want parent", got)
	}
}

func TestTreeMouseToggleSelectAndScroll(t *testing.T) {
	root := &loom.TreeNode{ID: "root", Label: "root", Children: []*loom.TreeNode{{ID: "child", Label: "child"}}}
	other := &loom.TreeNode{ID: "other", Label: "other"}
	tree := loom.NewTree([]*loom.TreeNode{root, other})
	c := loom.NewCanvas(20, 2)
	tree.Draw(c, loom.Rect{W: 20, H: 2})
	tree.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 0, Y: 0})
	if !root.Expanded {
		t.Fatal("disclosure click did not expand root")
	}
	tree.Draw(c, loom.Rect{W: 20, H: 2})
	tree.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 7, Y: 1})
	if tree.SelectedNode() != root.Children[0] {
		t.Fatal("row click did not select child")
	}
	tree.HandleKey(loom.KeyEvent{Key: "down"})
	if tree.ScrollY == 0 {
		t.Fatal("selection beyond viewport did not scroll")
	}
}

func TestTreeActivateLeafAndAliases(t *testing.T) {
	leaf := &loom.TreeNode{ID: "leaf", Label: "leaf"}
	tree := loom.NewTree([]*loom.TreeNode{leaf})
	var activated *loom.TreeNode
	tree.OnActivate = func(node *loom.TreeNode) { activated = node }
	tree.HandleKey(loom.KeyEvent{Key: "enter"})
	if activated != leaf {
		t.Fatal("Enter did not activate leaf")
	}
	for _, key := range []string{"pgdown", "pgdn", "pagedown", "pgup", "pageup"} {
		tree.HandleKey(loom.KeyEvent{Key: key})
	}
}
