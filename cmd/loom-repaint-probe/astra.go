// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"strings"
	"time"
)

// This small standalone renderer mirrors Loom's Astra background without
// importing Loom into the terminal repaint probe.
var astraGlyphs = [...]string{"⠁", "⠂", "⠄", "⠈", "⠐", "⠠", "⡀", "⢀"}

const (
	astraInterval = 150 * time.Millisecond
	astraDensity  = 0.16
	astraPeak     = 255
	rectSurface   = 64
)

type animationDefinition struct {
	name       string
	aliases    []string
	paint      func(*strings.Builder, int, int, int, int, time.Time)
	background func(int, int, time.Time) (uint8, uint8, uint8)
}

var animationModes = []animationDefinition{
	{name: "none", aliases: []string{"off"}, background: rectBackground},
	{name: "Astra", paint: writeAstra, background: rectBackground},
	{name: "rainbow", paint: writeRainbow, background: rainbowBackground},
}

func animationMode(index int) animationDefinition {
	if index < 0 || index >= len(animationModes) {
		return animationModes[animationNone]
	}
	return animationModes[index]
}

func writeAstra(out *strings.Builder, left, top, width, height int, now time.Time) {
	tick := now.UnixNano() / int64(astraInterval)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			h := astraHash(x, y)
			if float64(h%1000)/1000 >= astraDensity {
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
			channel := uint8(rectSurface + (astraPeak-rectSurface)*level/7)
			out.WriteString(cursor(top+y+1, left+x+1))
			out.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%d;48;2;%d;%d;%dm", channel, channel, channel, rectSurface, rectSurface, rectSurface))
			out.WriteString(astraGlyphs[int(h)%len(astraGlyphs)])
		}
	}
}

func writeRainbow(out *strings.Builder, left, top, width, height int, now time.Time) {
	if width <= 0 || height <= 0 {
		return
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b := rainbowBackground(x, width, now)
			out.WriteString(cursor(top+y+1, left+x+1))
			out.WriteString(fmt.Sprintf("\x1b[48;2;%d;%d;%dm ", r, g, b))
		}
	}
}

func animationBackground(animation, x, width int, now time.Time) (uint8, uint8, uint8) {
	mode := animationMode(animation)
	if mode.background == nil {
		return rectSurface, rectSurface, rectSurface
	}
	return mode.background(x, width, now)
}

func rectBackground(_, _ int, _ time.Time) (uint8, uint8, uint8) {
	return rectSurface, rectSurface, rectSurface
}

func rainbowBackground(x, width int, now time.Time) (uint8, uint8, uint8) {
	if width <= 0 {
		return rectSurface, rectSurface, rectSurface
	}
	shift := int(now.UnixMilli()/30) % 1536
	phase := (shift + x*1536/width) % 1536
	return rainbowColor(phase)
}

func rainbowColor(phase int) (uint8, uint8, uint8) {
	segment, offset := phase/256, uint8(phase%256)
	switch segment {
	case 0:
		return 255, offset, 0
	case 1:
		return 255 - offset, 255, 0
	case 2:
		return 0, 255, offset
	case 3:
		return 0, 255 - offset, 255
	case 4:
		return offset, 0, 255
	default:
		return 255, 0, 255 - offset
	}
}

func animationName(animation int) string {
	return animationMode(animation).name
}

func astraHash(x, y int) uint32 {
	return uint32(x*374761393+y*668265263) ^ uint32((x+y)*1274126177)
}
