# Loomoji terminal width measurements

Each terminal profile has a JSON store and a matching tab-separated text report.
Filenames use URL-escaped `$TERM` and `$TERM_PROGRAM` values, with `unknown` for
missing environment variables. The profile is part of the data because terminal
emulators may render the same glyph at different widths.

The JSON `measurements` object is keyed by the complete glyph string, so variation
selectors, joined emoji, and regional-indicator flags remain one record. Each record
stores the glyph, its Unicode codepoints, Loom's computed width, measured width, and
an optional comment. `measured_width` is `1` or `2` for an observation and `0` for
other/unsure. `answered: true` distinguishes an explicit unsure result from a saved
comment that still needs a width answer. Missing keys have not been measured for that
terminal profile.

Updates are saved incrementally. Existing glyph records are retained when new
records are merged; the text report is regenerated in glyph order from the JSON data.
