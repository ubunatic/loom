import re

with open("internal/session/session_pty_test.go", "r") as f:
    content = f.read()

# the string printed is ^[[<0;5;5M because cat -v converts ESC to ^[ so ESC[ becomes ^[[
content = content.replace('session.WaitFor("^[<0;5;5M"', 'session.WaitFor("^[[<0;5;5M"')
content = content.replace('session.WaitFor("^[<32;6;6M"', 'session.WaitFor("^[[<32;6;6M"')

with open("internal/session/session_pty_test.go", "w") as f:
    f.write(content)
