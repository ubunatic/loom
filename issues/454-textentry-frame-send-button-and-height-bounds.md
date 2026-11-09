# 454 — TextEntry frame, send button and height bounds

**Status**: Closed — delivered with the frame, send button, row bounds and gallery controls (owner request, taia wants these in loom)
**Priority**: P2 (Medium)
**Severity**: Minor
**Category**: Feature
**Related**: 448, 453

---

## 1. Problem & Motivation
taia draws its own send button and its own rounded box around `TextEntry` (`chat.go` `drawButtons`, `DrawBorder`). Other hosts want the same chrome, so it belongs in the library. Row bounds should be settable, with a share of the app height as the default maximum.

## 2. Technical Specification / Findings
Decisions (decide mode):
- `TextEntryOptions.Frame` / `SetFrame`: `TextEntryFrameNone` (zero), `TextEntryFrameBox` (rounded box, the taia.legacy look: title and hint in the top border, footer in the bottom border, one cell of padding inside the side borders), `TextEntryFrameSimple` (full-width rules above and below, no side borders). Title, Hint and Footer are plain texts, dropped or cut when the width is short.
- `BorderColor` and `LabelColor`: `""`/`dim` (default), `default`, `#RRGGBB`, a palette index, or a `ThemeColors` yaml name such as `border_fg`, resolved against the theme from `ApplyTheme`; unknown names fall back to dim.
- `SendButton` / `SendLabel` / `SetSendButton`: off by default; the label comes from `text_entry.send_label` (`" ➔ "`). It sits at the right of the last content row, takes its columns from the text on every row, and a left click calls `OnSubmit` like a submit key.
- Rows: `Height`, `MinHeight`, `MaxHeight` (spec defaults 1, 1, automatic). The automatic maximum is `max_height_percent` (30) of the canvas height at the latest `Draw`, or of the terminal height before one, or of `fallback_terminal_rows` (24) without a terminal. A negative `MaxHeight` is unbounded. `MinHeight` wins over `MaxHeight`. Breaking: `MaxHeight` 0 was unbounded and is now automatic (Upgrading row).
- `Measure` includes the frame; `Cursor()` and mouse coordinates account for it; a press on the border is ignored. A frame that does not fit (needs 3 rows and a few columns) is skipped and the content gets the whole rect.
- Values live in `spec/defaults.yaml` (`text_entry.*`), not in Go.

## 3. Implementation & Verification Plan
- `text_entry_frame.go`, `text_entry_frame_test.go`, gallery keys Alt+F (frame), Alt+B (button), Alt+C (border color), Alt+G / Alt+M / Alt+X (height, min, max).
- Docs: `docs/Widgets.md` TextEntry section; `docs/Upgrading.md` row; API snapshot refreshed.
- Unverified: the look in a real terminal (owner test in foot).
