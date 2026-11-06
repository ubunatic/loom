with open("internal/session/session.go", "r") as f:
    lines = f.readlines()

new_lines = []
for line in lines:
    if "type sessionTab struct {" in line:
        new_lines.append(line)
        new_lines.append("\tfgCmd            string\n")
        new_lines.append("\tcwd              string\n")
    else:
        new_lines.append(line)

with open("internal/session/session.go", "w") as f:
    f.writelines(new_lines)
