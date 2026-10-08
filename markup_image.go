// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// .tui image tags (issue 311):
//
//	[image src="logo.png" width=24 mode=half alt="logo"]
//
// An image tag stands alone on its line and has no closing tag. ParseMarkup
// hands it to MarkupOptions.Image, which returns the rendered rows; the root
// package does not decode images (loom/media renders them with cati). Without
// a hook the tag renders its alt text as one dim line. Modes and size limits
// are in spec/markup.yaml.

// MarkupImage is a parsed [image] tag.
type MarkupImage struct {
	Src, Alt      string
	Width, Height int    // cells; 0 = unset
	Mode          string // renderer from spec/markup.yaml: halfblock, quadblock, sextant
	Line          int
}

// MarkupOptions configures ParseMarkupWith.
type MarkupOptions struct {
	BaseDir string // resolves relative src; "" = current directory
	// Image renders an [image] tag into lines. nil: the tag renders its alt
	// text as one dim line.
	Image func(img MarkupImage, baseDir string) ([]RichLine, error)
}

type markupImageSpec struct {
	modes               []string // mode names, default first
	renderer            map[string]string
	maxWidth, maxHeight int
}

var markupImageSettings = func() markupImageSpec {
	var spec struct {
		Image struct {
			Modes     yaml.Node `yaml:"modes"`
			MaxWidth  int       `yaml:"max_width"`
			MaxHeight int       `yaml:"max_height"`
		} `yaml:"image"`
	}
	if err := yaml.Unmarshal(markupSpecYAML, &spec); err != nil {
		panic(fmt.Sprintf("loom: parse spec/markup.yaml: %v", err))
	}
	out := markupImageSpec{renderer: map[string]string{}, maxWidth: spec.Image.MaxWidth, maxHeight: spec.Image.MaxHeight}
	pairs := spec.Image.Modes.Content
	for i := 0; i+1 < len(pairs); i += 2 {
		out.modes = append(out.modes, pairs[i].Value)
		out.renderer[pairs[i].Value] = pairs[i+1].Value
	}
	if len(out.modes) == 0 || out.maxWidth < 1 || out.maxHeight < 1 {
		panic("loom: parse spec/markup.yaml: image needs modes, max_width and max_height")
	}
	return out
}()

// MarkupImageMaxSize returns the largest image size in cells that an
// [image] tag may request (spec/markup.yaml). Image hooks cap the native
// image size by it when width and height are unset.
func MarkupImageMaxSize() (width, height int) {
	return markupImageSettings.maxWidth, markupImageSettings.maxHeight
}

// parseMarkupImageTag parses an image tag line.
func parseMarkupImageTag(line string, lineNo int) (MarkupImage, error) {
	settings := markupImageSettings
	img := MarkupImage{Mode: settings.renderer[settings.modes[0]], Line: lineNo}
	options, err := parseMarkupBlockOptions(line, "image", lineNo)
	if err != nil {
		return img, err
	}
	hasSrc := false
	for _, opt := range options {
		fail := func(format string, args ...any) error {
			return &MarkupError{lineNo, opt.col, fmt.Sprintf(format, args...)}
		}
		size := func(limit int) (int, error) {
			n, err := strconv.Atoi(opt.value)
			if err != nil || n < 1 || n > limit {
				return 0, fail("%s must be a number from 1 to %d, got %q", opt.key, limit, opt.value)
			}
			return n, nil
		}
		switch opt.key {
		case "src":
			if opt.value == "" {
				return img, fail("src must not be empty")
			}
			img.Src, hasSrc = opt.value, true
		case "alt":
			img.Alt = opt.value
		case "width":
			if img.Width, err = size(settings.maxWidth); err != nil {
				return img, err
			}
		case "height":
			if img.Height, err = size(settings.maxHeight); err != nil {
				return img, err
			}
		case "mode":
			renderer, ok := settings.renderer[strings.ToLower(opt.value)]
			if !ok {
				return img, fail("unknown image mode %q (want %s)", opt.value, strings.Join(settings.modes, ", "))
			}
			img.Mode = renderer
		default:
			return img, fail("unknown image option %q", opt.key)
		}
	}
	if !hasSrc {
		return img, &MarkupError{lineNo, 1, "image tag needs src"}
	}
	return img, nil
}

// renderMarkupImage returns the lines for an image tag: the hook's output,
// or the alt text without a hook.
func renderMarkupImage(img MarkupImage, opts MarkupOptions) ([]RichLine, error) {
	if opts.Image == nil {
		alt := img.Alt
		if alt == "" {
			alt = "[image: " + img.Src + "]"
		}
		return []RichLine{{Spans: []RichSpan{{Text: alt, Style: Style{Dim: true}}}}}, nil
	}
	lines, err := opts.Image(img, opts.BaseDir)
	if err != nil {
		return nil, &MarkupError{img.Line, 1, fmt.Sprintf("image %q: %v", img.Src, err)}
	}
	return lines, nil
}
