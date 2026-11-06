**Once per session, before your first task, read all Harnez rules in one call: `harnez read .harnez/rules/Tools.md .harnez/rules/Issues.md .harnez/rules/Quota.md .harnez/rules/Subagents.md .harnez/rules/Output.md .harnez/rules/Local.md`. Skip this if you already read them in this session. This follows Index.md order; Local.md overrides the other rules.**

Adhere to the following conventions.

<!-- harnez:begin Project Summary -->
<!-- harnez:end Project Summary -->

## Development Scripts

Run from project root.
Mouse events reaching widgets are 0-based and child-local; never subtract 1 again (see docs/Widgets.md).
Validate every new or modified ANSI design using `loom eval` and `loom measure`, plus `loom check-box` when it contains box borders. Resolve geometry errors before presenting or committing the design.
Developer agents get the standard prompt in `.harnez/prompts/developer.md`.
A developer's "fixed but not rerun" counts as untested; the host reruns `make test-q1` before closing the ticket.
A release that breaks callers (API, semantics, module path) adds a row to the breaking-changes table in `docs/Upgrading.md`.
PRs go to `origin` (ubunatic/loom) by AGit, not via fork or upstream branches: `git push origin HEAD:refs/for/main/<topic> -o title="<line>" -o description="{base64}$(base64 -w0 body.md)"`. A fast-forward re-push updates the PR (rewrites need `-o force-push=true`). A topic whose PR is closed opens a new PR, so pick a fresh topic. Add labels via the API. Sync local `main` from `origin`.
Manual tests for the owner: open foot as a background task with a short checklist, then a login shell: `timeout 15m foot -- bash -c "printf '%s\n' '1. <step>' '2. <step>'; $SHELL -l"`. Keep full steps and expected results in chat. Drop the `timeout` only when the owner just replied asking for the test now.

<!-- harnez:begin Language Conventions -->
Adhere to the following conventions.

Docs in `./docs/` are managed by harnez. <!-- harnez:bundled -->

- Agentic Loop Practices @docs/AgenticLoop.md,
  5-phase loop (Advisory -> Dev -> Review -> Hygiene -> Retro), zero zombie guarantee
- Bash/Shell @docs/Bash.md,
  Read before multi-line shell: Make recipes, embedded scripts
  No ";", break before then/else/docs
  No "if [[]]", No "if []", Use "if test"
  smart indent!
  Use git -C/make -C, not cd
- Canary-first development @docs/Canary.md,
  probe external mechanisms before building features on them
- Git @docs/Git.md,
  conventional commits, work on the default branch, don't push unless asked
- Go/Golang @docs/Go.md,
  Modern Go; substantive features and tests in internal/<feature>, thin cmd/<app> for test caching; avoid deps, use runes and display width
- Release Pipeline @docs/GoRelease.md,
  harnez release, version.yaml spec, GoReleaser v2, non-interactive minisign (-W), Forgejo has_releases, language-agnostic (Go/Python/Zig/Rust/scripted)
- Issue Tracking Practices @docs/IssueTracking.md,
  P0-P3 priorities, metadata headers (Status, Priority, Severity, Category), tracker sync
- Make/Makefile @docs/Make.md,
  ⚙️ phony sentinel, self-doc help, build dependency pattern
- Man Pages for Go CLIs @docs/ManPages.md,
  cobra/doc GenManTree; <cmd> man / man --install; XDG ~/.local/share/man/man1
- Markdown @docs/Markdown.md,
  PascalCase for evergreens, kebab-case for ephemeral docs; ASCII art in chat, Mermaid only in docs/
- Search Practices @docs/Search.md,
  harnez find code/docs, finder configuration, partial results, and rg fallback
- Spec system @docs/Spec.md,
  YAML spec files as single source of truth; Go code must not duplicate spec values
- Test Tiers @docs/TestTiers.md,
  make lint / test-pkg / test-short / test / container-test; no sleeps, raceFactor budgets, PTY tests in the full tier
<!-- harnez:end Language Conventions -->
Fix the library, not the caller: no per-widget or per-app workarounds for library flaws. Change interfaces or the event approach only when needed, and then follow a proven key/mouse model (global or local routing).
- Tickets that touch event routing (keys, mouse, focus, EventResult) start on `terra:med`, not luna.
- Tab belongs to containers for focus moves; a widget consumes Tab only when it is its documented action (see Tab policy in docs/Widgets.md).
