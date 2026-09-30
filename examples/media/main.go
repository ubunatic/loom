// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command media displays a still image through Loom's cati-backed media widget.
package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codeberg.org/ubunatic/loom"
	"codeberg.org/ubunatic/loom/media"
	"github.com/spf13/cobra"
)

type demo struct {
	image   *media.Widget
	path    string
	mode    media.Mode
	video   bool
	message string
}

func (d *demo) Draw(c *loom.Canvas, r loom.Rect) {
	c.PaintSurface(r, loom.Style{BG: loom.ColorRGB(17, 24, 32)})
	if r.H < 2 {
		return
	}
	d.image.Draw(c, loom.Rect{X: r.X, Y: r.Y, W: r.W, H: r.H - 1})
	titleRect := loom.Rect{X: r.X, Y: r.Y, W: r.W, H: 1}
	c.PaintSurface(titleRect, loom.Style{BG: loom.ColorRGB(17, 24, 32)})
	c.Write(r.X+1, r.Y, "Media Demo  (q quits)", loom.Style{FG: loom.ColorRGB(240, 240, 240), Bold: true})
	status := fmt.Sprintf("Cols: %d  Rows: %d  Mode: %s  Media: %s", c.Cols(), c.Rows(), d.mode, d.path)
	if d.message != "" {
		status = d.message
	}
	c.Write(r.X, r.Y+r.H-1, status, loom.Style{FG: loom.ColorRGB(240, 240, 240), BG: loom.ColorRGB(38, 48, 60)})
}

func (d *demo) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	switch e.Rune() {
	case 'r', 'R':
		if !d.video {
			return loom.Ignored()
		}
		if err := d.image.Restart(); err != nil {
			d.message = err.Error()
		} else {
			d.message = ""
		}
	default:
		return d.image.ConsumeKey(e)
	}
	return loom.Handled()
}
func (d *demo) ConsumeMouse(e loom.MouseEvent) loom.EventResult { return d.image.ConsumeMouse(e) }

func (d *demo) TickInterval() time.Duration { return d.image.TickInterval() }
func (d *demo) Tick(now time.Time)          { d.image.Tick(now) }

func run(path string, mode media.Mode) error {
	return runWithPoster(path, mode, "")
}

func runWithPoster(path string, mode media.Mode, posterPath string) error {
	var widget *media.Widget
	var err error
	if isVideo(path) {
		if posterPath == "" {
			widget, err = media.NewVideo(path, mode, 24)
		} else {
			var poster image.Image
			poster, err = loadPoster(posterPath)
			if err == nil {
				widget, err = media.NewVideoWithPoster(path, mode, 24, poster)
			}
		}
	} else {
		if posterPath != "" {
			return fmt.Errorf("--poster can only be used with a video")
		}
		widget, err = media.LoadImage(path, mode)
	}
	if err != nil {
		return err
	}
	defer widget.Close()
	pane, err := loom.New(1 << 16)
	if err != nil {
		return err
	}
	pane.MaxCols = 0
	pane.EnableMouse()
	defer pane.Close()
	return pane.Run(&demo{image: widget, path: path, mode: mode, video: isVideo(path)})
}

func loadPoster(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode poster %q: %w", path, err)
	}
	return img, nil
}

func newCommand() *cobra.Command {
	return newCommandWithOptionsRun(runWithPoster)
}

func newCommandWithRun(runMedia func(string, media.Mode) error) *cobra.Command {
	return newCommandWithOptionsRun(func(path string, mode media.Mode, _ string) error {
		return runMedia(path, mode)
	})
}

func newCommandWithOptionsRun(runMedia func(string, media.Mode, string) error) *cobra.Command {
	var mode string
	var posterPath string
	cmd := &cobra.Command{
		Use:   "loom-media <image-or-video> [mode]",
		Short: "Display an image or video in the terminal",
		Long:  "Display an image or video in the terminal using Loom's media widget. For videos, press p to play or pause and r to restart. Use --poster to display a PNG, JPEG, or GIF poster before the first frame.",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 || len(args) > 2 {
				return fmt.Errorf("requires an image or video path, optionally followed by a render mode")
			}
			if len(args) == 2 && mode != string(media.ModeHalfblock) {
				return fmt.Errorf("render mode supplied both positionally and with --mode")
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 2 {
				mode = args[1]
			}
			return runMedia(args[0], media.Mode(mode), posterPath)
		},
	}
	cmd.Flags().StringVar(&mode, "mode", string(media.ModeHalfblock), "render mode (halfblock, quadblock, sextant)")
	cmd.Flags().StringVar(&posterPath, "poster", "", "poster image for a video (PNG, JPEG, or GIF)")
	return cmd
}

func isVideo(path string) bool {
	switch strings.ToLower(strings.TrimPrefix(filepath.Ext(path), ".")) {
	case "mp4", "m4v", "mov", "mkv", "webm", "avi", "mpeg", "mpg":
		return true
	default:
		return false
	}
}

func main() {
	cmd := newCommand()
	cmd.SetArgs(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
