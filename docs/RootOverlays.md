---
title: Root Overlays
weight: 43
---

# Root Overlays

Some UI (help, confirmation dialogs, global notices) must render modally over
the *entire* active canvas and capture input before the widget tree, even when
the requesting widget is nested arbitrarily deep inside `paintClipped`
sub-canvases (splits, tabs, boxes). A widget cannot draw outside its own
clipped region, so this cannot be solved by the widget itself — it needs a
cooperating root owner.

## The pattern

1. A package-level function-pointer hook (e.g. `paneHelpRequest
   func([]Cmd)`) is the injection point. It is `nil` by default.
2. The root owner (`Pane.run`) installs the hook for the duration of its event
   loop and restores the previous value via `defer` on exit — this keeps
   nested/hosted panes and tests from clobbering each other's hook.
3. A descendant widget (`cmdBar.showHelp`) calls the hook instead of
   drawing its own popup, when one is installed. The root owner stores the
   resulting state (e.g. `Pane.help *Popup`) and:
   - draws it last, after `root.Draw`, directly on the full canvas — not
     through any child's clipped sub-canvas;
   - checks it first in key/mouse handling, before forwarding to
     `root.HandleKey`/`HandleMouse`.
4. Widgets keep a local fallback (e.g. `cmdBar.help`, drawn via
   `drawHelp`/`handleHelp`) for when no root hook is installed — headless
   tests, custom hosting, or standalone widget use outside a `Pane`.

This gives three cooperating tiers, tried in order: an injected test/host
runner (`helpRunner`) → the root-overlay hook (`paneHelpRequest`) → the
widget-local inline popup. Only the tier that applies is active at once.

## Why not just draw from the widget

Loom's compositor gives every composite widget (`Box`, `Frame`, split panes) a
bounded sub-`Canvas` via `paintClipped`, merged back cell-by-cell into the
parent. A widget writing "outside" its own rect is structurally impossible —
the merge only copies the widget's own bounded region. A modal that must cover
a split layout's *other* pane has to be owned and drawn by something above
all children, which in Loom is the `Pane` running the event loop.

## Precedent: `:help` (issues 066, 069)

- Issue 066 first fixed a **safety** bug: the default help path opened a
  second `Pane` (second `/dev/tty`, second raw-mode session) while the outer
  one was still running, racing the tty reader. Fixed two ways: architecturally
  (default path no longer opens a second `Pane` at all — it renders an inline
  `Popup`) and defensively (`paneOwnership` guard rejects a nested `Pane.New`
  regardless).
- Issue 069 then fixed the **scope** gap the widget-local popup left behind:
  in a split layout, the inline popup was still confined to whichever child
  pane's `cmdBar` triggered it, not modal over the whole layout. This is the
  root-overlay pattern above, applied via `paneHelpRequest`.

Apply the same pattern to any future root-modal need (confirmation dialogs,
global notices) rather than inventing a new plumbing mechanism per feature.

## Known follow-up

Issue 070: the shared popup/text layout path does not wrap or clip long help
lines at narrow canvas widths, so text can overrun the modal border. This is a
text-layout bug in the popup content path, not a flaw in the overlay-ownership
pattern itself.
