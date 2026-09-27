// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package media

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestVideoStreamAdvancesAndCloses(t *testing.T) {
	for _, tool := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is unavailable: %v", tool, err)
		}
	}
	path := filepath.Join(t.TempDir(), "sample.mp4")
	cmd := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=c=red:s=16x16:d=0.5", "-an", "-c:v", "mpeg4", "-pix_fmt", "yuv420p", "-y", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("ffmpeg cannot create the test video: %v: %s", err, output)
	}

	widget, err := NewVideo(path, ModeHalfblock, 8)
	if err != nil {
		t.Fatal(err)
	}
	defer widget.Close()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && widget.image.Bounds().Dx() == 1 {
		widget.Tick(time.Now())
		time.Sleep(10 * time.Millisecond)
	}
	if got := widget.image.Bounds().Dx(); got != 16 {
		t.Fatalf("video frame width = %d, want 16", got)
	}
	widget.Close()
	if got := widget.TickInterval(); got != 0 {
		t.Fatalf("closed video tick interval = %s, want zero", got)
	}
}
