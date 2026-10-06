// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"math"
	"sort"
)

// TreemapLayout selects the partitioning strategy used by LayoutTreemap.
type TreemapLayout int

const (
	TreemapLayoutSliceDice TreemapLayout = iota
	TreemapLayoutSquarified
)

// LayoutTreemap returns positive-valued segments as deterministic, integral
// cells tiled within w by h. Non-positive, NaN, and infinite values are
// omitted.
func LayoutTreemap(segments []TreemapSegment, w, h int, layout TreemapLayout) []TreemapCell {
	if w <= 0 || h <= 0 {
		return nil
	}
	values := make([]float64, len(segments))
	for i, segment := range segments {
		values[i] = segment.Value
	}
	var rects []treemapRect
	if layout == TreemapLayoutSquarified {
		rects = layoutTreemapSquarified(values, treemapRect{W: w, H: h})
	} else {
		rects = layoutTreemap(values, treemapRect{W: w, H: h})
	}
	cells := make([]TreemapCell, 0, len(segments))
	for i, rect := range rects {
		if rect.W <= 0 || rect.H <= 0 {
			continue
		}
		cells = append(cells, TreemapCell{Rect: TreemapRect{rect.X, rect.Y, rect.W, rect.H}, Segment: segments[i], Label: segments[i].Name, Value: segments[i].Value, Index: i})
	}
	return cells
}

// layoutTreemap recursively partitions values into non-overlapping,
// integer-cell rectangles tiling rect with no gaps. See RenderTreemap for
// the algorithm. The returned slice has one rect per value, in the same
// order as values; segments with a non-positive value get a zero-size
// rect.
func layoutTreemap(values []float64, rect treemapRect) []treemapRect {
	rects := make([]treemapRect, len(values))
	idx := make([]int, 0, len(values))
	for i, v := range values {
		if isFinite(v) && v > 0 {
			idx = append(idx, i)
		}
	}
	var recurse func(idx []int, r treemapRect)
	recurse = func(idx []int, r treemapRect) {
		if len(idx) == 0 {
			return
		}
		if len(idx) == 1 {
			rects[idx[0]] = r
			return
		}
		sorted := append([]int(nil), idx...)
		sort.SliceStable(sorted, func(a, b int) bool { return values[sorted[a]] > values[sorted[b]] })

		total := 0.0
		for _, i := range sorted {
			total += values[i]
		}
		splitAt, bestDiff, cum := 1, math.Inf(1), 0.0
		for k := 1; k < len(sorted); k++ {
			cum += values[sorted[k-1]]
			if diff := math.Abs(cum - total/2); diff < bestDiff {
				bestDiff, splitAt = diff, k
			}
		}
		left, right := sorted[:splitAt], sorted[splitAt:]
		var leftSum, rightSum float64
		for _, i := range left {
			leftSum += values[i]
		}
		for _, i := range right {
			rightSum += values[i]
		}

		// Splitting along width (left/right) keeps the full height but
		// shrinks each child's width, making them relatively TALLER --
		// more portrait. Splitting along height (top/bottom) keeps the
		// full width and shrinks height, making children relatively
		// WIDER -- more landscape, which fits a horizontal text label
		// far more easily. So default to a height-split, and only cut
		// width once the rect is already comfortably landscape (see
		// treemapLandscapeBias) or height has no room left to give.
		splitWidth := r.H <= 1 || (r.W > 1 && float64(r.W) >= float64(r.H)*treemapLandscapeBias)
		if splitWidth {
			sizes := allocateCells([]float64{leftSum, rightSum}, r.W)
			recurse(left, treemapRect{r.X, r.Y, sizes[0], r.H})
			recurse(right, treemapRect{r.X + sizes[0], r.Y, sizes[1], r.H})
		} else {
			sizes := allocateCells([]float64{leftSum, rightSum}, r.H)
			recurse(left, treemapRect{r.X, r.Y, r.W, sizes[0]})
			recurse(right, treemapRect{r.X, r.Y + sizes[0], r.W, sizes[1]})
		}
	}
	recurse(idx, rect)
	return rects
}

