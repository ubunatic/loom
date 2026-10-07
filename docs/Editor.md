# Loom Editor

`loom edit [file]` hosts RichTextEdit for rich text and ANSI documents. When invoked without arguments, it opens an empty, Untitled buffer. This guide records editor design decisions and verification lessons; issue files own implementation status and reproductions.

## Design decisions

The [approved ANSI mockups](design/loom-edit-design-notes.md) are visual guidance. Final appearance may vary with standard Loom widgets. New UI must compose standard widgets; small tweaks are acceptable, heavy hacks are not. Always file or link issues for larger observed library gaps rather than force the mockup through app workarounds. See [Widgets](Widgets.md), [Root overlays](RootOverlays.md), and [Geometry](Geometry.md).

- **No-Argument Startup**: Running `loom edit` opens an empty Untitled buffer without requiring an existing file path ([302](../issues/302-loom-edit-ux-refinements-theme-default-mouse-disable-backdrop-clicks-and-untitled-buffer.md)).
- **Default Theme**: Default theme is `julia256` across all Loom CLI apps and the editor ([302](../issues/302-loom-edit-ux-refinements-theme-default-mouse-disable-backdrop-clicks-and-untitled-buffer.md)).
- **Mouse Capture & Temporary Grab**: In mouse-off mode (`"m"`), mouse reporting is fully disabled to allow unobstructed native terminal text selection. Interactive modal overlays (F1 Help, Popovers, Dialogs) temporarily engage mouse tracking, restoring `"m"` mode when closed ([303](../issues/303-temporary-mouse-grab-for-popups-menus-and-dialogs-in-mouse-off-mode.md)). Persistent global grab (`"M"`) is toggled via `--mousegrab` or clicking the status icon.
- **Modal Event Isolation**: Backdrop clicks on modal overlays dismiss the overlay as `loom.Handled()` without leaking mouse events into underlying text buffers or moving the cursor ([304](../issues/304-modal-overlay-mouse-event-isolation-backdrop-clicks-must-not-leak-into-underlying-editor.md)).
- **Bottom Chrome**: Compact three-line bottom chrome: Divider (`─`), Status line (`Ln X, Col Y`, compact indicators `◐`, `m`/`M`, `a`/`A`), and Hotkey bar (`^O Files`, `^F Search`, `^S Save`, `F1 Help`, `^D Box`, `⌃P Screenshot`, `F10 Quit`). All buttons and indicators are mouse-clickable ([300](../issues/300-loom-edit-replace-top-title-bar-with-compact-bottom-status-icons.md), [301](../issues/301-loom-edit-refine-hotkeys-nest-draw-under-box-and-make-bottom-bar-items-clickable.md)).
- **Side panel**: `^O` (and `F2`) toggles the file browser side panel open and closed ([301](../issues/301-loom-edit-refine-hotkeys-nest-draw-under-box-and-make-bottom-bar-items-clickable.md)).
- **Search**: `^F` (and `F3`) opens the search overlay ([301](../issues/301-loom-edit-refine-hotkeys-nest-draw-under-box-and-make-bottom-bar-items-clickable.md)).
- **Screenshot**: `^P` captures the complete visible editor to `~/Pictures/Screenshots/<num>-loom-edit-<details>.ansi` ([284](../issues/284-add-ctrl-p-full-editor-ansi-screenshots-to-loom-edit.md)).

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

Integration, 2026-10-07: PR #16 is merged on GitHub at `3d050e8`. Local main integrates it with host fixes in `6a33c29`: display-only RichTextEdit highlights, standard Box frames and HintBar mode controls, bounded search dimensions, container-owned Tab/Shift-Tab, error feedback, and collision-safe screenshots. These host fixes are committed locally and have not been pushed. The v0.3.2 release commit remains in history.

The final full `make test-q1` run passes. PTY checks passed for screenshot numbering after deleted captures, search, unsaved ^Q protection, dialog capture, browser cancellation/reopening and Discard/open. Search and dialog captures pass `loom eval`, `loom measure`, and `loom check-box` at 100×24. Approved designs are preserved; tests always write generated screenshots to temporary directories.

Search uses ^R to switch Normal/Regex; Tab/Shift-Tab moves among editor, browser and search. ^P works over the unsaved dialog. Screenshot numbers start at 01 and advance beyond existing numeric prefixes, using exclusive creation to avoid collisions. Details use the file basename with unsafe characters replaced by hyphens. The status row reports the saved path or error.

Remaining acceptance work stays in issues 280–284 and 287–290, particularly standard clipped layout containers, persistent hint-bar mouse hit regions, deterministic standalone evidence generation, and browser mouse/Save/Cancel coverage. Integration does not imply every feature ticket is closed.

A settings example:

```yaml
# yaml-language-server: $schema=/path/to/loom/spec/schemas/editor.schema.json
theme: plain
mousegrab: true
altscreen: true
```

The schema ships in the repository and is embedded for runtime validation; no external schema installation is performed. `--config`, `--theme`, `--mousegrab=false` and `--altscreen=false` override file preferences. Defaults are plain theme, mouse capture off, alternate screen on.


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
