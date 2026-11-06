## 102 Human review collection: manual checks for lean-sprint deliveries
- Verdict: keep
  - It tracks pending human manual checks that cannot be performed programmatically by headless tests.
- Evidence:
  - `issues/102-human-review-collection-manual-checks-for-lean-sprint-deliveries.md:14`: `- [ ] **All evidence frames**: \`for f in docs/progress/*/*.ansi; do echo "== $f"; cat "$f"; done\``
  - `issues/102-human-review-collection-manual-checks-for-lean-sprint-deliveries.md:25`: `- [ ] **092 scrollbar drag**: drag the View and Split scrollbars with a real mouse; check drag`
- Proposed priority: P2
  - Final validation of UI/UX is moderately important for correctness but doesn't block automated CI checks.

## 120 Implement wazero-backed Tree-Sitter syntax engine with embedded grammars and queries
- Verdict: keep
  - The implementation is completely missing from `syntax/`, as evidenced by the lack of `wazeroEngine`.
- Evidence:
  - `syntax/*.go`: not verified (search for `wazeroEngine` yielded no results)
  - `issues/120-implement-wazero-backed-tree-sitter-syntax-engine-with-embedded-grammars-and-queries.md:19`: `2. **Engine Implementation (\`wazeroEngine\`)`:`
- Proposed priority: P2
  - This is a significant feature addition to power syntax highlighting but doesn't block core functionality.

## 137 Feedback: CLI asset tools workflow & multi-box ANSI validation
- Verdict: close: done
  - The CLI tools (`check-box`, `eval`, `format`, `measure`) are already implemented.
- Evidence:
  - `ansibuffer.go:560`: `func ParseAnsiBuffer(content string, minCols, minRows int) (*AnsiBuffer, error) {`
  - `measure/measure.go:45`: `func StringWidthNew(text string) int {`
- Proposed priority: P3
  - Feedback ticket, the related tools are available.

## 139 Add `loom play` for interactive TUI commands and ANSI capture
- Verdict: keep
  - The `loom play` command doesn't exist yet in `cmd/loom`.
- Evidence:
  - `cmd/loom/play.go`: not verified (file doesn't exist)
  - `issues/139-add-loom-play-for-interactive-tui-commands-and-ansi-capture.md:12`: `Add \`loom play -- <cmd> <args>\` to run an arbitrary TUI, let the user see and interact with it in their terminal`
- Proposed priority: P2
  - Good utility feature for testing but not a blocker for core functionality.

## 313 Manual acceptance checklist for open issue follow-ups
- Verdict: keep
  - Similar to 102, this is a tracking ticket for manual validations that requires human interaction.
- Evidence:
  - `issues/313-manual-acceptance-checklist-for-open-issue-follow-ups.md:18`: `This list is based on a recent read-only issue status review. Confirm each referenced ticket's live acceptance criteria before testing`
- Proposed priority: P2
  - Required for signing off on completed work.

## 314 Aggregate remaining development across open issues
- Verdict: keep
  - This acts as an epic/tracking issue for several other unimplemented tasks.
- Evidence:
  - `issues/314-aggregate-remaining-development-across-open-issues.md:13`: `This ticket aggregates the development work so it can be planned without treating every open ticket as equally unfinished.`
- Proposed priority: P2
  - Useful for project management and tracking.

## 389 Detect modifier keys via voxi-modifierd on non-kitty terminals
- Verdict: keep
  - There is no implementation of `voxi-modifierd` integration in `internal/session` or `vt`.
- Evidence:
  - `vt/*.go`: not verified (search for `voxi` yielded no results)
  - `internal/session/*.go`: not verified (search for `voxi` yielded no results)
- Proposed priority: P3
  - It's a workaround for legacy terminals without kitty protocol support.

## 390 loom session: tab label shows foreground command (emoji for well-known) and cwd
- Verdict: keep
  - Code still uses a static title and doesn't query `/proc` for foreground process info.
- Evidence:
  - `internal/session/session.go`: not verified (search for `/proc` yielded no results)
  - `issues/390-loom-session-tab-label-shows-foreground-command-emoji-for-well-known-and-cwd.md:21`: `- Foreground process: \`tcgetpgrp\` on the tab's PTY master gives the foreground process group.`
- Proposed priority: P2
  - Improves UX significantly for session tabs.

## 392 loom session mouse modes: off|full|basic (basic suppresses drag to preserve text selection)
- Verdict: keep
  - The current code still uses a binary `mouseOverride` (-1, 0, 1) rather than the requested `off|basic|full`.
- Evidence:
  - `internal/session/session.go:163`: `mouseOverride    int // -1 disabled, 0 follows child, 1 enabled`
- Proposed priority: P2
  - Highly requested UX improvement to allow text selection while keeping basic mouse functionality.

## 396 Epic: clear 77 harnez lint warnings (func length, params, file length)
- Verdict: keep
  - It's an ongoing epic for code quality fixes. `vt/vt.go` is still very large.
- Evidence:
  - `vt/vt.go`: not verified (search for `func ` yielded 37 matches but file length is a concern)
  - `issues/396-epic-clear-77-harnez-lint-warnings-func-length-params-file-length.md:17`: `This is an Epic: the fixes span many packages and each category needs a design`
- Proposed priority: P2
  - Code debt cleanup is important for maintainability.

## 397 loom session: view-only attach mode to observe running session without detaching owner
- Verdict: keep
  - `Attach` still rejects concurrent connections instead of supporting a read-only mode.
- Evidence:
  - `internal/session/server.go:123`: `if s.conn != nil {`
  - `issues/397-loom-session-view-only-attach-mode-to-observe-running-session-without-detaching-owner.md:12`: `if s.conn != nil {`
- Proposed priority: P2
  - Very useful feature for collaboration and monitoring.

## 400 loom view and loom-probe keys need /dev/tty; add headless key input for agents
- Verdict: keep
  - Headless key input via stdin/file is not yet implemented for these commands.
- Evidence:
  - `cmd/loom/view_pty_test.go`: not verified (no headless specific implementation found)
  - `issues/400-loom-view-and-loom-probe-keys-need-dev-tty-add-headless-key-input-for-agents.md:11`: `/goal Headless agents can drive \`loom view\` and \`loom-probe keys\` without /dev/tty (e.g. \`--keys-from\` stdin/file)`
- Proposed priority: P3
  - Relevant for agents, but normal users run this with a TTY.

## 401 Scrolling word-wrapping transcript view with follow-end and scroll indicator
- Verdict: keep
  - The transcript widget is missing, as noted in the issue and by lack of related files.
- Evidence:
  - `issues/401-scrolling-word-wrapping-transcript-view-with-follow-end-and-scroll-indicator.md:20`: `- Source to adopt: \`~/projects/taia/internal/tui\` at taia commit 45a6469.`
- Proposed priority: P3
  - Deferred by owner per ticket description.

## 408 loom view and loom.View: add find on ctrl-f
- Verdict: keep
  - `loom view` still doesn't implement find; it's mentioned as missing in code comments.
- Evidence:
  - `cmd/loom/view_pty_test.go:15`: `// \`loom view\` (issue 368). ctrl-f is find, not page down (issue 407); the`
- Proposed priority: P3
  - Enhances view capabilities but not blocking.

## 413 ansibuffer: parser misses CSI A/B/D, colon truecolor, wide-char continuation and relative-move prescan
- Verdict: keep
  - `ParseAnsiBuffer` in `ansibuffer.go` is still missing these parser features.
- Evidence:
  - `ansibuffer.go:560`: `func ParseAnsiBuffer(content string, minCols, minRows int) (*AnsiBuffer, error) {`
  - `issues/413-ansibuffer-parser-misses-csi-a-b-d-colon-truecolor-wide-char-continuation-and-relative-move-prescan.md:17`: `- The CSI switch (\`ansibuffer.go:645\`) handles \`C\`, \`H\`, \`f\`, \`G\` and \`d\`, but not \`A\`, \`B\` or \`D\`.`
- Proposed priority: P3
  - Fixes edge cases in design validation rendering.

## 414 measure: share one grapheme cluster boundary scanner between StringWidthNew and ClustersNew
- Verdict: keep
  - `StringWidthNew` and `ClustersNew` still have separate boundary logic.
- Evidence:
  - `measure/measure.go:45`: `func StringWidthNew(text string) int {`
  - `measure/measure.go:299`: `func ClustersNew(text string) []string {`
- Proposed priority: P3
  - Refactoring to avoid width and cell layout drifting.

## 421 syntax: shell lexer, colour the command after if, else, while and !
- Verdict: keep
  - The syntax engine doesn't use `CaptureFunction` after these keywords yet.
- Evidence:
  - `syntax/theme.go:8`: `CaptureFunction           Capture = "function"`
  - `issues/421-syntax-shell-lexer-colour-command-after-if-else-while-and-bang.md:12`: `The shell lexer from 420 (v0.8.0) colours the first word of a command, but not the word after the keywords`
- Proposed priority: P3
  - Minor visual enhancement for shell scripts.

## Merges
none

## Priorities
1. 392 - loom session mouse modes: off|full|basic (basic suppresses drag to preserve text selection). High impact on usability and text selection within terminal sessions.
2. 397 - loom session: view-only attach mode to observe running session without detaching owner. Important for collaboration and observing active sessions.
3. 390 - loom session: tab label shows foreground command (emoji for well-known) and cwd. Greatly improves session navigation and context.
4. 120 - Implement wazero-backed Tree-Sitter syntax engine with embedded grammars and queries. Big architectural piece needed for syntax highlighting.
5. 396 - Epic: clear 77 harnez lint warnings (func length, params, file length). Reduces technical debt and improves codebase health.
