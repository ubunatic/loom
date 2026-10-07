# Upgrading Loom in Dependent Projects

How to move a project that imports loom to a newer release, and the breaking changes that need code
changes. Field data: the 2026-10-01 migration of 8 workspace repos
([study](studies/2026-10-01-loom-module-migration.md), issue 235).

## Breaking changes by release

| Release | Change | What callers do |
|---|---|---|
| Unreleased | Modal overlay capture and temporary mouse grab: overlays implement `ModalTarget` and capture all mouse events while active; backdrop clicks dismiss the overlay as `loom.Handled()` without leaking to underlying widgets. In mouse-off (`"m"`) mode, opening an overlay temporarily engages mouse tracking in the terminal and restores `"m"` on close (issues 303, 304) | No caller workarounds needed; containers and `Pane` route directly to active modal targets. Embed static previews with `Modeless: true` if modal capture is not desired |
| Unreleased | `RichTextEdit` hotkey and popover structure: box drawing bound to `ctrl-d` (`^D`) with Enter/Esc exiting box mode; "Draw" nested under "Box" submenu; bottom bar keycaps and status icons clickable (issue 301) | Use `ctrl-d` for box drawing or toggle via the Box submenu |
| Unreleased | Default editor and widget styles use `julia256`; inline/full primary-screen panes can use the terminal's last row. `RichTextEditDefaults.HotkeySaveAsBinding` was removed and Save as is menu-only. Popups dismiss on backdrop clicks; dialogs invoke their Cancel button (issue 302) | Apply `Theme("plain")` explicitly to retain plain styling; remove references to the deleted binding field. Set `Popup.DismissOnOutsideClick = false` to retain a popup on backdrop clicks. Hosts route input to `RichTextEdit` first while `ModalOpen()` is true, including mouse events outside its bounds |
| Unreleased | Keycaps derive from bindings: the spec `*_key` fields `Hotkey{Files,Box,Screenshot}Key` (editor) and `Hotkey{Help,Save,SaveAs,ViewEdit,Quit}Key` (rich text edit) were removed, `HintEntry.Key` may be empty and then renders `KeyCap(Binding)` (`⌃S`, `⌃⌥S`, `⇧Tab`), menu shortcuts accept the same glyphs, `KeyHelp` renders `KeyCap`, and the Save-as binding changed from `ctrl-shift-s` to `ctrl-alt-s`; the decoder now reports `ESC ^S` and `ESC[27;7;115~` as `ctrl-alt-s` instead of `alt-\x13` (issue 295) | Drop the removed fields (leave `HintEntry.Key` empty or call `loom.KeyCap(binding)`), match `ctrl-alt-s` for Save as, expect `^S` caps to now read `⌃S` in text assertions |
| Unreleased | Default quit keys: spec `fallback_quit_keys` is replaced by `quit_keys: [ctrl-q, f10]` and `escape_quits: false`; an unhandled `Esc`, `q` or `ctrl-d` no longer quits a pane, and `LibDefaults.FallbackQuitKeys` became `QuitKeys` (issue 297) | Quit via `^Q`/`F10`, handle `q`/`Esc` in your widget (`QuitResult()`), or set `Pane.EscapeQuits`; use `Pane.OnCloseRequest` to veto or confirm a close and `Pane.Quit()` to close later |
| Unreleased | Key decoding: `shift-insert`, `ctrl-insert`, `shift-delete` and `ctrl-delete` are no longer reported as plain `insert`/`delete` (issue 257, 40c153a) | Match the modified key names (`ctrl-insert`, `shift-insert`, `shift-delete`, `ctrl-delete`) where you handled `insert`/`delete` for them |
| Unreleased | Key decoding: modified Home/End (`ESC[1;2H`, `ESC[1;5F`, `ESC[1;2~`, `ESC[4;2~`, ...) are no longer reported as plain `home`/`end` (or dropped); they decode as `shift-home`, `ctrl-end`, `alt-home`, `ctrl-shift-end` and so on (issue 263) | Match the modified key names where you handled `home`/`end` for them |
| Unreleased | `RichTextEdit`: `ctrl-shift-b` no longer toggles box mode; F5 wraps an active selection in a box and toggles draw mode only without a selection; an open popover consumes arrows, Space and Esc (issue 264) | Bind box drawing to F5 or the popover Draw button; send arrows after closing the popover |
| Unreleased | `loom.Style` gained `Italic`, `Strike`, and `Invert` fields (issue 254 M1) | Use keyed `Style` literals or add values to positional literals |
| v0.2.18 | Module path `codeberg.org/ubunatic/loom` → `ubunatic.com/loom` (issue 235) | Rewrite imports, require `ubunatic.com/loom` |
| v0.2.15 | `Widget` lost the bool `HandleKey`/`HandleMouse`; only `ConsumeKey`/`ConsumeMouse` returning `EventResult` remain (9dd183b) | Map results as below |
| v0.2.6 | `EventResult` introduced (issue 126) | — |

Add a row here with every release that breaks callers.

## Module path

- `ubunatic.com/loom` is a vanity path: `https://ubunatic.com/loom?go-get=1` (301 to `/loom/`) serves
  `go-import "ubunatic.com/loom git https://codeberg.org/ubunatic/loom"`. Source: `~/projects/ubunatic.com/loom/index.html`.
- The code stays on Codeberg: repository URLs (`https://codeberg.org/ubunatic/loom`, issues, releases)
  do not change. Tags before v0.2.18 declare the old path, so `go get ubunatic.com/loom` needs ≥ v0.2.18.
- Rewrite only module/import paths, never URLs:
  `perl -pi -e 's{(?<![/@])codeberg\.org/ubunatic/loom}{ubunatic.com/loom}g' <files>`.

## Switching a dependent

1. Rewrite `.go` files and live docs; leave closed `issues/` and dated `docs/studies/` as history.
2. `go mod edit -droprequire codeberg.org/ubunatic/loom`, `go get ubunatic.com/loom@<tag>`, `go mod tidy`.
   Never text-edit `go.sum` or `go.work.sum`: a rewritten line is a checksum for a module that does not exist.
3. Repos that build against the checkout keep it: `replace ubunatic.com/loom => ../loom`.
4. Verify with `GOWORK=off go build ./... && GOWORK=off go vet ./...`: a `go.work` that uses `../loom`
   hides problems with the released module.

## bool → EventResult

In the old API, `HandleKey`/`HandleMouse` returning **`true` meant quit**, not "consumed".

| Old | New |
|---|---|
| `return true` | `return loom.Quit()` |
| `return false`, key was used (navigation, state change) | `return loom.Handled()` |
| `return false`, key not used | `return loom.Ignored()` |
| `return child.HandleKey(e)` | `return child.ConsumeKey(e)` (forward the result unchanged) |
| child confirmed a selection | check `result.Done` (see [Widgets](Widgets.md) §8) |

Pitfalls seen in the migration: mapping old `true` to `Handled()` (q no longer quits), and falling
through to `Ignored()` after a key that changed state, e.g. a wizard step. Add a test that the quit keys
return `Quit` and an unrelated key returns `Ignored`.

## Releasing so dependents can fetch

The Go module zip rejects paths that differ only in case; such a tag can never be fetched (v0.2.16,
issue 234). A test guards tracked paths. After a release, check from a scratch module:
`go mod init x && go get ubunatic.com/loom@<tag> && go build ./...`.
