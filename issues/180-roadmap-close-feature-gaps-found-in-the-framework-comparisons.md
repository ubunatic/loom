# 180 — Roadmap: close feature gaps found in the framework comparisons

**Status**: Closed — roadmap delivered, execution tracked in docs/progress/roadmap.ansi
**Priority**: P1 (High)
**Severity**: Moderate
**Category**: Planning
**Related**: issues/177, issues/178, issues/179

---

## 1. Problem & Motivation
177-179 list features Bubble Tea, ncurses, tview/Ratatui/Textual have and loom lacks. We need one ordered plan to close the gaps that fit loom's design.

## 2. Scope & Rules
- Input: the Findings sections of 177-179. Merge duplicates; drop items marked "does not fit design".
- Group into phases by dependency (foundations like input/terminal capabilities before widgets that need them) and value.
- Every roadmap item maps to exactly one ticket: reuse an existing open ticket or file a new one (`harnez issues new`, lean, with a `/goal` and exit clause). Each ticket must be small enough for one lean sprint.
- Write the roadmap as `## Roadmap` in this ticket: phase, ticket number, title, size, depends-on.

## 3. Implementation & Verification Plan
/goal A committed roadmap in this ticket where every item points to a lean-sprint-sized ticket, in execution order. Stop and report when blocked on a user decision or denied permission.

## Roadmap

Execution order, top to bottom. Every item is one ticket with a `/goal` and exit clause, sized for one lean sprint (S = small, M = medium). Tickets that need a human to judge the look or interaction are listed separately under **Human feedback**, so the host can schedule them when a reviewer is available; nothing in the main list depends on them.

### Main (agent-verifiable)

| # | Phase | Ticket | Title | Size | Depends on |
|---|---|---|---|---|---|
| 1 | 1 Foundations | 154 | Dynamic cadence and timer control on Ticker | S | — |
| 2 | 1 Foundations | 158 | KeyMap and action key aliasing (with help labels) | S | — |
| 3 | 1 Foundations | 181 | Bracketed paste event for TextInput and TextArea | S | — |
| 4 | 1 Foundations | 182 | Terminal color profile detection and downsampling | M | — |
| 5 | 1 Foundations | 156 | Canvas SubCanvas and Blit | S | — |
| 6 | 2 Input polish | 173 | Masked password mode for TextInput | S | — |
| 7 | 2 Input polish | 184 | TextInput horizontal scrolling | S | 173 |
| 8 | 2 Input polish | 185 | TextArea content-driven height bounds | S | — |
| 9 | 2 Input polish | 189 | Fuzzy ranking in Choice filter | S | — |
| 10 | 2 Input polish | 174 | Standalone NumberInput and Toggle | S | — |
| 11 | 2 Input polish | 183 | Compact key help from KeyMap | S | 158 |
| 12 | 3 Status widgets | 186 | Spinner widget | S | 154 |
| 13 | 3 Status widgets | 167 | ProgressBar remainder (indeterminate, labels, styles) | S | 154 |
| 14 | 3 Status widgets | 187 | Timer and Stopwatch widgets | S | 154 |
| 15 | 3 Status widgets | 188 | Paginator widget | S | — |
| 16 | 4 Composites | 190 | Viewport for arbitrary child content | M | 156 |
| 17 | 4 Composites | 168 | Tree widget | M | — |
| 18 | 4 Composites | 172 | FilePicker (with extension filter) | M | 168 (optional) |
| 19 | 4 Composites | 169 | Form and FieldGroup | M | 173, 174 |
| 20 | 4 Composites | 191 | Table cell cursor and frozen columns | M | — |
| 21 | 5 Media | 160 | Media playback controls M2 (fix PTY test) | S | — |

| 22 | 5 Extras | 194 | OSC 52 clipboard copy | S | 158 (optional) |

### Human feedback (host decides look and behavior, builds after the main list; review later in 102)

| # | Ticket | Title | Size | Depends on | What needs judging |
|---|---|---|---|---|---|
| H1 | 155 | Dialog on top of Popup | M | — | button row and placement look |
| H2 | 170 | Menu and MenuBar, one level | M | 155, 158 | menu interaction and look |
| H3 | 193 | Nested submenus | S | 170 | submenu placement and hover behavior |
| H4 | 175 | DatePicker | M | — | layout, week start, time-of-day scope |
| H5 | 192 | Chart widget (line, grouped bars) | M | — | axis labels and glyph choice |
| H6 | 112 | Media controls (play/pause, zoom, pan) | M | 160 | control layout |

### Not scheduled

- **System clipboard (OSC 52)** — 177 says optional, 179 says fit unclear and no user need established; decide before filing.
- **terminfo database, curses API, Elm app model, string styling DSL, DOM/CSS, async workers, graphics protocols** — marked "does not fit design" in 177-179.
- **Choice status messages and built-in spinner, Table rich cells** — optional extras; revisit after 186 and 191.
- **Popup-as-modal, layers, mouse, resize, Unicode/ACS, layouts, text views** — no gap per 177-179.
