# Terminal Repaint Probe

The standalone [`loom-repaint-probe`](../cmd/loom-repaint-probe/README.md) command compares terminal behavior under different repaint strategies. It writes directly to a raw terminal and does not depend on Loom's widget renderer. It is separate from `loom-probe`, which measures terminal cursor advance after emoji sequences.

## Rendering model

The preferred rectangle is 160 columns by 67 rows, centered and clamped to the available terminal size. A terminal resize updates the geometry and clears the screen once before repainting, even when per-frame clearing is disabled.

The background and clear controls are independent:

- `--clear 1` emits a full-screen clear before each frame. This overrides incremental savings because the screen must then be repainted.
- `--clear 0` is the default. The rectangle is still painted every frame. With `--background fill`, the outside is painted too; with `--background clear`, the outside is left as-is.
- Resize performs a one-time clear independently of the `--clear` setting.

The default paint strategy is `rows`. `spans` and `cells` provide alternative output patterns. The alternate screen and incremental row rendering are enabled by default. `--incremental=false` selects full output, while the `i` key toggles the mode interactively.

Incremental rendering splits the probe's cursor-addressed ANSI frame into terminal rows, compares each row with its previous output, and writes only changed rows. The row cache is invalidated on resize or render-setting changes. Keep screen writes addressed with the shared cursor helper so the row splitter can associate output with terminal rows. A full-screen clear forces all rows to be emitted.

This is intentionally an optional mode. Full-frame output remains useful for observing terminals under large writes, while incremental output models a common application rendering strategy.

## Animations

`--anim` parsing, aliases, animation painting, and HUD background colors use the registry in `cmd/loom-repaint-probe/astra.go`. Add new modes there with their accepted names and rendering functions; do not add mode-specific cases to the parser or frame loop. Astra is the default animation. Tuning keeps the selected animation fixed so animation cost does not vary between mode candidates.

## Measurements

HUD values are assembled before the current frame is written, so sampled values describe completed frames through the preceding frame.

| Value | Computation |
|---|---|
| Measured FPS | Completed frames divided by elapsed time since the last settings change or resize. This is cumulative, not a moving average. |
| Compose and write main values | Arithmetic mean of per-frame compose or write durations sampled in the preceding one-second window. |
| Compose and write parenthesized averages | Cumulative average duration since the last settings change or resize. |
| Maximum write | Largest completed output-write duration since the last settings change or resize. |
| Incremental rows | The last completed frame's emitted row count over terminal height, plus the arithmetic mean of per-frame row percentages from the preceding second. A full clear counts as all rows. |
| Bytes | Byte length of the preceding composed frame. |

These measurements cover work in the process and the terminal write call. They cannot tell when a terminal or physical display has finished presenting the frame.

Changing a render setting or terminal size resets the measurement window and row cache. `--tune` compares the baseline with one-at-a-time changes to unpinned paint, background, clear, incremental, and alternate-screen modes. Explicit mode flags pin those dimensions. `--out` saves the ranked report while the same report is printed to the terminal.

## Session notes

`make install` installs `loom-repaint-probe`. Its command name distinguishes the full-screen repaint benchmark from the existing cursor-advance `loom-probe`. To build it without installing, run `go build -o /tmp/loom-repaint-probe ./cmd/loom-repaint-probe`.
