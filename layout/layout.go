// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package layout contains renderer-independent allocation primitives.
package layout

import "fmt"

// Constraint bounds one layout item in cells. A zero Preferred uses Min.
type Constraint struct {
	Min       int
	Preferred int
	Max       int
	HasMax    bool
}

// Item is one participant in a plan. Invisible items consume no cells and
// retain a zero allocation in the result.
type Item struct {
	Constraint Constraint
	Visible    bool
}

// Allocation is the offset and size assigned to an item.
type Allocation struct {
	Offset int
	Size   int
}

// Plan allocates total cells among visible items separated by gap cells.
// Preferred sizes are packed first, then items shrink from the end to their
// minima or stretch from the beginning to their maxima. Both passes are
// deterministic and preserve declaration order.
func Plan(total, gap int, items []Item) ([]Allocation, error) {
	if total < 0 || gap < 0 {
		return nil, fmt.Errorf("layout: total and gap must be non-negative")
	}
	result := make([]Allocation, len(items))
	visible := make([]int, 0, len(items))
	for i, item := range items {
		if !item.Visible {
			continue
		}
		if err := validate(item.Constraint); err != nil {
			return nil, fmt.Errorf("layout item %d: %w", i, err)
		}
		visible = append(visible, i)
	}
	if len(visible) == 0 {
		return result, nil
	}
	if len(visible) > 1 && gap > total/(len(visible)-1) {
		return nil, fmt.Errorf("layout: gaps exceed total space")
	}
	available := total - gap*(len(visible)-1)
	if available < 0 {
		return nil, fmt.Errorf("layout: gaps exceed total space")
	}
	used := 0
	for _, index := range visible {
		preferred := items[index].Constraint.Preferred
		if preferred == 0 {
			preferred = items[index].Constraint.Min
		}
		result[index].Size = preferred
		if preferred > int(^uint(0)>>1)-used {
			return nil, fmt.Errorf("layout: preferred sizes overflow")
		}
		used += preferred
	}
	if used > available {
		for i := len(visible) - 1; i >= 0 && used > available; i-- {
			index := visible[i]
			floor := items[index].Constraint.Min
			cut := min(result[index].Size-floor, used-available)
			result[index].Size -= cut
			used -= cut
		}
		if used > available {
			return nil, fmt.Errorf("layout: minimum sizes exceed available space")
		}
	}
	if used < available {
		for i := 0; used < available; i++ {
			index := visible[i%len(visible)]
			limit := available - used
			if items[index].Constraint.HasMax {
				limit = min(limit, items[index].Constraint.Max-result[index].Size)
			}
			if limit > 0 {
				result[index].Size++
				used++
			}
			if i >= len(visible) && allCapped(result, items, visible) {
				break
			}
		}
	}
	offset := 0
	for _, index := range visible {
		result[index].Offset = offset
		offset += result[index].Size + gap
	}
	return result, nil
}

func allCapped(allocations []Allocation, items []Item, visible []int) bool {
	for _, index := range visible {
		constraint := items[index].Constraint
		if !constraint.HasMax || allocations[index].Size < constraint.Max {
			return false
		}
	}
	return true
}

func validate(c Constraint) error {
	if c.Min < 0 || c.Preferred < 0 || c.Max < 0 {
		return fmt.Errorf("constraints must be non-negative")
	}
	preferred := c.Preferred
	if preferred == 0 {
		preferred = c.Min
	}
	if preferred < c.Min || c.HasMax && (c.Max < c.Min || c.Max < preferred) {
		return fmt.Errorf("invalid min/preferred/max ordering")
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
