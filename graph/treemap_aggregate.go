// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import "sort"

// MaxTreemapNodes is a sensible default node budget for AggregateTreemap,
// e.g. when reducing a full process tree to a glanceable treemap.
const MaxTreemapNodes = 100

// TreemapNode is one node of a hierarchical usage tree, such as a process
// tree with per-process CPU usage. Value is this node's own contribution,
// excluding children; a node's total (used for sizing, sorting, and the
// TreemapSegment it may collapse to) is its own Value plus every
// descendant's Value.
type TreemapNode struct {
	Name     string
	Value    float64
	Children []TreemapNode
}

// TreemapSegment is one aggregated entry ready for RenderTreemap or
// RenderStackedBar: a label plus the combined value of everything it
// represents.
type TreemapSegment struct {
	Name  string
	Value float64
}

// AggregateTreemap reduces a hierarchical usage tree, such as a process
// tree with per-process CPU usage, to at most maxNodes segments suitable
// for RenderTreemap or RenderStackedBar.
//
// Same-name nodes are merged wherever they appear in the working set,
// summing their values and pooling their children, so recurring node types
// (repeated browser-tab renderer processes, for example) collapse into one
// segment instead of each counting separately toward the node budget.
//
// Detail is prioritized by weight: starting from root's (merged) children,
// the heaviest node that still has children to reveal is repeatedly
// replaced by its own merged children -- including a synthetic child
// carrying the node's own value, so expanding a node never silently drops
// it -- until node count would exceed maxNodes or no node has children
// left to reveal. A node whose full expansion would overflow the budget
// (a process with dozens of direct children, say) is not skipped outright:
// it expands partially, keeping only its heaviest children and folding the
// rest into a same-level "other" child, so one bushy branch can never
// permanently block all further detail the way an all-or-nothing expansion
// would. A node left unexpanded (down to a single leaf, or because the
// overall budget is exhausted) is reported as one segment summing its own
// value and every descendant's value: "aggregating lower nodes." If root
// already has more than maxNodes (merged) children, the lightest are
// combined into a single trailing "other" segment before expansion starts,
// so the budget is never exceeded on input either.
//
// maxNodes <= 0 clamps to 1. Segments with a non-positive total are
// dropped.
func AggregateTreemap(root TreemapNode, maxNodes int) []TreemapSegment {
	if maxNodes <= 0 {
		maxNodes = 1
	}
	frontier := capTreemapNodes(mergeTreemapNodes(root.Children), maxNodes)

	for len(frontier) < maxNodes {
		best := -1
		for i := range frontier {
			if len(frontier[i].Children) == 0 {
				continue
			}
			if best == -1 || subtreeTotal(frontier[i]) > subtreeTotal(frontier[best]) {
				best = i
			}
		}
		if best == -1 {
			break
		}
		expanded := expandTreemapNode(frontier[best])
		available := maxNodes - (len(frontier) - 1)
		if available < 1 {
			available = 1
		}
		expanded = capTreemapNodes(expanded, available)

		candidate := make([]TreemapNode, 0, len(frontier)-1+len(expanded))
		candidate = append(candidate, frontier[:best]...)
		candidate = append(candidate, expanded...)
		candidate = append(candidate, frontier[best+1:]...)
		frontier = mergeTreemapNodes(candidate)
	}

	segments := make([]TreemapSegment, 0, len(frontier))
	for _, n := range frontier {
		if total := subtreeTotal(n); total > 0 {
			segments = append(segments, TreemapSegment{Name: n.Name, Value: total})
		}
	}
	return segments
}

// capTreemapNodes shrinks nodes to at most maxNodes entries when it already
// exceeds that budget, keeping the maxNodes-1 heaviest nodes intact (by
// subtree total) and combining the rest into a single trailing "other" node
// carrying their summed total. A no-op when nodes already fits.
func capTreemapNodes(nodes []TreemapNode, maxNodes int) []TreemapNode {
	if len(nodes) <= maxNodes {
		return nodes
	}
	sorted := append([]TreemapNode(nil), nodes...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return subtreeTotal(sorted[i]) > subtreeTotal(sorted[j])
	})
	keep := maxNodes - 1
	if keep < 0 {
		keep = 0
	}
	result := append([]TreemapNode(nil), sorted[:keep]...)
	var otherTotal float64
	for _, n := range sorted[keep:] {
		otherTotal += subtreeTotal(n)
	}
	if otherTotal > 0 {
		result = append(result, TreemapNode{Name: "other", Value: otherTotal})
	}
	return mergeTreemapNodes(result)
}

// expandTreemapNode replaces a node with its merged children, folding the
// node's own value in as a synthetic same-named child first so expansion
// never loses it.
func expandTreemapNode(n TreemapNode) []TreemapNode {
	set := n.Children
	if n.Value != 0 {
		set = append([]TreemapNode{{Name: n.Name, Value: n.Value}}, set...)
	}
	return mergeTreemapNodes(set)
}

// subtreeTotal sums a node's own value and every descendant's value.
func subtreeTotal(n TreemapNode) float64 {
	total := n.Value
	for _, c := range n.Children {
		total += subtreeTotal(c)
	}
	return total
}

// mergeTreemapNodes combines nodes sharing a Name (in first-seen order),
// summing their values and pooling their children.
func mergeTreemapNodes(nodes []TreemapNode) []TreemapNode {
	if len(nodes) == 0 {
		return nil
	}
	order := make([]string, 0, len(nodes))
	byName := make(map[string]*TreemapNode, len(nodes))
	for _, n := range nodes {
		existing, ok := byName[n.Name]
		if !ok {
			merged := TreemapNode{Name: n.Name, Value: n.Value, Children: append([]TreemapNode(nil), n.Children...)}
			byName[n.Name] = &merged
			order = append(order, n.Name)
			continue
		}
		existing.Value += n.Value
		existing.Children = append(existing.Children, n.Children...)
	}
	result := make([]TreemapNode, len(order))
	for i, name := range order {
		result[i] = *byName[name]
	}
	return result
}
