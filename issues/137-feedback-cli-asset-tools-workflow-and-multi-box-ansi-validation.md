# 137 — Feedback: CLI asset tools workflow & multi-box ANSI validation

**Status**: Open
**Priority**: P3 (Low)
**Severity**: Minor
**Category**: Feedback
**Related**: `cmd/loom` (`check-box`, `eval`, `format`, `measure`), `ValidateAnsiBox` (issue 128, 130)

---

## 1. Context & Feedback Summary

During development of a split-pane ANSI mockup and TUI prototype in `cati` (`examples/mediabrowse/mockup/mediabrowse.ansi`), the new Loom CLI tools (`loom check-box`, `loom eval`, `loom format`, `loom measure`) were used to validate and fix a 2-box horizontal split ANSI mockup.

### What Worked Well
1. **`loom eval` & `loom measure`**: Immediately provided accurate terminal column metrics and detected ragged row line widths across truecolor/styled ANSI rows.
2. **`loom check-box`**: Accurately flagged misalignment in Unicode box borders, identifying exactly which line numbers had boundaries diverging from the reference top row.
3. **Rich Widget & Media ecosystem**: Loom's `Frame`, `Box`, `Choice`, and `media.Widget` allowed quick construction of a rich split-pane media browser CLI with live video streaming and keyboard/mouse focus handling.

---

## 2. How the Mockup Was Fixed

### The Initial Issues
- When generating a split-pane layout with two horizontal boxes (Left: `Files`, Right: `Media Preview`), the top frame line was shorter than the inner rows (104 columns vs 106 columns).
- `loom check-box` reported:
  ```text
  box validation failed:
  examples/mediabrowse/mockup/mediabrowse.ansi: boxed line 2: boundaries 0..105 width 106, want 0..103 width 104 from line 1
  ```
- The inner row calculation accounted for:
  - Left box width: 44 cols (border + 42 content + border)
  - Inter-box gap: 2 spaces
  - Right box width: 60 cols (border + 58 content + border)
  - Total row width: 44 + 2 + 60 = 106 cols.
- The top header row had hardcoded dash counts (`"┌─" + title + dashes + "┐"`) that under-calculated the required dashes relative to the title's visual width (`StringWidth(title)`).

### The Fix
1. Computed dynamic dash padding for top borders:
   - `leftTopDashes = leftBoxWidth - 2 (borders) - 1 (prefix dash) - StringWidth(leftTitle)`
   - `rightTopDashes = rightBoxWidth - 2 (borders) - 1 (prefix dash) - StringWidth(rightTitle)`
2. Aligned all inner content lines using `loom.StringWidth()` to exact box inner widths (42 for left, 58 for right).
3. Padded the bottom status bar to the exact total width (106 cols).
4. Verified with `loom check-box` (`ok`) and `loom eval` (0 ragged rows, 106 columns x 13 rows).

---

## 3. Suggestions & Feature Requests for Loom CLI

1. **Multi-Box Row Detection in `loom format --align-box`**:
   - `loom format --align-box` currently aligns single box outer widths, but multi-box split layouts (e.g. `│...│  │...│`) could benefit from recognizing side-by-side box columns and re-padding inner gaps or boxes individually.
2. **Detailed Diff / Diagnostic on `loom check-box`**:
   - When a border fails, printing an inline visual pointer (e.g., showing where the expected `│` or `┐` was expected vs where it was encountered in terminal cells) would make manual fixes even faster.
3. **Interactive Preview in `loom view`**:
   - `loom view` displays `.ansi` files interactively. Supporting a split or ruler inspection overlay in `loom view` to inspect per-column coordinate markers would be a valuable design aid.
