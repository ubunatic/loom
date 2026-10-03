// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"strings"
	"testing"
	"time"
)

func TestFramePaintStrategies(t *testing.T) {
	for _, mode := range []int{0, 1, 2} {
		t.Run(paintModeName(mode), func(t *testing.T) {
			d := newDemo(30)
			d.width, d.height = minimumWidth+16, minimumHeight+6
			d.paintMode = mode
			frame := string(d.frame())
			if !strings.Contains(frame, "\x1b[48;2;0;0;139m") {
				t.Fatal("frame does not paint the dark-blue outside")
			}
			if !strings.Contains(frame, "\x1b[48;2;64;64;64m") {
				t.Fatal("frame does not paint the dark-gray area")
			}
			visible := stripANSI(frame)
			if !strings.Contains(visible, "FPS target 30") || !strings.Contains(visible, "measured") || !strings.Contains(visible, "avg") || !strings.Contains(visible, "q quit") {
				t.Fatal("frame does not include timing and hotkey HUD")
			}
		})
	}
}

func TestBackgroundClearOnlyPaintsArea(t *testing.T) {
	d := newDemo(30)
	d.width, d.height = minimumWidth+16, minimumHeight+6
	d.background = 1
	d.clearEach = true
	frame := string(d.frame())
	if !strings.Contains(frame, "\x1b[48;2;0;0;139m\x1b[2J") {
		t.Fatal("background-plus-clear strategy does not set blue background and clear")
	}
	if !strings.Contains(frame, cursor(4, 9)+"\x1b[48;2;64;64;64m") {
		t.Fatal("clear strategy does not paint the centered area")
	}
	if strings.Contains(frame, cursor(1, 1)) {
		t.Fatal("clear strategy unexpectedly paints the outside row")
	}
}

func TestHUDExplainsPartialRepaintWhenClearIsOff(t *testing.T) {
	d := newDemo(30)
	d.width, d.height = minimumWidth+16, minimumHeight+6
	d.background = 1
	d.clearEach = false
	if !strings.Contains(stripANSI(string(d.frame())), "outside retains previous contents") {
		t.Fatal("HUD does not explain partial repaint behavior when clear is disabled")
	}
}

func TestResizeAndAreaControls(t *testing.T) {
	d := newDemo(30)
	d.width, d.height = minimumWidth+16, minimumHeight+6
	d.key('+')
	if d.areaWidth != minimumWidth+16 || d.areaHeight != minimumHeight+6 {
		t.Fatalf("grow area = %dx%d", d.areaWidth, d.areaHeight)
	}
	d.key('-')
	if d.areaWidth != minimumWidth || d.areaHeight != minimumHeight {
		t.Fatalf("shrink area = %dx%d", d.areaWidth, d.areaHeight)
	}
	d.key('p')
	d.key('b')
	d.key('c')
	if d.paintMode != 1 || d.background != 1 || !d.clearEach {
		t.Fatalf("mode controls not applied: paint=%d background=%d clear=%t", d.paintMode, d.background, d.clearEach)
	}
}

func stripANSI(input string) string {
	var output strings.Builder
	for i := 0; i < len(input); {
		if input[i] == '\x1b' && i+1 < len(input) && input[i+1] == '[' {
			i += 2
			for i < len(input) && (input[i] < '@' || input[i] > '~') {
				i++
			}
			if i < len(input) {
				i++
			}
			continue
		}
		output.WriteByte(input[i])
		i++
	}
	return output.String()
}

func TestTimingAndByteFormatting(t *testing.T) {
	if got := formatMillis(0); got != "0.1ms" {
		t.Fatalf("formatMillis(0) = %q", got)
	}
	if got := formatMillis(1567 * time.Microsecond); got != "1.6ms" {
		t.Fatalf("formatMillis(1.567ms) = %q", got)
	}
	if got := formatBytes(32 * 1024); got != "32k" {
		t.Fatalf("formatBytes(32 KiB) = %q", got)
	}
}

func TestTargetFPSControls(t *testing.T) {
	d := newDemo(30)
	d.key(keyPageUp)
	if d.fps != 35 {
		t.Fatalf("FPS after increase = %d, want 35", d.fps)
	}
	d.fps = 120
	d.key(keyPageUp)
	if d.fps != 120 {
		t.Fatalf("FPS after upper bound = %d, want 120", d.fps)
	}
	d.fps = 1
	d.key(keyPageDown)
	if d.fps != 1 {
		t.Fatalf("FPS after lower bound = %d, want 1", d.fps)
	}
	d.fps = 35
	d.key(keyPageDown)
	if d.fps != 30 {
		t.Fatalf("FPS after decrease = %d, want 30", d.fps)
	}
}
