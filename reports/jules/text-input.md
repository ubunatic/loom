1. **Prevent out-of-bounds crash on word boundary navigation in `TextArea`**
   - **Path**: `textarea.go:373`
   - **Change**: In `moveWordLeft`, it scans backwards inside the current line using `t.lines[t.row]`. If `t.col` reaches `0`, it does `t.moveLeft()` but only once, and if the previous line is empty or we haven't hit a word character yet, it stops. Repeatedly pressing `ctrl-left` fails to scan across multiple empty or non-word lines. `moveWordLeft` and `moveWordRight` need to wrap around correctly.
   - **Why it helps**: Fixes an unhandled edge case when navigating by words near empty lines or at the start/end of the document.
   - **Smallest step**: Update the `for` loops in `moveWordLeft` and `moveWordRight` to use a helper that safely reads the rune before/after the caret across line boundaries, wrapping to the next line when necessary.

2. **Sanitize control characters on `ConsumePaste` in `TextArea`**
   - **Path**: `textarea.go:81`
   - **Change**: `TextArea.ConsumePaste` directly takes `event.Text` and uses it to do `strings.Split(event.Text, "\n")`. It should filter out non-printable ASCII control characters (like `\x00` - `\x1f`, except `\n` and `\t`) which can crash or corrupt the terminal renderer when drawn.
   - **Why it helps**: Prevents the application from crashing or rendering broken terminal escape sequences when a user accidentally pastes binary data or garbage text containing terminal control codes.
   - **Smallest step**: Pre-process `event.Text` in `ConsumePaste` using `strings.Map` to remove or replace control characters before inserting them.

3. **Ensure combining characters are kept together during `Backspace` in `TextInput`**
   - **Path**: `textinput.go:108`
   - **Change**: `ConsumeKey` for `backspace` simply removes the previous `rune` (`t.runes[t.caret-1]`). If the text contains a base character followed by combining characters (e.g., `é`), it will delete the combining mark, leaving a broken base character.
   - **Why it helps**: Makes the text editor robust against modern Unicode input, behaving like standard text editors that treat a grapheme cluster as a single unit during deletion.
   - **Smallest step**: Use a grapheme cluster boundary detector (or step back until a non-combining rune is found) to remove the entire grapheme cluster during backspace instead of a single `rune`.

4. **Fix clipboard slice alias bug in `TextArea.ConsumeKey` on copy**
   - **Path**: `textarea.go:240`
   - **Change**: In `ConsumeKey`'s `copy` handling, it converts `t.Value()` into a `[]rune` and slices it. `t.Value()` is an expensive allocation of the entire document. It then converts it to `[]rune` which is another massive allocation for large documents, just to copy a small selection.
   - **Why it helps**: Improves performance and prevents out-of-memory crashes on large text documents when copying a small selection.
   - **Smallest step**: Implement a targeted string builder loop over `t.lines` that extracts the selected text directly from the line buffers without calling the expensive `t.Value()`.

5. **Fix wide-rune clipping visual glitch in `TextInput.Draw`**
   - **Path**: `textinput.go:146`
   - **Change**: When scrolling past wide runes, the window logic in `Draw` measures width but tracks `t.view` and `t.caret` simply by rune index instead of using display width correctly across the visible window, leading to cursor rendering drift and clipping errors.
   - **Why it helps**: Solves the UI bug where editing past the window width corrupts the render and misplaces the cursor when wide characters (like `界` or emoji) are present.
   - **Smallest step**: Modify `Draw` to use `StringWidth` for calculating `available` space accurately taking into account whether a character at the boundary is a double-width character.
