// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"testing"
)

func segmentValue(t *testing.T, segments []TreemapSegment, name string) (float64, bool) {
	t.Helper()
	for _, s := range segments {
		if s.Name == name {
			return s.Value, true
		}
	}
	return 0, false
}

func TestAggregateTreemapMergesSameNameSiblings(t *testing.T) {
	// Three "chrome" renderer processes under one browser should collapse
	// into a single "chrome" segment (the browser-tab case).
	root := TreemapNode{
		Name: "root",
		Children: []TreemapNode{
			{Name: "chrome", Value: 5, Children: []TreemapNode{
				{Name: "chrome-tab", Value: 10},
				{Name: "chrome-tab", Value: 20},
				{Name: "chrome-tab", Value: 15},
			}},
			{Name: "sshd", Value: 1},
		},
	}
	// Force expansion of "chrome" by giving a budget that fits the expansion.
	segments := AggregateTreemap(root, 5)
	tabTotal, ok := segmentValue(t, segments, "chrome-tab")
	if !ok {
		t.Fatalf("expected merged chrome-tab segment, got %+v", segments)
	}
	if tabTotal != 45 {
		t.Errorf("chrome-tab total = %v, want 45", tabTotal)
	}
	chromeOwn, ok := segmentValue(t, segments, "chrome")
	if !ok || chromeOwn != 5 {
		t.Errorf("chrome own segment = %v, ok=%v, want 5", chromeOwn, ok)
	}
	sshd, ok := segmentValue(t, segments, "sshd")
	if !ok || sshd != 1 {
		t.Errorf("sshd = %v, ok=%v, want 1", sshd, ok)
	}
}

func TestAggregateTreemapCapsNodeCount(t *testing.T) {
	root := TreemapNode{Name: "root"}
	for i := 0; i < 500; i++ {
		root.Children = append(root.Children, TreemapNode{Name: "proc", Value: 1})
	}
	// All same name, so it should merge to 1 regardless of the cap.
	segments := AggregateTreemap(root, MaxTreemapNodes)
	if len(segments) != 1 {
		t.Fatalf("len(segments) = %d, want 1", len(segments))
	}
	if segments[0].Value != 500 {
		t.Errorf("segments[0].Value = %v, want 500", segments[0].Value)
	}
}

func TestAggregateTreemapNeverExceedsMaxNodes(t *testing.T) {
	root := TreemapNode{Name: "root"}
	for i := 0; i < 500; i++ {
		root.Children = append(root.Children, TreemapNode{
			Name: distinctName(i), Value: float64(i + 1),
		})
	}
	for _, max := range []int{-5, 0, 1, 10, 100, 500, 1000} {
		segments := AggregateTreemap(root, max)
		wantMax := max
		if wantMax <= 0 {
			wantMax = 1
		}
		if len(segments) > wantMax {
			t.Errorf("maxNodes=%d: len(segments) = %d, exceeds budget", max, len(segments))
		}
	}
}

func TestAggregateTreemapPreservesTotal(t *testing.T) {
	root := TreemapNode{
		Name: "root",
		Children: []TreemapNode{
			{Name: "a", Value: 3, Children: []TreemapNode{
				{Name: "a1", Value: 4},
				{Name: "a2", Value: 5, Children: []TreemapNode{
					{Name: "a2x", Value: 6},
				}},
			}},
			{Name: "b", Value: 7},
		},
	}
	want := subtreeTotal(root)
	for _, max := range []int{1, 2, 3, 4, 5, 100} {
		segments := AggregateTreemap(root, max)
		var got float64
		for _, s := range segments {
			got += s.Value
		}
		if got != want {
			t.Errorf("maxNodes=%d: total = %v, want %v (segments=%+v)", max, got, want, segments)
		}
	}
}

func TestAggregateTreemapPreservesRootValue(t *testing.T) {
	leaf := TreemapNode{Name: "root", Value: 7}
	if got := AggregateTreemap(leaf, 3); len(got) != 1 || got[0].Name != "root" || got[0].Value != 7 {
		t.Errorf("leaf root = %+v, want one root segment worth 7", got)
	}
	root := TreemapNode{Name: "root", Value: 7, Children: []TreemapNode{
		{Name: "a", Value: 3}, {Name: "b", Value: 5},
	}}
	for _, max := range []int{1, 2, 3, 10} {
		var total float64
		for _, segment := range AggregateTreemap(root, max) {
			total += segment.Value
		}
		if total != 15 {
			t.Errorf("maxNodes=%d: total = %v, want 15", max, total)
		}
	}
}

