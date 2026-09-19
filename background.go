// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"time"
)

// Background paints behind a widget tree. Backgrounds must only write inside
// the supplied rectangle; Pane invokes them before the foreground widget.
type Background interface {
	DrawBackground(*Canvas, Rect)
}

// ImageBackground renders an image as a dimmed Braille texture. It is safe to
// reuse across frames and resizes because the source image is immutable.
type ImageBackground struct {
	image     image.Image
	dim       float64
	threshold float64
	glyphs    []string
}

// AstraBackground renders a deterministic, slowly changing star field.
type AstraBackground struct{}

// NewAstraBackground returns the spec-driven Astra star field background.
func NewAstraBackground() *AstraBackground { return &AstraBackground{} }

func (AstraBackground) BackgroundInterval() time.Duration {
	return SpeccedBackground.RedrawInterval
}

// DrawBackground renders one Astra animation frame.
func (AstraBackground) DrawBackground(c *Canvas, r Rect) {
	AstraBackground{}.DrawBackgroundAt(c, r, time.Now())
}

// DrawBackgroundAt renders a deterministic frame at now. Cell placement and
// glyph selection are stable; only the visibility phase changes over time.
func (AstraBackground) DrawBackgroundAt(c *Canvas, r Rect, now time.Time) {
	if len(SpeccedBackground.Glyphs) == 0 {
		return
	}
	tick := now.UnixNano() / int64(SpeccedBackground.RedrawInterval)
	for y := 0; y < r.H; y++ {
		for x := 0; x < r.W; x++ {
			h := astraHash(x, y)
			if float64(h%1000)/1000 >= SpeccedBackground.Density {
				continue
			}
			phase := (tick + int64(h%20)) % 20
			if phase >= 14 {
				continue
			}
			level := phase
			if level > 7 {
				level = 14 - phase
			}
			// Fade toward the actual surface color at the dim end instead of a
			// fixed dark gray, so a star disappears into whatever background
			// it sits on (e.g. mc's bright blue panes) rather than showing a
			// dark smudge. The bright peak must also scale with the surface:
			// a fixed mid-gray reads as barely visible on a bright bluish
			// panel, so the peak is pushed a fixed margin above each of the
			// background's own channels (clamped to 255), staying at least
			// as bright as the old fixed peak on dark/neutral surfaces.
			const peakFloor = 105
			const peakMargin = 130
			t := float64(level) / 7
			bgR, bgG, bgB, ok := c.Get(r.X+x, r.Y+y).Style.BG.RGB()
			var fr, fg, fb uint8
			if ok {
				peakR := peakChannel(bgR, peakFloor, peakMargin)
				peakG := peakChannel(bgG, peakFloor, peakMargin)
				peakB := peakChannel(bgB, peakFloor, peakMargin)
				fr = uint8(float64(bgR) + (peakR-float64(bgR))*t)
				fg = uint8(float64(bgG) + (peakG-float64(bgG))*t)
				fb = uint8(float64(bgB) + (peakB-float64(bgB))*t)
			} else {
				v := uint8(42 + level*9)
				fr, fg, fb = v, v, v
			}
			glyph := SpeccedBackground.Glyphs[int(h)%len(SpeccedBackground.Glyphs)]
			c.PaintDecoration(r.X+x, r.Y+y, Cell{Text: glyph, Style: Style{FG: ColorRGB(fr, fg, fb)}})
		}
	}
}

// peakChannel returns the bright-phase target for one color channel: at
// least floor, and at least margin above the surface's own channel value,
// clamped to the representable range.
func peakChannel(bg uint8, floor, margin float64) float64 {
	peak := float64(bg) + margin
	if peak < floor {
		peak = floor
	}
	if peak > 255 {
		peak = 255
	}
	return peak
}

func astraHash(x, y int) uint32 {
	return uint32(x*374761393+y*668265263) ^ uint32((x+y)*1274126177)
}

// LoadImageBackground loads PNG or JPEG data from path and returns a terminal
// background. The image is fitted to the target rectangle while preserving
// its aspect ratio.
func LoadImageBackground(path string) (*ImageBackground, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load background image: %w", err)
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode background image: %w", err)
	}
	return newImageBackground(src), nil
}

func newImageBackground(src image.Image) *ImageBackground {
	return &ImageBackground{
		image: src, dim: SpeccedBackground.DimFactor,
		threshold: SpeccedBackground.Threshold,
		glyphs:    append([]string(nil), SpeccedBackground.Glyphs...),
	}
}

// DrawBackground paints the image using the eight specced single-dot Braille
// glyphs. Each terminal cell represents a 2x4 sample block.
func (b *ImageBackground) DrawBackground(c *Canvas, r Rect) {
	if b == nil || b.image == nil || r.W <= 0 || r.H <= 0 {
		return
	}
	bounds := b.image.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		return
	}
	for y := 0; y < r.H; y++ {
		for x := 0; x < r.W; x++ {
			var peak, floor uint8
			floor = 255
			var peakIndex int
			for py := 0; py < 4; py++ {
				for px := 0; px < 2; px++ {
					ix := bounds.Min.X + (x*2+px)*bounds.Dx()/(r.W*2)
					iy := bounds.Min.Y + (y*4+py)*bounds.Dy()/(r.H*4)
					lum := luminance(b.image.At(ix, iy))
					index := py*2 + px
					if lum > peak {
						peak, peakIndex = lum, index
					}
					if lum < floor {
						floor = lum
					}
				}
			}
			contrast := peak - floor
			if float64(contrast)/255 < b.threshold || len(b.glyphs) == 0 {
				continue
			}
			if peakIndex >= len(b.glyphs) {
				peakIndex = len(b.glyphs) - 1
			}
			v := uint8(float64(contrast) * b.dim)
			c.PaintDecoration(r.X+x, r.Y+y, Cell{Text: b.glyphs[peakIndex], Style: Style{FG: ColorRGB(v, v, v), Dim: true}})
		}
	}
}

func luminance(c color.Color) uint8 {
	r, g, b, _ := c.RGBA()
	return uint8((299*r + 587*g + 114*b) / 1000 >> 8)
}
