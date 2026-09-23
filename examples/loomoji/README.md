# loomoji

An inline searchable emoji, symbol, and box-drawing picker built with Loom.

## Installation

Install the command and Zsh widget script from the repository root:

```sh
make install
```

Add the widget to Zsh by sourcing `loomoji.zsh` from `.zshrc`, then choose a
key binding:

```zsh
source ~/.local/share/loomoji/loomoji.zsh
bindkey '^X^M' loomoji-insert-widget
```

Invoke the binding while editing a command to insert the selected emoji or symbol
at the current command-line cursor. Running `loomoji` as a regular command prints
the chosen character to standard output instead.

## Controls

- **Arrow Keys** (`↑` `↓` `←` `→`): Navigate the character grid (focused on start).
- **Tab**: Switch focus between character grid and search input.
- **`1`–`9`, `0`**: Jump directly to category tabs 1 through 10.
- **`[` / `]`** or **`PgUp` / `PgDn`**: Cycle categories backward and forward.
- **Typing text**: Filter and fuzzy-search across all emojis, symbols, and box characters.
- **Enter**: Copy the selected emoji / character and exit.
- **Esc** / **Ctrl-C**: Cancel and exit without output.
- **Mouse**: Click category tabs to switch categories, hover to preview, or click any grid cell to select immediately.

## Categories

- **Faces** (`😀`): Smileys, emotions, gestures, reactions, and creatures.
- **Hands** (`👋`): Hand gestures, body parts, and poses.
- **Animals** (`🐾`): Mammals, birds, reptiles, aquatic life, bugs, and mythical creatures.
- **Food** (`🍔`): Fruits, meals, drinks, snacks, and tableware.
- **Sports** (`⚽`): Activities, equipment, medals, games, and music.
- **Travel** (`🚀`): Vehicles, places, maps, landmarks, and buildings.
- **Objects** (`💡`): Tools, hardware, electronics, office supplies, and household items.
- **Hearts** (`❤️`): Hearts, love symbols, zodiac signs, and spiritual marks.
- **Nature** (`🌿`): Sky, celestial bodies, weather, plants, and flora.
- **Symbols** (`🔣`): Status marks, warnings, traffic indicators, and math symbols.
- **Arrows** (`➔`): Directional, double, wave, paired, logic, and flow arrows.
- **Blocks** (`█`): Block elements, shades, quadrants, geometric shapes, and symbols.
- **Lines** (`─`): Single, double, dashed, dotted, and heavy line rules.
- **Box** (`┼`): Box-drawing corners, junctions, tees, crosses, and frames.
