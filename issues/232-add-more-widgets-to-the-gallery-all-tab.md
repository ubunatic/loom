# 232 — Add more widgets to the gallery All tab

**Status**: Closed — All tab shows 24 demos (M1 cc85872/00b3a19); M2 43ed8bc fixed PTY tests broken by 231 key move (F2->F9/F10) and Tree wait race from 232. Host: make test-q1 green (exit 0, no FAIL), make install, loom widgets --show All checked in a PTY: all 24 widgets render in their cells.
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 229 (All tab, 3-column grid)

---

/goal Show every gallery demo widget that fits a grid cell on the "All" tab of the loom widget gallery,
with tests passing, or stop and report when blocked on a user decision (e.g. which large widgets to leave out).

## 1. Problem & Motivation
The "All" tab (`newAllDemo` in `gallery/gallery.go`, issue 229) shows only 12 of the 26 gallery
demos: Button, Toggle, Checkbox, NumberInput, Badge, PillCluster, ProgressBar, Sparkline, Spinner,
Stopwatch, Timer, Paginator. The user wants more widgets there so the overview is representative.

## 2. Technical Specification / Findings
Demos not yet on the All tab (as of commit 641bd44): Chart, DatePicker, Choice, Dialog, FilePicker,
Form, KeyHelp, MenuBar, Media, PaintCanvas, Popup, Table, Tabs, TextArea, TextInput, Tree, Viewport.

- Compact candidates (one grid cell): TextInput, Choice, DatePicker, KeyHelp, MenuBar, small Chart/Table/Tree.
- Large or modal widgets (Dialog, Popup, FilePicker, Media, PaintCanvas, Form, TextArea, Viewport)
  may need a reduced size, a taller row, or be left out; record the choice in the ticket.
- Focus and mouse routing on the All tab must keep working for every added widget (see 229, 199, 204).

Re-check the live demo list in `gallery/gallery.go` before starting.

## 3. Implementation & Verification Plan
- Extend `newAllDemo` (grid may grow more rows); keep 3 columns.
- Extend `gallery/gallery_test.go`: each added widget draws inside its cell and receives focus/keys/mouse.
- Run `make test-q1`, `make install`, and check `loom widgets --show` manually.

## M1 delivered

The All tab now has 24 demos in a 3-column grid: the original 12 plus TextInput,
Choice, DatePicker, KeyHelp, MenuBar, Chart, Table, Tree, Dialog, Popup, TextArea,
and Viewport. Dialog and Popup are reduced to fit a cell; they return unconsumed
Tab and Shift-Tab events so gallery tab navigation remains available.

FilePicker is omitted because its file listing needs more room; Form is omitted
because its fields and actions need more vertical space; Media and PaintCanvas
are omitted because their visual content needs a larger cell; Tabs is omitted
because a nested tab bar is not useful in the compact overview.

The follow-up fix uses public mouse routing to focus cells in tests. `go vet
./gallery/` and `go test -count=1 -run XXX ./gallery/` passed. The subsequent
`make test-q1` run failed in PTY-related tests (`TestPaneFirstDrawUsesScreenBounds`,
`TestGalleryChoiceQuitContractPTY`, `TestGalleryThemeFooterSurfacePTY`,
`TestGalleryNumberInputRangeAlignmentPTY`, `TestWidgetsPTYClickTabAndTreeDisclosure`,
and `TestWidgetsPTYSizeAndF2ThemePropagation`). `TestAllTabAddedWidgetsStayInCellsAndRouteInput`
passed; no `vet:` errors occurred in that run.

