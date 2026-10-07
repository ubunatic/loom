# Loom Editor

`loom edit <file>` hosts RichTextEdit for rich text and ANSI documents. This guide records editor design decisions and verification lessons; issue files own implementation status and reproductions.

## Design decisions

The [approved ANSI mockups](design/loom-edit-design-notes.md) are visual guidance. Final appearance may vary with standard Loom widgets. New UI must compose standard widgets; small tweaks are acceptable, heavy hacks are not. Always file or link issues for larger observed library gaps rather than force the mockup through app workarounds. See [Widgets](Widgets.md), [Root overlays](RootOverlays.md), and [Geometry](Geometry.md).

- **Mouse capture**: an explicit boolean `--mousegrab` flag, including `--mousegrab=false` ([280](../issues/280-add-mouse-capture-flag-to-loom-edit.md)).
- **Settings**: `~/.config/loom/editor.yaml`, following XDG conventions, with a schema and initial `theme`, `mousegrab`, and `altscreen` preferences. YAML booleans represent on/off; explicit CLI settings override the file ([281](../issues/281-add-schema-backed-loom-edit-settings-in-editor-yaml.md)). Defaults belong in the spec; the mockups do not choose them.
- **Side panel**: F2 toggles a file browser; opening another file must preserve Save/Discard/Cancel protection ([282](../issues/282-add-f2-side-panel-with-file-browser-to-loom-edit.md)).
- **Search**: F3 or ^F opens a small top-right panel with Normal/Regex and next/previous navigation. Enter/Shift-Enter are proposed navigation keys. Containers own Tab focus; displayed controls must have working input routes ([283](../issues/283-add-f3-and-ctrl-f-search-panel-to-loom-edit.md)).
- **Screenshot**: ^P captures the complete visible editor, including chrome and open panels/overlays, to `~/Pictures/Screenshots/<num>-loom-edit-<details>.ansi`. Numeric padding and filename details remain choices to document ([284](../issues/284-add-ctrl-p-full-editor-ansi-screenshots-to-loom-edit.md)).

## ANSI assets and evidence

[AGENTS.md](../AGENTS.md) requires `loom eval` and `loom measure` for every new or modified ANSI design, plus `loom check-box` for box borders. Resolve geometry errors before presenting or committing. Use `loom view` for visual inspection. The approved editor mockups are 100×24 cells; terminal width must leave room to avoid wrapping.

Geometry checks verify rows and borders, not input behavior. Search routing, focus, unsaved edits and widget lifecycle require interaction with the actual composed app in a PTY.

Keep approved mockups distinct from generated implementation evidence. Ordinary tests must not rewrite tracked assets. Explicit generation should use stable paths, fixture directory contents and repeatable output; see [289](../issues/289-keep-editor-tests-from-overwriting-tracked-ansi-design-files.md).

## Integration pitfalls

- **Unicode search offsets**: case conversion can change UTF-8 byte lengths. Offsets in transformed text cannot index the original string safely. Preserve original rune positions and test length-changing mappings ([285](../issues/285-fix-unicode-literal-search-panic-in-loom-edit.md)).
- **Child termination**: a search widget's QuitResult must not escape into application termination without unsaved-change protection ([286](../issues/286-prevent-search-panel-quit-keys-from-losing-unsaved-editor-changes.md)).
- **Overlay input**: render order and input order must agree. A visible search field cannot pass clicks through to the document. Search highlighting must not reuse editing selection if that opens formatting controls or changes editing semantics ([287](../issues/287-fix-search-overlay-mouse-routing-and-selection-isolation-in-loom-edit.md)).
- **Responsive composition**: clamp panels to available bounds and clip their children. Prefer standard frame/layout/control widgets over manually drawing application chrome ([288](../issues/288-compose-loom-edit-panels-with-standard-widgets-and-clipped-responsive-layout.md)).
- **Embedded picker lifecycle**: connect selection/cancellation and preserve reusability. Keyboard and mouse activation must share the same file-open and unsaved-change paths ([290](../issues/290-keep-the-loom-edit-file-browser-usable-after-cancel-and-mouse-activation.md)).

## Review evidence and status

Snapshot, 2026-10-07: [PR #15](https://github.com/ubunatic/loom/pull/15) merged the initial editor command and is present locally; [279](../issues/279-add-loom-edit-command-to-open-a-file-in-richtextedit.md) remains open pending its acceptance verification. [PR #16](https://github.com/ubunatic/loom/pull/16), updated head `9f41045`, is draft/unmerged and proposes issues 280–284. Recheck live code, PR state and recent history before beginning ticket work.

The initial review at `b8a59ca` passed the full `make test-q1` suite in a disposable worktree. Its five rendered designs passed `loom check-box`. Independent PTY probes still reproduced a Unicode search panic, search-field clicks modifying the underlying document, and ^Q exiting with unsaved changes without a prompt. Findings are recorded in issues 285–290.

The updated review at `a9f1a7d` again passed the full suite. PTY probes confirmed fixes for the panic, search-field click-through, and ^Q/^C/^D unsaved-change prompts. Ordinary tests now leave tracked designs unchanged; all five assets pass `loom eval`, `loom measure`, and `loom check-box`. Remaining findings include search selection/focus isolation, standard-widget/responsive composition, browser lifecycle, and deterministic explicit evidence generation. The PR remains unmerged; tickets record partial fixes without claiming integrated completion. No application architecture change from that PR was accepted.

These results establish that green unit tests and aligned ANSI files are insufficient merge evidence. Review the actual root composition, key/mouse results, saved document bytes, and post-test working-tree status.

The review at `9f41045` passed the full suite with tracked assets unchanged. Browser callbacks now share activation and cancellation paths; a PTY probe verified cancel/reopen and opening another file. Search text clipping improved, but its minimum dimensions still exceed very small editor bounds, and manual frames/controls and search editing-selection behavior remain. New ^P screenshots produce valid 100×24 ANSI output, but a PTY probe reproduced overwriting an existing capture after deleting an earlier one. Screenshot errors, overlay capture, filename sanitization and spec binding remain unfinished in issue 284. The PR remains unmerged.

## Agent workflow lessons

After adding the ANSI rule to AGENTS.md, a `luna:med` leaf developer was asked to make an alternative search design without a CLI-validation reminder. Its command stream showed validation before delivery; a read-only follow-up identified the exact commands and cited AGENTS.md. Host checks also passed. This is evidence from one bounded task, not a guarantee for future workers. The experimental design was later removed at the user's request; the validation rule remains.

For event routing, focus or EventResult work, start developers on `codex:sol:med` as required by AGENTS.md. The proposed lean sprint order is mouse capture, settings, file browser, then search, with one writer at a time; those sprints were discussed, not executed in this session. See [Lean sprints](LeanSprints.md).

For future reviews, check live PR and tracker state early, verify actual changed file paths before searching, and use precise `rg` fallback when a finder reports partial/error results. Do not equate a merged PR with completed acceptance verification, or a developer's reported checks with host evidence.

Harness improvements proposed for user decision, not applied to AGENTS.md or developer prompts: add a post-test check for unexpected tracked-file changes, and require composed-app PTY probes for overlay input and unsaved-change protection before declaring editor work complete. This session initially guessed nonexistent file/spec paths and discovered the merged-command/open-ticket mismatch only after checking GitHub; earlier live-state and path discovery would reduce that wasted work.
