---
title: Loom CLI
weight: 75
---

# Loom CLI

`loom info` prints the loom version, terminal dimensions when available,
terminal environment variables, detected colour profile, a Unicode width
sample, and the detected graphics rendering path. Pixel dimensions are shown
as unknown when the terminal does not report them.

The mouse tracking line describes Loom's SGR 1003 mode request; it does not
claim the current terminal supports every mouse reporting mode.

`loom info --watch` opens a full-screen view, enables SGR any-motion mouse
tracking, and refreshes terminal and environment values once per second and
after resizes. The pointer line
shows the widget's resolved 0-based cell coordinates and the original 1-based
SGR report coordinates. A `+` marks the resolved cell in the view. Press `q` or
`Esc` to exit.

The command reuses `loom.DetectColorProfile` and `measure.DetectRenderPath` so
reported capabilities follow the same detection rules as Loom rendering.
