# Ideas for VT Parser Improvements

## 1. Fix parsing of SGR truecolor when color space ID is included
**Location:** `vt/vt.go:858`
**Change:** When parsing SGR colon subparameters for RGB (e.g., `38:2:id:r:g:b`), recognize that a color space identifier parameter is optional but present when there are 4 components. If `len(color) >= 4`, parse `r, g, b` from indices 1, 2, 3 instead of 0, 1, 2. If it's `38:2::r:g:b` (`len(color) == 3`), use 0, 1, 2.
**Why it helps:** The ITU T.416 standard for 24-bit color specifies `38:2:[colorSpaceId]:r:g:b`. The current logic blindly takes the first 3 numbers from the split parameters. Consequently, `38:2:1:10:20:30` extracts `1, 10, 20` as the RGB components (the ID becomes the red channel), incorrectly parsing standard valid truecolor sequences.
**Smallest step to do it:** Check `len(color)` in the `mode == 2` block: if `len(color) >= 4`, set SGR color using `color[1]`, `color[2]`, `color[3]`. If `len(color) >= 3`, use `color[0]`, `color[1]`, `color[2]`.

## 2. Default omitted parameters in SGR truecolor to zero
**Location:** `vt/vt.go:862`
**Change:** In the loop parsing colon-separated SGR subparameters `for _, part := range parts[2:]`, if `part == ""` (omitted parameter), treat its value as `0` and append it to the `color` slice rather than using `continue` to skip it entirely.
**Why it helps:** Omitting a subparameter in SGR colon notation (e.g., `38:2:::20:30` meaning Red=0, Green=20, Blue=30) should default the parameter to zero according to the specification. Currently, skipping empty strings reduces the `color` slice length, which then fails the `len(color) >= 3` check and discards the color code completely.
**Smallest step to do it:** Remove `if part == "" { continue }` and replace the integer conversion logic with `value, err := strconv.Atoi(part); if err != nil && part != "" { continue }; color = append(color, value)`.

## 3. Handle truncated semicolon SGR sequences instead of ignoring them
**Location:** `vt/vt.go:917`
**Change:** For `38;2;r;g;b` parameter parsing, if `i+4 < len(p)` fails, check if the slice has fewer parameters and default the missing values to `0` rather than skipping the entire sequence.
**Why it helps:** Some terminal output might clip or incorrectly formulate SGR color codes (e.g., `38;2;100;200` omitting the blue channel). Instead of dropping the foreground color altogether (leaving it unstyled), it is robust to parse what is available and pad the rest with `0` (as missing parameters default to `0`), matching xterm fallback behavior.
**Smallest step to do it:** In the `case 2:` for semicolon parsing, don't check `i+4 < len(p)`. Instead, read up to three subsequent elements in `p`, using `0` if out of bounds, and then increment `i` by the number of parameters actually consumed.

## 4. Retain and properly decode OSC escape sequences missing terminating characters
**Location:** `vt/vt.go:506`
**Change:** When processing an escape sequence character that begins an OSC string (e.g., `]`), ensure `discardControl` is set appropriately to capture the string if a subsequent `ESC` terminates it, rather than blindly dropping data if `maxPendingBytes` is reached.
**Why it helps:** The current implementation of `WriteFramed` successfully retains `\x1b]...` without the terminating character due to the `discardControl` mechanism. However, if the text inside the OSC hits `maxPendingBytes`, the `ok` logic in `escape()` drops the state early, corrupting large OSC buffers (e.g., long titles or clipboard sequences) before the terminating character can arrive.
**Smallest step to do it:** In the `escape()` switch cases for OSC `]`, don't return `0, false` unconditionally on missing sequence termination. Maintain a dedicated OSC buffering state or increase the pending buffer limit to allow complete extraction of large terminal sequences.

## 5. Properly erase full-width characters that span boundaries
**Location:** `vt/vt.go:618`
**Change:** When writing a character that overwrites part of a multi-cell (wide) character, the adjacent cell(s) belonging to the wide character should be cleared (replaced with a single-width space) to avoid rendering artifacts.
**Why it helps:** Currently, moving the cursor left and overwriting the first half of a wide character (e.g., `中`) leaves the second half intact (which stores `Rune: 0` but points to the same wide grapheme in some rendering engines). This causes display corruption and invalid cell states where a continuation cell exists without its base character.
**Smallest step to do it:** Before assigning `v.cells[v.y][v.x] = Cell{...}` in `putCluster`, check if the existing cell is a continuation cell (e.g., `Rune == 0 && Text != ""`) and clear its base cell (at `x-1`), or if it's the base cell of a wide character, clear its continuation cell (at `x+1`).
