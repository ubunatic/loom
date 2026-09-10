// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package measure

import "fmt"

// Insets reserves cells around content. All values must be non-negative.
type Insets struct {
	Top, Right, Bottom, Left int
}

// Width returns the horizontal cells occupied by the insets.
func (i Insets) Width() int { return i.Left + i.Right }

// Height returns the vertical cells occupied by the insets.
func (i Insets) Height() int { return i.Top + i.Bottom }

// Limit describes a minimum, preferred, and optional maximum dimension.
type Limit struct {
	Min       int
	Preferred int
	Max       int
	HasMax    bool
}

// Constraints bounds measured outer dimensions. Preferred values are hints;
// Min and Max are hard limits.
type Constraints struct {
	Width  Limit
	Height Limit
}

// Validate checks dimension ordering and rejects ambiguous negative values.
func (c Constraints) Validate() error {
	for name, l := range map[string]Limit{"width": c.Width, "height": c.Height} {
		if l.Min < 0 || l.Preferred < 0 || l.Max < 0 {
			return fmt.Errorf("%s constraints must be non-negative", name)
		}
		if l.Preferred > 0 && l.Preferred < l.Min {
			return fmt.Errorf("%s preferred %d is below minimum %d", name, l.Preferred, l.Min)
		}
		if l.HasMax && (l.Max < l.Min || l.Max < l.Preferred) {
			return fmt.Errorf("%s maximum %d is below minimum/preferred", name, l.Max)
		}
	}
	return nil
}

// TextSize returns the visible size of text. When maxWidth is positive,
// lines wrap at terminal-cell boundaries; explicit newlines are preserved.
// A wide cluster is kept intact even when it exceeds maxWidth.
func TextSize(text string, maxWidth int) Size {
	if maxWidth <= 0 {
		return Lines(text)
	}
	lines := plainTerminalLines(text)
	result := Size{Height: 0}
	for _, line := range lines {
		clusters := Clusters(line)
		if len(clusters) == 0 {
			result.Height++
			continue
		}
		used := 0
		for _, cluster := range clusters {
			w := StringWidth(cluster)
			if used > 0 && used+w > maxWidth {
				if used > result.Width {
					result.Width = used
				}
				result.Height++
				used = 0
			}
			used += w
		}
		if used > result.Width {
			result.Width = used
		}
		result.Height++
	}
	if result.Height == 0 {
		result.Height = 1
	}
	return result
}

// MeasureText returns an outer size satisfying c. availableWidth, when
// positive, is a hard width supplied by a parent planner. Insets are included
// in the returned dimensions.
func MeasureText(text string, availableWidth int, insets Insets, c Constraints) (Size, error) {
	if insets.Top < 0 || insets.Right < 0 || insets.Bottom < 0 || insets.Left < 0 {
		return Size{}, fmt.Errorf("insets must be non-negative")
	}
	if err := c.Validate(); err != nil {
		return Size{}, err
	}
	if availableWidth > 0 && availableWidth < c.Width.Min {
		return Size{}, fmt.Errorf("available width %d is below minimum %d", availableWidth, c.Width.Min)
	}
	if c.Width.HasMax && c.Width.Max < insets.Width() || c.Height.HasMax && c.Height.Max < insets.Height() {
		return Size{}, fmt.Errorf("constraints cannot contain insets")
	}
	contentWidth := availableWidth - insets.Width()
	if contentWidth <= 0 {
		contentWidth = c.Width.Preferred - insets.Width()
	}
	natural := TextSize(text, contentWidth)
	width := natural.Width + insets.Width()
	if c.Width.Preferred > width {
		width = c.Width.Preferred
	}
	if availableWidth > 0 && width > availableWidth {
		width = availableWidth
	}
	width = clamp(width, c.Width.Min, c.Width.Max, c.Width.HasMax)
	contentWidth = width - insets.Width()
	if contentWidth < 0 {
		return Size{}, fmt.Errorf("available width cannot contain insets")
	}
	natural = TextSize(text, contentWidth)
	height := natural.Height + insets.Height()
	if c.Height.Preferred > height {
		height = c.Height.Preferred
	}
	height = clamp(height, c.Height.Min, c.Height.Max, c.Height.HasMax)
	return Size{Width: width, Height: height}, nil
}

func clamp(value, minValue, maxValue int, hasMax bool) int {
	if value < minValue {
		value = minValue
	}
	if hasMax && value > maxValue {
		value = maxValue
	}
	return value
}
