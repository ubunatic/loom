import re

with open("internal/session/session_pty_test.go", "r") as f:
    content = f.read()

# Let's change fixture to something that reads exact bytes:
fixture = r"""fixture := "printf '\\033[?1003h\\033[?1006hREADY\n'; cat -v" """

content = re.sub(r'fixture := "printf \'\\\\033\[\?1003h\\\\033\[\?1006hREADY\\\\n\';.*?"', fixture, content)

content = content.replace('session.WaitFor("^[\\\[<0;5;5M", time.Second*5)', 'session.WaitFor("^[<0;5;5M", time.Second*5)')
content = content.replace('session.WaitFor("^[\\x5b<0;5;5M", time.Second*5)', 'session.WaitFor("^[<0;5;5M", time.Second*5)')
content = content.replace('session.WaitFor("RECEIVED_M", time.Second*5)', 'session.WaitFor("^[<0;5;5M", time.Second*5)')

content = content.replace('strings.Contains(strings.Join(session.Screen(), "\\n"), "\\x1b[<32")', 'strings.Contains(strings.Join(session.Screen(), "\\n"), "^[<32;6;6M")')
content = content.replace('session.WaitFor("\\x1b[<32;6;6M", time.Second*5)', 'session.WaitFor("^[<32;6;6M", time.Second*5)')

# fix syntax error manually
content = content.replace('session.WaitFor("^[\\x5b<0;5;5M", time.Second*5)', 'session.WaitFor("^[<0;5;5M", time.Second*5)')
# find the line with syntax error
lines = content.split('\n')
for i, line in enumerate(lines):
    if 'session.WaitFor("^[\\[' in line:
        lines[i] = '	session.WaitFor("^[<0;5;5M", time.Second*5)'

with open("internal/session/session_pty_test.go", "w") as f:
    f.write("\n".join(lines))
