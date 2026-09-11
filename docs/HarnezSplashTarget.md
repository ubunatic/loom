---
title: Harnez Splash UI Target
weight: 41
---

# Harnez Splash UI Target

This document records the startup splash and loading screen target for Loom.
It serves as the reference target for centered flex-layouts, braille activity
animations, determinate progress bars, status pill clusters, and transition
lifecycles.

## Screenshot

![Harnez usage splash screen](assets/harnez-usage-splash.png)

## Monochrome reference

In-progress fetch state:

```text
                    ⠙  harnez usage

              [⣿⣿⣿⣿⡇                   ]

                  fetching claude...
               ● mic  ✳ claude  ֍ codex  Λ agy

                      Esc to skip
```

Completed initialization state:

```text
                    ⠛  harnez usage

              [::::::::::::::::::::::::::::::::]

                      agy done
               ● mic  ✳ claude  ֍ codex  Λ agy

                      Esc to skip
```

## UI model implied by the target

The splash screen is a vertically and horizontally centered viewport container
with discrete stateful sub-components:

```text
splash screen
└── centered viewport container (flex column, align center)
    ├── title line
    │   ├── animated spinner / activity glyph (braille sequence)
    │   └── application title ("harnez usage")
    ├── progress / activity bar
    │   ├── left bracket "["
    │   ├── determinate / animated braille fill
    │   └── right bracket "]"
    ├── status line
    │   └── current task / step description (e.g. "fetching claude...", "agy done")
    ├── provider status row (flex row, centered)
    │   └── status pill items (glyph + service name + state color)
    │       ├── mic (●)
    │       ├── claude (✳)
    │       ├── codex (֍)
    │       └── agy (Λ)
    └── footer / dismissal hint
        └── keyboard action hint ("Esc to skip" / "Press Enter to continue")
```

## State & transition lifecycle

The splash screen models an asynchronous initialization pipeline:

```text
Init Start ──► Fetch Providers (parallel/sequential) ──► Complete ──► Transition to Dashboard
     │                    │                                   │
     └────────────────────┴─── [Esc / Keypress] ──────────────┴──────► Immediate Dashboard
```

1. **Spinner & Progress**: The header spinner animates continuously at a fixed
   tick rate while the bar reflects either aggregate loading percentage or an
   activity pulse.
2. **Provider Indicators**: Each service pill transitions through lifecycle
   states:
   - *Pending*: Dim / muted gray
   - *Active / Fetching*: Accent / animated glyph
   - *Ready / Success*: Green symbol
   - *Failed / Skipped*: Warning / error color
3. **Completion & Dismissal**: When all providers finish (or the user presses
   `Esc`), the splash view transitions cleanly to the main dashboard view.

## Shipped Loom implementation

The splash target is implemented in the Loom library core and demonstrated in `examples/splash/`:

- **Layout & Centering**: [`loom.AlignBox`](../align.go), [`loom.NewCenter`](../align.go), [`layout.AlignOffset`](../layout/layout.go)
- **Animation & Graphs**: [`graph.SpinnerGlyph`](../graph/spinner.go), [`graph.RenderBracketedBar`](../graph/bracketed.go)
- **Status Badges**: [`loom.ProviderPill`](../pill.go), [`loom.PillCluster`](../pill.go)
- **Lifecycle Coordination**: [`loom.SplashController`](../splash.go)
- **View Widget**: [`loom.SplashView`](../splash_view.go)
- **Specced Defaults**: [`spec/defaults.yaml`](../spec/defaults.yaml), [`loom.SpeccedDefaults`](../defaults.go)
- **Runnable Demo**: [`examples/splash/main.go`](../examples/splash/main.go)
- **Case Study**: [`studies/2026-09-11-splash-screen-architecture-and-runtime-safeguards.md`](studies/2026-09-11-splash-screen-architecture-and-runtime-safeguards.md)

```text
examples/splash/
├── main.go        # interactive CLI with --watch, --width, --height
├── main_test.go   # CLI argument and show-once golden tests
└── README.md      # usage guide
```

## Acceptance verification

The Loom splash screen implementation is verified by automated tests:

1. **Centering**: The splash cluster remains centered both vertically and
   horizontally across changing terminal dimensions (`TestSplashViewCenteringDimensions`).
2. **Braille & Unicode Safety**: Braille glyphs and unicode symbols (`⠙`, `⣿`, `●`,
   `✳`, `֍`, `Λ`) do not cause layout skew or ANSI column calculation bugs (`TestPillDimensions`, `TestSpinnerFrames`).
3. **Width Stability**: Step message updates (e.g. transitioning from `fetching claude...`
   to `agy done`) do not alter the bar width or horizontal center (`TestSplashViewGoldenInProgress`, `TestSplashViewGoldenCompleted`).
4. **Key Dismissal**: Pressing `Esc`, `q`, `Enter`, or `Ctrl-C` immediately dismisses the
   splash screen via controller delegation and library-level driver safeguards (`TestSplashViewKeyHandling`, `TestPaneHandleKeyFallback`).
5. **Decoupled Animation**: Spinner ticks and background provider initialization
   routines execute independently without blocking or starvation (`TestSplashControllerProgression`).

