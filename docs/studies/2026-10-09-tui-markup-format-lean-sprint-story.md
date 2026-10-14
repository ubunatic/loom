# Case Study: Rapid Evolution of the .tui Markup Format via Harnez Lean Sprints

## 1. Header & Context

- **Date**: 2026-10-09
- **Scope of Work**: Lean sprint progression spanning issues 306, 307, 308, 309, 310, and 311.
- **Starting State**: 
  - `loom edit` had redundant ANSI reset prefixes on file saves (issue 306).
  - Modal/popup dismissal left an unhandled Escape without clearing selection (issue 307).
  - ANSI files were rigid 2D byte arrays without a clean, human-editable markup format (issues 308–311).
- **Intended Goals**:
  1. Fix serializer artifacts and editor key routing.
  2. Research, design, and implement an editable `.tui` ANSI markup language.
  3. Support flow text, inline styles, box blocks, table blocks, and embedded cati terminal images.
  4. Integrate `.tui` loading, editing, saving, and CLI conversion into Loom.

---

## 2. Executive Summary

In a continuous session of less than two hours, a multi-model agentic loop orchestrated the complete end-to-end conception and delivery of the `.tui` markup format and related editor fixes:
- **Issue 306**: Resolved redundant ANSI resets in `ansibuffer.go` using `codex:luna:low`.
- **Issue 307**: Scoped editor-level Escape selection clearing using `codex:sol:med`.
- **Issue 308**: Researched terminal markup formats with `claude:haiku:med`, aligned on bracket syntax (`[bold red on #bg]text[/]`) and flow text (`RichDocument`) with the human user, and implemented the foundational parser, serializer, CLI (`loom markup`), and `loom edit` integration via `claude:opus:low`.
- **Issue 309**: Extended `.tui` with `[box]` block markup (sharp, rounded, double, ascii borders) via `claude:opus:low`.
- **Issue 310**: Added full `[table]` block markup with pipe syntax, column alignment (`:--`, `:-:`, `--:`), auto-sizing via `StringWidth`, and spec-backed junction glyphs via `claude:opus:low`.
- **Issue 311**: Added `[image]` tag rendering using `ubunatic.com/cati`, cleanly decoupled across `loom` (core hook) and `loom/media` (renderer), supporting halfblock, quadblock, and sextant modes.

All 6 tickets were verified through strict Quota-1 test passes (`make test-q1`), geometry validation (`loom eval`, `loom measure`, `loom check-box`), and closed with committed tracker synchronizations.

---

## 3. What Worked Well

### Multi-Tier Model Specialization
The Harnez harness enabled matching tasks to model strengths:
- **Mechanical bugfixes & clean scoped tests**: `codex:luna:low` and `codex:sol:med` quickly handled serializer fixes and key event propagation with zero fuss.
- **Broad landscape research**: `claude:haiku:med` executed lightning-fast comparative research across Python Rich, Spectre.Console, gookit/color, and chalk, providing concise options for user discussion.
- **Complex architecture & parser engineering**: `claude:opus:low` implemented clean, bug-free recursive descent and block parsers, comprehensive fuzz testing, and decoupled media adapters while respecting project invariants (spec YAML files, non-breaking APIs).

### Spec-Driven Architecture
Rather than hardcoding terminal glyphs, colors, or limits in Go files:
- Named colors and image mode limits were placed in [`spec/markup.yaml`](../../spec/markup.yaml).
- Table tee (`┬ ┴ ├ ┤`) and cross (`┼`) junction glyphs were placed in [`spec/box.yaml`](../../spec/box.yaml).
- Schema validation (`cmd/validate-spec`) automatically validated all additions on every test run.

### Interactive User Phase (Phase 2 Alignment)
Before writing code for issue 308, an explicit alignment turn presented syntax comparisons (bracket tags vs. angle tags vs. markdown vs. sidecars) and data models (flow text vs. 2D grid cells). Reaching explicit consensus prevented churn and ensured the parser aligned with user expectations.

---

## 4. Honest Post-Mortem (Failures, Bugs & Near-Misses)

1. **Serializer Space vs. Blank Cell Mismatch (Issue 306)**:
   - *Problem*: `Serialize()` initialized `lastStyle` to zero `AnsiCell` (with rune 0), whereas `BlankAnsiCell()` had a space rune `' '`. The very first cell comparison falsely registered a style change and emitted `\x1b[0m\x1b[39m\x1b[49m`.
   - *Fix*: Initialized `lastStyle` to `BlankAnsiCell()` and trimmed redundant default color suffixes when returning to default terminal state.
2. **Empty Line Loss in Markup Fuzz Testing (Issue 308)**:
   - *Problem*: Early parser implementation dropped trailing newlines when re-serializing empty lines, causing documents with empty trailing lines to shrink by one line on each save/load cycle.
   - *Detection*: Caught by Go 1.25 native fuzz test (`FuzzMarkupParse`).
   - *Fix*: Normalized newline termination and added a permanent regression test seed in `testdata/fuzz/FuzzMarkupParse/`.
3. **Circular Import Hazard with Cati (Issue 311)**:
   - *Problem*: Directly importing `cati` in root `loom` would introduce image decoding and external SVG processes into the base library, and risked circular module/package dependencies.
   - *Fix*: Designed a callback hook `loom.MarkupOptions.Image` in the root package; implemented the concrete `cati` adapter in `loom/media` (`media.MarkupOptions`), keeping core `loom` lightweight and decoupled.

