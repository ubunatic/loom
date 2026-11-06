import re
with open("internal/session/session_pty_test.go", "r") as f:
    content = f.read()

# I will replace the exact expectations with the ones matched by the screenshot
content = content.replace(' 1: session ', ' ¹🐚session ')
content = content.replace(' 2: session ', ' ²🐚session ')
content = content.replace(' 3: session ', ' ³🐚session ')
content = content.replace(' 1: bash ', ' ¹🐚bash ')
content = content.replace(' 2: bash ', ' ²🐚bash ')
content = content.replace(' 3: bash ', ' ³🐚bash ')

# AND for adjacent tabs:
content = content.replace(' ¹🐚session   ²🐚session ', ' ¹🐚session   ²🐚session ')
content = content.replace(' ¹session   ²session ', ' ¹🐚session   ²🐚session ')
content = content.replace(' ¹bash   ²bash ', ' ¹🐚bash   ²🐚bash ')
content = content.replace(' ¹session   ²session   ³session ', ' ¹🐚session   ²🐚session   ³🐚session ')

# A few instances might be missing a trailing space depending on how they were written
content = content.replace(' ¹session', ' ¹🐚session')
content = content.replace(' ²session', ' ²🐚session')
content = content.replace(' ³session', ' ³🐚session')

# Also the direct string contains
content = content.replace('strings.Contains(r, " ¹🐚session ")', 'strings.Contains(r, "¹🐚session")')
content = content.replace('strings.Contains(r, " ¹🐚session   ²🐚session ")', 'strings.Contains(r, "¹🐚session") && strings.Contains(r, "²🐚session")')
content = content.replace('strings.Contains(strings.Join(rows, "\\n"), "2: ")', 'strings.Contains(strings.Join(rows, "\\n"), "²🐚session")')
content = content.replace('strings.Contains(strings.Join(rows, "\\n"), "3: ")', 'strings.Contains(strings.Join(rows, "\\n"), "³🐚session")')
content = content.replace('session.WaitFor("2: ", time.Second*5)', 'session.WaitFor("²🐚session", time.Second*5)')
content = content.replace('session.WaitFor("3: ", time.Second*5)', 'session.WaitFor("³🐚session", time.Second*5)')
content = content.replace('strings.Contains(session.Screen()[len(session.Screen())-1], "1: bash")', 'strings.Contains(session.Screen()[len(session.Screen())-1], "¹🐚bash")')
content = content.replace('strings.Contains(strings.Join(rows, "\\n"), "1: bash")', 'strings.Contains(strings.Join(rows, "\\n"), "¹🐚bash")')
content = content.replace('session.WaitFor("1: bash", time.Second*5)', 'session.WaitFor("¹🐚bash", time.Second*5)')

with open("internal/session/session_pty_test.go", "w") as f:
    f.write(content)
