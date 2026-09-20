---
title: Filebrowser Feature Gap Analysis
---
# Filebrowser Feature Gap Analysis

## Scope

Review of `examples/filebrowser/filebrowser/browser.go` and its tests for
logic that is likely reusable Loom framework functionality.

## Gaps and opportunities

| Example code | Gap | Rough Loom API |
| --- | --- | --- |
| `browser.open`, `os.ReadDir`, `filepath.Dir`, `filepath.Join` | No directory-browser model or parent/child navigation helper. The example rebuilds `loom.Item` rows, path mapping, and selection restoration. | `loom.ReadDirectory(path string, opts DirectoryOptions) (Directory, error)`; `Directory.Parent() (string, bool)`; `Directory.EntryPath(name string) string`; `loom.NewFileBrowser(model Directory, onOpen func(FileEntry) error) *FileBrowser` |
| `browser.open` selection callback | File-vs-directory dispatch, `os.Stat`, and navigation/open error handling are application code. | `FileBrowser.OnActivate(func(FileEntry) error)` with default directory navigation and `FileEntry.Kind` (`Directory`, `Regular`, `Symlink`, `Special`) |
| `launchFile` | Desktop-file opening through `exec.Command("xdg-open", ...)`, detached process setup, and swallowed child output are hard-coded platform behavior. | `loom.OpenFile(path string) error` or `loom.FileOpener` with platform-specific implementations; allow injection for tests. |
| `displayName` | Control-character escaping is implemented locally with `unicode.IsPrint` and `strconv.QuoteToASCII`. | `loom.DisplayPath(path string) string` / `loom.QuoteUnprintable(value string) string`, with terminal-safe width/truncation options. |
| `metadata` | File type, size, mode, mtime, symlink target, and error formatting are duplicated presentation logic. | `loom.FileDetails(path string) ([]string, error)` or a `FileInfoView` widget with configurable fields and formatter. Use `os.Lstat` semantics explicitly. |
| `detailView`, `updateDetails`, `equalLines` | A scrollable detail pane synchronized with the selected row requires custom state updates and redraw deduplication. | `loom.Inspector` / `loom.DetailPane` bound to `Selection[T]`, with `SetLines` or `SetValue` and automatic refresh. |
| `frame` setup in `newBrowser` and `Draw` | The common file-list plus preview/detail split, focus titles, breakpoint, borders, and status affordances are manually composed. | `loom.SplitBrowser` or `loom.FileBrowserLayout{List, Detail, Preview, Breakpoint}`; expose `List`, `Detail`, and `Preview` slots rather than hard-coding a policy. |

## Recommended scope

The strongest framework candidates are a reusable `FileBrowser` widget,
directory-entry model/navigation helpers, and an injectable `OpenFile` helper.
Path escaping and metadata formatting should begin as small standalone helpers;
the detail/preview layout can then consume them without forcing one browser UX.

Theme cycling, command-line flag parsing, `xdg-open` status messages, and the
choice of metadata fields remain example/application policy.

## Boundary notes

The API should preserve filesystem semantics: use `Lstat` when describing a
symlink, avoid following links merely to render metadata, retain the original
entry name separately from its display label, and return errors to the widget's
caller rather than embedding notices such as `"Open failed: ..."`.
