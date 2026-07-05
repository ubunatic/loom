Adhere to the following conventions.

<!-- claudeconfig:begin Project Summary -->
<!-- claudeconfig:end Project Summary -->

## Development Scripts

Run from project root.

<!-- claudeconfig:begin Language Conventions -->
Adhere to the following conventions.

Docs in `./docs/` are managed by claudeconfig. <!-- claudeconfig:bundled -->

- Go/Golang @docs/Go.md,
  Modern Go, avoid deps but use Cobra, add tests
- Make/Makefile @docs/Make.md,
  ⚙️ phony sentinel, self-doc help, build dependency pattern
- Markdown @docs/Markdown.md,
  PascalCase for evergreens, kebab-case for ephemeral docs
- Git @docs/Git.md,
  conventional commits, work on the default branch, don't push unless asked
- Canary-first development @docs/Canary.md,
  probe external mechanisms before building features on them
- Spec system @docs/Spec.md,
  YAML spec files as single source of truth; Go code must not duplicate spec values
<!-- claudeconfig:end Language Conventions -->
