# Gallery manual test runbook

Run `loom widgets --show` in a terminal. Use Tab and Shift-Tab to change gallery tabs; click a tab when visible. F9 changes theme and F10 or Ctrl+Q exits. A single sample can be opened with `loom widgets --show <Name>`. Resize the terminal as needed. Record observations below and return this file to the host; use `ok`, `bug: …`, or `idea: …`.

Each step should produce the visible result described. If a control does not respond or the display clips unexpectedly, record the terminal size and input in feedback.

## All

1. Open this tab.
2. Move among the grid controls with arrows or clicks; activate a toggle and select a table cell.

Expected: The grid shows several controls; the active control and selected table cell move, and the toggle changes state.

Feedback: `ok` / `bug: …` / `idea: …`

## AlignBox

1. Open this tab.
2. Resize the terminal and observe the centered content.

Expected: “Centered content” stays centered in the 24-by-3 sample as the terminal is resized.

Feedback: `ok` / `bug: …` / `idea: …`

## AnsiEditor

1. Open this tab.
2. Move with arrow keys, type text, and click a cell to move the cursor.

Expected: The cursor moves to the requested cell and typed characters appear on the canvas.

Feedback: `ok` / `bug: …` / `idea: …`

## Box

1. Open this tab.
2. Inspect the border and the differently styled title segments.

Expected: A border surrounds the sample; the menu hotkey, Load label and host metadata have distinct styles.

Feedback: `ok` / `bug: …` / `idea: …`

## Chart

1. Open this tab.
2. Inspect the labeled request and error series.

Expected: The Requests and Errors series appear with their labels and different plotted values.

Feedback: `ok` / `bug: …` / `idea: …`

## Choice

1. Open this tab.
2. Type a filter, move with arrows, and select an item with Enter or a click.

Expected: The initial fbw filter shows matching choices; editing the filter changes the list and selection follows arrows or a click.

Feedback: `ok` / `bug: …` / `idea: …`

## ColorPicker

1. Open this tab.
2. Use arrows and clicks to choose foreground and background colors; inspect the preview.

Expected: The foreground and background selections move, and the preview reflects the chosen colors.

Feedback: `ok` / `bug: …` / `idea: …`

## Confirm

1. Open this tab.
2. Press Left or Right to move the selection, then press y or n to answer.

Expected: The selected answer moves between Yes and No; y or n chooses the corresponding answer.

Feedback: `ok` / `bug: …` / `idea: …`

## DatePicker

1. Open this tab.
2. Move between dates with arrows and select a date.

Expected: The highlighted day moves through the calendar and selecting a day marks it as selected.

Feedback: `ok` / `bug: …` / `idea: …`

## Dialog

1. Open this tab.
2. Move between buttons, click or press Enter, then press Enter to reopen.

Expected: The Save changes dialog highlights the selected button, closes on activation, and reopens with Enter.

Feedback: `ok` / `bug: …` / `idea: …`

## FilePicker

1. Open this tab.
2. Navigate a directory with arrows and enter a directory or select a file.

Expected: The directory listing follows the chosen directory; a selected matching file is highlighted.

Feedback: `ok` / `bug: …` / `idea: …`

## Form

1. Open this tab.
2. Edit both fields and move focus with Tab.

Expected: The Name and Role values change as typed, and Tab moves focus between the fields.

Feedback: `ok` / `bug: …` / `idea: …`

## Frame

1. Open this tab.
2. Inspect the title, status, and two bordered panes while resizing.

Expected: The Frame title and status surround First and Second bordered panes; the panes adjust to the available width.

Feedback: `ok` / `bug: …` / `idea: …`

## Gauge

1. Open this tab.
2. Inspect the gauge, stacked bar and labeled treemap.

Expected: A 65% gauge, a stacked bar and a treemap labeled API, Worker and Cache are visible.

Feedback: `ok` / `bug: …` / `idea: …`

## HintBar

1. Open this tab.
2. Click a key cap or press its indicated key and observe the result.

Expected: Key caps and their labels are visible; activating a cap or its key triggers its displayed action.

Feedback: `ok` / `bug: …` / `idea: …`

## KeyHelp

1. Open this tab.
2. Inspect the key labels and resize to check clipping.

Expected: Back, Next and Open appear with their corresponding keys; the display clips cleanly in a narrow terminal.

Feedback: `ok` / `bug: …` / `idea: …`

## Media

1. Open this tab.
2. Inspect the generated color gradient and resize the terminal.

Expected: A colored gradient appears and adjusts to the available terminal area when resized.

Feedback: `ok` / `bug: …` / `idea: …`

## MenuBar

1. Open this tab.
2. Open a menu, move through items and submenus, and toggle Autosave.

Expected: The menu opens with a moving highlight, Recent files reveals a submenu, and Autosave changes its check mark.

Feedback: `ok` / `bug: …` / `idea: …`

## Notif

1. Open this tab.
2. Inspect the saved message at the bottom of the panel.

Expected: “Changes saved” appears at the bottom of the sample panel.

Feedback: `ok` / `bug: …` / `idea: …`

## NumberInput

1. Open this tab.
2. Change the value with arrows and type a new number.

Expected: The initial 7.50 changes in 0.50 steps with arrows and reflects the number entered.

Feedback: `ok` / `bug: …` / `idea: …`

## PaintCanvas

1. Open this tab.
2. Drag the mouse to paint; use the displayed controls to clear.

Expected: Dragging adds marks to the canvas; the clear control removes them.

Feedback: `ok` / `bug: …` / `idea: …`

## Paginator

1. Open this tab.
2. Use arrows and click a page; the selected page should move.

Expected: The highlighted page changes when using arrows or clicking another page.

