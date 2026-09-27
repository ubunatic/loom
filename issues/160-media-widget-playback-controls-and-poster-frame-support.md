# 160 — Media Widget Playback Controls and Poster Frame Support

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `media/widget.go`, [cati #066](https://codeberg.org/ubunatic/cati)

---

## 1. Problem & Motivation

Loom's `media.Widget` provides `LoadImage`, `NewImage`, and `NewVideo`. However, when building interactive media applications (such as `cati`'s `mediabrowse`), several playback and UX limitations require awkward workarounds outside Loom:

1. **Initial Blank/Empty Frame (Poster Frame Missing)**:
   - `NewVideo` initializes `still.image` with a 1x1 blank RGBA image (`image.NewRGBA(image.Rect(0, 0, 1, 1))`).
   - When a video widget is rendered before the ffmpeg frame pipeline delivers its first decoded frame via `Tick()`, the canvas clears to blank spaces, causing visible flicker/disappearance.
   - Applications that have an initial preview thumbnail (poster frame) have no way to pass it to `NewVideo` or `media.Widget`.

2. **No Exported Playback State Controls**:
   - `Widget` holds an unexported `playing bool` field, but does not provide methods like `Play()`, `Pause()`, `IsPlaying()`, `Rewind()`, or `Restart()`.
   - Host applications must manually track external playing flags, wrap ticker intervals, and destroy/recreate widgets to restart ended streams.

## 2. Technical Specification / Findings

- In `media/widget.go`:
  - `NewVideo(path string, mode Mode, fps float64)` currently opens the stream and assigns a 1x1 empty placeholder image.
  - Adding `NewVideoWithPoster(path string, mode Mode, fps float64, poster image.Image) (*Widget, error)` (or allowing `SetImage` / poster configuration) allows instant rendering of the first frame with zero blank frame flicker.
  - Adding playback control methods:
    - `(w *Widget) Pause()`
    - `(w *Widget) Play()`
    - `(w *Widget) IsPlaying() bool`
    - `(w *Widget) Restart() error`

## 3. Implementation & Verification Plan

### Goal
Add native poster frame support and playback controls to `codeberg.org/ubunatic/loom/media.Widget`.

### Acceptance Criteria
- [ ] `NewVideoWithPoster` (or equivalent option) renders the poster image immediately before the first stream frame arrives.
- [ ] `Play()`, `Pause()`, and `IsPlaying()` cleanly toggle frame advancement and ticker interval without destroying the widget.
- [ ] Unit tests in `media/widget_test.go` verifying poster display before first tick, pause/resume behavior, and stream restart.
