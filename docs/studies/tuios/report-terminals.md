# Topic 6: modern terminal best practices (terminal specs)

Scope: terminal escape-sequence specs and per-terminal docs (kitty, foot, Ghostty, Contour specs,
xterm). Sources are read-only. Rows are in `features-terminals.yaml` (20 rows).

## Summary

- Terminal specs we could fetch: foot's ctlseqs (full page), Contour's synchronized-output and
  grapheme specs, Ghostty's features and VT reference pages, and the kitty keyboard, graphics and
  clipboard pages (all three truncated or summarised).
- Not reachable: the xterm ctlseqs page (HTTP 403) and the Contour 2031 colour-notification page
  (HTTP 404). Mode 2031 unsolicited report format is therefore open.
- Web search returned mostly git mirrors and unrelated pages (one query returned only Beckhoff
  hardware pages). Treat search snippets as leads, not evidence.

## Findings that matter for loom (P1 first)

1. **Mode 2026 (synchronized output)** is a frame batch: `CSI ? 2026 h` begins, `CSI ? 2026 l` ends
   (Contour spec, verified). Support is detected with `CSI ? 2026 $ p`. Reply values: 0 and 4 mean
   unsupported, 1 and 2 mean supported (spec text, verified). If no reply comes, that also means
   unsupported, so the probe must time out. The spec says there is no consensus on how long an
   unterminated batch may wait. Loom should gate it on a DECRQM reply and keep an ESU timeout as a
   safety net. Canary: send BSU, a frame, ESU, in a PTY on foot and kitty, and check the output
   stays whole.

2. **Mode 2027 (grapheme clusters)** is the knob for emoji and ZWJ width. The Contour spec says
   terminals MUST cluster per UTS-29 when enabled, and that extending a cluster does not move the
   cursor except for VS16 (verified). It does not define `wcwidth`. Foot lists 2027 as "Grapheme
   cluster processing" (verified). Loom's `measure.DetectZWJMode` should be checked against this:
   it has to query 2027 with DECRQM, not infer from the terminal name. Our measure probes are the
   baseline and the difference is that 2027 is an explicit mode, not a guess.

3. **Kitty keyboard flags.** Loom sends `CSI > 1 u`. The spec adds bits 2 (event types), 4
   (alternate keys), 8 (all keys as escape codes) and 16 (associated text), plus push/pop/query
   forms (fetcher summary of kitty docs; not checked verbatim). Under disambiguate, Ctrl+C no longer
   raises SIGINT, while Enter, Tab and Backspace keep legacy bytes (summary). Loom's Ctrl+C copy
   binding and its paste/quit handling need a test under the mode. Ctrl+Z is not stated in the text
   we read.

4. **Mode 2048 (in-band resize)** sends the size inside the byte stream instead of SIGWINCH.
   Foot lists it with origin "TODO", so there is no settled spec yet (verified). Relevant to loom's
   SIGWINCH robustness (P1). Do not build on it until a spec is settled. Track it, do not adopt it.

5. **Detect the terminal, not the environment.** XTVERSION `CSI > 0 q` replies
   `DCS > | name(version) ST` (foot doc, verified). DA1 `CSI c` and DA2 `CSI > c` are the
   standard feature probes (verified). Loom should use XTVERSION plus DECRQM to pick features,
   and fall back to a conservative path on no reply. The "barrier" use of DA1 is unverified.

6. **Clipboard.** OSC 52 write with `c`, `s` or `p` targets is in the foot doc (verified). Read
   is `OSC 52 ; Pc ; ? ST` (verified). Kitty's OSC 5522 asks the user for permission on read and
   returns EPERM on denial (summary, unverified verbatim). WezTerm reportedly ignores reads (search
   snippet only). Loom: keep OSC 52 write as is; read is P2 to P3 and needs a permission policy.

7. **Light/dark.** Query with `CSI ? 996 n`, reply `CSI ? 997 ; 1|2 n` where 1 is dark and 2 is
   light (foot doc, verified). Subscribe with mode 2031 ("Request color theme updates", verified).
   The unsolicited report format is not fetched. Loom theme detection (P2) should prefer 996 plus
   2031 and fall back to OSC 11 (background colour query) where 2031 is absent. OSC 11 query form is
   not in the foot doc, so the fallback is unverified.

