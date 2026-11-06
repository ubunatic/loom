import re

with open("internal/session/session.go", "r") as f:
    content = f.read()

# Let's fix applyTabModes since it's the only remaining piece that can be causing the timeouts if it's sending tracking off sequences instead of basic sequences on switch, which disrupts wait states.
# In `applyTabModes`, the outer terminal mode gets reset according to the current active tab.
# Wait, `applyTabModes` resets the mode to child's mode!
# If the mode is "basic", the child might have asked for motion tracking, but we must clamp it!
# I didn't change `applyTabModes` to clamp in my latest iteration! That's why tests time out - because `applyTabModes` sends `\x1b[?1003h` while `syncSessionMouse` sends `\x1b[?1000h`!
# They conflict or test waits for 1000h but gets 1003h from applyTabModes!

# In applyTabModes, if newTab.mouseMode == MouseModeBasic, we must only send 1000h and 1006h, ignoring newMouse, newSGR etc.
# Actually, syncSessionMouse is called at the end of applyTabModes, which sends 1000h. But `applyTabModes` sent 1003h just before! This sends two sequences and confuses the test which looks for exact states.

content = re.sub(
    r"newMouse := newTab\.vt\.MouseMode\s+newSGR := newTab\.vt\.MouseSGR\s+newUTF8 := newTab\.vt\.MouseUTF8\s+newURXVT := newTab\.vt\.MouseURXVT",
    """newMouse := newTab.vt.MouseMode
	newSGR := newTab.vt.MouseSGR
	newUTF8 := newTab.vt.MouseUTF8
	newURXVT := newTab.vt.MouseURXVT
	if sessionTabMouseMode(newTab) == MouseModeBasic {
		if newMouse != 0 {
			newMouse = 1000
			newSGR = true
			newUTF8 = false
			newURXVT = false
		}
	} else if sessionTabMouseMode(newTab) == MouseModeOff {
		newMouse = 0
		newSGR = false
		newUTF8 = false
		newURXVT = false
	}""", content
)

with open("internal/session/session.go", "w") as f:
    f.write(content)
