# 105 — Unify filebrowser navigation pane across examples

**Status**: Open
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: `examples/ansiviewer/ansiviewer/viewer.go`,
`examples/filebrowser/filebrowser/browser.go`,
`examples/filebrowser/filebrowser/filebrowser.go`,
[063](063-convert-filebrowser-example-to-a-hostable-widget.md)

## Goal

Provide one shared filebrowser navigation-pane implementation and use it in
all file-browser examples, so navigation behavior and interaction contracts do
not diverge between `ansiviewer`, `filebrowser`, and future consumers.

## Required behavior

- ESC goes back one directory; at the starting/filesystem root it retains the
  established quit behavior.
- Returning to a parent lands with the directory just left selected.
- `/` enters file filtering/search mode and filters the visible files.
- Mouse-clicking an item selects it and updates the preview/metadata.
- Keyboard navigation, directory opening, preview handling, and quit behavior
  remain available through the shared pane API.

## Acceptance

- The examples no longer maintain separate copies of the filebrowser
  navigation logic.
- Shared tests cover ESC back, selection restoration, slash filtering, and
  mouse selection, including the hosted/pane key-consumption boundary.
- Existing standalone and hosted example tests pass, and a manual smoke pass
  confirms the same interactions in each example.

When work starts, verify the live implementations and recent history first;
the current examples may have evolved beyond the paths listed above.

## Milestones (lean sprint, developer luna → flash37 → terra → opus ladder)

Preflight (HEAD 505455e): 777b664 aligned behavior, but ansiviewer `viewer.go` and
filebrowser `browser.go` still carry duplicated navigation logic.

- **M1 — Shared pane** in `examples/filebrowser/filebrowser` (063 builds on it): navigation,
  selection, `/` filter, ESC back + parent-selection restore, mouse select (0-based child-local),
  quit request, key-consumption report. API: follow existing loom widget conventions (callbacks vs
  observable state — pick the one the repo already uses and say which). Pane-level unit tests.
- **M2 — filebrowser on the pane**: delete its duplicate navigation; keep metadata/theme.
- **M3 — ansiviewer on the pane**: delete its duplicate navigation; keep ANSI preview.
- **M4 — Validate**: standalone + hosted example tests; note manual smoke items for the user.

M1 first pass (8e1105a, luna): shared `NavigationPane` with callbacks (OnSelection/OnActivate/OnOpen/OnQuit). Host suite run: 4 pane tests fail. Escalated to flash37.

Pre-Work / Required Refinements (M1):

- Make the four `TestNavigationPane*` tests green without weakening them; failures:

```text
    navigation_test.go:44: activation callback name = "..", want activated:beta
    navigation_test.go:61: opened directory = "/tmp/TestNavigationPaneOpensDirectoryAndRestoresParentSelection1232952098", want "/tmp/TestNavigationPaneOpensDirectoryAndRestoresParentSelection1232952098/001/child"
    navigation_test.go:81: root escape = quit:false consumed:true, want true,true
    navigation_test.go:100: selection after local row 1 click = {Name:alpha Path:/tmp/TestNavigationPaneMouseSelectionUsesChildLocalCoordinates1360201472/001/alpha Kind:0 IsParent:false}, ok=true; want beta
FAIL
FAIL	codeberg.org/ubunatic/loom/examples/filebrowser/filebrowser	0.606s
```

M1 delivered (8e1105a + flash37 fix): shared `NavigationPane` (callbacks OnSelection/OnActivate/OnOpen/OnQuit,
`ConsumeKey` boundary); mouse maps child-local back into the frame `Choice` drew in (Choice subtracts its own rect).
Host suite green.

Pre-Work for M4 (docs, user request): update the evergreen docs this touches — `docs/Widgets.md` (shared
pane, callback API, the Choice rect bridge), example docs/READMEs, `docs/README.md` index if a doc is added.
Only edit non-managed sections of AGENTS.md; harnez-managed blocks stay untouched.

M2 first pass (luna, uncommitted WIP in `browser.go`/`navigation.go`): browser integration + PTY click
tests red. Escalated to flash37.

Pre-Work / Required Refinements (M2):

- Root cause per luna: the pane replaces its `Choice` on navigation while the frame caches the old child,
  and filebrowser treats the pane's start dir as quit root. Fix the ownership, e.g. the pane keeps one
  stable `Choice` (swap items, not the instance) and the frame hosts the pane itself, not `pane.List()`.
- Keep existing filebrowser tests unweakened; the M1 pane tests must stay green.

M2 second pass (flash37, timed out at 10 min, uncommitted WIP): added `Choice.SetItems`/`SelectIndex` so the pane keeps one stable `Choice`; frame hosts the pane. Status unknown. Escalated to terra.

M2 delivered (80d9e8d, terra): filebrowser hosts the shared pane; duplicate navigation removed; `Choice.SetItems`/`SelectIndex` with unit tests. Host suite green.

M3 first pass (luna, uncommitted WIP in ansiviewer `viewer.go`/`viewer_test.go`): host suite red. Escalated to flash37.

Pre-Work / Required Refinements (M3):

- File list renders empty (framed recording shows no rows; `TestRecordWritesOneSnapshotAfterDelay`).
- ESC in a child dir does not go to parent (`TestBrowserEscapeGoesToParentDirectory`); framed ESC at
  a non-root dir quits (`TestBrowserFilterAndMouseSelection` wants quit:false).
- `q` no longer quits: `TestViewerPTYShowsFilesAndQuits` and the ansiviewer PTY smoke hang.
- Luna routed navigation keys from the host around frame focus; prefer filebrowser's 80d9e8d wiring.

M3 delivered (c046cc2, flash37 after luna): ansiviewer hosts the shared pane; duplicate navigation removed.

Pre-Work / Required Refinements (M4):

- `TestBrowserFilterAndMouseSelection` now only checks `Query() == "beta"`; restore the check that the
  visible list is exactly `beta.txt` (via the pane's filtered items), and drive the filter through
  the browser/frame boundary as before, not the pane directly.
- Docs (see M1 note): `docs/Widgets.md` (shared `NavigationPane`, callbacks, `ConsumeKey`, the Choice
  rect bridge, `Choice.SetItems`/`SelectIndex`), example READMEs if they describe navigation.
- Write `docs/manual/105-smoke.md`-style checklist? No: list the manual smoke items for the user in
  your report instead (ESC back, parent reselect, `/` filter, click select, in both examples).
