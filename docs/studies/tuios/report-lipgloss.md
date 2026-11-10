# Topic 4: Lip Gloss v2 and Glamour against loom

Ticket 457, searcher output. Scope: layers, Z-order, popovers, borders, styles, theming and Markdown,
compared with loom's RootOverlays and astra. Rows are in `features-lipgloss.yaml`.
Fetched on 2026-11-10 from `main`/`master` (unpinned; no commit SHA was resolved, so line numbers are
approximate, +-3).

## Summary

- Lip Gloss v2 (`charm.land/lipgloss/v2`, released 23 February 2026 per the Charm announcement at
  https://charm.land/blog/v2/) has a real compositor. It stacks layers by an absolute Z value, can
  hit-test the stack, and can draw onto an ultraviolet screen. This is the closest Go-lib match to loom's
  RootOverlays and astra claim-cell model.
- The overlap is partial. Lip Gloss composes strings or cell buffers. It does not route events, keep
  focus, or own modal capture. Loom's claim-cell routing and ModalTarget stay loom's job.
- Lip Gloss v2 removed the global renderer and moved adaptive colours into `compat`. Background detection
  is now explicit, and the app listens for `tea.BackgroundColorMsg`.
- Glamour v2 (`charm.land/glamour/v2`) is a Markdown-to-ANSI renderer with seven named styles. It has no
  verified auto-style option in v2 (see Gaps).

## 1. Layers and Z-order (Lip Gloss v2)

Source: `layer.go` in https://github.com/charmbracelet/lipgloss (raw fetch of `main`).

- A `Layer` holds content, a size, absolute X/Y/Z and child layers (`type Layer struct`, ~L9-17).
- Builders are chainable: `NewLayer(content, children...).X(4).Y(2).Z(1)` (README, "Compositing").
- Children are placed relative to their parent. The compositor flattens the tree into absolute
  positions, then sorts by Z: `// Sort by absolute z-index (lowest to highest for drawing)` (~L220).
- Z is per layer after flattening, so a child cannot escape a higher sibling by its parent's Z. Loom's
  RootOverlays order is the thing to compare; loom's exact rule was not checked (host to fill `loom`).

Minimal example, from `examples/canvas/main.go` (~L62-79):

```go
layers := []*lipgloss.Layer{ /* ...NewLayer(...).X(..).Y(..).Z(..) */ }
comp := lipgloss.NewCompositor(layers...)
lipgloss.Println(comp.Render())
```

Note: `Render()` returns a plain string. The caller still owns the frame write, so there is no built-in
diff against the last frame. This matters for loom's minimal-diff goal.

## 2. Hit-testing (events)

- `func (c *Compositor) Hit(x, y int) LayerHit` (~L277-292) walks layers from highest Z to lowest:
  `// Check from highest z to lowest (reverse order)`.
- `LayerHit` returns the layer ID, the layer and its bounds. This gives the topmost-layer answer for a
  mouse position in one call.
- Gap for loom: Lip Gloss does not decide what a click means (click-through, modal capture, focus
  change). Loom's claim-cell model already does that. Hit-testing is useful only if loom keeps the
  claim-cell rule as the authority and uses Lip Gloss purely as the geometry source.

## 3. Canvas (lower-level cell buffer)

- `type Canvas struct { scr uv.ScreenBuffer }` (`canvas.go` ~L17-19). It implements `uv.Screen`.
- Drawables are composed in call order, later ones on top: `// composed, meaning later drawables will
  appear "on top" of earlier ones.` (~L8-9).
- `NewCanvas` sets `c.scr.Method = ansi.GraphemeWidth` (~L26). This is a per-canvas width method and
  is a point to compare with loom's `measure.DetectZWJMode`. Verified only by reading the source line;
  behaviour under real terminals was not tested.
- Canvas and Compositor overlap. The README and example use Compositor; the source comment is the
  Canvas one. Pick one in a loom spike, not both.

## 4. Upgrade notes that matter for loom (Lip Gloss v2)

From `UPGRADE_GUIDE_V2.md`
(https://github.com/charmbracelet/lipgloss/blob/main/UPGRADE_GUIDE_V2.md):

| Change | v2 | Why it matters |
|---|---|---|
| Import path | `charm.land/lipgloss/v2` | v1 and v2 cannot share a module; check transitive deps |
| `Color` | a function, not a type | `TerminalColor` interface removed |
| Adaptive colours | moved to `compat` | reads stdin/stdout globally; README discourages it for new code |
| `HasDarkBackground()` | removed; `HasDarkBackground(in, out)` | detection must pass the streams |
| `NewStyle()` | not tied to a Renderer | styles are plain values |
| `Render()` | always full-fidelity ANSI | downsampling is the caller's step |

Example of the explicit style (README, "Advanced Color Usage"):

```go
hasDarkBG := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
lightDark := lipgloss.LightDark(hasDarkBG)
```

For Bubble Tea apps the README says: "request the background color" and handle `tea.BackgroundColorMsg`,
calling `msg.IsDark()` (README, "Advanced Color Usage"). The app must re-render on that message.
Inconsistency: the guide's standalone example writes to stdout, the README's writes to stderr.
Pick one in loom's theme code and test it.

## 5. Glamour v2 (Markdown)

- `import "charm.land/glamour/v2"` (README, ~L20).
- Rendering: `NewTermRenderer(options ...TermRendererOption)` and package-level `Render(in, stylePath)`
  (`glamour.go`). Options seen: `WithWordWrap`, `WithTableWrap`, `WithPreservedNewLines`, `WithEmoji`,
  `WithChromaFormatter`, `WithStylesFromJSONFile`, `WithStylesFromJSONBytes`, `WithStyles`,
  `WithStylePath`, `WithBaseURL`, `WithOptions`.
- Built-in styles (`styles/styles.go` ~L12-20): `ascii`, `dark`, `light`, `notty`, `pink`, `dracula`,
  `tokyo-night`. Example: `TokyoNightStyle = "tokyo-night"`.
- Environment: `GLAMOUR_STYLE` with `glamour.WithEnvironmentConfig()` (README ~L95). The fetched
  `glamour.go` text says it defaults to `styles.DarkStyle` when unset; not a verbatim line.

What loom has: TextEntry already offers bold and Markdown. Glamour would be a second Markdown path, so
the question is whether loom needs ANSI Markdown output at all, or only inline emphasis. Host decision.

## 6. Gaps and unverified claims

- `glamour.WithAutoStyle()` appears in third-party mirrors and skill docs only. It is not in the v2
  README, `glamour.go` or `styles/styles.go`. `options.go` returned HTTP 404. Treat it as a v1
  claim until the v2 `options.go` is read. Row `glamour.v2.auto-style` is `verified: false`.
- `WithStandardStyle` signature came from converted text, not a verbatim line. Row marked
  `verified: false`.
- The Lip Gloss `compat/compat.go` URL returned 404. The compat facts come from the README only.
- No Lip Gloss v2 release notes list (v2.0.0 to v2.0.3) were read. Search results mention a Table patch
  (v2.0.2) and a docs patch (v2.0.3), not verified.
- Line numbers are approximate. Re-check before citing in a PR.
- Bubble Tea v2 `tea.BackgroundColorMsg` is cross-listed here; topic 2 covers it in depth.
- Loom column is `unknown` in every row. The host fills it from `docs/`, `spec/` and code.

## 7. Proposed follow-ups for the host (not filed)

1. P1: spike Compositor hit-test and Z-order against loom RootOverlays using one popover and one
   modal. Measure: does a click through the top layer reach the layer below? Expected: no, unless
   loom routes it. Canary first.
2. P2: decide the theme path. Options: explicit `HasDarkBackground(in, out)` plus
   `tea.BackgroundColorMsg`, or loom's own detection. Confirm which one loom uses today.
3. P3: Glamour only if loom needs ANSI Markdown output. Otherwise skip.

Rows written: 22 (features-lipgloss.yaml). Rows verified: 20 (`verified: true`). Open gaps: section 6.
