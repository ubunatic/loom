## Item vt-diff (test-draft)

### Verdict
Loom and upstream charmbracelet/x/vt disagree on 8 out of 25 terminal streams, highlighting differences in terminal escape sequence support, grapheme boundaries, UTF-8 robustness, and corner cases like reverse wrap and scroll regions.

### Changes
`experiments/vtdiff/go.mod`
`experiments/vtdiff/vtdiff_test.go`
`reports/jules/vt-diff.md`

### Disagreements

| Stream | Quoted Input | Loom Result | Upstream Result | Class |
|--------|--------------|-------------|-----------------|-------|
| insert/delete chars and lines | `A\x1b[1@B\x1b[1PC\x1b[1LD\x1b[1ME` | screen:`[]string{"A", "D", "", ""}` | screen:`[]string{"EBC", "", "", ""}` | upstream bug: Upstream incorrectly executes character/line insertion and deletion sequences. |
| VS16 | `❤️` | screen:`[]string{"❤️", "", "", ""}` | screen:`[]string{"❤️", "", "", ""}` | upstream bug: Upstream correctly renders the VS16 emoji but incorrectly calculates the cursor width as 2 columns instead of 1. |
| skin tones | `👍🏽` | screen:`[]string{"👍🏽", "", "", ""}` | screen:`[]string{"👍🏽", "", "", ""}` | upstream bug: Upstream renders the skin tone modifier but incorrectly leaves the cursor position at 2 rather than handling the grapheme cluster properly width-wise. |
| DEC line drawing | `\x1b(0lqqk\x1b(B` | screen:`[]string{"┌──┐", "", "", ""}` | screen:`[]string{"lqqk", "", "", ""}` | upstream bug: Upstream does not support `ESC ( 0` to designate the G0 Character Set to DEC Special Character and Line Drawing Set. |
| ZWJ emoji | `👨‍🌾` | screen:`[]string{"👨\u200d🌾", "", "", ""}` | screen:`[]string{"👨\u200d🌾", "", "", ""}` | upstream bug: Upstream incorrectly handles ZWJ sequence grapheme clusters width across boundaries, resulting in cursor position 2 instead of 4 for whole write, and visual artifacts for byte-by-byte write `👨🌾`. |
| incomplete UTF-8 | `a\xc2` | screen:`[]string{"", "", "", ""}` | screen:`[]string{"a", "", "", ""}` | policy difference: Loom holds the preceding character in the parser buffer waiting for the UTF-8 sequence to complete, whereas upstream flushes valid sequences immediately before the incomplete byte. |
| reverse wrap | `abcdefghij\b\b` | cursor:(8,0) screen:`[]string{"abcdefghij", "", "", ""}` | cursor:(7,0) screen:`[]string{"abcdefghij", "", "", ""}` | policy difference: Upstream wraps backwards when backspace hits the left margin for a wrapped line, which is non-standard but often desired in terminal emulators, while Loom strictly stops at the left margin. |
| scroll regions | `\x1b[2;3r1\n2\n3\n4` | cursor:(4,3) screen:`[]string{"1", " 2", "  3", "   4"}` | cursor:(4,2) screen:`[]string{"1", "  3", "   4", ""}` | upstream bug: Upstream scrolling region `\x1b[2;3r` logic improperly scrolls lines outside the specified top and bottom margins. |

### Tests

```bash
cd experiments/vtdiff && go test ./...
```

```text
FAIL	ubunatic.com/loom/experiments/vtdiff	0.547s
FAIL
```

Failing subtests (8 streams):
- TestVTDiff/insert/delete_chars_and_lines
- TestVTDiff/VS16
- TestVTDiff/skin_tones
- TestVTDiff/DEC_line_drawing
- TestVTDiff/ZWJ_emoji
- TestVTDiff/incomplete_UTF-8
- TestVTDiff/reverse_wrap
- TestVTDiff/scroll_regions

Passed tests (17 streams):
- TestVTDiff/plain_text
- TestVTDiff/CR/LF
- TestVTDiff/wrap_at_the_last_column
- TestVTDiff/scroll
- TestVTDiff/SGR_colours
- TestVTDiff/CUP/CUU/CUD/CUF/CUB
- TestVTDiff/erase_in_line_and_display
- TestVTDiff/tabs
- TestVTDiff/wide_CJK
- TestVTDiff/combining_marks
- TestVTDiff/alt_screen
- TestVTDiff/CUP_wrap
- TestVTDiff/erase_display_0
- TestVTDiff/erase_display_1
- TestVTDiff/erase_line_0
- TestVTDiff/erase_line_1
- TestVTDiff/tab_stop

