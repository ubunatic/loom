// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package media

import (
	"context"
	"image"
	"testing"
	"time"

	"ubunatic.com/loom"
)

// TestMediaSpecDefaultsAreConsumed overrides the media spec values. A consumer
// that reverts to a literal fails, because the override no longer comes back.
func TestMediaSpecDefaultsAreConsumed(t *testing.T) {
	saved := loom.SpeccedDefaults
	t.Cleanup(func() { loom.SpeccedDefaults = saved })
	loom.SpeccedDefaults.Media.DefaultFPS = 12
	loom.SpeccedDefaults.Media.PlayLabel = "[go]"
	loom.SpeccedDefaults.Media.PauseLabel = "[halt]"

	poster := image.NewRGBA(image.Rect(0, 0, 2, 2))
	widget, err := newVideoWithOpener("test-video", ModeHalfblock, 0, poster, func(ctx context.Context, _ string, _ float64) (<-chan image.Image, func(), error) {
		frames := make(chan image.Image)
		go func() {
			<-ctx.Done()
			close(frames)
		}()
		return frames, func() {}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(widget.Close)

	if widget.fps != 12 {
		t.Fatalf("fps = %v, want spec media.default_fps 12", widget.fps)
	}
	if want := time.Second / 12; widget.interval != want {
		t.Fatalf("interval = %v, want %v", widget.interval, want)
	}
	if got := playControlLabel(false); got != "[go]" {
		t.Fatalf("play label = %q, want spec label", got)
	}
	if got := playControlLabel(true); got != "[halt]" {
		t.Fatalf("pause label = %q, want spec label", got)
	}
}
