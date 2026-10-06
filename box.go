// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"bytes"
	"embed"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	"ubunatic.com/loom/measure"
)

//go:embed spec/box.yaml
var frameSpecs embed.FS

// BoxBorder specifies the one-cell border glyphs and title decoration.
type BoxBorder struct {
	TopLeft     string `yaml:"top_left"`
	TopRight    string `yaml:"top_right"`
	BottomLeft  string `yaml:"bottom_left"`
	BottomRight string `yaml:"bottom_right"`
	Horizontal  string `yaml:"horizontal"`
	Vertical    string `yaml:"vertical"`
	TitlePrefix string `yaml:"title_prefix"`
	TitleSuffix string `yaml:"title_suffix"`
}

// BoxStyle controls a box's background, border, title, and footer.
type BoxStyle struct {
	Background Style
	Border     Style
	Title      Style
	Footer     Style
}

// BoxBorderStyle specifies which glyph set to use for drawing borders.
// Values: Sharp (┌─┐│└┘), Rounded (╭─╮│╰╯), Double (╔═╗║╚╝), ASCII (+-+||++)
type BoxBorderStyle int

const (
	BoxBorderStyleSharp BoxBorderStyle = iota
	BoxBorderStyleRounded
	BoxBorderStyleDouble
	BoxBorderStyleASCII
)

// String returns the name of the border style.
func (s BoxBorderStyle) String() string {
	switch s {
	case BoxBorderStyleSharp:
		return "sharp"
	case BoxBorderStyleRounded:
		return "rounded"
	case BoxBorderStyleDouble:
		return "double"
	case BoxBorderStyleASCII:
		return "ascii"
	default:
		return "unknown"
	}
}

// ParseBoxBorderStyle parses a string into a BoxBorderStyle.
func ParseBoxBorderStyle(s string) (BoxBorderStyle, error) {
	switch strings.ToLower(s) {
	case "sharp":
		return BoxBorderStyleSharp, nil
	case "rounded":
		return BoxBorderStyleRounded, nil
	case "double":
		return BoxBorderStyleDouble, nil
	case "ascii":
		return BoxBorderStyleASCII, nil
	default:
		return BoxBorderStyleSharp, fmt.Errorf("unknown box border style: %q", s)
	}
}

// getBoxBorderGlyphs retrieves the border glyphs for a given BoxBorderStyle.
func getBoxBorderGlyphs(style BoxBorderStyle) (BoxBorder, error) {
	data, err := frameSpecs.ReadFile("spec/box.yaml")
	if err != nil {
		return BoxBorder{}, fmt.Errorf("box border spec: %w", err)
	}

	var rawSpec map[string]interface{}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&rawSpec); err != nil {
		return BoxBorder{}, fmt.Errorf("box border spec: %w", err)
	}

	// Get the styles map
	stylesRaw, ok := rawSpec["styles"]
	if !ok {
		return BoxBorder{}, fmt.Errorf("styles section not found in box border spec")
	}

	stylesMap, ok := stylesRaw.(map[string]interface{})
	if !ok {
		return BoxBorder{}, fmt.Errorf("styles is not a map in box border spec")
	}

	styleName := style.String()
	styleRaw, ok := stylesMap[styleName]
	if !ok {
		return BoxBorder{}, fmt.Errorf("box border style %q not found in spec", styleName)
	}

	styleMap, ok := styleRaw.(map[string]interface{})
	if !ok {
		return BoxBorder{}, fmt.Errorf("box border style %q is not a map", styleName)
	}

	border := BoxBorder{
		TopLeft:     getStringField(styleMap, "top_left", ""),
		TopRight:    getStringField(styleMap, "top_right", ""),
		BottomLeft:  getStringField(styleMap, "bottom_left", ""),
		BottomRight: getStringField(styleMap, "bottom_right", ""),
		Horizontal:  getStringField(styleMap, "horizontal", ""),
		Vertical:    getStringField(styleMap, "vertical", ""),
	}
	return border, nil
}

// getStringField extracts a string field from a map with a default fallback.
func getStringField(m map[string]interface{}, key string, defaultVal string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultVal
}