## M2 — Pre-Work / Required Refinements (host, 2026-10-01)
Host review of `make test-q1` on c9a185c (log: /tmp/loom-234-test-q1.log) found gallery PTY failures:
`TestGalleryChoiceQuitContractPTY` (Choice, Choice+Popup, popup_consumes_first_escape — timed out waiting
for "Theme: julia256"), `TestGalleryThemeFooterSurfacePTY` (Choice, FilePicker, Media, Table, Tree),
`TestGalleryNumberInputRangeAlignmentPTY`, `TestWidgetsPTYClickTabAndTreeDisclosure` ("click inside Tree
did not collapse src"), `TestWidgetsPTYSizeAndF2ThemePropagation`. Also root
`TestPaneFirstDrawUsesScreenBounds/wrapped-request-auto-alt` (W 99 vs 100).
1. Bisect: run the failing tests (targeted `go test -run`) on 641bd44 (before 232), cc85872 and HEAD to
   attribute each failure to 231, 232 or pre-existing/environment.
2. Fix every failure caused by 232 or 231 in the library or gallery (not by loosening assertions);
   failures that are pre-existing and unrelated get their attribution recorded here, not fixed.
3. Then one `make test-q1` run; record the real result.

## M2 delivered

Ran each listed test individually with `go test -count=1 -run '^<Name>$'
`./gallery/` (root bounds test: `./`) in isolated worktrees at 641bd44,
cc85872, and HEAD (479d3d6). Also tested 856cd6d, the parent of the issue 231
implementation, to distinguish 231 regressions from older failures.

| Test | Before 231 (856cd6d) | 641bd44 | cc85872 (*) | HEAD | Attribution |
|---|---|---|---|---|---|
| `TestGalleryChoiceQuitContractPTY` | Pass | Fail: three theme waits | Same failure | Same failure | 231: stale F2 input; Choice text keys now reach the child |
| `TestGalleryThemeFooterSurfacePTY` | Pass | Fail: all five theme waits | Same failure | Same failure | 231: stale F2 input |
| `TestGalleryNumberInputRangeAlignmentPTY` | Pass | Fail: theme wait | Same failure | Same failure | 231: stale F2 input |
| `TestWidgetsPTYClickTabAndTreeDisclosure` | Pass | Pass | Fail: src stays expanded | Same failure | 232: stale screen coordinates after an ambiguous content wait |
| `TestWidgetsPTYSizeAndF2ThemePropagation` | Pass | Fail: theme wait | Same failure | Same failure | 231: stale F2 input |
| `TestPaneFirstDrawUsesScreenBounds` | Pass | Pass | Pass | Pass | Pre-existing/environment category: reported failure not reproduced; unrelated to 231/232 |

(*) Unmodified cc85872 cannot compile the gallery tests: its new All-tab test
calls private `Grid.setFocus` and nonexistent `Canvas.Width`. Every requested
gallery diagnostic was attempted and hit that compile failure, attributable to
232 and already fixed by 00b3a19. The runtime results above use cc85872 with
only 00b3a19's existing test-compilation repair applied in the scratch worktree;
gallery production code remains exactly cc85872. Root tests pass unmodified too.

Updated PTY theme inputs to F9 and renamed the theme propagation test accordingly.
Choice confirmation still must leave the gallery running; F10 then verifies
unconditional exit, matching 231's contract that consumed text keys do not quit.
The existing unconsumed q, F10, and Escape exit cases remain intact. Tree switching
now waits for standalone-only `tree.go` before locating the disclosure marker:
232 added `app.go` to All, so the old wait could return on the previous panel.
All color, size, range, collapse, and quit assertions remain in place.
These were PTY test input/synchronization defects, not library routing defects;
no widget workaround or event-routing design change was needed.

The root bounds test and its Pane/resize implementation are unchanged from before
231. Its reported 99-column failure did not recur in any targeted diagnostic or
the final suite; no unrelated root changes were made.

Validation: `go vet ./gallery/` passed before the final suite. All six targeted
post-fix tests passed with `-count=1` (including the renamed F9 test). Exactly one
final `make test-q1` passed, exit 0, including spec validation, geometry replay,
repository-wide vet, and tests. Gallery passed in 24.938s; `grep -n -- '--- FAIL'`
found no matches. Full output: `/tmp/loom-232-m2.F99puH/test-q1.log`; per-revision
diagnostic logs and the compile-repair patch are in the same scratch directory.
Ticket remains open for host review; no release, tag, or push performed.
