// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

// allocateCells divides `total` integer cells among weights proportionally
// to their magnitude using the largest-remainder method, guaranteeing the
// returned sizes always sum to exactly total. NaN and negative weights are
// treated as zero. Returns nil when weights is empty or sums to zero, and
// an all-zero slice when total <= 0.
//
// Shared by RenderStackedBar (dividing a row's width among segments) and
// RenderTreemap's box layout (dividing a rectangle's width or height
// between two branches at each split).
func allocateCells(weights []float64, total int) []int {
	n := len(weights)
	if n == 0 {
		return nil
	}
	clamped := make([]float64, n)
	var sum float64
	for i, v := range weights {
		if !isFinite(v) || v < 0 {
			v = 0
		}
		clamped[i] = v
		sum += v
	}
	if sum <= 0 {
		return nil
	}
	if total <= 0 {
		return make([]int, n)
	}

	sizes := make([]int, n)
	remainders := make([]float64, n)
	allocated := 0
	for i, v := range clamped {
		ideal := v / sum * float64(total)
		sizes[i] = int(ideal)
		remainders[i] = ideal - float64(sizes[i])
		allocated += sizes[i]
	}
	remaining := total - allocated
	for remaining > 0 {
		best := -1
		for i := range sizes {
			if clamped[i] <= 0 {
				continue
			}
			if best == -1 || remainders[i] > remainders[best] {
				best = i
			}
		}
		if best == -1 {
			break
		}
		sizes[best]++
		remainders[best] = -1
		remaining--
	}
	return sizes
}
