# 252 — Audit library defaults that are missing from the spec

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Bug
**Related**:

---

## 1. Problem & Motivation
Loom already embeds YAML specs for defaults, themes, borders, cursor effects, resize policy, emoji width, and the widget catalog, but the library still owns several user-visible defaults only in Go. This creates multiple sources of truth and makes behavior harder to review consistently.

**Goal**: Audit and resolve the library-owned spec gaps below, moving only values judged to be durable product defaults into YAML with matching schema and runtime consumption. Done when accepted candidates are consumed from specs, rejected candidates have a recorded reason, and spec integrity checks pass; stop and report if ownership or spec-file placement requires a user decision or permission.

## 2. Technical Specification / Findings
Read-only audit of the reusable Loom library and bundled `graph`, `media`, and `collector` packages, using the boundary in `docs/Spec.md` and `issues/055`. Ratings indicate how strong a candidate each set is for specification; they are recommendations, not a commitment to move every literal.

| Candidate set | Recommendation | Candidate values / examples | Likely files |
|---|---|---|---|
| Default widget key bindings and aliases | **Strong** for stable library defaults. Spec action names, keys, and aliases; keep dispatch and interaction logic in Go. | Widget-local `case` key strings and aliases are distributed across Choice, View, Tabs, Table, Tree, Form, TextInput/TextArea, NumberInput, SearchBar, Viewport, Menu, Dialog, Confirm, Grid, Split, Timer, PaintCanvas, and Media. `spec/defaults.yaml` currently centralizes fallback quit keys but not these widget bindings. | Those widget `*.go` files; `spec/defaults.yaml` or a focused keys/actions YAML; matching schema; `defaults.go`/loader and key documentation. `event.go` normalization of terminal protocols should stay code. `key_defaults.go` is a documentation summary, not the dispatch source; reconcile or generate it from the eventual authoritative data. |
| Shared widget chrome, glyphs, and fixed labels | **Strong** for stable library-owned display defaults; **moderate** for per-widget hints and validation wording. Keep formatting and dynamic text generation in Go. | Examples: Choice's `DefaultPlaceholder` duplicates `search_bar.placeholder`; NumberInput step markers; Choice/Tree selection markers; Toggle checkmark; Paginator dots; Table sort arrows; Split separators; Chart axes/series markers; Form headings/required text; fixed validation strings. | `choice.go`, `numberinput.go`, `toggle.go`, `paginator.go`, `table.go`, `split.go`, `tree.go`, `chart.go`, `form.go`, and related widgets; `spec/defaults.yaml` plus schema for strings/glyphs, and possibly `spec/themes.yaml` plus schema and `theme.go` for themed visual roles. |
| Semantic colors and default palettes | **Strong** for library presentation colors; **weak / keep** for terminal-standard color conversion tables. | Chart series palette, semantic error/selection colors in Form and ANSI editor, Pill status colors, and debug boundary color are fixed in code. The xterm/ANSI palette math is a terminal standard and should remain implemented in Go. | `chart.go`, `form.go`, `pill.go`, `ansieditor.go`, `style.go`; `spec/themes.yaml`, its schema, and `theme.go` if these become theme roles. Review whether `internal/ptytest/style.go` intentionally mirrors the standard palette. |
| Public widget/package defaults and timing policies | **Strong** where a zero/omitted option selects a documented user-visible behavior; **moderate** where it is an algorithm or resource policy. | NumberInput's zero Step becomes 1; Timer's one-second widget cadence; ProgressBar's 100ms fallback interval; graph render defaults (width/range/wrappers/glyph sequences/background ANSI, spinner frames); collector's 15-minute omitted retention. Existing `progress_bar.demo_interval` is separate from the runtime ProgressBar fallback. | `numberinput.go`, `timer.go`, `progressbar.go`; `graph/options.go`, `bar.go`, `sparkline.go`, `stackedbar.go`, `treemap.go`, `spinner.go`; `collector/collector.go`; `defaults.go`, `spec/defaults.yaml`/schema, or package-specific spec/schema/loader if ownership calls for that split. |
| Spec identity and presentation order duplicated in Go | **Moderate**. Keep behavior in Go, but spec explicit identity/order when it is presentation data. | `SpeccedResizeModeIDs` repeats the mode IDs from `spec/resize.yaml`; the Go slice supplies stable UI order because the loaded mapping is unordered. | `resize.go`, `spec/resize.yaml`, `spec/schemas/resize.schema.json`, and any presentation consumer such as `examples/winch`. Consider an explicit ordered ID list in YAML. |
| Timings and numeric guardrails that are implementation policy | **Weak to moderate; classify case by case**. Spec values only if maintainers need to review/change them as product policy; leave mechanical safeguards in Go otherwise. | Examples to assess include the resize-rate minimum span and pane cursor-query/shutdown timeouts. | `resize.go`, `pane.go`, matching YAML/schema only if accepted. |

Keep in Go: algorithm flow and validation predicates; loop/index/buffer mechanics; ANSI/terminal protocol constants and platform identifiers; dynamic content and application composition in examples; test fixtures. Do not move values solely because they are numeric or string literals. The scan found no missing schema companion among the current `spec/*.yaml` files; the gap is primarily uncentralized defaults and Go-side duplicates/metadata.

## 3. Implementation & Verification Plan
Resolve the recommendations set by set. For accepted values, add YAML keys and schema declarations, load them from the embedded spec, remove duplicate Go defaults, and add integrity tests proving runtime consumption. Preserve code for implementation mechanics and record any consciously retained policy constants. Run spec validation and the relevant test/build targets after implementation; this ticket records the audit only and makes no source/spec changes.