func layoutTreemapSquarified(values []float64, rect treemapRect) []treemapRect {
	return layoutTreemapSquarifiedFloat(values, rect)
	/*
		rects := make([]treemapRect, len(values))
		indexes := make([]int, 0, len(values))
		for i, value := range values {
			if isFinite(value) && value > 0 {
				indexes = append(indexes, i)
			}
		}
		sort.SliceStable(indexes, func(i, j int) bool { return values[indexes[i]] > values[indexes[j]] })
		var place func([]int, treemapRect)
		place = func(items []int, area treemapRect) {
			if len(items) == 0 || area.W <= 0 || area.H <= 0 {
				return
			}
			if len(items) == 1 {
				rects[items[0]] = area
				return
			}
			if area.W == 1 || area.H == 1 {
				weights := make([]float64, len(items))
				for i, index := range items {
					weights[i] = values[index]
				}
				if area.W == 1 {
					sizes := allocateCells(weights, area.H)
					y := area.Y
					for i, index := range items {
						rects[index] = treemapRect{area.X, y, 1, sizes[i]}
						y += sizes[i]
					}
				} else {
					sizes := allocateCells(weights, area.W)
					x := area.X
					for i, index := range items {
						rects[index] = treemapRect{x, area.Y, sizes[i], 1}
						x += sizes[i]
					}
				}
				return
			}
			bestK, bestScore, bestWidth := 1, math.Inf(1), area.W >= area.H
			for k := 1; k < len(items); k++ {
				for _, horizontal := range []bool{true, false} {
					left := items[:k]
					ls, total := 0.0, 0.0
					for _, i := range left {
						ls += values[i]
					}
					for _, i := range items {
						total += values[i]
					}
					if horizontal {
						width := int(math.Round(float64(area.W) * ls / total))
						if width < 1 || width >= area.W {
							continue
						}
						score := math.Max(float64(width)/float64(area.H), float64(area.H)/float64(width))
						score = math.Max(score, math.Max(float64(area.W-width)/float64(area.H), float64(area.H)/float64(area.W-width)))
						if score < bestScore {
							bestK, bestScore, bestWidth = k, score, true
						}
					} else {
						height := int(math.Round(float64(area.H) * ls / total))
						if height < 1 || height >= area.H {
							continue
						}
						score := math.Max(float64(area.W)/float64(height), float64(height)/float64(area.W))
						if score < bestScore {
							bestK, bestScore, bestWidth = k, score, false
						}
					}
				}
			}
			left, right := items[:bestK], items[bestK:]
			ls := 0.0
			for _, i := range left {
				ls += values[i]
			}
			total := 0.0
			for _, i := range items {
				total += values[i]
			}
			if bestWidth {
				width := int(math.Round(float64(area.W) * ls / total))
				sizes := allocateCellsForValues(left, values, area.H, ls)
				y := area.Y
				for n, i := range left {
					rects[i] = treemapRect{area.X, y, width, sizes[n]}
					y += sizes[n]
				}
				place(right, treemapRect{area.X + width, area.Y, area.W - width, area.H})
			} else {
				height := int(math.Round(float64(area.H) * ls / total))
				sizes := allocateCellsForValues(left, values, area.W, ls)
				x := area.X
				for n, i := range left {
					rects[i] = treemapRect{x, area.Y, sizes[n], height}
					x += sizes[n]
				}
				place(right, treemapRect{area.X, area.Y + height, area.W, area.H - height})
			}
		}
		place(indexes, rect)
		for _, index := range indexes {
			if rects[index].W == 0 || rects[index].H == 0 {
				return layoutTreemap(values, rect)
			}
		}
		return rects
	*/
}