```bash
go test -short ./...
```

```text
ok  	ubunatic.com/loom	1.578s
ok  	ubunatic.com/loom/ansi	0.008s
ok  	ubunatic.com/loom/cmd/loom	1.674s
ok  	ubunatic.com/loom/cmd/loom-bench	0.090s
ok  	ubunatic.com/loom/cmd/loom-demo	0.072s
ok  	ubunatic.com/loom/cmd/loom-probe	0.030s
ok  	ubunatic.com/loom/cmd/loom-repaint-probe	0.026s
?   	ubunatic.com/loom/cmd/validate-spec	[no test files]
ok  	ubunatic.com/loom/collector	0.032s
?   	ubunatic.com/loom/examples/ansicanvas_demo	[no test files]
ok  	ubunatic.com/loom/examples/ansicanvas_demo/ansicanvas_demo	0.042s
?   	ubunatic.com/loom/examples/ansiedit	[no test files]
ok  	ubunatic.com/loom/examples/ansiedit/ansiedit	0.040s
?   	ubunatic.com/loom/examples/ansiviewer	[no test files]
ok  	ubunatic.com/loom/examples/ansiviewer/ansiviewer	0.219s
?   	ubunatic.com/loom/examples/background	[no test files]
ok  	ubunatic.com/loom/examples/background/background	0.048s
?   	ubunatic.com/loom/examples/filebrowser	[no test files]
ok  	ubunatic.com/loom/examples/filebrowser/filebrowser	0.113s
ok  	ubunatic.com/loom/examples/loomoji	0.040s
ok  	ubunatic.com/loom/examples/loomoji/loomoji	0.075s
ok  	ubunatic.com/loom/examples/media	0.033s
?   	ubunatic.com/loom/examples/monitor	[no test files]
ok  	ubunatic.com/loom/examples/monitor/monitor	0.101s
?   	ubunatic.com/loom/examples/screens	[no test files]
ok  	ubunatic.com/loom/examples/screens/screens	0.043s
?   	ubunatic.com/loom/examples/splash	[no test files]
ok  	ubunatic.com/loom/examples/splash/splash	0.033s
?   	ubunatic.com/loom/examples/split	[no test files]
ok  	ubunatic.com/loom/examples/split/split	0.039s
?   	ubunatic.com/loom/examples/suspend	[no test files]
ok  	ubunatic.com/loom/examples/suspend/suspend	0.025s
?   	ubunatic.com/loom/examples/tabs	[no test files]
ok  	ubunatic.com/loom/examples/tabs/tabs	0.027s
?   	ubunatic.com/loom/examples/textedit	[no test files]
ok  	ubunatic.com/loom/examples/textedit/textedit	0.097s
?   	ubunatic.com/loom/examples/textrender	[no test files]
ok  	ubunatic.com/loom/examples/textrender/textrender	0.065s
?   	ubunatic.com/loom/examples/treemap	[no test files]
ok  	ubunatic.com/loom/examples/treemap/treemap	0.076s
?   	ubunatic.com/loom/examples/usage	[no test files]
ok  	ubunatic.com/loom/examples/usage/usage	0.075s
?   	ubunatic.com/loom/examples/winch	[no test files]
ok  	ubunatic.com/loom/examples/winch/winch	0.067s
ok  	ubunatic.com/loom/gallery	0.319s
ok  	ubunatic.com/loom/graph	0.078s
ok  	ubunatic.com/loom/internal/apitest	6.593s
ok  	ubunatic.com/loom/internal/boxcheck	0.011s
?   	ubunatic.com/loom/internal/canary/treesitter	[no test files]
ok  	ubunatic.com/loom/internal/clitest	0.028s
ok  	ubunatic.com/loom/internal/editor	0.781s
ok  	ubunatic.com/loom/internal/examplesreg	0.041s
ok  	ubunatic.com/loom/internal/ptytest	0.008s
ok  	ubunatic.com/loom/internal/session	0.194s
ok  	ubunatic.com/loom/internal/specdefaults	5.166s
ok  	ubunatic.com/loom/layout	0.003s
ok  	ubunatic.com/loom/loomtest	0.014s
ok  	ubunatic.com/loom/measure	0.050s
ok  	ubunatic.com/loom/media	0.180s
?   	ubunatic.com/loom/scripts/canary095	[no test files]
?   	ubunatic.com/loom/spec	[no test files]
ok  	ubunatic.com/loom/specdefaults	0.015s
ok  	ubunatic.com/loom/syntax	0.018s
ok  	ubunatic.com/loom/terminal	0.028s
ok  	ubunatic.com/loom/vt	0.014s
```
