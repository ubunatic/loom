## 014 Configurable graph colors and glyph presentation
- Verdict: keep, it lacks declarative level ranges and color stops.
- Evidence: `graph/sparkline.go:90` `ForegroundANSI func(value float64) string`; `graph/bar.go:36` `type BarOptions struct {` (no color hook).
- Proposed priority: P3, UI styling enhancement with low impact on core logic.

## 090 Support image-backed app backgrounds and background theme switching
- Verdict: keep, image mode exists but lacks ubunatic.com/cati usage and Astra/cat theme switch in examples/background.
- Evidence: `background.go:131` `func (c *Canvas) PaintDecoration` (sampling lacks image averaging); `examples/background/background/background.go:128` `b.background = loom.NewAstraBackground()` (no switch).
- Proposed priority: P3, visual enhancement example with no blocker for core functionality.

## 095 Add human-observable PTY test view mode and feedback flow
- Verdict: keep, no interactive feedback mode in pty test yet.
- Evidence: not verified.
- Proposed priority: P1, improves developer experience and testing observability.

## 157 ANSI Golden Mockup Visual Test Comparator
- Verdict: keep, CompareWidget exists but misses bold/dim attributes and continuation cells comparison.
- Evidence: `loomtest/compare.go:23` `func CompareWidget(t testing.TB, reference string, width, height int, render func() string, options ...CompareOptions) {`; `loomtest/compare.go:71` `fmt.Fprintf(&report, "row=%d col=%d want={glyph=%q fg=%s bg=%s} got={glyph=%q fg=%s bg=%s}\n", y+1, x+1, string(w.Rune), formatColor(wfg), formatColor(wbg), string(g.Rune), formatColor(gfg), formatColor(gbg))`
- Proposed priority: P2, robust visual tests are critical but basic comparison works.

## 161 Align media example with redesign 005
- Verdict: keep, media example layout is not aligned with redesign 005 mockup.
- Evidence: `examples/media/main.go:36` `canvas.DrawImage(img, loom.Rect{W: termCols, H: termRows - 1})`; `examples/media/main.go:40` `status := fmt.Sprintf("%dx%d | %s | %s", termCols, termRows, mode, filepath.Base(*mediaPath))`
- Proposed priority: P3, visual alignment of a demo application.

## 241 Add public grab package (SDL3 transparent input grabber) and examples/grabber demo
- Verdict: keep, package and demo do not exist.
- Evidence: not verified.
- Proposed priority: P2, useful for input handling but might depend on SDL3 integration.

## 259 OSC 8 terminal hyperlink support (human-assisted)
- Verdict: keep, hyperlink attribute is missing from styles and not emitted.
- Evidence: `ansi/style.go:227` `type Style struct {`; `ansi/style.go:259` `func AppendANSIFor(dst []byte, s, ref Style) []byte {` (no OSC 8 handling).
- Proposed priority: P2, feature parity with modern terminals but not blocking.

## 265 RichTextEdit right-click context menu and clear-style popover button
- Verdict: keep, context menu and clear-style button are missing.
- Evidence: `colorpicker.go:202` `c.Write(cr.X, cr.Y, "█", defs.Glyphs.Swatch.Style)` (no focus marker); `richtextedit_clipboard.go:9` `func (e *RichTextEdit) copySelectionOrWord() {` (internal clipboard exists).
- Proposed priority: P2, user interaction completeness for the editor widget.

## 266 RichTextEdit S-F5 box-push drawing mode
- Verdict: keep, push mode not implemented.
- Evidence: not verified.
- Proposed priority: P3, specialized drawing feature.

## 274 Overlay and modal layer for Canvas to eliminate local popup stacks
- Verdict: keep, local modal fields and popup management are still scattered.
- Evidence: `richtextedit.go:76` `helpPopup *Popup`; `richtextedit_input.go:19` `if e.helpPopup != nil && e.helpPopup.Open {`
- Proposed priority: P1, architecture debt causing duplicated effort and potential bugs.

## 275 Decouple focus state updates from Popup and Widget Draw rendering
- Verdict: keep, `Draw` methods in widgets still mutate focus state.
- Evidence: `popup.go:70` `f.SetFocus(p.Open)`; `tabs.go:304` `f.SetFocus(i == t.focus)`; `grid.go:210` `f.SetFocus(i == g.focus)`
- Proposed priority: P1, architecture debt violating unidirectional rendering and causing bugs.

## 277 Standardize click-only focus and cursor invariants across compound widgets
- Verdict: keep, focus invariants are partially addressed (Choice hover sets highlight but FilePicker is fixed).
- Evidence: `filepicker.go:300` `p.nameFocus = true`; `choice.go:647` `c.sel = fi`
- Proposed priority: P2, UX consistency across widgets.

## 288 Compose loom edit panels with standard widgets and clipped responsive layout
- Verdict: keep, standard widgets are still not composed in `loom edit` panels.
- Evidence: not verified.
- Proposed priority: P2, layout responsiveness and code reuse.

## 305 Add Ctrl+N to create a new file in loom edit
- Verdict: keep, missing shortcut Ctrl+N for new file creation in `loom edit`.
- Evidence: not verified.
- Proposed priority: P3, missing editor feature but low effort/impact.

## 312 Preserve .tui block source (box, table, image) on save
- Verdict: keep, blocks lose original tags and save as plain glyphs.
- Evidence: `richtext.go:13` `type RichDocument struct {`; `markup.go:348` `func (d RichDocument) ToMarkup() string {`
- Proposed priority: P2, data loss on save limits the utility of the editor.

## Merges
none

## Priorities
1. 275 Decouple focus state updates from Popup and Widget Draw rendering, because it fixes architecture debt causing bugs across multiple widgets.
2. 274 Overlay and modal layer for Canvas to eliminate local popup stacks, because it deduplicates modal boilerplate and reduces stale modal bugs.
3. 095 Add human-observable PTY test view mode and feedback flow, because testing observability directly impacts developer velocity.
4. 157 ANSI Golden Mockup Visual Test Comparator, because rigorous visual tests are needed to prevent regressions in layout and rendering.
5. 312 Preserve .tui block source (box, table, image) on save, because it prevents data loss during saving in the rich text editor.
