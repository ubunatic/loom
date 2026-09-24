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