// layoutTreemapSquarifiedFloat lays out rows in continuous coordinates first;
// integer edges are rounded cumulatively when each row is emitted.
func layoutTreemapSquarifiedFloat(values []float64, rect treemapRect) []treemapRect {
	type frect struct{ x, y, w, h float64 }
	result := make([]treemapRect, len(values))
	ids := make([]int, 0, len(values))
	total := 0.0
	for i, v := range values {
		if isFinite(v) && v > 0 {
			ids = append(ids, i)
			total += v
		}
	}
	sort.SliceStable(ids, func(i, j int) bool { return values[ids[i]] > values[ids[j]] })
	var place func([]int, frect, float64)
	place = func(items []int, a frect, sum float64) {
		if len(items) == 0 || a.w <= 0 || a.h <= 0 || sum <= 0 {
			return
		}
		if len(items) == 1 {
			x0, y0 := math.Round(a.x), math.Round(a.y)
			x1, y1 := math.Round(a.x+a.w), math.Round(a.y+a.h)
			result[items[0]] = treemapRect{int(x0), int(y0), int(x1 - x0), int(y1 - y0)}
			return
		}
		short := math.Min(a.w, a.h*2)
		row := make([]int, 0, len(items))
		rowSum := 0.0
		worst := math.Inf(1)
		for _, id := range items {
			next := append(row, id)
			ns := rowSum + values[id]
			area := a.w * a.h * ns / sum
			maxv, minv := values[next[0]], values[next[0]]
			for _, n := range next {
				maxv = math.Max(maxv, values[n])
				minv = math.Min(minv, values[n])
			}
			score := math.Max(short*short*maxv/(area*area), area*area/(short*short*minv))
			if len(row) > 0 && score > worst {
				break
			}
			row, rowSum, worst = next, ns, score
		}
		rowSet := make(map[int]bool, len(row))
		for _, id := range row {
			rowSet[id] = true
		}
		rowArea := a.w * a.h * rowSum / sum
		if a.w <= a.h*2 {
			h := rowArea / a.w
			x := a.x
			for _, id := range row {
				w := a.w * values[id] / rowSum
				x0, x1 := math.Round(x), math.Round(x+w)
				y0, y1 := math.Round(a.y), math.Round(a.y+h)
				result[id] = treemapRect{int(x0), int(y0), int(x1 - x0), int(y1 - y0)}
				x += w
			}
			place(filterIDs(items, rowSet), frect{a.x, a.y + h, a.w, a.h - h}, sum-rowSum)
		} else {
			w := rowArea / a.h
			y := a.y
			for _, id := range row {
				h := a.h * values[id] / rowSum
				x0, x1 := math.Round(a.x), math.Round(a.x+w)
				y0, y1 := math.Round(y), math.Round(y+h)
				result[id] = treemapRect{int(x0), int(y0), int(x1 - x0), int(y1 - y0)}
				y += h
			}
			place(filterIDs(items, rowSet), frect{a.x + w, a.y, a.w - w, a.h}, sum-rowSum)
		}
	}
	place(ids, frect{float64(rect.X), float64(rect.Y), float64(rect.W), float64(rect.H)}, total)
	return result
}

func filterIDs(ids []int, used map[int]bool) []int {
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if !used[id] {
			out = append(out, id)
		}
	}
	return out
}

func squarifiedBetter(row []int, next int, values []float64, area treemapRect) bool {
	if len(row) == 0 {
		return true
	}
	short := math.Min(float64(area.W), float64(area.H))
	areaSize := float64(area.W * area.H)
	rowSum := 0.0
	for _, i := range row {
		rowSum += values[i]
	}
	old := squarifiedWorst(rowSum, values, row, short, areaSize)
	with := append(append([]int(nil), row...), next)
	return squarifiedWorst(rowSum+values[next], values, with, short, areaSize) <= old
}

func squarifiedWorst(sum float64, values []float64, row []int, short, areaSize float64) float64 {
	if sum <= 0 || short <= 0 {
		return math.Inf(1)
	}
	maxValue, minValue := values[row[0]], values[row[0]]
	for _, i := range row {
		maxValue = math.Max(maxValue, values[i])
		minValue = math.Min(minValue, values[i])
	}
	rowArea := areaSize * sum / (areaSize + sum)
	return math.Max(short*short*maxValue/(rowArea*rowArea), rowArea*rowArea/(short*short*minValue))
}

func allocateCellsForValues(indexes []int, values []float64, total int, sum float64) []int {
	weights := make([]float64, len(indexes))
	for i, index := range indexes {
		weights[i] = values[index]
	}
	return allocateCells(weights, total)
}

// treemapLandscapeBias sets how much wider than tall (in character cells)
// a rectangle must already be before layoutTreemap is willing to split it
// along its width instead of its height. Terminal character cells are
// themselves roughly twice as tall as they are wide, and it's horizontal
// text that needs to fit inside a box, so this is deliberately set above
// 1 (which would just aim for a square in cell *count*, already a
// landscape rectangle on screen) to favor genuinely wide, short boxes.
const treemapLandscapeBias = 3.0