8. **Mouse.** Mode 1016 (SGR-Pixels) reports pixel coordinates (foot doc, verified). Loom is
   cell-based. Pixel mode is P2 and needs a cell-size source, which is unverified here.

9. **Images.** Kitty graphics uses APC `ESC _G ... ESC \` with media t=d, f, t, s and Unicode
   placeholder U+10EEEE (kitty page summary). Ghostty supports it (verified, features page). P3.

10. **Shell marks.** OSC 133 A/C/D (FinalTerm origin, foot doc, verified). Used for prompt jumps.
    OSC 8 links, OSC 9/777 notify and OSC 22 pointer shape are all in foot's doc (verified). All P3.

## Per-terminal differences (only where sourced)

- **foot**: the most complete doc we read. Has 2026, 2027, 2031, 2048 (TODO origin), 1016, XTVERSION,
  DA1/DA2, OSC 52 p/s, OSC 8/9/22/133/777.
- **Ghostty**: kitty graphics supported (verified). Synchronized output, kitty keyboard, 2027,
  2031 and OSC 133 are reportedly supported by search snippets and terminfo-dev, but its
  features page does not list them by name. Verify against Ghostty source before relying on them.
- **Contour**: hosts the 2026 spec in vt-extensions and the 2027 spec in terminal-unicode-core
  (verified). Foot tags 2027 and 2031 with origin "contour" (verified); it tags 2026 with origin
  "terminal-wg", so the 2026 origin is not one author.
- **WezTerm**: OSC 52 read ignored (search snippet only, unverified).
- **kitty**: the reference for keyboard flags and OSC 5522. Its pages were truncated, so the
  quotes need a direct check.

## Gaps and what was not checked

- xterm ctlseqs: not reached (403). Items (a)-(f) of the xterm check are open.
- Contour 2031 spec: not reached (404). Unsolicited 2031 report format is open.
- Mode 2048 report format: not fetched.
- Kitty keyboard and graphics pages: truncated at about 40k characters; verbatim quotes are
  fetcher summaries, marked `verified: false` in the YAML.
- Terminal support matrix: no primary matrix was fetched. Per-terminal claims above are sourced
  only where marked verified.
- Crash restore (what a terminal resets on crash) and tmux pass-through: not researched.
- VTE (GNOME): no VTE primary source was fetched. Gnome Terminal is not covered.
- Bracketed paste (2004) and focus reporting (1004): not verified in the fetched pages.

## Proposed next steps (no tickets filed)

1. Canary for 2026 gating: one PTY test per terminal (foot, kitty, Ghostty) that sends BSU, a
   frame, ESU and checks the bytes. Gate on DECRQM. Owner: host decides.
2. Check `measure.DetectZWJMode` against the 2027 DECRQM query. Read-only check in loom.
3. Fetch the kitty keyboard section on "Disambiguate" and "Report event types" in two pieces to
   verify the quotes marked `verified: false`.
4. Fetch Contour's color-palette notification file (URL unknown; search for it by name) to close
   the 2031 gap.

Rows written: 20. Rows verified: 16. Rows with `verified: false`: 4 (kitty keyboard flags, kitty
keyboard exceptions, kitty OSC 5522 read permission, kitty graphics APC). Open gaps: xterm ctlseqs
(403), Contour 2031 spec (404), 2031 unsolicited report format, mode 2048 report format, Ghostty
support for 2026/2027/OSC 133 (not on its features page).

Sources: https://codeberg.org/dnkl/foot/raw/branch/master/doc/foot-ctlseqs.7.scd ;
https://github.com/contour-terminal/vt-extensions/blob/master/synchronized-output.md ;
https://github.com/contour-terminal/terminal-unicode-core/blob/master/spec/terminal-unicode-core.tex ;
https://ghostty.org/docs/features ; https://ghostty.org/docs/vt/reference ;
https://sw.kovidgoyal.net/kitty/keyboard-protocol/ ; https://sw.kovidgoyal.net/kitty/graphics-protocol/ ;
https://sw.kovidgoyal.net/kitty/clipboard/
