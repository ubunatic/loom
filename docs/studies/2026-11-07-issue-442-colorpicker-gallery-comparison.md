# Issue 442: ColorPicker gallery comparison

## Scope

Add a gallery-only visual comparison below the existing interactive ColorPicker. Keep one picker and preserve its current foreground/background selection and keyboard, focus, and mouse behavior. Do not add a semantic dim-background color or change color encoding, theme roles, or persistence.

## Rendering decisions

- **Default:** `Style{FG: ColorReset(), BG: ColorReset()}`; Dim is false.
- **Faint text:** terminal-reset foreground and background with `Dim: true`.
- **Subdued BG (theme):** terminal-reset foreground and the active theme's `NormalBG.Color()`; Dim is false. This is a sample of the existing theme surface, not a dim-background role. A theme with reset `NormalBG` may look identical to Default.
- Sample labels use the theme's `NormalFG` and `NormalBG`. Sample glyphs explicitly claim their cells so reset colors remain owned. The wrapper stores and delegates theme changes without replacing or resetting the picker.
- At widths of 64 cells or more, samples occupy three columns. Below that threshold they stack. The wrapper measures the picker plus the sample rows, confines child drawing through `Canvas.SubCanvas` for tiny allocations, and routes mouse events only within its recorded picker bounds.

## Verification plan and status

- Focused gallery and CLI PTY tests check per-cell FG/BG/Dim/claim, reset fallback, layout, bounds, theme refresh, picker state, mouse locality, and sample noninteraction.
- `go vet ./gallery ./cmd/loom` passed after fixing the gallery constructor's concrete return type.
- `make test-q1` passed; the complete output is logged at `/tmp/loom-issue442-test-q1-final.log`. The gallery package and CLI PTY package passed, along with the rest of the suite.
- `make install` passed and installed the CLI and companion applications.
- Implementation commit: `b9dc934` (`feat: add ColorPicker gallery comparisons (issue 442)`).
- Known baseline: a prior full suite hit an unrelated `internal/session` screenshot timeout. If it recurs, leave that code and assertions untouched.
