# 041 — v0 split-pane focus + key routing for Frame/Box

**Status**: Closed — implemented and verified in fec5aa7
**Priority**: P2 (Medium)
**Severity**: Enhancement
**Category**: Feature
**Related**: `frame.go` (`Frame`, `Frame.HandleKey`, `Frame.Layout`, `Box`, `Box.HandleKey`, `Box.Draw`), `choice.go` (`Choice.Focused`/`SetFocus`, `viewOffset` virtual scrolling), `view.go` (`View.Scroll`/`HandleKey`), `widget.go` (`Widget` interface), `pane.go` (`Pane.Run`, `RunPane`), `examples/monitor`, sibling project `../fdapps` (`internal/tui`, hand-rolled split list+details TUI built on `ubunatic.com/cati`, not on loom)

---

## 1. Problem & Motivation

Raised (2026-09-14) while scoping a `cati/examples` image-browser demo: a
file list on the left, a live image preview on the right, arrow keys move
the list selection and update the preview, Tab (or similar) moves focus
between panes. The sibling project `fdapps` already builds exactly this
shape (`internal/tui`: filtered app list + icon-preview details pane,
`list`/`full`/`split` layout toggle) but hand-rolls it directly against
`golang.org/x/term` instead of using loom — because loom currently has no
support for it.

Investigation shows loom already has most of the hard parts:

- **Layout** — `Frame`/`Box` already arrange boxes side by side or stacked
  (`Frame.Layout`, dynamic width/height via `layout.Plan`, `Breakpoint` for
  narrow-terminal stacking) and compose nested widget rendering
  (`Box.Draw` → `paintClipped` → `b.Child.Draw`).
- **Per-widget scroll** — `View` already has `Scroll`/`HandleKey`
  (up/down/j/k, proportional `▐` indicator); `Choice` already has
  `viewOffset` virtual scrolling for long lists.
- **Per-widget focus flag** — `Choice` already exposes `Focused()` /
  `SetFocus(bool)` and dims its border when unfocused. Nothing sets it to
  `false` today because nothing composes two focusable widgets yet.

What's missing is entirely in `Frame`/`Box`, not in the individual widgets:

- `Box.HandleKey` is a permanent stub — `func (b *Box) HandleKey(KeyEvent)
  bool { return false }` ("leaves static boxes inert" by design).
- `Frame.HandleKey` only matches declared `FrameAction`s (`quit`, `toggle`
  box visibility) and drops every other key — it never forwards to
  `Box.Child.HandleKey`.
- `Frame` has no concept of "which box is focused" — no focus index, no
  cycle-focus key, no call to a focused child's `SetFocus(true)` /
  `SetFocus(false)` on its siblings.

So today a `Frame` can *display* two boxes side by side (as the monitor
example does, statically), but cannot *drive* two independent interactive
widgets — arrow keys typed into a two-box Frame currently go nowhere.

## 2. Scope

**In scope (v0):**
- A focus index on `Frame` (or a small wrapper type) over the *visible*
  `Boxes` slice, with a configurable cycle-focus key (Tab / Shift-Tab is
  the obvious default) that:
  - skips `Hidden` boxes,
  - calls `SetFocus(true)`/`SetFocus(false)` on children that implement an
    optional `Focuser` interface (`Focused() bool; SetFocus(bool)` — the
    shape `Choice` already has) via a type assertion, no-op otherwise,
  - re-picks a sane focus target if the currently focused box becomes
    hidden (e.g. via an existing `toggle` `FrameAction`).
- `Frame.HandleKey` falling through to the focused box's
  `Child.HandleKey(k)` for any key not matched by a declared
  `FrameAction`, and propagating its `quit` return value.
- `Box.HandleKey` forwarding to `b.Child.HandleKey(k)` when `Child != nil`
  (replacing the current always-`false` stub) so a `Box` used outside a
  multi-box `Frame` also works.
- Tests: focus cycles in visible-box order and wraps; hidden boxes are
  skipped and never receive keys; unhandled keys reach only the focused
  child, not siblings; a toggle that hides the focused box reassigns focus
  instead of routing keys to a hidden widget.
- One example wired up (extend `examples/monitor` or add a new
  `examples/split` demo) showing two `Choice`-like widgets — or a `Choice`
  + a static preview `View` — in one `Frame`, Tab switching focus, arrow
  keys scrolling the focused pane.

**Out of scope (defer to a later ticket if wanted):**
- Cross-pane reactivity as a loom primitive (e.g. an
  observer/on-select-changed hook so a preview pane auto-refreshes when a
  sibling list's selection changes). v0 does not need this: the composing
  app already gets a redraw tick via `Pane.RunWatch`'s `collect` callback
  and can just read `list.Selected()` each tick and push new content into
  the preview widget itself — document this pattern rather than build a
  new callback mechanism.
- Mouse-driven focus switching (click-to-focus a box) — v0 is keyboard-only
  to match the existing `FrameAction`/`Choice` keyboard-first model.
- A generic N-way split (more than the two-pane case) is not excluded by
  this design (it falls out of iterating `Boxes` regardless of count) but
  is not the driving use case and doesn't need dedicated testing here.

## 3. Difficulty Estimate

**Small–medium.** The expensive parts (box geometry/layout, canvas
sub-rect composition, per-widget scrolling, the focus flag itself) already
exist and are already tested; this ticket is glue:

1. Add a focus index + cycle-key handling to `Frame` (~30–50 lines).
2. Replace `Box.HandleKey`'s stub with a one-line forward to `Child`.
3. Extend `Frame.HandleKey` to fall through to the focused child after
   checking declared actions (~10–15 lines).
4. Define (or reuse, if one already exists elsewhere in the package) a
   small `Focuser` interface and type-assert for it, so `Frame` doesn't
   need to import anything `Choice`-specific.
5. Tests covering focus cycling, hidden-box skipping, and key routing
   isolation (~4–6 table-driven cases, following the style already used in
   `frame_actions_test.go`/`frame_layout_test.go`).
6. Update one example (or add a minimal new one) to prove it end-to-end.

Rough sizing: comparable to issue 040 (treemap theming, closed) —
half a day to a day of focused work for someone already familiar with
`frame.go`, most of it in step 5 (getting the focus/visibility edge cases
right) rather than step 1–3 (which are mechanical). No changes needed to
`choice.go`, `view.go`, `canvas.go`, or `pane.go` — this is entirely a
`Frame`/`Box` wiring change.

## 4. Acceptance Criteria

- [x] `Frame` tracks a focused box among visible boxes; a configurable key
      (default Tab) cycles focus forward, wrapping; Shift-Tab (or a second
      configurable key) cycles backward.
- [x] Unhandled keys route to the focused box's `Child.HandleKey`; keys
      matched by a declared `FrameAction` are never forwarded.
- [x] Hidden boxes are never focus targets and never receive key events;
      toggling the focused box hidden moves focus to the next visible box
      (or clears focus if none remain).
- [x] `Box.HandleKey` forwards to `Child.HandleKey` when `Child != nil`.
- [x] New tests pass; existing `frame_test.go`, `frame_actions_test.go`,
      `frame_layout_test.go` still pass unchanged (no behavior change for
      single-box or non-interactive Frames).
- [x] One example demonstrates two focusable panes with Tab-driven focus
      switching and per-pane scrolling.
