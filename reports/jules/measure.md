## 1. Deduplicate grapheme cluster bound logic into a shared iterator

**Path:** `measure/measure.go:117`
**Context:** `StringWidthNew` implements an inline, hand-rolled grapheme cluster boundary scanner (handling ZWJ, base + modifiers, variation selectors) directly in its width calculation loop.
**Change:** Extract the cluster advancing loop into a standalone `nextClusterBoundary(text string, start int) int` function and call it from `StringWidthNew`, `ClustersNew`, and anywhere else iterating clusters.
**Why it helps:** Keeps cluster definition identical between width measurement and cluster extraction. It simplifies `StringWidthNew` and eliminates the risk of `ClustersNew` returning a different boundary than `StringWidthNew` measured.
**Smallest step:** Move the `i += sz ... for i < n` modifier/ZWJ loop into a private `nextClusterBoundary` function, replacing the loop in `StringWidthNew` with a call to it.

## 2. Resume slow path from first non-ASCII byte in `StringWidthNew`

**Path:** `measure/measure.go:49`
**Context:** `StringWidthNew` uses a fast path that scans the entire string to see if it is pure ASCII printable (`b < 0x20 || b > 0x7e`). If it finds even one non-ASCII or control byte, it drops the scan result and starts the slow path from `i = 0`.
**Change:** Instead of a separate full-string scan loop that returns early, integrate the ASCII fast-path into the main loop, or keep the index `i` where the scan stopped and start the slow path from there.
**Why it helps:** Strings with a mix of ASCII and Unicode or ANSI (e.g., "Hello 🌍") currently scan the "Hello " prefix twice. Reusing the scan index avoids this redundant work.
**Smallest step:** Modify the `isPureASCII` loop to break and retain the index `i`. If `i < len(text)`, start the slow path from `i` and set the initial width accumulator to `w = i`.

## 3. Short-circuit ANSI parsing in `StringWidthNew` for common reset sequences

**Path:** `measure/measure.go:76`
**Context:** In the slow path, when `StringWidthNew` encounters an ESC byte (`27`), it drops into a `switch` and a `for` loop to skip the ANSI sequence byte-by-byte until the terminator.
**Change:** Add a fast-path for the most common SGR reset sequences (`\x1b[0m` and `\x1b[m`).
**Why it helps:** In styled TUI text, reset sequences are extremely common. Skipping them with a fast string prefix check (`text[i:i+3] == "[0m"`) skips the byte-by-byte switch/loop overhead entirely.
**Smallest step:** Immediately after checking `text[i] == 27`, add `if strings.HasPrefix(text[i:], "\x1b[0m") { i += 4; continue }` (and similarly for `\x1b[m`).

## 4. Skip utf8 decoding for ASCII in the `plainTerminalTextNew` slow path

**Path:** `measure/measure.go:489`
**Context:** In the fallback loop of `plainTerminalTextNew`, after checking for ANSI escapes, every remaining character is decoded using `utf8.DecodeRuneInString` and then checked with `unicode.IsControl` and `unicode.Is(unicode.Cf)`.
**Change:** Add a quick check `if s[i] < utf8.RuneSelf` inside the loop before calling `utf8.DecodeRuneInString`.
**Why it helps:** Even in a string containing some Unicode or ANSI, the vast majority of characters are typically plain ASCII. Bypassing `utf8.DecodeRuneInString` and `unicode` package checks for these characters saves CPU cycles.
**Smallest step:** Before `r, sz := utf8.DecodeRuneInString(s[i:])`, insert `if s[i] < utf8.RuneSelf { if s[i] >= 0x20 && s[i] != 0x7F { b.WriteByte(s[i]) }; i++; continue }`.

## 5. Avoid allocations in `plainTerminalTextNew` when text is clean

**Path:** `measure/measure.go:437`
**Context:** `plainTerminalTextNew` has a fast path for pure ASCII, but if it finds *any* Unicode, it does a second scan checking for ANSI and controls. If it determines the string has no ANSI but does have Unicode, it does a full pass checking `unicode.IsControl`. If that passes, it returns `s`. However, the initial control check `if b == 27 || b < 0x20 || b == 0x7f` misses C1 controls and unicode format characters, meaning strings containing them will drop to the slow path and allocate a `strings.Builder`.
**Change:** Make the initial scan robust enough to detect *all* characters that would be stripped, and if none are found, return `s` without allocating a `strings.Builder`.
**Why it helps:** Pure text with characters like `\n` or some unicode symbols might not contain any format or control characters that need stripping. Avoiding the `strings.Builder` allocation entirely reduces GC pressure for clean Unicode strings.
**Smallest step:** Combine the two `hasEscOrControl` and `clean` loops into a single pass that identifies if *any* stripping is required, and if not, `return s`.