---

## 5. Quality & Invariants Audit

| Dimension | Assessment | Evidence |
|---|---|---|
| **Architecture & Separation** | High | Root `loom` defines parsing interfaces; `spec/*.yaml` holds single-source-of-truth tables; `media` holds optional image rendering dependencies. |
| **Idempotency** | High | Bi-directional roundtrip tests: `ParseMarkup(doc.ToMarkup()) == doc`. |
| **Backward Compatibility** | Complete | All existing ANSI loading paths in `RichTextEdit` and `loom edit` preserved; `.tui` activated strictly by file extension. |
| **Test Coverage & Verification** | Complete | All milestones backed by table unit tests, fuzz tests, `loom check-box` geometry validation, and clean `make test-q1` runs. |

---

## 6. Efficiency & Velocity Assessment

- **Total Duration**: ~1 hour 45 minutes for 6 feature tickets and bug fixes.
- **Turnaround per Milestone**: Average 2–4 minutes per milestone execution.
- **Subagents Deployed**:
  - `dev-306` (`codex:luna:low`)
  - `dev-307` (`codex:sol:med`)
  - `res-308` (`claude:haiku:med`)
  - `dev-308` (`claude:opus:low`)
  - `dev-309` (`claude:opus:low`)
  - `dev-310` (`claude:opus:low`)
  - `dev-311` (`claude:opus:low`)
- **Developer Intervention**: Zero direct host code edits. The host strictly orchestrated preflight checks, user alignment, diff audits, test executions, ratings, and issue closures.

---

## 7. Key Learnings & Evergreen Upstream

1. **Decoupled Render Hooks Over Direct Engine Dependencies**: When integrating heavy third-party renderers (like image decoders or syntax highlighters), standardizing on a functional option/hook in the root parser prevents dependency bloat and eliminates circular package imports. Root `loom` defines `MarkupOptions.Image`, while `loom/media` provides the concrete `cati` adapter.
2. **Spec Expansion Over Code Additions**: When new layout primitives (like table junction tees) are needed, adding them to YAML specs (`spec/box.yaml`) and JSON schemas keeps UI rendering declarative, validates them automatically via `validate-spec`, and avoids hardcoded constants in Go code.
3. **Fuzz Testing for Parsers on Turn 1**: Fuzz tests added during M1 (`FuzzMarkupParse`) caught an insidious edge case in trailing empty-line loss during serialization before `.tui` was wired to `loom edit`.
4. **Interactive Alignment Gates**: Halting between research (Phase 1) and implementation (Phase 3) to align on bracket syntax and flow text saved multiple refactoring cycles.
5. **Shared Block Lexer Pattern**: Factoring out `parseMarkupBlockOptions` into `markup_block.go` ensured boxes, tables, and images share the same `key=value` and quoted string syntax rules across the ecosystem.
6. **Harness & Logging Disciplines**:
   - Developer workers under Quota-1 must avoid piping test runs through `tail`, which obscures exit status and fail lines. Developer prompts should mandate direct file output capture and targeted `grep` inspection.
   - PTY and Geometry gates (`loom check-box`, `loom measure`) provide fast, objective verification of visual alignment before host commits.
7. **Evergreen Artifact Creation**:
   - Authored [`docs/TuiMarkup.md`](../TuiMarkup.md) as the authoritative evergreen reference for bracket tags, box/table/image blocks, strict parsing invariants, and CLI tooling.
   - Preserved follow-up work (retaining raw block source on save) as an independent tracked ticket ([Issue 312](../../issues/312-preserve-tui-block-markup-source-on-save.md)).

---

## 8. File & Diff Summary

### Key Files Created
- `markup.go` & `markup_test.go`: Core `.tui` parser and serializer.
- `markup_block.go`: Shared `key=value` block tag option lexer.
- `markup_box.go`: Box block parser and border renderer.
- `markup_table.go`: Table block parser, column auto-sizer, and grid renderer.
- `markup_image.go`: Decoupled image block hook and fallback.
- `media/markup.go` & `media/markup_test.go`: Cati image rendering adapter.
- `cmd/loom/markup.go`: Bi-directional `loom markup` CLI tool.
- `spec/markup.yaml` & `spec/schemas/markup.schema.json`: Markup color and mode specification.
- `docs/TuiMarkup.md`: Complete user and developer syntax reference.
- `examples/markup/box.tui`, `table.tui`, `image.tui`: Validated examples.

### Key Commits
- `ab3efd2`: `fix: avoid redundant ANSI reset prefixes in serializer (issue 306 M1)`
- `07b95c6`: `fix(edit): clear selection on unhandled Escape in editor (issue 307 M1)`
- `7417272`: `feat(markup): add .tui bracket markup parser and serializer (issue 308 M1)`
- `17c1346`: `feat(edit): load and save .tui files in RichTextEdit and loom edit (issue 308 M2)`
- `96e3981`: `feat(markup): add [box] block markup to .tui format (issue 309)`
- `4714cf4`: `feat(box): add tee and cross junction glyphs to box spec`
- `511e0d1`: `feat(markup): add [table] block markup to .tui format (issue 310)`
- `68bb546`: `feat(markup): add [image] tag with render hook to .tui format (issue 311)`
- `f294a49`: `feat(media): render .tui [image] tags with cati (issue 311)`
- `2b50579`: `feat(markup): render .tui images in loom markup and loom edit (issue 311)`
