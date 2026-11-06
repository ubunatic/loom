import re

with open("internal/session/session_pty_test.go", "r") as f:
    content = f.read()

# Fix WaitFor usage, which doesn't return an error, it just logs and times out if it fails or panics the test.
content = content.replace("""	// Verify nothing is printed.
	err := session.WaitFor("\\x1b[<32;6;6M", time.Second*1)
	if err == nil {
		t.Fatalf("Drag event was not filtered in basic mode!")
	}""", """	// Verify nothing is printed (Wait for a bit to ensure it doesn't arrive).
	// We'll just sleep.
	time.Sleep(1 * time.Second)
	// check screen buffer
	if strings.Contains(strings.Join(session.Screen(), "\\n"), "\\x1b[<32") {
		t.Fatalf("Drag event was not filtered in basic mode!")
	}""")

content = content.replace("""	// Verify it is printed.
	err := session.WaitFor("\\x1b[<32;6;6M", time.Second*5)
	if err != nil {
		t.Fatalf("Drag event was incorrectly filtered in full mode!")
	}""", """	// Verify it is printed.
	session.WaitFor("\\x1b[<32;6;6M", time.Second*5)""")

with open("internal/session/session_pty_test.go", "w") as f:
    f.write(content)