// Box is a titled border with padding and an optional isolated child.
// Width and Height are preferred outer dimensions used by Frame.
// YAML frames load Border from Loom's embedded spec; Go callers supply it.
type Box struct {
	ID         string    `yaml:"id"`
	Title      string    `yaml:"title"`
	Width      int       `yaml:"width"`
	Height     int       `yaml:"height"`
	Dynamic    bool      `yaml:"dynamic"`
	FillHeight bool      `yaml:"fill_height"`
	MinWidth   int       `yaml:"min_width"`
	MaxWidth   int       `yaml:"max_width"`
	MinHeight  int       `yaml:"min_height"`
	MaxHeight  int       `yaml:"max_height"`
	Padding    int       `yaml:"padding"`
	Border     BoxBorder `yaml:"-"`
	Style      BoxStyle  `yaml:"-"`
	Child      Widget    `yaml:"-"`
	Hidden     bool      `yaml:"hidden"`
	Footer     string    `yaml:"footer"`
	Rows       *Rows     `yaml:"rows"`
	childRect  Rect
}

// Measure returns the preferred outer size of the box, including border,
// padding, title/footer and known child content. A positive width wraps the
// child content before calculating its height.
func (b *Box) Measure(width int) measure.Size {
	const borderWidth, borderHeight = 2, 2
	padding := max(0, b.Padding)
	innerInsets := measure.Insets{Top: borderHeight + padding*2, Bottom: 0, Left: borderWidth + padding*2, Right: 0}
	contentWidth, contentHeight := 0, 1
	if b.Rows != nil {
		contentWidth, contentHeight = b.Rows.ContentWidth(), b.Rows.ContentHeight()
	} else {
		if child, ok := b.Child.(ContentWidther); ok {
			contentWidth = child.ContentWidth()
		}
		if child, ok := b.Child.(ContentHeighter); ok {
			contentHeight = child.ContentHeight()
		}
	}
	contentWidth = max(contentWidth, StringWidth(b.Title)+StringWidth(b.Border.TitlePrefix)+StringWidth(b.Border.TitleSuffix))
	contentWidth = max(contentWidth, StringWidth(b.Footer))
	if width > 0 {
		contentWidth = max(1, width-innerInsets.Width())
	}
	return measure.Size{Width: contentWidth + innerInsets.Width(), Height: contentHeight + innerInsets.Height() + boolInt(b.Footer != "")}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

// SetRowsValues updates the box's rows values dynamically if rows are present.
func (b *Box) SetRowsValues(values [][]string) {
	if b.Rows != nil {
		b.Rows.SetValues(values)
	}
}

// Draw paints a box. Bounds smaller than a complete border are left blank.
func (b *Box) Draw(c *Canvas, r Rect) {
	b.childRect = Rect{}
	paintClipped(c, r, func(local *Canvas) {
		w, h := local.Cols(), local.Rows()
		if w < 2 || h < 2 {
			return
		}
		local.PaintSurface(local.Bounds(), b.Style.Background)

		// Draw border and title using the new primitive.
		// The border is drawn, then the title is overlaid on the top edge.
		local.DrawBorder(Rect{X: 0, Y: 0, W: w, H: h}, BoxBorderStyleSharp, b.Style.Border)
		if b.Title != "" {
			// Write title with prefix/suffix (matching original behavior)
			titleStr := b.Border.TitlePrefix + b.Title + b.Border.TitleSuffix
			// Title is positioned at (1, 0) and styled with Title style
			fit := 0
			for _, cluster := range textClusters(titleStr) {
				clusterWidth := StringWidth(cluster)
				// Keep title within border: w-2 is the interior width (excluding corners)
				if fit+clusterWidth > w-2 {
					break
				}
				n := local.Write(1+fit, 0, cluster, b.Style.Title)
				if n <= 0 {
					break
				}
				fit += clusterWidth
			}
		}
		padding := max(0, b.Padding)
		// Compare before doubling to avoid overflow for programmatic inputs.
		if b.Child != nil && padding < (w-1)/2 && padding < (h-1)/2 {
			inner := Rect{X: 1 + padding, Y: 1 + padding, W: w - 2 - 2*padding, H: h - 2 - 2*padding}
			b.childRect = inner
			paintClipped(local, inner, func(child *Canvas) { b.Child.Draw(child, child.Bounds()) })
			if b.Footer != "" {
				writeBoundedStyled(local, inner.X, inner.Y+inner.H-1, inner.W, b.Footer, b.Style.Footer)
			}
		}
	})
}

// ConsumeKey forwards input to the child, if present.
func (b *Box) ConsumeKey(k KeyEvent) EventResult {
	if b.Child == nil {
		return Ignored()
	}
	return b.Child.ConsumeKey(k)
}

// ConsumeMouse forwards events inside the last drawn child bounds.
func (b *Box) ConsumeMouse(e MouseEvent) EventResult {
	if b.Child == nil || !b.childRect.Contains(e.X, e.Y) {
		return Ignored()
	}
	e.X -= b.childRect.X
	e.Y -= b.childRect.Y
	return b.Child.ConsumeMouse(e)
}
