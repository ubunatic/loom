# Loom edit design proposals

Static ANSI mockups for review; the corresponding features are tracked in issues 280–283.
The user approved these designs as visual guidance. The final app may look different depending on how standard Loom widgets render and compose. New UI elements must use standard Loom widgets; small tweaks are acceptable, heavy hacks are not. For larger observed gaps in Loom, developers must always file or link a library issue rather than add an app workaround or force the mockup's exact appearance.

Each file is 100 columns by 24 rows. Open in a terminal at least 106 columns wide:

```bash
loom view docs/design/loom-edit-01-editor.ansi
loom view docs/design/loom-edit-02-file-browser.ansi
loom view docs/design/loom-edit-03-search-normal.ansi
loom view docs/design/loom-edit-04-search-regex.ansi
loom view docs/design/loom-edit-05-settings.ansi
```

- **Editor**: file action, filename and active mouse/theme status above the document; path, cursor and alternate-screen status below it; compact key hints at the bottom.
- **File browser**: a left panel toggled by F2, with the selected file highlighted. Tab moves focus; Enter opens. Initial panel visibility remains a design choice.
- **Search**: a compact panel at the top right, shown by F3 or ^F. Proposed navigation is Enter / Shift-Enter for next / previous; Tab reaches the Normal/Regex controls, and Esc closes the panel. A yellow document span marks the current match.
- **Settings**: the document displays the requested config file with `theme: mc-dark`, `mousegrab: true`, and `altscreen: true`. The adjacent schema path is illustrative; issue 281 must supply and document the actual schema. This proposes YAML booleans for on/off. The settings preview is an editable document, not a new preferences dialog.

Mouse capture and alternate-screen status show the enabled state as a proposal, not a decision about defaults. Colors are illustrative. The mockups preserve existing Save/Save as/Box/Quit hints.

Validation: `loom check-box` passes for all five assets; `loom eval` reports 24 rows, 100 columns and zero ragged rows for each; `loom measure` confirms every row occupies 100 terminal cells.
