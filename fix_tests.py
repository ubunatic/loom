import re

with open("internal/session/session_pty_test.go", "r") as f:
    content = f.read()

# wait, `cat` would process the input, but `session.go` might swallow it since it parses mouse events?
# If `session.go` passes the mouse event, `cat` will see it and print it.
# The reason it's timing out is probably because `session.WaitFor` waits for the string to appear in the rendered screen,
# but `\x1b[<0;5;5M` is an ANSI sequence that won't be rendered directly as text!
# It will be parsed by `vt.VT`!
# Let's send regular printable text to verify passthrough if we can't observe mouse events easily,
# or we can inspect `session.VT().ANSI()`? No, `session.VT()` is the terminal emulator parsing it.
# If `cat` prints `\x1b[<0;5;5M`, `vt.VT` will parse it as a mouse event report. It won't show up in the screen text!
# Instead, we can run a bash script that reads the sequence using `read` and then prints "CLICK RECEIVED" and "DRAG RECEIVED"!

content = content.replace(
    """fixture := "printf '\\033[?1003h\\033[?1006hREADY\\n'; cat\"""",
    """fixture := "printf '\\033[?1003h\\033[?1006hREADY\\n'; while IFS= read -r -n 1 c; do if [[ \\"$c\\" == \\"M\\" ]]; then echo \\"RECEIVED_M\\"; fi; done\""""
)

# And wait for "RECEIVED_M"
content = content.replace("session.WaitFor(\"\\x1b[<0;5;5M\", time.Second*5)", "session.WaitFor(\"RECEIVED_M\", time.Second*5)")
content = content.replace("session.WaitFor(\"\\x1b[<32;6;6M\", time.Second*5)", "session.WaitFor(\"RECEIVED_M\", time.Second*5)")

with open("internal/session/session_pty_test.go", "w") as f:
    f.write(content)