Feedback: `ok` / `bug: …` / `idea: …`

## PillCluster

1. Open this tab.
2. Move selection through the status pills and inspect their labels.

Expected: API, Worker and Cache pills display their status symbols; the selection moves among them.

Feedback: `ok` / `bug: …` / `idea: …`

## Popup

1. Open this tab.
2. Press Esc to close the popup and Enter to reopen it.

Expected: The Gallery popup disappears on Esc and is displayed again on Enter.

Feedback: `ok` / `bug: …` / `idea: …`

## ProgressBar

1. Open this tab.
2. Watch the bar fill, dim, disappear, and restart.

Expected: The progress bar advances, dims, disappears and then starts its cycle again.

Feedback: `ok` / `bug: …` / `idea: …`

## RichTextEdit

1. Open this tab.
2. Move the caret, type, select text, and scroll the document.

Expected: The caret moves, typed text appears, selected text is highlighted and scrolling reveals other document lines.

Feedback: `ok` / `bug: …` / `idea: …`

## Router

1. Open this tab.
2. Press Enter to switch views; press Esc to return through history.

Expected: Enter switches between “Router main view” and “Router details view”; Esc returns to the preceding view.

Feedback: `ok` / `bug: …` / `idea: …`

## Rows

1. Open this tab.
2. Inspect left and right aligned columns; resize to see clipping.

Expected: Requests and Errors appear in the left column with 1200 and 4 right aligned; narrow widths clip the columns.

Feedback: `ok` / `bug: …` / `idea: …`

## SearchBar

1. Open this tab.
2. Type and edit a query; inspect the placeholder when empty.

Expected: Typed query text appears and can be edited; clearing it reveals “type to filter”.

Feedback: `ok` / `bug: …` / `idea: …`

## Settings

1. Open this tab.
2. Use arrows to choose a setting and Enter or arrows to edit each type.

Expected: The highlight moves among Enabled, Name, Theme and Retries; editing changes the selected value.

Feedback: `ok` / `bug: …` / `idea: …`

## Spinner

1. Open this tab.
2. Watch the activity indicator animate.

Expected: The activity indicator animates beside “Syncing workspace”.

Feedback: `ok` / `bug: …` / `idea: …`

## SplashView

1. Open this tab.
2. Inspect the centered title, progress bar, status and pill.

Expected: “Starting Loom”, a 65% progress bar, “Loading workspace” and the API status pill are visible.

Feedback: `ok` / `bug: …` / `idea: …`

## Split

1. Open this tab.
2. Resize and drag the divider to adjust the two panes.

Expected: Left pane and Right pane remain on opposite sides of a divider; dragging it changes their widths.

Feedback: `ok` / `bug: …` / `idea: …`

## Stack

1. Open this tab.
2. Inspect the two vertically arranged panes when resizing.

Expected: Top pane appears above Bottom pane; their layout adjusts when the terminal is resized.

Feedback: `ok` / `bug: …` / `idea: …`

## Stopwatch

1. Open this tab.
2. Use the displayed controls to stop and restart the elapsed time.

Expected: The elapsed time advances, stops when paused and advances again after restart.

Feedback: `ok` / `bug: …` / `idea: …`

## StyledRows

1. Open this tab.
2. Inspect the plain, red, and bold text.

Expected: Plain text, red text and bold text appear with their respective styles.

Feedback: `ok` / `bug: …` / `idea: …`

## Table

1. Open this tab.
2. Move among cells with arrows and click a cell.

Expected: The highlighted cell moves with arrows and jumps to the clicked cell.

Feedback: `ok` / `bug: …` / `idea: …`

## Tabs

1. Open this tab.
2. Switch between Overview and Details with Tab or a click.

Expected: The active tab changes between Overview and Details, displaying the corresponding content.

Feedback: `ok` / `bug: …` / `idea: …`

## TextArea

1. Open this tab.
2. Inspect Go syntax colors, move the caret and type text.

Expected: Go keywords and strings are syntax colored; the caret moves and typed text appears in the editor.

Feedback: `ok` / `bug: …` / `idea: …`

## TextEntry

1. Open this tab.
2. Press F2 to change and show the placeholder, F3 to accept it, and F4 to move the caret to the start. Type, select with the mouse, and submit with Enter.

Expected: F2 clears the text and shows “Write an update…” as the placeholder; F3 inserts it as text, and F4 moves the caret to offset 0. Typing and mouse selection edit the text; Enter submits and clears it.

Feedback: `ok` / `bug: …` / `idea: …`

## TextInput

1. Open this tab.
2. Edit the name, placeholder and masked fields; use Tab to move.

Expected: The Name and Placeholder fields display edits, the Masked field hides its characters, and Tab moves focus.

Feedback: `ok` / `bug: …` / `idea: …`

## Timer

1. Open this tab.
2. Use the displayed controls to pause and resume the countdown.

Expected: The countdown decreases, holds its value while paused and continues decreasing when resumed.

Feedback: `ok` / `bug: …` / `idea: …`

## Toggle

1. Open this tab.
2. Press Enter or click the toggle and observe its value change.

Expected: The toggle changes between on and off with Enter or a click.

Feedback: `ok` / `bug: …` / `idea: …`

## Tree

1. Open this tab.
2. Expand and collapse folders with arrows or clicks.

Expected: The src and docs folders expand to reveal children and collapse to hide them.

Feedback: `ok` / `bug: …` / `idea: …`

## Viewport

1. Open this tab.
2. Scroll down and up with the wheel or navigation keys.

Expected: The visible lines move down and back up as the viewport scrolls.

Feedback: `ok` / `bug: …` / `idea: …`
