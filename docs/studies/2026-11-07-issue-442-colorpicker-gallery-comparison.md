# Issue 442: ColorPicker gallery comparison

## Scope

Add a gallery-only visual comparison below the existing interactive ColorPicker. Keep one picker and preserve its current foreground/background selection and keyboard, focus, and mouse behavior. Do not add a semantic dim-background color or change color encoding, theme roles, or persistence.

## Rendering decisions

- **Default:** `Style{FG: ColorReset(), BG: ColorReset()}`; Dim is false.
- **Faint text:** terminal-reset foreground and background with `Dim: true`.
- **Subdued BG (theme):** terminal-reset foreground and the active theme's `NormalBG.Color()`; Dim is false. This is a sample of the existing theme surface, not a dim-background role. A theme with reset `NormalBG` may look identical to Default.
- Sample labels use the theme's `NormalFG` and `NormalBG`. Sample glyphs explicitly claim their cells so reset colors remain owned. The wrapper stores and delegates theme changes without replacing or resetting the picker.
- At widths of 64 cells or more, samples occupy three columns. Below that threshold they stack. The wrapper measures the picker plus the sample rows and routes mouse events only within its recorded picker bounds; the library picker must honor its own supplied rectangle.

## Verification plan and status

- Focused gallery and CLI PTY tests check per-cell FG/BG/Dim/claim, reset fallback, layout, bounds, theme refresh, picker state, mouse locality, and sample noninteraction.
- `go vet ./gallery ./cmd/loom` passed after fixing the gallery constructor's concrete return type.
- `make test-q1` passed; the complete output is logged at `/tmp/loom-issue442-test-q1-final.log`. The gallery package and CLI PTY package passed, along with the rest of the suite.
- `make install` passed and installed the CLI and companion applications.
- Implementation commit: `b9dc934` (`feat: add ColorPicker gallery comparisons (issue 442)`).
- Known baseline: a prior full suite hit an unrelated `internal/session` screenshot timeout. If it recurs, leave that code and assertions untouched.

## M2: library bounds regression

- Before changing library code, a direct `ColorPicker.Draw` test with sentinel cells reproduced row-3 writes outside offset rectangles with heights 1 and 2. A zero-width draw also retained the previous mouse hitbox. The regression cases cover zero width/height, widths and heights 1/2, a normal-size offset rect, and a two-cell wide edge glyph at the right edge.
- The fix guards the preview-gap and swatch-padding writes by the local height, and records even empty rectangles so mouse input cannot target stale geometry. The gallery again draws the child directly into its caller's canvas; the standalone regression and gallery tiny/offset checks remain strict.
- Pending M2 verification: focused checks, vet, one logged `make test-q1`, and install.
