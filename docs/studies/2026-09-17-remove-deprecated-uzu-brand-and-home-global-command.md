---
title: Remove Deprecated uzu Brand and :home Global Command
---
# Remove Deprecated `uzu` Brand and `:home` Global Command

## Header & Context

Date: 2026-0917. Resolved ticket [issues/047](../../issues/047-remove-deprecated-uzu-brand-and-default-global-commands-from-cmdbar.md) (P3, Cleanup/Refactor). This is a focused change summary, not a sprint retrospective — it records one small cleanup: dropping the last hardcoded references to the deprecated `uzu` tool brand from `loom` runtime code and re-scoping the built-in command bar so `:home` is host-declared rather than a hard default.

Status at time of writing: committed (`e561e44`) and issue closed. The diffs below are taken verbatim from that commit.

## Executive Summary

The `uzu` CLI was renamed away, but `loom/cmd.go` still shipped a default global command titled `"go to uzu main menu"`, and a comment in `pane.go` used `$(uzu)` as its example. Both were stale brand references with no functional meaning left. This change removes the `:home` entry from the built-in global command list (so `:help` no longer advertises an outdated title), generalizes the `pane.go` comment, and keeps the underlying `NavHome` navigation capability fully intact for hosts that opt in via `AddCmd`.

The change is deliberately small: one line deleted from production code, one comment generalized, and test updates that preserve coverage of the `NavHome` plumbing by registering `:home` locally where it was previously relied on by default.

## What Changed

### 1. Default globals no longer include `:home` (`cmd.go`)

Before, `newCmdBar()` seeded three global commands; after, two:

```diff
-			{Name: "home", Title: "go to uzu main menu"},
```

:help and :back remain because they are generic and host-agnostic. :home is removed only because its *title* named a deprecated product; the mechanism it exercised (`NavHome`, which unwinds the widget chain to the application root) is unchanged and still reachable. A host that wants a home action now declares it explicitly:

```go
c.AddCmd(loom.Cmd{Name: "home", Title: "go to root"})
```

This matches the ticket's guidance to either remove or make such commands host-configurable, while keeping the built-in surface minimal and brand-neutral.

### 2. Stale comment generalized (`pane.go`)

The `Pane.New` doc comment used a dead brand as its example of why `/dev/tty` (not `os.Stdin`) matters:

```diff
-// Always uses /dev/tty — never os.Stdin — so the pane works inside ZSH
-// command substitution (result=$(uzu)) where stdin may be a pipe.
+// Always uses /dev/tty — never os.Stdin — so the pane works inside shells
+// and command substitution (e.g. `foo=$(bar)`), where stdin may be a pipe.
```

The explanation is identical; only the example is now brand-neutral.

### 3. Tests keep `NavHome` coverage by registering `:home` locally (`cmd_test.go`)

Two tests previously depended on `:home` being present by default. They now register it themselves, so they still exercise the exact code path under test (command dispatch returning `cmdHome`, which `Choice`/`Table` map to `NavHome`):

- `TestCmdHomeSetsNavHost`: adds `c.AddCmd(loom.Cmd{Name: "home", Title: "go to root"})`.
- `TestTableCmdHomeViaColumn`: adds the same local registration on the table.

A third assertion in `TestCmdHelpInvokesRunner` expected the rendered help list to contain `":home"`. Since `:home` is no longer a default global, that string was removed from the expectation while the rest of the check (`:help`, `:back`, the host-added `:custom`, and the footer) stays intact.

## What Was Not Changed

- The `NavHome` constant, the `cmdHome` result value, and the `case cmdHome:` handlers in `choice.go` and `table.go`. These are the real capability; only the *default advertisement* of them changed.
- No public API signature changed. `AddCmd` already existed and is the supported way for hosts to add commands.
- No behavior change to key routing, matching, or help rendering beyond the removed default entry.

## Verification

- `grep -rn uzu --include=*.go .` returns no matches (was two before).
- `go vet ./...` clean.
- Full suite passes: `go test ./...` → all packages `ok`.
- Targeted run confirms the affected tests still pass:
  - `TestCmdHomeSetsNavHost` PASS
  - `TestTableCmdHomeViaColumn` PASS
  - `TestCmdHelpInvokesRunner` PASS
- No new gofmt issues introduced; the files touched were not newly flagged by `gofmt -l` (the pre-existing non-gofmt-clean files were already so at HEAD).

## File & Diff Summary

| Commit | Change |
|---|---|
| `e561e44` | Remove `:home` global from `newCmdBar()`, generalize `pane.go` comment, update three tests in `cmd_test.go`. |
| `06ab421` | Close issue 047 and update the issue index (`harnez issues done`). |

Files changed in the code commit:

- `cmd.go`: −1 line (removed the `:home` global entry).
- `pane.go`: 2-line comment rewrite (brand-neutral example).
- `cmd_test.go`: two tests now register `:home` locally; one help-render expectation drops `":home"`.

No source files deleted. The study and its README index entry are recorded in a subsequent documentation commit.
