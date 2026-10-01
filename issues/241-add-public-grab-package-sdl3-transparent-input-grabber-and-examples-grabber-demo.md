# 241 — Add public grab package (SDL3 transparent input grabber) and examples/grabber demo

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: loom-games 036 (capture mode for loom-doom, where the grabber was developed)

---

/goal A public package `ubunatic.com/loom/grab` that captures raw keyboard and mouse input for a TUI through an
invisible SDL3 surface, plus an `examples/grabber` demo TUI showing fine-grained mouse position and detailed
key-down/key-up events, with tests; or stop and report when blocked on a user decision.

## 1. Problem & Motivation
Terminals deliver no key-up events, coarse cell-based mouse positions and no relative mouse motion. loom-games
issue 036 solved this for loom-doom with a small C helper (`doom-capture`) that opens an invisible SDL3 window
over the screen and forwards raw input. It lives in `loom-games/internal/doom`, uses Doom's SDL2-event ring
layout, and cannot be imported. The user wants it as a generic, public loom component.

## 2. Technical Specification / Findings
Source to extract (loom-games, after 036 M10): `doom/capture/loomdoom_capture.c` (C helper, loads SDL via
`dlopen`, no SDL headers or build-time SDL dependency, binary embedded in Go), `internal/doom/capture.go`,
`internal/doom/shm.go` (SHM ring + 64-byte status header: status, key and mouse counters, regrab request, error
text), `docs/InputCapture.md`.

Design that works on GNOME Wayland (found live in 036; keep it):
- SDL3 borderless `SDL_WINDOW_TRANSPARENT` window sized to the display bounds ("windowed fullscreen"). The
  fullscreen window state is drawn opaque by GNOME. Fill with pixel value 0: Wayland uses premultiplied alpha, so
  non-zero RGB at alpha 0 shows as a glow.
- Re-present on RESIZED / PIXEL_SIZE_CHANGED / EXPOSED, or the visible and grabbable area stays stale.
- Wayland pointer lock only holds while the pointer is over the surface, hence the full-screen surface; shrinking
  the window loses the mouse, keyboard focus survives.
- Release on focus loss and on a release key (hide window, stop relative mode); regrab on request; quit keys.
- SDL2 path (via `sdl2-compat`) is a fallback without transparency.

Public API sketch (to refine in the plan step): `grab.Start(ctx, grab.Options) (*Grabber, error)`,
`Grab()`, `Release()`, `Close()`, `Events() <-chan Event` with key down/up (scancode, keycode, modifiers, repeat),
mouse motion (absolute position and float relative motion), buttons, wheel; `Status()` (grab, focus, relative,
counters, error). Options: release/quit keys, tint and trace for debugging. Linux only; other platforms return a
clear "unsupported" error.

Open points for the plan: the ring/event format (a grab-owned format vs. keeping SDL2 event bytes for Doom), how
the embedded helper binary is built and kept in sync (Makefile target, `go generate`), and whether loom-games
then imports `ubunatic.com/loom/grab` (follow-up in loom-games, via go.work during development).

## 3. Implementation & Verification Plan
- M1: `grab` package + C helper moved into loom, Go API, unit tests (ring encoding, status header, option
  handling, SDL3 constants/function names pinned against upstream headers as in loom-games 036).
- M2: `examples/grabber` demo TUI: live mouse position (absolute and accumulated relative, sub-pixel), button and
  wheel state, and a scrolling log of key-down/key-up events with timestamps, scancode, key name, modifiers and
  repeat flag; status line with grab/focus state and the controls (grab, release, quit).
- Live acceptance (user, GNOME Wayland): the demo grabs at start without a click, shows key-up events and smooth
  mouse values, releases on focus loss and release key, regrabs, quits cleanly.
- Follow-up (separate, loom-games): switch loom-doom capture onto `ubunatic.com/loom/grab`.