func TestAggregateTreemapExpandsHeaviestFirst(t *testing.T) {
	root := TreemapNode{
		Name: "root",
		Children: []TreemapNode{
			{Name: "heavy", Children: []TreemapNode{
				{Name: "heavy-a", Value: 50},
				{Name: "heavy-b", Value: 50},
			}},
			{Name: "light", Children: []TreemapNode{
				{Name: "light-a", Value: 1},
				{Name: "light-b", Value: 1},
			}},
		},
	}
	// Budget for exactly one expansion (2 top nodes -> 3 nodes after
	// expanding one of them).
	segments := AggregateTreemap(root, 3)
	if len(segments) != 3 {
		t.Fatalf("len(segments) = %d, want 3 (segments=%+v)", len(segments), segments)
	}
	if _, ok := segmentValue(t, segments, "light"); !ok {
		t.Errorf("expected unexpanded light segment, got %+v", segments)
	}
	if _, ok := segmentValue(t, segments, "heavy-a"); !ok {
		t.Errorf("expected heavy to expand into heavy-a, got %+v", segments)
	}
}

func TestAggregateTreemapPartiallyExpandsBushyNode(t *testing.T) {
	// Regression: a node with far more direct children than the budget
	// allows (e.g. a real init process with 80+ services) must not lock
	// up as one giant unexpanded blob just because its FULL expansion
	// doesn't fit. It should partially expand -- revealing its heaviest
	// children up to budget and folding the rest into an "other" child --
	// so a single bushy branch can never block all further detail.
	root := TreemapNode{Name: "root"}
	var bushyChildren []TreemapNode
	for i := 0; i < 90; i++ {
		bushyChildren = append(bushyChildren, TreemapNode{Name: distinctName(i), Value: 1})
	}
	root.Children = []TreemapNode{
		{Name: "init", Children: bushyChildren}, // subtree total 90
		{Name: "other-proc", Value: 5},
	}

	segments := AggregateTreemap(root, 10)
	if len(segments) > 10 {
		t.Fatalf("len(segments) = %d, exceeds budget of 10", len(segments))
	}
	if len(segments) < 3 {
		t.Fatalf("expected 'init' to partially expand into several segments, got %+v", segments)
	}
	if _, ok := segmentValue(t, segments, "init"); ok {
		t.Errorf("'init' should have been replaced by its (partially expanded) children, not kept as one blob: %+v", segments)
	}
	otherTotal, ok := segmentValue(t, segments, "other")
	if !ok {
		t.Fatalf("expected init's excess children folded into an 'other' segment, got %+v", segments)
	}
	if otherTotal <= 0 || otherTotal >= 90 {
		t.Errorf("'other' segment = %v, want a partial remainder strictly between 0 and 90", otherTotal)
	}
	otherProc, ok := segmentValue(t, segments, "other-proc")
	if !ok || otherProc != 5 {
		t.Errorf("other-proc = %v, ok=%v, want 5 (a small sibling should not get crowded out)", otherProc, ok)
	}
	var total float64
	for _, s := range segments {
		total += s.Value
	}
	if total != 95 {
		t.Errorf("total = %v, want 95 (90 init subtree + 5 other-proc)", total)
	}
}

func TestAggregateTreemapDropsNonPositiveSegments(t *testing.T) {
	root := TreemapNode{
		Name: "root",
		Children: []TreemapNode{
			{Name: "zero", Value: 0},
			{Name: "positive", Value: 1},
		},
	}
	segments := AggregateTreemap(root, 100)
	if _, ok := segmentValue(t, segments, "zero"); ok {
		t.Errorf("expected zero-value node dropped, got %+v", segments)
	}
	if _, ok := segmentValue(t, segments, "positive"); !ok {
		t.Errorf("expected positive segment present, got %+v", segments)
	}
}

func TestAggregateTreemapEmptyRoot(t *testing.T) {
	segments := AggregateTreemap(TreemapNode{Name: "root"}, MaxTreemapNodes)
	if len(segments) != 0 {
		t.Errorf("AggregateTreemap(empty) = %+v, want empty", segments)
	}
}

func TestAggregateTreemapFeedsRenderStackedBar(t *testing.T) {
	root := TreemapNode{
		Name: "root",
		Children: []TreemapNode{
			{Name: "a", Value: 3},
			{Name: "b", Value: 1},
		},
	}
	segments := AggregateTreemap(root, MaxTreemapNodes)
	values := make([]float64, len(segments))
	for i, s := range segments {
		values[i] = s.Value
	}
	out := RenderStackedBar(values, StackedBarOptions{Width: 4, NoWrapper: true})
	if got := len([]rune(out)); got != 4 {
		t.Errorf("RenderStackedBar(aggregated) width = %d, want 4 (out=%q)", got, out)
	}
}

func distinctName(i int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	return string(letters[i%len(letters)]) + string(letters[(i/len(letters))%len(letters)]) + string(rune('0'+i%10))
}
