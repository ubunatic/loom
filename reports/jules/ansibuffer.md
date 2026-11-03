## 1. Support basic cursor movement CSI commands (A, B, D) in `ParseAnsiBuffer`.

- **Citation**: `ansibuffer.go:645`
- **Change/Why**: The parser handles `C` (Cursor Forward) and absolute positions like `H`, `f`, `G`, and `d`, but ignores `A` (Up), `B` (Down), and `D` (Left). Missing these relative cursor movements means it cannot correctly parse ANSI files containing cursor repositioning.
- **Smallest step**: Add cases for `'A'`, `'B'`, and `'D'` to the `switch cmd` block, updating `x` and `y` using `parseAnsiCSINum(params, 1)` just like `C`.

## 2. Parse SGR truecolor sequences using colons (38:2::r:g:b) as well as semicolons.

- **Citation**: `ansibuffer.go:838`
- **Change/Why**: The `parseAnsiSGR` function splits parameters purely by `;`. It expects `38;2;r;g;b` but standard xterm truecolor sequences often use colons `38:2:[id]:r:g:b`. The parser fails to apply truecolor styles when colons are used.
- **Smallest step**: In `parseAnsiSGR`, normalize colon separators to semicolons before splitting, or handle `strings.Split(params, ":")` as a fallback when `38:2` or `48:2` is detected.

## 3. Properly represent CJK/wide characters without inserting an extraneous trailing space.

- **Citation**: `ansibuffer.go:724`
- **Change/Why**: When measuring a rune with `w > 1` (e.g., CJK characters), `ParseAnsiBuffer` skips `x` by `w` without writing a continuation marker (like `0`) to the right cell. During `Serialize`, the loop evaluates this skipped right cell as a standard blank and outputs a literal space, inserting spurious spaces into the text output.
- **Smallest step**: If `w > 1`, set the continuation cell at `buf.cells[y][x+1]` with a special marker (like `Rune: 0` as used in `vt`) and modify `Serialize` to skip rendering when `Rune == 0`.

## 4. Recognize proper whitespace in `NextWord` instead of just space and NUL.

- **Citation**: `ansibuffer.go:321`
- **Change/Why**: `NextWord` and `PrevWord` hardcode word boundaries to `Rune == ' '` and `Rune == 0`. It does not break on tabs or punctuation, leading to incorrect text selection behavior in editors when jumping words.
- **Smallest step**: Replace the manual space checks with `unicode.IsSpace` (and optionally `unicode.IsPunct`) from the standard `unicode` package.

## 5. Include CSI `A`, `B`, and `C` in `ParseAnsiBuffer` dimension prescan.

- **Citation**: `ansibuffer.go:579`
- **Change/Why**: The `ParseAnsiBuffer` prescan function computes `maxRows` and `maxCols` before allocation by examining absolute positioning commands (`H`, `f`, `d`, `G`, `r`). However, large relative cursor movements (`B` down, `C` right) could push the required buffer size past `maxCols` or `maxRows` and are currently ignored, causing out-of-bounds truncations.
- **Smallest step**: Track simulated `x` and `y` during the prescan for `A`, `B`, `C`, and `D` commands, updating `maxCols` and `maxRows` dynamically.
