# 043 — Widget Color Themes: `plain` and `mc` (Midnight Commander)

**Status**: Closed
**Priority**: P2 (Medium)
**Severity**: Moderate
**Category**: Feature
**Related**: [style.go](file:///home/uwe/projects/loom/style.go),
[choice.go](file:///home/uwe/projects/loom/choice.go),
[table.go](file:///home/uwe/projects/loom/table.go),
[grid.go](file:///home/uwe/projects/loom/grid.go),
[spec/themes.yaml](file:///home/uwe/projects/loom/spec/themes.yaml),
[spec/schemas/themes.schema.json](file:///home/uwe/projects/loom/spec/schemas/themes.schema.json),
[#037](file:///home/uwe/projects/loom/issues/037-canvas-drawborder-and-drawbox-primitives-with-configurable-boxstyles.md)

---

## 1. Problem & Motivation

Loom widgets carry their own disconnected style structs (`ChoiceStyle`,
`TableStyle`, `Grid.FocusBG`, …) and each `Default*Style()` factory hard-codes
monochrome terminal-default colors.  There is no shared concept of a *theme* —
a named palette that can be applied to the whole widget tree in one call.

Two concrete themes are wanted as first deliverables:

| Theme | Description |
|-------|-------------|
| `plain` | Terminal-default colors — identical to today's hard-coded defaults |
| `mc` | Midnight Commander blue-panel palette: white-on-blue body, cyan selected row, yellow header |

Scope is **color only** — border glyphs, rounded corners, and special rune sets
are deferred to #037.

---

## 2. Spec

Color roles are captured in [`spec/themes.yaml`](file:///home/uwe/projects/loom/spec/themes.yaml)
and validated by [`spec/schemas/themes.schema.json`](file:///home/uwe/projects/loom/spec/schemas/themes.schema.json).

### Color roles (per theme entry)

| Role | Applies to |
|------|-----------|
| `normal_fg` / `normal_bg` | Unselected rows in Choice and Table |
| `selected_fg` / `selected_bg` / `selected_bold` | Highlighted / active row |
| `header_fg` / `header_bg` / `header_bold` | Table column headers |
| `prompt_fg` / `prompt_bg` | Filter / command prompt row |
| `border_fg` / `border_bg` | Box border lines and title |
| `focus_bg` | Focused cell background in Grid |

Color values are **256-color palette indices**.  `0` is the sentinel for
*terminal default* (translates to `ColorReset()`).  True-color (RGB) overrides
are left for application code.

### `plain` snapshot

```yaml
normal_fg: 0 / normal_bg: 0   # terminal default
selected_fg: 0 / selected_bg: 0 / selected_bold: true
header_fg: 0  / header_bg: 0  / header_bold: true
focus_bg: 238
```

### `mc` snapshot

```yaml
normal_fg: 15 / normal_bg: 27     # white on blue
selected_fg: 0  / selected_bg: 51 # black on cyan
header_fg: 226  / header_bg: 27   # yellow on blue
prompt_fg: 0    / prompt_bg: 51   # black on cyan
border_fg: 15   / border_bg: 27   # white on blue
focus_bg: 33
```

---

## 3. Proposed Implementation

### 3.1 `ThemeColors` struct and `spec/themes.yaml` loader

```go
// ThemeColors is the Go representation of one theme entry in spec/themes.yaml.
type ThemeColors struct {
    NormalFG     uint8 `yaml:"normal_fg"`
    NormalBG     uint8 `yaml:"normal_bg"`
    SelectedFG   uint8 `yaml:"selected_fg"`
    SelectedBG   uint8 `yaml:"selected_bg"`
    SelectedBold bool  `yaml:"selected_bold"`
    HeaderFG     uint8 `yaml:"header_fg"`
    HeaderBG     uint8 `yaml:"header_bg"`
    HeaderBold   bool  `yaml:"header_bold"`
    PromptFG     uint8 `yaml:"prompt_fg"`
    PromptBG     uint8 `yaml:"prompt_bg"`
    BorderFG     uint8 `yaml:"border_fg"`
    BorderBG     uint8 `yaml:"border_bg"`
    FocusBG      uint8 `yaml:"focus_bg"`
}
```

The `uint8` zero value naturally maps to `ColorReset()` (index sentinel = 0).
Load via `//go:embed spec/themes.yaml` in the same pattern as `defaults.yaml`.

### 3.2 Helper: convert `uint8` → `Color`

```go
// themeColor converts a spec palette index to a Color.
// 0 is the "terminal default" sentinel and returns ColorReset().
func themeColor(index uint8) Color {
    if index == 0 {
        return ColorReset()
    }
    return ColorIndex(index)
}
```

### 3.3 `ThemeColors` → `ChoiceStyle` / `TableStyle`

```go
func (t ThemeColors) ChoiceStyle() ChoiceStyle {
    return ChoiceStyle{
        Normal:   Style{FG: themeColor(t.NormalFG),   BG: themeColor(t.NormalBG)},
        Selected: Style{FG: themeColor(t.SelectedFG), BG: themeColor(t.SelectedBG), Bold: t.SelectedBold},
        Prompt:   Style{FG: themeColor(t.PromptFG),   BG: themeColor(t.PromptBG)},
        Border:   Style{FG: themeColor(t.BorderFG),   BG: themeColor(t.BorderBG)},
    }
}

func (t ThemeColors) TableStyle() TableStyle {
    return TableStyle{
        Normal:     Style{FG: themeColor(t.NormalFG),   BG: themeColor(t.NormalBG)},
        Selected:   Style{FG: themeColor(t.SelectedFG), BG: themeColor(t.SelectedBG), Bold: t.SelectedBold},
        Header:     Style{FG: themeColor(t.HeaderFG),   BG: themeColor(t.HeaderBG),   Bold: t.HeaderBold},
        SortHeader: Style{FG: themeColor(t.HeaderFG),   BG: themeColor(t.HeaderBG),   Bold: t.HeaderBold, Underline: true},
        Prompt:     Style{FG: themeColor(t.PromptFG),   BG: themeColor(t.PromptBG)},
    }
}

func (t ThemeColors) FocusBGColor() Color {
    return themeColor(t.FocusBG)
}
```

### 3.4 Named theme accessor

```go
// SpeccedThemes holds all themes loaded from spec/themes.yaml at init time.
var SpeccedThemes map[string]ThemeColors  // keys: "plain", "mc", …

// Theme returns the named theme, falling back to "plain" if name is unknown.
func Theme(name string) ThemeColors {
    if t, ok := SpeccedThemes[name]; ok {
        return t
    }
    return SpeccedThemes["plain"]
}
```

---

## 4. Out of Scope

- Border glyph variants (rounded, double, ASCII) → #037
- True-color / RGB theme entries
- Runtime theme switching or hot-reload
- Per-box overrides; the theme provides defaults, callers may still override
  individual `Style` fields afterwards

---

## 5. Verification & Acceptance

- `spec/themes.yaml` validates against `spec/schemas/themes.schema.json` with
  a schema-validator test (or CI lint step).
- `Theme("plain").ChoiceStyle()` equals `DefaultChoiceStyle()` exactly.
- `Theme("plain").TableStyle()` equals `DefaultTableStyle()` exactly.
- `Theme("unknown")` falls back to `"plain"` without panic.
- A smoke-test or example that renders a `Choice` and a `Table` under `mc`
  theme visually matches the Midnight Commander blue-panel palette (golden or
  manual review).
- No existing golden tests break — `plain` must be byte-for-byte identical to
  current defaults when rendered to a canvas with ANSI disabled.
